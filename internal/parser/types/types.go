package types

// AnnotatedFields holds the set of field names that should be converted from pointer slices to value slices
type AnnotatedFields struct {
	fields map[string]bool
}

// NewAnnotatedFields creates a new AnnotatedFields instance
func NewAnnotatedFields() *AnnotatedFields {
	return &AnnotatedFields{
		fields: make(map[string]bool),
	}
}

// Add adds a field name to the set
func (af *AnnotatedFields) Add(fieldName string) {
	af.fields[fieldName] = true
}

// Contains checks if a field name is in the set
func (af *AnnotatedFields) Contains(fieldName string) bool {
	return af.fields[fieldName]
}

// All returns all field names
func (af *AnnotatedFields) All() map[string]bool {
	// Return a copy to prevent external modification
	result := make(map[string]bool)
	for k, v := range af.fields {
		result[k] = v
	}
	return result
}

// Count returns the number of annotated fields
func (af *AnnotatedFields) Count() int {
	return len(af.fields)
}