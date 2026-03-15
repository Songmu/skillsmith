package agentskills

import "fmt"

// ValidationResult holds the warnings and errors produced by Validate.
type ValidationResult struct {
	Warnings []string
	Errors   []string
}

// OK reports whether there are no errors (warnings are allowed).
func (v *ValidationResult) OK() bool {
	return len(v.Errors) == 0
}

// Validate performs lenient validation of a Skill against its directory name.
// Warnings do not prevent installation; errors cause the skill to be skipped.
func Validate(s *Skill, dirName string) *ValidationResult {
	result := &ValidationResult{}

	// description empty/missing → Error (skip skill)
	if s.Description == "" {
		result.Errors = append(result.Errors, "description is empty or missing")
	}

	// name mismatch with directory → Warning (not error)
	if s.Name != "" && dirName != "" && s.Name != dirName {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("skill name %q does not match directory name %q", s.Name, dirName))
	}

	// name > 64 chars → Warning
	if len(s.Name) > 64 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("skill name %q exceeds 64 characters", s.Name))
	}

	return result
}
