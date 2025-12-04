package parser

func validateDependencies(tasks TaskMap) error {
	for name, task := range tasks {
		for _, dep := range task.Dependencies {
			if !tasks.Has(dep) {
				return &NotFoundError{
					TaskName:     dep,
					ReferencedBy: name,
					Available:    tasks.Names(),
				}
			}
		}
	}
	return nil
}
