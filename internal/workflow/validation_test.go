package workflow

import "testing"

func TestValidateWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		workflow  Workflow
		shouldErr bool
	}{
		{
			name: "valid workflow",
			workflow: Workflow{
				Name: "Test Workflow",
				Steps: []Step{
					{
						Name: "hello",
						Type: "log",
					},
				},
			},
			shouldErr: false,
		},
		{
			name: "missing workflow name",
			workflow: Workflow{
				Steps: []Step{
					{
						Name: "hello",
						Type: "log",
					},
				},
			},
			shouldErr: true,
		},
		{
			name: "no steps",
			workflow: Workflow{
				Name: "Test Workflow",
			},
			shouldErr: true,
		},
		{
			name: "duplicate step names",
			workflow: Workflow{
				Name: "Test Workflow",
				Steps: []Step{
					{Name: "hello", Type: "log"},
					{Name: "hello", Type: "log"},
				},
			},
			shouldErr: true,
		},
		{
			name: "unsupported step type",
			workflow: Workflow{
				Name: "Test Workflow",
				Steps: []Step{
					{Name: "hello", Type: "random"},
				},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.workflow)

			if tt.shouldErr && err == nil {
				t.Fatal("expected validation error")
			}

			if !tt.shouldErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}
