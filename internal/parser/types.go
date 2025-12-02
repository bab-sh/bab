package parser

type Command struct {
	Cmd       string
	Platforms []string
}

func (c Command) ShouldRunOnPlatform(platform string) bool {
	if len(c.Platforms) == 0 {
		return true
	}
	for _, p := range c.Platforms {
		if p == platform {
			return true
		}
	}
	return false
}

type Task struct {
	Name         string
	Description  string
	Commands     []Command
	Dependencies []string
}

type TaskMap map[string]*Task

type IncludeMap map[string]string

type ParseContext struct {
	visited map[string]bool
	stack   []string
}

func NewParseContext() *ParseContext {
	return &ParseContext{
		visited: make(map[string]bool),
		stack:   make([]string, 0),
	}
}
