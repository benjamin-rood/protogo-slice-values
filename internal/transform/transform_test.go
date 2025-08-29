package transform

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/benjamin-rood/protogo-values/internal/parser/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestTransformField(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		fieldName string
		expected  string
	}{
		{
			name:      "basic field transformation",
			content:   "type Message struct {\n\tUsers []*User `protobuf:\"bytes,1,rep,name=users\"`\n}",
			fieldName: "Users",
			expected:  "type Message struct {\n\tUsers []User `protobuf:\"bytes,1,rep,name=users\"`\n}",
		},
		{
			name:      "getter method transformation",
			content:   "func (m *Message) GetUsers() []*User {\n\treturn m.Users\n}",
			fieldName: "Users",
			expected:  "func (m *Message) GetUsers() []User {\n\treturn m.Users\n}",
		},
		{
			name:      "both field and getter",
			content:   "type Message struct {\n\tUsers []*User\n}\nfunc (m *Message) GetUsers() []*User {\n\treturn m.Users\n}",
			fieldName: "Users",
			expected:  "type Message struct {\n\tUsers []User\n}\nfunc (m *Message) GetUsers() []User {\n\treturn m.Users\n}",
		},
		{
			name:      "no transformation needed",
			content:   "type Message struct {\n\tUsers []User\n}",
			fieldName: "Users",
			expected:  "type Message struct {\n\tUsers []User\n}",
		},
		{
			name:      "field not present",
			content:   "type Message struct {\n\tProducts []*Product\n}",
			fieldName: "Users",
			expected:  "type Message struct {\n\tProducts []*Product\n}",
		},
		{
			name:      "multiple occurrences",
			content:   "Users []*User\nUsers []*User\nGetUsers() []*User",
			fieldName: "Users",
			expected:  "Users []User\nUsers []User\nGetUsers() []User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformField(tt.content, tt.fieldName)
			if result != tt.expected {
				t.Errorf("transformField() failed:\nInput:\n%s\nExpected:\n%s\nGot:\n%s",
					tt.content, tt.expected, result)
			}
		})
	}
}

func TestTransformPointerSlices(t *testing.T) {
	fields := types.NewAnnotatedFields()
	fields.Add("Users")
	fields.Add("Products")

	content := `type Message struct {
	Users []*User
	Products []*Product
	Tags []string
}

func (m *Message) GetUsers() []*User {
	return m.Users
}

func (m *Message) GetProducts() []*Product {
	return m.Products
}`

	expected := `type Message struct {
	Users []User
	Products []Product
	Tags []string
}

func (m *Message) GetUsers() []User {
	return m.Users
}

func (m *Message) GetProducts() []Product {
	return m.Products
}`

	result := transformPointerSlices(content, fields)
	if result != expected {
		t.Errorf("transformPointerSlices() failed:\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestApplyTransformations(t *testing.T) {
	fields := types.NewAnnotatedFields()
	fields.Add("Users")

	content := "type Message struct {\n\tUsers []*User\n}"
	expected := "type Message struct {\n\tUsers []User\n}"

	resp := &pluginpb.CodeGeneratorResponse{
		File: []*pluginpb.CodeGeneratorResponse_File{
			{
				Name:    proto.String("test.pb.go"),
				Content: &content,
			},
		},
	}

	err := ApplyTransformations(resp, fields)
	if err != nil {
		t.Errorf("ApplyTransformations() returned error: %v", err)
	}

	if resp.File[0].GetContent() != expected {
		t.Errorf("ApplyTransformations() failed:\nExpected: %s\nGot: %s",
			expected, resp.File[0].GetContent())
	}
}

func TestApplyTransformationsNoContent(t *testing.T) {
	fields := types.NewAnnotatedFields()
	fields.Add("Users")

	resp := &pluginpb.CodeGeneratorResponse{
		File: []*pluginpb.CodeGeneratorResponse_File{
			{
				Name: proto.String("test.pb.go"),
				// Content is nil
			},
		},
	}

	err := ApplyTransformations(resp, fields)
	if err != nil {
		t.Errorf("ApplyTransformations() returned error for nil content: %v", err)
	}
}

// Test error handling and edge cases
func TestApplyTransformationsEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		resp   *pluginpb.CodeGeneratorResponse
		fields *types.AnnotatedFields
		wantErr bool
	}{
		{
			name:    "nil response",
			resp:    nil,
			fields:  types.NewAnnotatedFields(),
			wantErr: true,
		},
		{
			name: "nil fields",
			resp: &pluginpb.CodeGeneratorResponse{
				File: []*pluginpb.CodeGeneratorResponse_File{},
			},
			fields:  nil,
			wantErr: true,
		},
		{
			name: "response with nil file slice",
			resp: &pluginpb.CodeGeneratorResponse{
				File: nil,
			},
			fields:  types.NewAnnotatedFields(),
			wantErr: false,
		},
		{
			name: "empty response",
			resp: &pluginpb.CodeGeneratorResponse{
				File: []*pluginpb.CodeGeneratorResponse_File{},
			},
			fields:  types.NewAnnotatedFields(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ApplyTransformations(tt.resp, tt.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyTransformations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test transformation with various field name patterns
func TestTransformFieldAdvanced(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		fieldName string
		expected  string
	}{
		{
			name:      "field with same prefix",
			content:   "Users []*User\nUsersList []*User",
			fieldName: "Users",
			expected:  "Users []User\nUsersList []*User", // Only exact match should be transformed
		},
		{
			name:      "field in struct with tabs",
			content:   "type Message struct {\n\tUsers\t[]*User\t`protobuf:\"bytes,1,rep,name=users\"`\n}",
			fieldName: "Users",
			expected:  "type Message struct {\n\tUsers\t[]User\t`protobuf:\"bytes,1,rep,name=users\"`\n}",
		},
		{
			name:      "multiple spaces around field",
			content:   "  Users   []*User  ",
			fieldName: "Users",
			expected:  "  Users   []User  ",
		},
		{
			name:      "field in comment should not be transformed",
			content:   "// Users []*User is the field\nUsers []*User",
			fieldName: "Users",
			expected:  "// Users []*User is the field\nUsers []User",
		},
		{
			name:      "empty field name should not transform anything",
			content:   "Users []*User",
			fieldName: "",
			expected:  "Users []*User",
		},
		{
			name:      "field name with special characters",
			content:   "Field_With_Underscores []*Type",
			fieldName: "Field_With_Underscores",
			expected:  "Field_With_Underscores []Type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformField(tt.content, tt.fieldName)
			if result != tt.expected {
				t.Errorf("transformField() failed:\nInput:\n%s\nField: %s\nExpected:\n%s\nGot:\n%s",
					tt.content, tt.fieldName, tt.expected, result)
			}
		})
	}
}

// Test transformation with large content
func TestTransformFieldPerformance(t *testing.T) {
	// Create large content to test performance
	var content strings.Builder
	fieldNames := []string{"Users", "Products", "Items", "Data", "Messages"}
	
	for i := 0; i < 1000; i++ {
		for _, field := range fieldNames {
			content.WriteString(fmt.Sprintf("type Message%d struct {\n", i))
			content.WriteString(fmt.Sprintf("\t%s []*%sType\n", field, field))
			content.WriteString("}\n")
			content.WriteString(fmt.Sprintf("func (m *Message%d) Get%s() []*%sType {\n", i, field, field))
			content.WriteString(fmt.Sprintf("\treturn m.%s\n", field))
			content.WriteString("}\n\n")
		}
	}

	largeContent := content.String()
	fields := types.NewAnnotatedFields()
	for _, field := range fieldNames {
		fields.Add(field)
	}

	// Time the transformation
	start := time.Now()
	result := transformPointerSlices(largeContent, fields)
	duration := time.Since(start)

	t.Logf("Transformed %d characters in %v", len(largeContent), duration)

	// Verify some transformations occurred
	if !strings.Contains(result, "Users []UsersType") {
		t.Error("Expected transformation not found in large content")
	}

	if strings.Contains(result, "Users []*UsersType") {
		t.Error("Untransformed content found - transformation incomplete")
	}
}

// Test concurrent access to transformation (if this becomes relevant)
func TestTransformFieldConcurrent(t *testing.T) {
	content := `type Message struct {
		Users []*User
		Products []*Product
		Items []*Item
	}
	func (m *Message) GetUsers() []*User { return m.Users }
	func (m *Message) GetProducts() []*Product { return m.Products }
	func (m *Message) GetItems() []*Item { return m.Items }`

	fields := types.NewAnnotatedFields()
	fields.Add("Users")
	fields.Add("Products")
	fields.Add("Items")

	// Run transformation concurrently
	const numGoroutines = 10
	var wg sync.WaitGroup
	results := make([]string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = transformPointerSlices(content, fields)
		}(i)
	}

	wg.Wait()

	// All results should be identical
	expected := results[0]
	for i, result := range results {
		if result != expected {
			t.Errorf("Concurrent run %d produced different result", i)
		}
	}

	// Verify transformation worked
	if !strings.Contains(expected, "Users []User") {
		t.Error("Expected transformation not found")
	}
}
