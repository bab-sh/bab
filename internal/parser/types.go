package parser

type Task struct {
	Name        string
	Description string
	Commands    []string
}

type TaskMap map[string]*Task
