package transform

import (
	"testing"

	"github.com/benjamin-rood/protoc-gen-go-value-slices/internal/parser/types"
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