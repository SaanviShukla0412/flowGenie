package execution

import "testing"

func TestContextSetAndGetOutput(t *testing.T) {
	ctx := NewContext()

	ctx.SetOutput(
		"Create User",
		map[string]interface{}{
			"id":   "123",
			"name": "Saanvi",
		},
	)

	output, ok := ctx.GetOutput("Create User")

	if !ok {
		t.Fatal("expected output to exist")
	}

	outputMap, ok := output.(map[string]interface{})
	if !ok {
		t.Fatalf("expected output to be a map, got %T", output)
	}

	if outputMap["id"] != "123" {
		t.Errorf("expected id=123, got %v", outputMap["id"])
	}

	if outputMap["name"] != "Saanvi" {
		t.Errorf("expected name=Saanvi, got %v", outputMap["name"])
	}
}

func TestContextGetMissingOutput(t *testing.T) {
	ctx := NewContext()

	_, ok := ctx.GetOutput("Unknown Step")

	if ok {
		t.Fatal("expected missing output to return false")
	}
}
