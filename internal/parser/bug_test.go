package parser

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// TestFieldsWithoutOptionsNotTransformed tests that fields without field options are not transformed
func TestFieldsWithoutOptionsNotTransformed(t *testing.T) {
	// Create a proto file descriptor that mimics bug_test.proto
	protoFile := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("bug_test.proto"),
		Package: proto.String("bug_test"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("TestResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("results"),
						Number:   proto.Int32(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".bug_test.TestResult"),
						// NO OPTIONS - this is the key test case
					},
				},
			},
			{
				Name: proto.String("TestResult"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("name"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
					{
						Name:   proto.String("passed"),
						Number: proto.Int32(2),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum(),
					},
				},
			},
		},
	}

	req := &pluginpb.CodeGeneratorRequest{
		ProtoFile: []*descriptorpb.FileDescriptorProto{protoFile},
	}

	// Parse annotated fields
	annotatedFields, err := FindAnnotatedFields(req)
	if err != nil {
		t.Fatalf("FindAnnotatedFields failed: %v", err)
	}

	// The key test: fields without options should NOT be in the annotated fields
	if annotatedFields.Contains("Results") {
		t.Error("Field 'Results' should NOT be annotated (has no field options)")
	}

	// Verify the annotated fields collection is empty since no fields have options
	if annotatedFields.Count() != 0 {
		t.Errorf("Expected 0 annotated fields, got %d", annotatedFields.Count())
	}
}