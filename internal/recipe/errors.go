package recipe

import "strings"

type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	if e == nil || len(e.Problems) == 0 {
		return "recipe validation failed"
	}

	var builder strings.Builder
	builder.WriteString("recipe validation failed:")
	for _, problem := range e.Problems {
		builder.WriteString("\n- ")
		builder.WriteString(problem)
	}
	return builder.String()
}

func validationFailed(problems ...string) error {
	return &ValidationError{Problems: problems}
}
