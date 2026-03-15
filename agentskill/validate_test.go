package agentskill

import "testing"

func TestValidate_Valid(t *testing.T) {
	s := &Skill{Name: "foo", Description: "some description"}
	result := Validate(s, "foo")
	if !result.OK() {
		t.Errorf("expected OK, got errors: %v", result.Errors)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("expected no warnings, got: %v", result.Warnings)
	}
}

func TestValidate_EmptyDescription(t *testing.T) {
	s := &Skill{Name: "foo", Description: ""}
	result := Validate(s, "foo")
	if result.OK() {
		t.Error("expected error for empty description, got OK")
	}
}

func TestValidate_NameMismatch(t *testing.T) {
	s := &Skill{Name: "other-name", Description: "some description"}
	result := Validate(s, "foo")
	if !result.OK() {
		t.Errorf("expected OK (name mismatch is a warning), got errors: %v", result.Errors)
	}
	if len(result.Warnings) == 0 {
		t.Error("expected warning for name mismatch, got none")
	}
}

func TestValidate_NameTooLong(t *testing.T) {
	longName := "a"
	for i := 0; i < 64; i++ {
		longName += "a"
	}
	s := &Skill{Name: longName, Description: "some description"}
	result := Validate(s, longName)
	if !result.OK() {
		t.Errorf("expected OK (long name is a warning), got errors: %v", result.Errors)
	}
	if len(result.Warnings) == 0 {
		t.Error("expected warning for long name, got none")
	}
}

func TestValidate_NoNameNoDir(t *testing.T) {
	s := &Skill{Name: "", Description: "desc"}
	result := Validate(s, "")
	if !result.OK() {
		t.Errorf("expected OK, got errors: %v", result.Errors)
	}
}
