package executor

import (
	"context"

	"github.com/bab-sh/bab/internal/parser"
)

type Executor interface {
	Execute(ctx context.Context, task *parser.Task) error
}

type DefaultExecutor struct{}

func NewExecutor() Executor {
	return &DefaultExecutor{}
}

func (e *DefaultExecutor) Execute(ctx context.Context, task *parser.Task) error {
	return Execute(ctx, task)
}
