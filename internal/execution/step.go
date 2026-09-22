package execution

import "time"

type ExecutionStep struct {
	ID          string     `json:"id"`
	ExecutionID string     `json:"execution_id"`
	StepName    string     `json:"step_name"`
	StepType    string     `json:"step_type"`
	Status      string     `json:"status"`
	Output      *string    `json:"output,omitempty"`
	Error       *string    `json:"error,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
