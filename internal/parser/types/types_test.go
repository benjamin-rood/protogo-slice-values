package types

import "testing"

func TestAnnotatedFields(t *testing.T) {
	fields := NewAnnotatedFields()

	// Test initial state
	if fields.Count() != 0 {
		t.Errorf("Expected count 0, got %d", fields.Count())
	}

	if fields.Contains("NonExistent") {
		t.Error("Expected false for non-existent field")
	}

	// Test adding fields
	fields.Add("Users")
	fields.Add("Products")

	if fields.Count() != 2 {
		t.Errorf("Expected count 2, got %d", fields.Count())
	}

	if !fields.Contains("Users") {
		t.Error("Expected true for Users field")
	}

	if !fields.Contains("Products") {
		t.Error("Expected true for Products field")
	}

	if fields.Contains("NonExistent") {
		t.Error("Expected false for non-existent field")
	}

	// Test adding duplicate
	fields.Add("Users")
	if fields.Count() != 2 {
		t.Errorf("Expected count to remain 2 after duplicate add, got %d", fields.Count())
	}

	// Test All() method
	all := fields.All()
	if len(all) != 2 {
		t.Errorf("Expected All() to return 2 fields, got %d", len(all))
	}

	if !all["Users"] || !all["Products"] {
		t.Error("All() should contain both Users and Products")
	}

	// Test that All() returns a copy (modification shouldn't affect original)
	all["NewField"] = true
	if fields.Contains("NewField") {
		t.Error("Modifying result of All() should not affect original fields")
	}
}

func TestAnnotatedFieldsEmpty(t *testing.T) {
	fields := NewAnnotatedFields()
	all := fields.All()

	if len(all) != 0 {
		t.Errorf("Expected empty map, got %d items", len(all))
	}
}