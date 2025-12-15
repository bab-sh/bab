package interpolate

import (
	"os"
	"regexp"
	"strings"
)

var Pattern = regexp.MustCompile(`\$\{\{\s*([a-zA-Z_][a-zA-Z0-9_.]*)\s*\}\}`)
var EscapePattern = regexp.MustCompile(`\$\$\{\{`)

const escapePlaceholder = "\x00ESCAPED_BRACE\x00"

type Context struct {
	Vars map[string]string
	Path string
	Line int
}

func NewContext(vars map[string]string) *Context {
	if vars == nil {
		vars = make(map[string]string)
	}
	return &Context{Vars: vars}
}

func NewContextWithLocation(vars map[string]string, path string, line int) *Context {
	ctx := NewContext(vars)
	ctx.Path = path
	ctx.Line = line
	return ctx
}

func Interpolate(input string, ctx *Context) (string, error) {
	if ctx == nil {
		ctx = NewContext(nil)
	}

	var errs []error

	result := EscapePattern.ReplaceAllString(input, escapePlaceholder)

	result = Pattern.ReplaceAllStringFunc(result, func(match string) string {
		submatch := Pattern.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}
		name := submatch[1]

		value, err := resolveVar(name, ctx)
		if err != nil {
			errs = append(errs, err)
			return match
		}
		return value
	})

	result = strings.ReplaceAll(result, escapePlaceholder, "${{")

	if len(errs) > 0 {
		return result, errs[0]
	}
	return result, nil
}

func resolveVar(name string, ctx *Context) (string, error) {
	if strings.HasPrefix(name, "env.") {
		envName := strings.TrimPrefix(name, "env.")
		return os.Getenv(envName), nil
	}

	if value, ok := ctx.Vars[name]; ok {
		return value, nil
	}

	available := make([]string, 0, len(ctx.Vars))
	for k := range ctx.Vars {
		available = append(available, k)
	}

	return "", &VarNotFoundError{
		Path:      ctx.Path,
		Line:      ctx.Line,
		Name:      name,
		Available: available,
	}
}

func ResolveVars(vars map[string]string, parentVars map[string]string) (map[string]string, error) {
	return ResolveVarsWithLocation(vars, parentVars, "", 0)
}

func ResolveVarsWithLocation(vars map[string]string, parentVars map[string]string, path string, line int) (map[string]string, error) {
	if vars == nil {
		if parentVars == nil {
			return make(map[string]string), nil
		}
		return parentVars, nil
	}

	resolved := make(map[string]string)
	for k, v := range parentVars {
		resolved[k] = v
	}

	resolving := make(map[string]bool)
	resolvedSet := make(map[string]bool)
	var resolvingStack []string

	allVarNames := func() []string {
		names := make([]string, 0, len(vars)+len(parentVars))
		for k := range vars {
			names = append(names, k)
		}
		for k := range parentVars {
			if _, exists := vars[k]; !exists {
				names = append(names, k)
			}
		}
		return names
	}

	var resolveOne func(name string) (string, error)
	resolveOne = func(name string) (string, error) {
		if resolvedSet[name] {
			return resolved[name], nil
		}

		if resolving[name] {
			chain := make([]string, len(resolvingStack)+1)
			copy(chain, resolvingStack)
			chain[len(resolvingStack)] = name
			return "", &VarCycleError{Path: path, Line: line, Name: name, Chain: chain}
		}

		raw, exists := vars[name]
		if !exists {
			if v, ok := parentVars[name]; ok {
				return v, nil
			}
			return "", &VarNotFoundError{Path: path, Line: line, Name: name, Available: allVarNames()}
		}

		resolving[name] = true
		resolvingStack = append(resolvingStack, name)

		matches := Pattern.FindAllStringSubmatch(raw, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			refName := match[1]

			if strings.HasPrefix(refName, "env.") {
				continue
			}

			if _, ok := vars[refName]; ok {
				refValue, err := resolveOne(refName)
				if err != nil {
					return "", err
				}
				resolved[refName] = refValue
				resolvedSet[refName] = true
			}
		}

		ctx := NewContextWithLocation(resolved, path, line)
		value, err := Interpolate(raw, ctx)
		if err != nil {
			return "", err
		}

		resolvingStack = resolvingStack[:len(resolvingStack)-1]
		delete(resolving, name)
		resolved[name] = value
		resolvedSet[name] = true

		return value, nil
	}

	for name := range vars {
		if _, err := resolveOne(name); err != nil {
			return nil, err
		}
	}

	return resolved, nil
}

func ContainsVarRef(s string) bool {
	return Pattern.MatchString(s)
}

func ExtractVarRefs(s string) []string {
	matches := Pattern.FindAllStringSubmatch(s, -1)
	refs := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 2 {
			refs = append(refs, match[1])
		}
	}
	return refs
}
