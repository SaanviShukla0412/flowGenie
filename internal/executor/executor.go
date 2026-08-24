package executor

import (
	"context"
	"fmt"
	"log"

	"github.com/SaanviShukla0412/flowGenie/internal/workflow"
)

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Execute(
	ctx context.Context,
	step workflow.Step,
) error {
	switch step.Type {
	case "log":
		log.Printf("LOG step executed: %s", step.Name)
		return nil

	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}
}
