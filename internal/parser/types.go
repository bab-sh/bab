package parser

type Task struct {
	Name         string
	Description  string
	Commands     []string
	Dependencies []string
}

type TaskMap map[string]*Task
