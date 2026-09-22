package workflow

import "fmt"

var supportedStepTypes = map[string]bool{
	"log":  true,
	"http": true,
}

func Validate(wf Workflow) error {
	if wf.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(wf.Steps) == 0 {
		return fmt.Errorf("workflow must contain at least one step")
	}

	stepNames := make(map[string]bool)

	for _, step := range wf.Steps {
		if step.Name == "" {
			return fmt.Errorf("step name is required")
		}

		if stepNames[step.Name] {
			return fmt.Errorf("duplicate step name: %s", step.Name)
		}

		stepNames[step.Name] = true

		if step.Type == "" {
			return fmt.Errorf("step type is required for step: %s", step.Name)
		}

		if !supportedStepTypes[step.Type] {
			return fmt.Errorf("unsupported step type: %s", step.Type)
		}
	}

	return nil
}
