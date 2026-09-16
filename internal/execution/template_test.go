package execution

import "testing"

func TestResolveTemplateWholeOutput(t *testing.T) {
	ctx := NewContext()

	ctx.SetOutput(
		"Create User",
		map[string]interface{}{
			"id":   "123",
			"name": "Saanvi",
		},
	)

	result := ResolveTemplate(
		"{{Create User}}",
		ctx,
	)

	expected := `{"id":"123","name":"Saanvi"}`

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestResolveTemplateNestedField(t *testing.T) {
	ctx := NewContext()

	ctx.SetOutput(
		"Create User",
		map[string]interface{}{
			"json": map[string]interface{}{
				"name": "Saanvi",
				"role": "backend-engineer",
			},
		},
	)

	result := ResolveTemplate(
		"{{Create User.json.name}}",
		ctx,
	)

	if result != "Saanvi" {
		t.Errorf("expected Saanvi, got %s", result)
	}
}

func TestResolveTemplateMultipleValues(t *testing.T) {
	ctx := NewContext()

	ctx.SetOutput(
		"Create User",
		map[string]interface{}{
			"json": map[string]interface{}{
				"name": "Saanvi",
				"role": "backend-engineer",
			},
		},
	)

	result := ResolveTemplate(
		"User={{Create User.json.name}}, Role={{Create User.json.role}}",
		ctx,
	)

	expected := "User=Saanvi, Role=backend-engineer"

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestResolveTemplateMissingField(t *testing.T) {
	ctx := NewContext()

	ctx.SetOutput(
		"Create User",
		map[string]interface{}{
			"json": map[string]interface{}{
				"name": "Saanvi",
			},
		},
	)

	result := ResolveTemplate(
		"{{Create User.json.email}}",
		ctx,
	)

	if result != "{{Create User.json.email}}" {
		t.Errorf(
			"expected unresolved template to remain unchanged, got %s",
			result,
		)
	}
}
