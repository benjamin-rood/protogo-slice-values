package parser

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestPathToString(t *testing.T) {
	tests := []struct {
		name     string
		path     []int32
		expected string
	}{
		{"empty path", []int32{}, ""},
		{"single element", []int32{4}, "4"},
		{"multiple elements", []int32{4, 0, 2, 1}, "4.0.2.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pathToString(tt.path)
			if result != tt.expected {
				t.Errorf("pathToString(%v) = %q, expected %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestToGoFieldName(t *testing.T) {
	tests := []struct {
		name      string
		protoName string
		expected  string
	}{
		{"simple name", "user", "User"},
		{"snake case", "user_name", "UserName"},
		{"complex name", "user_full_profile_data", "UserFullProfileData"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toGoFieldName(tt.protoName)
			if result != tt.expected {
				t.Errorf("toGoFieldName(%q) = %q, expected %q", tt.protoName, result, tt.expected)
			}
		})
	}
}

func TestIsFieldAnnotated(t *testing.T) {
	tests := []struct {
		name     string
		comments string
		expected bool
	}{
		{"no annotation", "This is a regular field", false},
		{"nullable false", "@nullable=false", true},
		{"valueslice", "@valueslice", true},
		{"nullable false in sentence", "This field has @nullable=false annotation", true},
		{"valueslice in sentence", "Use @valueslice for this field", true},
		{"both annotations", "@nullable=false and @valueslice", true},
		{"empty comments", "", false},
		{"similar but not exact", "@nullable=true", false},
		{"case sensitive", "@VALUESLICE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFieldAnnotated(tt.comments)
			if result != tt.expected {
				t.Errorf("isFieldAnnotated(%q) = %t, expected %t", tt.comments, result, tt.expected)
			}
		})
	}
}

func TestBuildLocationMap(t *testing.T) {
	sourceInfo := &descriptorpb.SourceCodeInfo{
		Location: []*descriptorpb.SourceCodeInfo_Location{
			{
				Path:            []int32{4, 0},
				LeadingComments: proto.String("Message comment"),
			},
			{
				Path:            []int32{4, 0, 2, 0},
				LeadingComments: proto.String("Field comment"),
			},
		},
	}

	locationMap := buildLocationMap(sourceInfo)

	if len(locationMap) != 2 {
		t.Errorf("Expected 2 locations, got %d", len(locationMap))
	}

	if loc, ok := locationMap["4.0"]; !ok || loc.GetLeadingComments() != "Message comment" {
		t.Error("Message location not found or incorrect")
	}

	if loc, ok := locationMap["4.0.2.0"]; !ok || loc.GetLeadingComments() != "Field comment" {
		t.Error("Field location not found or incorrect")
	}
}

func TestBuildLocationMapNil(t *testing.T) {
	locationMap := buildLocationMap(nil)
	if len(locationMap) != 0 {
		t.Errorf("Expected empty map for nil input, got %d items", len(locationMap))
	}
}

func TestFindAnnotatedFieldsIntegration(t *testing.T) {
	// Create a minimal proto file descriptor with annotated fields
	protoFile := &descriptorpb.FileDescriptorProto{
		Name: proto.String("test.proto"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("users"),
						Number: proto.Int32(1),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
					},
					{
						Name:   proto.String("products"),
						Number: proto.Int32(2),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
					},
					{
						Name:   proto.String("tags"),
						Number: proto.Int32(3),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
				},
			},
		},
		SourceCodeInfo: &descriptorpb.SourceCodeInfo{
			Location: []*descriptorpb.SourceCodeInfo_Location{
				{
					Path:            []int32{4, 0, 2, 0}, // Path to users field
					LeadingComments: proto.String("@valueslice"),
				},
				{
					Path:            []int32{4, 0, 2, 1}, // Path to products field
					LeadingComments: proto.String("@nullable=false"),
				},
				{
					Path:            []int32{4, 0, 2, 2}, // Path to tags field (no annotation)
					LeadingComments: proto.String("Regular field"),
				},
			},
		},
	}

	req := &pluginpb.CodeGeneratorRequest{
		ProtoFile: []*descriptorpb.FileDescriptorProto{protoFile},
	}

	fields, err := FindAnnotatedFields(req)
	if err != nil {
		t.Errorf("FindAnnotatedFields() returned error: %v", err)
	}

	if fields.Count() != 2 {
		t.Errorf("Expected 2 annotated fields, got %d", fields.Count())
	}

	if !fields.Contains("Users") {
		t.Error("Expected Users field to be annotated")
	}

	if !fields.Contains("Products") {
		t.Error("Expected Products field to be annotated")
	}

	if fields.Contains("Tags") {
		t.Error("Tags field should not be annotated")
	}
}