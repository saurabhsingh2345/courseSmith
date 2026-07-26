package pipeline

import "testing"

func TestValidateMermaidSyntax(t *testing.T) {
	ok := []string{
		"flowchart TD\n  A[Start] --> B{Empty?}\n  B -->|yes| C[Grow]\n  B -->|no| D[Append]",
		"graph LR\n A-->B",
		"sequenceDiagram\n  Alice->>Bob: Hi",
		"stateDiagram-v2\n  [*] --> Idle",
		"classDiagram\n  Animal <|-- Dog",
		"erDiagram\n  CUSTOMER ||--o{ ORDER : places",
		"%%{init: {'theme':'neutral'}}%%\nflowchart TD\n A-->B", // directive before header is allowed
	}
	for _, s := range ok {
		if err := validateMermaidSyntax(s); err != nil {
			t.Errorf("validateMermaidSyntax rejected valid source %q: %v", firstWords(s, 3), err)
		}
	}

	bad := []string{
		"",
		"   \n  \n",
		"Here is a flowchart for you:\nflowchart TD\n A-->B", // prose first
		"just some prose about lists",
	}
	for _, s := range bad {
		if err := validateMermaidSyntax(s); err == nil {
			t.Errorf("validateMermaidSyntax accepted invalid source %q", s)
		}
	}
}

func TestValidateCompiledSVG(t *testing.T) {
	if err := validateCompiledSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><style>.a{fill:red}</style><g><rect/></g></svg>`)); err != nil {
		t.Errorf("valid compiled SVG rejected: %v", err)
	}
	// No viewBox is fine for compiled output (mermaid omits it).
	if err := validateCompiledSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><text>x</text></svg>`)); err != nil {
		t.Errorf("viewBox-less SVG rejected: %v", err)
	}

	bad := map[string]string{
		"not-svg-root": `<div><svg/></div>`,
		"has-script":   `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`,
		"malformed":    `<svg><g></svg>`,
		"empty":        ``,
	}
	for name, s := range bad {
		if err := validateCompiledSVG([]byte(s)); err == nil {
			t.Errorf("validateCompiledSVG accepted %s: %q", name, s)
		}
	}
}
