package bundle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

type Delivery struct {
	Mode            string `json:"mode"`
	Instructions    string `json:"instructions"`
	SkillsRoot      string `json:"skills_root,omitempty"`
	CompiledContext string `json:"compiled_context,omitempty"`
}

type Manifest struct {
	Format      string   `json:"format"`
	Role        string   `json:"role"`
	Personality string   `json:"personality"`
	Color       string   `json:"color,omitempty"`
	Density     string   `json:"density"`
	Sources     []string `json:"sources"`
	Delivery    Delivery `json:"delivery"`
}

type Trace struct {
	Format    string              `json:"format"`
	Decisions []resolver.Decision `json:"decisions"`
}

type Result struct {
	Dir    string
	Key    string
	Reused bool
}

// ReadTrace loads the retained decision evidence; it is also the stable
// machine-readable explanation surface - there is no second format.
func ReadTrace(dir string) (*Trace, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "trace.json"))
	if err != nil {
		return nil, fmt.Errorf("read bundle trace: %w", err)
	}
	var tr Trace
	if err := json.Unmarshal(raw, &tr); err != nil {
		return nil, fmt.Errorf("parse bundle trace: %w", err)
	}
	if tr.Format != "agent-compose.trace" {
		return nil, fmt.Errorf("%s holds no agent-compose trace (format %q)", dir, tr.Format)
	}
	return &tr, nil
}

// ReadManifest is the consumer entry point: bundles are opaque except for
// manifest.json.
func ReadManifest(dir string) (*Manifest, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("read bundle manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("parse bundle manifest: %w", err)
	}
	if m.Format != "agent-compose.bundle" {
		return nil, fmt.Errorf("%s is not an agent-compose bundle (format %q)", dir, m.Format)
	}
	return &m, nil
}

// Materialize writes the resolved bundle beneath outDir, atomically and at
// most once per input key. Identical inputs reuse the existing tree.
func Materialize(res *resolver.Resolution, outDir string) (*Result, error) {
	if !safeSegment(res.Skill.Source) || !safeSegment(res.Skill.ID) {
		return nil, fmt.Errorf("selected identity path %q/%q is unsafe", res.Skill.Source, res.Skill.ID)
	}
	key, err := cacheKey(res)
	if err != nil {
		return nil, err
	}
	short := key[:16]
	target := filepath.Join(outDir, short)
	if _, err := os.Stat(filepath.Join(target, "manifest.json")); err == nil {
		if _, err := Verify(target); err != nil {
			return nil, fmt.Errorf("cached bundle %s failed verification: %w", target, err)
		}
		return &Result{Dir: target, Key: short, Reused: true}, nil
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	staging, err := os.MkdirTemp(outDir, ".stage-"+short+"-")
	if err != nil {
		return nil, err
	}
	if err := write(res, staging); err != nil {
		os.RemoveAll(staging)
		return nil, err
	}
	if _, err := Verify(staging); err != nil {
		os.RemoveAll(staging)
		return nil, fmt.Errorf("verify staged bundle: %w", err)
	}
	if err := os.Rename(staging, target); err != nil {
		os.RemoveAll(staging)
		if _, statErr := os.Stat(filepath.Join(target, "manifest.json")); statErr == nil {
			if _, verifyErr := Verify(target); verifyErr != nil {
				return nil, fmt.Errorf("concurrent cached bundle %s failed verification: %w", target, verifyErr)
			}
			return &Result{Dir: target, Key: short, Reused: true}, nil
		}
		return nil, fmt.Errorf("finalize bundle: %w", err)
	}
	return &Result{Dir: target, Key: short, Reused: false}, nil
}

func write(res *resolver.Resolution, root string) error {
	if !safeSegment(res.Skill.Source) || !safeSegment(res.Skill.ID) {
		return fmt.Errorf("selected identity path %q/%q is unsafe", res.Skill.Source, res.Skill.ID)
	}
	instructions, err := joinInstructions(res)
	if err != nil {
		return err
	}
	if err := writeFile(filepath.Join(root, "content", "instructions.md"), instructions); err != nil {
		return err
	}

	skillRoot := filepath.Join(root, "content", "skills", res.Skill.Source, res.Skill.ID)
	if err := copyTree(res.SkillDir, skillRoot); err != nil {
		return err
	}

	delivery := Delivery{Mode: res.Request.Delivery, Instructions: "content/instructions.md"}
	switch res.Request.Delivery {
	case schema.DeliveryNativeSkills:
		delivery.SkillsRoot = "content/skills"
	case schema.DeliveryCompiled:
		body, err := os.ReadFile(res.CompiledBody)
		if err != nil {
			return err
		}
		compiled := append(append([]byte{}, instructions...), '\n')
		compiled = append(compiled, body...)
		if err := writeFile(filepath.Join(root, "delivery", "compiled.md"), compiled); err != nil {
			return err
		}
		delivery.CompiledContext = "delivery/compiled.md"
	}

	trace, err := json.MarshalIndent(Trace{Format: "agent-compose.trace", Decisions: res.Decisions}, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFile(filepath.Join(root, "trace.json"), append(trace, '\n')); err != nil {
		return err
	}

	manifest, err := json.MarshalIndent(Manifest{
		Format:      "agent-compose.bundle",
		Role:        res.Request.Role,
		Personality: res.Request.Personality,
		Color:       res.FavoriteColor,
		Density:     res.Request.Density,
		Sources:     res.SourceIDs,
		Delivery:    delivery,
	}, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(root, "manifest.json"), append(manifest, '\n'))
}

func joinInstructions(res *resolver.Resolution) ([]byte, error) {
	var out []byte
	for i, sel := range res.Instructions {
		raw, err := os.ReadFile(sel.Path)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			out = append(out, '\n')
		}
		out = append(out, raw...)
	}
	if len(out) > 0 && out[len(out)-1] != '\n' {
		out = append(out, '\n')
	}
	return out, nil
}

// cacheKey hashes every input that can change bundle bytes: the request,
// the person source, the decisions, and every referenced content file.
func cacheKey(res *resolver.Resolution) (string, error) {
	h := sha256.New()
	fmt.Fprintf(h, "request\x00%s\x00%s\x00%s\x00%s\x00",
		res.Request.Role, res.Request.Personality, res.Request.Delivery, res.Request.Density)
	fmt.Fprintf(h, "person\x00%d\x00", len(res.Person.Raw))
	h.Write(res.Person.Raw)
	for _, d := range res.Decisions {
		fmt.Fprintf(h, "decision\x00%s\x00%s\x00%s\x00%s\x00", d.Subject, d.Source, d.Outcome, d.Reason)
	}
	for _, sel := range res.Instructions {
		raw, err := os.ReadFile(sel.Path)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "instruction\x00%s\x00%s\x00%d\x00", sel.Source, sel.ID, len(raw))
		h.Write(raw)
	}
	err := filepath.WalkDir(res.SkillDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(res.SkillDir, p)
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "skill\x00%s\x00%d\x00", filepath.ToSlash(rel), len(raw))
		h.Write(raw)
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%s: symlinks are invalid inside a bundle", p)
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		out := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return writeFile(out, raw)
	})
}

func writeFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}
