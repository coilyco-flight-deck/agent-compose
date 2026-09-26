// Package bashguard decides whether one Bash tool command may run bare kubectl
// or helm. It parses the command rather than matching its text, so a flag before
// the verb, a nested shell, or a wrapper does not change the answer. It is a
// guard on ordinary use, not a sandbox. See docs/claude-native-ui-surfaces.md.
package bashguard

import (
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// Policy names the kubectl verbs a role may still run bare. Helm has none.
type Policy struct {
	KubectlVerbs []string
}

var clusterWord = regexp.MustCompile(`(^|[^A-Za-z0-9_-])(kubectl|helm)([^A-Za-z0-9_-]|$)`)

var unquote = strings.NewReplacer(`"`, "", `'`, "", `\`, "")

// wrappers run the command that follows them, so the guard looks past them.
var wrappers = map[string]bool{
	"sudo": true, "doas": true, "env": true, "command": true, "exec": true,
	"time": true, "nice": true, "nohup": true, "timeout": true, "stdbuf": true,
	"xargs": true, "watch": true, "parallel": true, "builtin": true,
}

var shells = map[string]bool{
	"bash": true, "sh": true, "zsh": true, "dash": true, "ksh": true, "fish": true,
}

// interpreters run a string argument the guard cannot parse as shell.
var interpreters = map[string]bool{
	"python": true, "python3": true, "perl": true, "ruby": true, "node": true,
	"ssh": true, "su": true, "osascript": true, "script": true, "source": true, ".": true,
}

// kubectlValueFlags are the global flags that take a separate value. An
// unlisted flag before the verb is refused rather than guessed at.
var kubectlValueFlags = map[string]bool{
	"-n": true, "--namespace": true, "--context": true, "--cluster": true,
	"--user": true, "-s": true, "--server": true, "--token": true, "--as": true,
	"--as-group": true, "--as-uid": true, "--kubeconfig": true, "--cache-dir": true,
	"--certificate-authority": true, "--client-certificate": true,
	"--client-key": true, "--request-timeout": true, "-v": true, "--v": true,
	"--tls-server-name": true, "--username": true, "--password": true,
}

var kubectlBoolFlags = map[string]bool{
	"--insecure-skip-tls-verify": true, "--match-server-version": true,
	"--warnings-as-errors": true, "--disable-compression": true,
}

// Check returns an empty string when command may run, or why it may not.
func Check(command string, p Policy) string {
	g := guard{policy: p, mentions: clusterWord.MatchString(unquote.Replace(command))}
	return g.check(command)
}

type guard struct {
	policy   Policy
	mentions bool
}

func (g guard) check(command string) string {
	file, err := syntax.NewParser(syntax.Variant(syntax.LangBash)).Parse(strings.NewReader(command), "")
	if err != nil {
		if g.mentions {
			return "the command does not parse and mentions kubectl or helm"
		}
		return ""
	}
	reason := ""
	syntax.Walk(file, func(node syntax.Node) bool {
		if reason != "" {
			return false
		}
		switch n := node.(type) {
		case *syntax.Stmt:
			reason = g.stmt(n)
		case *syntax.BinaryCmd:
			reason = g.pipe(n)
		}
		return reason == ""
	})
	return reason
}

func (g guard) stmt(s *syntax.Stmt) string {
	call, ok := s.Cmd.(*syntax.CallExpr)
	if !ok || len(call.Args) == 0 {
		return ""
	}
	// After a wrapper, any later word may be the command it runs, so the scan
	// continues to the end of the line rather than stopping at an option value.
	wrapped := false
	for i, word := range call.Args {
		name, literal := lit(word)
		if !literal {
			if i == 0 || wrapped {
				return g.dynamic()
			}
			continue
		}
		base := path.Base(name)
		switch {
		case base == "kubectl" || base == "helm":
			return g.cluster(base, call.Args[i+1:])
		case shells[base] || base == "eval":
			return g.nested(base, call.Args[i+1:], s.Redirs)
		case interpreters[base]:
			return g.interpreted(call.Args[i+1:], s.Redirs)
		case wrappers[base] && (i == 0 || wrapped):
			wrapped = true
		case !wrapped:
			return ""
		}
	}
	return ""
}

func (g guard) cluster(base string, args []*syntax.Word) string {
	if base == "helm" {
		return "bare helm is refused for this role. Change the cluster through aosguard (teable:coilyco/agentic-os#8282)"
	}
	for i := 0; i < len(args); i++ {
		arg, literal := lit(args[i])
		if !literal {
			return "bare kubectl with a computed argument before its verb is refused"
		}
		switch {
		case kubectlBoolFlags[arg] || (strings.HasPrefix(arg, "--") && strings.Contains(arg, "=")):
			continue
		case kubectlValueFlags[arg]:
			i++
			continue
		case strings.HasPrefix(arg, "-"):
			return fmt.Sprintf("bare kubectl flag %q before the verb is not recognized, so the command is refused", arg)
		}
		if slices.Contains(g.policy.KubectlVerbs, arg) {
			return ""
		}
		return g.refuseKubectl(arg)
	}
	return g.refuseKubectl("")
}

func (g guard) refuseKubectl(verb string) string {
	allowed := "no bare verb"
	if len(g.policy.KubectlVerbs) > 0 {
		allowed = "only bare kubectl " + strings.Join(g.policy.KubectlVerbs, ", ")
	}
	if verb == "" {
		verb = "without a verb"
	}
	return fmt.Sprintf("bare kubectl %s is refused: this role keeps %s. Use aosguard ops kubectl (teable:coilyco/agentic-os#8282)", verb, allowed)
}

// nested parses the script a shell or eval would run and checks it in turn.
func (g guard) nested(base string, args []*syntax.Word, redirs []*syntax.Redirect) string {
	var script []string
	for i := 0; i < len(args); i++ {
		arg, literal := lit(args[i])
		if !literal {
			return g.dynamic()
		}
		if base == "eval" {
			script = append(script, arg)
			continue
		}
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && strings.Contains(arg, "c") && i+1 < len(args) {
			body, literal := lit(args[i+1])
			if !literal {
				return g.dynamic()
			}
			return g.check(body)
		}
	}
	for _, r := range redirs {
		if r.Hdoc != nil {
			body, literal := lit(r.Hdoc)
			if !literal {
				return g.dynamic()
			}
			script = append(script, body)
		}
	}
	return g.check(strings.Join(script, "\n"))
}

func (g guard) interpreted(args []*syntax.Word, redirs []*syntax.Redirect) string {
	for _, r := range redirs {
		if r.Hdoc != nil {
			args = append(args, r.Hdoc)
		}
	}
	for _, arg := range args {
		value, literal := lit(arg)
		if !literal && g.mentions || clusterWord.MatchString(value) {
			return "an interpreter or remote shell argument mentions kubectl or helm, so the command is refused"
		}
	}
	return ""
}

// pipe refuses text piped into a shell or interpreter that mentions a cluster CLI.
func (g guard) pipe(b *syntax.BinaryCmd) string {
	if b.Op != syntax.Pipe && b.Op != syntax.PipeAll {
		return ""
	}
	call, ok := b.Y.Cmd.(*syntax.CallExpr)
	if !ok || len(call.Args) == 0 {
		return ""
	}
	name := path.Base(litOr(call.Args[0]))
	if !shells[name] && !interpreters[name] && name != "xargs" {
		return ""
	}
	var out strings.Builder
	if err := syntax.NewPrinter().Print(&out, b.X); err != nil || clusterWord.MatchString(unquote.Replace(out.String())) {
		return "text piped into a shell or interpreter mentions kubectl or helm, so the command is refused"
	}
	return ""
}

// dynamic refuses a computed command only on a line that names a cluster CLI.
func (g guard) dynamic() string {
	if g.mentions {
		return "a computed command on a line that mentions kubectl or helm is refused"
	}
	return ""
}

// lit returns a word's value when every part of it is literal text.
func lit(w *syntax.Word) (string, bool) {
	var out strings.Builder
	for _, part := range w.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			if strings.ContainsAny(p.Value, "*?[") {
				return "", false
			}
			out.WriteString(p.Value)
		case *syntax.SglQuoted:
			out.WriteString(p.Value)
		case *syntax.DblQuoted:
			for _, inner := range p.Parts {
				l, ok := inner.(*syntax.Lit)
				if !ok {
					return "", false
				}
				out.WriteString(l.Value)
			}
		default:
			return "", false
		}
	}
	return out.String(), true
}

func litOr(w *syntax.Word) string {
	value, _ := lit(w)
	return value
}
