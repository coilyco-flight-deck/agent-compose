// Package remoteskills hydrates Git-backed skill catalogs into agent-compose's
// state directory without making native skill projection depend on live network
// access after the first successful checkout.
package remoteskills

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const DefaultTTL = 10 * time.Minute
const defaultCatalogPath = ".agents/skills"

// Source is one remote repository containing a skill catalog.
type Source struct {
	URL string
	Ref string
}

// UnmarshalYAML keeps the public configuration at one source locator per list
// item. URL and Git object details stay internal after parsing.
func (s *Source) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
		return fmt.Errorf(
			"remote skill source must be a scalar locator such as owner/repo/path@ref",
		)
	}
	source, err := ParseLocator(node.Value)
	if err != nil {
		return err
	}
	*s = source
	return nil
}

// ParseLocator expands a GitHub Actions-style owner/repo/path@ref locator or a
// generic Git URL whose optional catalog path follows its .git boundary.
func ParseLocator(raw string) (Source, error) {
	locator := strings.TrimSpace(raw)
	if locator == "" {
		return Source{}, fmt.Errorf("remote skill source locator cannot be empty")
	}

	address, revision := splitLocatorRevision(locator)
	repository, catalogPath, err := splitRepositoryPath(address)
	if err != nil {
		return Source{}, err
	}
	source := Source{
		URL: repository,
		Ref: revision + ":" + catalogPath,
	}
	if err := ValidateSource(source); err != nil {
		return Source{}, err
	}
	return source, nil
}

func splitLocatorRevision(locator string) (string, string) {
	index := strings.LastIndex(locator, "@")
	if index < 0 || locatorAtIsURLUser(locator, index) {
		return locator, "HEAD"
	}
	revision := strings.TrimSpace(locator[index+1:])
	if revision == "" {
		return locator[:index], ""
	}
	return locator[:index], revision
}

func locatorAtIsURLUser(locator string, index int) bool {
	if scheme := strings.Index(locator, "://"); scheme >= 0 {
		authority := locator[scheme+3:]
		if slash := strings.Index(authority, "/"); slash >= 0 {
			return index < scheme+3+slash
		}
		return true
	}
	// Preserve the user separator in an SCP-style Git URL without an explicit
	// locator ref. An explicit ref follows the repository's .git suffix.
	return !strings.Contains(locator[:index], "/") &&
		strings.Contains(locator[index+1:], ":")
}

func splitRepositoryPath(address string) (string, string, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return "", "", fmt.Errorf("remote skill source needs a repository")
	}
	if marker := strings.LastIndex(address, ".git/"); marker >= 0 {
		repository := address[:marker+len(".git")]
		catalogPath := address[marker+len(".git/"):]
		clean, err := cleanCatalogPath(catalogPath)
		return repository, clean, err
	}
	if strings.HasSuffix(address, ".git") ||
		strings.Contains(address, "://") ||
		strings.HasPrefix(address, "/") ||
		strings.HasPrefix(address, "./") ||
		strings.HasPrefix(address, "../") ||
		strings.Contains(address, `\`) ||
		strings.Contains(address, ":") {
		return address, defaultCatalogPath, nil
	}

	parts := strings.Split(strings.Trim(address, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf(
			"remote skill source %q must use owner/repo/path@ref or a Git URL",
			address,
		)
	}
	ownerIndex := 0
	if parts[0] == "github.com" {
		if len(parts) < 3 {
			return "", "", fmt.Errorf(
				"remote skill source %q needs a GitHub owner and repository",
				address,
			)
		}
		ownerIndex = 1
	}
	repository := "https://github.com/" +
		parts[ownerIndex] + "/" + parts[ownerIndex+1] + ".git"
	catalogPath := defaultCatalogPath
	if len(parts) > ownerIndex+2 {
		var err error
		catalogPath, err = cleanCatalogPath(
			strings.Join(parts[ownerIndex+2:], "/"),
		)
		if err != nil {
			return "", "", err
		}
	}
	return repository, catalogPath, nil
}

func cleanCatalogPath(catalogPath string) (string, error) {
	if strings.TrimSpace(catalogPath) == "" {
		return "", fmt.Errorf("remote skill source catalog path cannot be empty")
	}
	clean := path.Clean(catalogPath)
	if strings.HasPrefix(catalogPath, "/") ||
		clean == "." ||
		clean == ".." ||
		strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf(
			"remote skill source catalog path %q must stay inside its repository",
			catalogPath,
		)
	}
	return clean, nil
}

// Catalog is one verified, hydrated skill root ready for native projection.
type Catalog struct {
	Path string
}

// State describes the network and cache work used to resolve one catalog.
type State string

const (
	StateCached    State = "cached"
	StateHydrated  State = "hydrated"
	StateRefreshed State = "refreshed"
	StateFallback  State = "fallback"
)

// Result reports one source without exposing cache paths in the stable summary.
type Result struct {
	Source  Source
	Catalog Catalog
	State   State
	Warning string
}

// Options controls the cache location, freshness window, and test clock.
type Options struct {
	StateDir string
	TTL      time.Duration
	Now      func() time.Time
}

// Hydrate resolves every source in declaration order. A first-use failure
// aborts. A stale refresh failure reuses the last verified checkout.
func Hydrate(ctx context.Context, sources []Source, opts Options) ([]Result, error) {
	if len(sources) == 0 {
		return nil, nil
	}
	if strings.TrimSpace(opts.StateDir) == "" {
		return nil, fmt.Errorf("remote skill cache needs a state directory")
	}
	if opts.TTL < 0 {
		return nil, fmt.Errorf("remote skill cache TTL cannot be negative")
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}

	results := make([]Result, 0, len(sources))
	for _, source := range sources {
		result, err := hydrateOne(ctx, source, opts)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func hydrateOne(ctx context.Context, source Source, opts Options) (Result, error) {
	source.URL = strings.TrimSpace(source.URL)
	source.Ref = strings.TrimSpace(source.Ref)
	if err := ValidateSource(source); err != nil {
		return Result{}, err
	}
	revision, catalogPath := splitRef(source.Ref)
	source.Ref = revision + ":" + catalogPath
	localCatalogPath := filepath.FromSlash(catalogPath)

	cacheRoot := filepath.Join(opts.StateDir, "cache", "remote-skills", cacheKey(source))
	mirror := filepath.Join(cacheRoot, "mirror.git")
	work := filepath.Join(cacheRoot, "work")
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return Result{}, fmt.Errorf("create remote skill cache: %w", err)
	}
	unlock, err := lockCache(cacheRoot)
	if err != nil {
		return Result{}, err
	}
	defer unlock()

	state := StateCached
	warning := ""
	mirrorExists := isDir(mirror)
	if !mirrorExists {
		if err := cloneMirror(ctx, source.URL, mirror, opts.Now()); err != nil {
			return Result{}, fmt.Errorf("hydrate remote skill source %s: %w", displaySource(source), err)
		}
		mirrorExists = true
		state = StateHydrated
	} else if mirrorStale(mirror, opts.TTL, opts.Now()) {
		if immutableRefPresent(ctx, mirror, revision) {
			touchFetchHead(mirror, opts.Now())
		} else if err := runGit(ctx, "-C", mirror, "remote", "update", "--prune"); err != nil {
			state = StateFallback
			warning = fmt.Sprintf(
				"remote skill refresh failed for %s, using cached state: %v",
				displaySource(source),
				err,
			)
			if checkoutValid(work, localCatalogPath) {
				return Result{
					Source: source,
					Catalog: Catalog{
						Path: filepath.Join(work, localCatalogPath),
					},
					State:   state,
					Warning: warning,
				}, nil
			}
		} else {
			touchFetchHead(mirror, opts.Now())
			state = StateRefreshed
		}
	}

	commit, err := resolveCommit(ctx, mirror, revision)
	if err != nil {
		return Result{}, fmt.Errorf("resolve remote skill source %s: %w", displaySource(source), err)
	}
	if err := verifyCatalogTree(ctx, mirror, revision, catalogPath); err != nil {
		return Result{}, fmt.Errorf("resolve remote skill source %s: %w", displaySource(source), err)
	}
	if !checkoutMatches(ctx, work, commit, localCatalogPath) {
		if err := dropCheckout(ctx, source, mirror, work, commit, localCatalogPath); err != nil {
			return Result{}, fmt.Errorf("materialize remote skill source %s: %w", displaySource(source), err)
		}
	}

	catalog := Catalog{
		Path: filepath.Join(work, localCatalogPath),
	}
	return Result{Source: source, Catalog: catalog, State: state, Warning: warning}, nil
}

// ValidateSource checks one normalized remote locator without touching disk or
// the network.
func ValidateSource(source Source) error {
	source.URL = strings.TrimSpace(source.URL)
	source.Ref = strings.TrimSpace(source.Ref)
	if source.URL == "" {
		return fmt.Errorf("remote skill source needs url")
	}
	revision, catalogPath := splitRef(source.Ref)
	if revision == "" {
		return fmt.Errorf("remote skill source %s needs a revision before ':'", displaySource(source))
	}
	if strings.HasPrefix(catalogPath, "/") {
		return fmt.Errorf("remote skill source %s path must be relative", displaySource(source))
	}
	clean := path.Clean(catalogPath)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("remote skill source %s path escapes its checkout", displaySource(source))
	}
	return nil
}

// splitRef applies Git's <tree-ish>:<path> syntax while preserving the common
// shorthand where a bare revision selects the conventional skill root.
func splitRef(locator string) (string, string) {
	if locator == "" {
		return "HEAD", defaultCatalogPath
	}
	revision, catalogPath, found := strings.Cut(locator, ":")
	if !found {
		return locator, defaultCatalogPath
	}
	return revision, path.Clean(catalogPath)
}

func displaySource(source Source) string {
	if source.Ref == "" {
		return source.URL
	}
	return source.URL + "@" + source.Ref
}

func cacheKey(source Source) string {
	h := sha256.New()
	fmt.Fprintf(h, "url\x00%s\x00ref\x00%s\x00", source.URL, source.Ref)
	return hex.EncodeToString(h.Sum(nil))
}

func cloneMirror(ctx context.Context, remote, mirror string, now time.Time) error {
	parent := filepath.Dir(mirror)
	stage, err := os.MkdirTemp(parent, ".mirror-*")
	if err != nil {
		return err
	}
	if err := os.Remove(stage); err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err := runGit(ctx, "clone", "--mirror", remote, stage); err != nil {
		return err
	}
	touchFetchHead(stage, now)
	return os.Rename(stage, mirror)
}

func resolveCommit(ctx context.Context, mirror, ref string) (string, error) {
	target := "HEAD^{commit}"
	if ref != "" {
		target = ref + "^{commit}"
	}
	out, err := gitOutput(ctx, "-C", mirror, "rev-parse", "--verify", target)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func verifyCatalogTree(ctx context.Context, mirror, revision, catalogPath string) error {
	objectType, err := gitOutput(ctx, "-C", mirror, "cat-file", "-t", revision+":"+catalogPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(objectType) != "tree" {
		return fmt.Errorf("skill path %q is not a directory", catalogPath)
	}
	return nil
}

func checkoutMatches(ctx context.Context, work, commit, catalogPath string) bool {
	if !checkoutValid(work, catalogPath) {
		return false
	}
	out, err := gitOutput(ctx, "-C", work, "rev-parse", "--verify", "HEAD^{commit}")
	return err == nil && strings.TrimSpace(out) == commit
}

func checkoutValid(work, catalogPath string) bool {
	return isDir(filepath.Join(work, ".git")) && isDir(filepath.Join(work, catalogPath))
}

func dropCheckout(
	ctx context.Context,
	source Source,
	mirror, work, commit, catalogPath string,
) error {
	parent := filepath.Dir(work)
	stage, err := os.MkdirTemp(parent, ".work-*")
	if err != nil {
		return err
	}
	if err := os.Remove(stage); err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err := runGit(ctx, "clone", "--quiet", mirror, stage); err != nil {
		return err
	}
	if err := runGit(ctx, "-C", stage, "remote", "set-url", "origin", source.URL); err != nil {
		return err
	}
	if err := runGit(
		ctx,
		"-C", stage,
		"-c", "advice.detachedHead=false",
		"checkout", "--quiet", commit,
	); err != nil {
		return err
	}
	if !isDir(filepath.Join(stage, catalogPath)) {
		return fmt.Errorf("skill path %q is not a directory", catalogPath)
	}
	return replaceCheckout(stage, work)
}

func replaceCheckout(stage, work string) error {
	backup := work + ".previous"
	if !isDir(work) && isDir(backup) {
		if err := os.Rename(backup, work); err != nil {
			return err
		}
	}
	_ = os.RemoveAll(backup)
	hadWork := isDir(work)
	if hadWork {
		if err := os.Rename(work, backup); err != nil {
			return err
		}
	}
	if err := os.Rename(stage, work); err != nil {
		if hadWork {
			_ = os.Rename(backup, work)
		}
		return err
	}
	_ = os.RemoveAll(backup)
	return nil
}

var fullSHA = regexp.MustCompile(`^(?:[0-9a-fA-F]{40}|[0-9a-fA-F]{64})$`)

func immutableRefPresent(ctx context.Context, mirror, ref string) bool {
	if !fullSHA.MatchString(ref) {
		return false
	}
	return runGit(ctx, "-C", mirror, "cat-file", "-e", ref+"^{commit}") == nil
}

func mirrorStale(mirror string, ttl time.Duration, now time.Time) bool {
	info, err := os.Stat(filepath.Join(mirror, "FETCH_HEAD"))
	if err != nil {
		return false
	}
	return now.Sub(info.ModTime()) >= ttl
}

func touchFetchHead(mirror string, now time.Time) {
	path := filepath.Join(mirror, "FETCH_HEAD")
	if err := os.Chtimes(path, now, now); err != nil {
		_ = os.WriteFile(path, nil, 0o644)
		_ = os.Chtimes(path, now, now)
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func runGit(ctx context.Context, args ...string) error {
	_, err := gitOutput(ctx, args...)
	return err
}

func gitOutput(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	raw, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(raw))
		if message == "" {
			return "", err
		}
		return "", fmt.Errorf("%w: %s", err, message)
	}
	return string(raw), nil
}
