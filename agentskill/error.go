package agentskill

import "fmt"

// SkillError is an error associated with a specific skill directory.
// It is returned (potentially wrapped in a joined error) by [Discover].
// Callers can use [errors.As] to extract the Dir field.
type SkillError struct {
	Dir string
	Err error
}

func (e *SkillError) Error() string {
	return fmt.Sprintf("%s: %v", e.Dir, e.Err)
}

func (e *SkillError) Unwrap() error {
	return e.Err
}
