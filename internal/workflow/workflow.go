package workflow

type Step struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Workflow struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Steps []Step `json:"steps"`
}