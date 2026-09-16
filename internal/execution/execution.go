package execution

import "time"

type Execution struct {
	ID         string    `json:"id"`
	WorkflowID string    `json:"workflow_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
