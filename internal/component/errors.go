package component

import "strings"

type (
	UnknownComponentError struct {
		ID string
	}

	MissingDependencyError struct {
		ComponentID  string
		DependencyID string
	}

	MissingRequirementError struct {
		ComponentID string
		Requires    []string
	}

	ConflictError struct {
		ComponentID string
		ConflictID  string
	}

	DependencyCycleError struct {
		ComponentID string
	}
)

func (e *MissingRequirementError) Error() string {
	return "component " + e.ComponentID + " requires one of: " + strings.Join(e.Requires, ", ")
}

func (e *UnknownComponentError) Error() string {
	return "unknown component " + e.ID
}

func (e *MissingDependencyError) Error() string {
	return "component " + e.ComponentID + " requires missing component " + e.DependencyID
}

func (e *ConflictError) Error() string {
	return "component " + e.ComponentID + " conflicts with " + e.ConflictID
}

func (e *DependencyCycleError) Error() string {
	return "component dependency cycle includes " + e.ComponentID
}
