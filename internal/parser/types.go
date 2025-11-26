package parser

type Task struct {
	Name         string
	Description  string
	Commands     []string
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
