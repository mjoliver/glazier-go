package template

import (
	"bytes"
	"fmt"
	"text/template"
)

// Process applies Go text/template to the input data using BuildInfo context.
func Process(data []byte, info *BuildInfo) ([]byte, error) {
	if info == nil {
		// No template processing if BuildInfo is nil
		return data, nil
	}

	tmpl, err := template.New("config").Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, info); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
