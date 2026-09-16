package execution

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ResolveTemplate(
	value string,
	ctx *Context,
) string {
	result := value

	for stepName := range ctx.Outputs {
		prefix := "{{" + stepName
		start := strings.Index(result, prefix)

		for start != -1 {
			end := strings.Index(result[start:], "}}")
			if end == -1 {
				break
			}

			end = start + end + 2

			token := result[start:end]
			path := strings.TrimSuffix(
				strings.TrimPrefix(token, "{{"),
				"}}",
			)

			resolved, ok := resolvePath(path, ctx)
			if !ok {
				start = strings.Index(result[start+1:], prefix)

				if start != -1 {
					start += 1
				}

				continue
			}

			result = result[:start] + resolved + result[end:]

			start = strings.Index(result, prefix)
		}
	}

	return result
}

func resolvePath(
	path string,
	ctx *Context,
) (string, bool) {
	parts := strings.Split(path, ".")

	if len(parts) == 0 {
		return "", false
	}

	output, ok := ctx.GetOutput(parts[0])
	if !ok {
		return "", false
	}

	current := output

	for _, part := range parts[1:] {
		object, ok := current.(map[string]interface{})
		if !ok {
			return "", false
		}

		current, ok = object[part]
		if !ok {
			return "", false
		}
	}

	switch value := current.(type) {
	case string:
		return value, true

	default:
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Sprint(value), true
		}

		return string(data), true
	}
}
