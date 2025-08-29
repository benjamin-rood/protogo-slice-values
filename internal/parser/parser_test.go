package parser

import (
	"testing"

	"github.com/benjamin-rood/protogo-values/internal/parser/types"
	"github.com/benjamin-rood/protogo-values/proto/protogo_values"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

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

func TestShouldUseValueSlice(t *testing.T) {
	tests := []struct {
		name     string
		field    *descriptorpb.FieldDescriptorProto
		expected bool
	}{
		{
			name: "no options",
			field: &descriptorpb.FieldDescriptorProto{
				Name: proto.String("test_field"),
			},
			expected: false,
		},
		{
			name: "value_slice true",
			field: &descriptorpb.FieldDescriptorProto{
				Name: proto.String("test_field"),
				Options: func() *descriptorpb.FieldOptions {
					opts := &descriptorpb.FieldOptions{}
					proto.SetExtension(opts, protogo_values.E_ValueSlice, true)
					return opts
				}(),
			},
			expected: true,
		},
		{
			name: "value_slice false",
			field: &descriptorpb.FieldDescriptorProto{
				Name: proto.String("test_field"),
				Options: func() *descriptorpb.FieldOptions {
					opts := &descriptorpb.FieldOptions{}
					proto.SetExtension(opts, protogo_values.E_ValueSlice, false)
					return opts
				}(),
			},
			expected: false,
		},
		{
			name: "structured field_opts true",
			field: &descriptorpb.FieldDescriptorProto{
				Name: proto.String("test_field"),
				Options: func() *descriptorpb.FieldOptions {
					opts := &descriptorpb.FieldOptions{}
					fieldOpts := &protogo_values.FieldOptions{
						ValueSlice: proto.Bool(true),
					}
					proto.SetExtension(opts, protogo_values.E_FieldOpts, fieldOpts)
					return opts
				}(),
			},
			expected: true,
		},
		{
			name: "structured field_opts false",
			field: &descriptorpb.FieldDescriptorProto{
				Name: proto.String("test_field"),
				Options: func() *descriptorpb.FieldOptions {
					opts := &descriptorpb.FieldOptions{}
					fieldOpts := &protogo_values.FieldOptions{
						ValueSlice: proto.Bool(false),
					}
					proto.SetExtension(opts, protogo_values.E_FieldOpts, fieldOpts)
					return opts
				}(),
			},
			expected: false,
		},
		{
			name: "structured field_opts nil value_slice",
			field: &descriptorpb.FieldDescriptorProto{
				Name: proto.String("test_field"),
				Options: func() *descriptorpb.FieldOptions {
					opts := &descriptorpb.FieldOptions{}
					fieldOpts := &protogo_values.FieldOptions{
						ValueSlice: nil, // nil means not set
					}
					proto.SetExtension(opts, protogo_values.E_FieldOpts, fieldOpts)
					return opts
				}(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldUseValueSlice(tt.field)
			if result != tt.expected {
				t.Errorf("shouldUseValueSlice() = %t, expected %t", result, tt.expected)
			}
		})
	}
}

func TestFindAnnotatedFieldsWithFieldOptions(t *testing.T) {
	// Create a proto file descriptor with field options
	protoFile := &descriptorpb.FileDescriptorProto{
		Name: proto.String("test.proto"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("users_with_option"),
						Number: proto.Int32(1),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						Options: func() *descriptorpb.FieldOptions {
							opts := &descriptorpb.FieldOptions{}
							proto.SetExtension(opts, protogo_values.E_ValueSlice, true)
							return opts
						}(),
					},
					{
						Name:   proto.String("products_with_struct_option"),
						Number: proto.Int32(2),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						Options: func() *descriptorpb.FieldOptions {
							opts := &descriptorpb.FieldOptions{}
							fieldOpts := &protogo_values.FieldOptions{
								ValueSlice: proto.Bool(true),
							}
							proto.SetExtension(opts, protogo_values.E_FieldOpts, fieldOpts)
							return opts
						}(),
					},
					{
						Name:   proto.String("users_without_option"),
						Number: proto.Int32(3),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						// No field options
					},
					{
						Name:   proto.String("products_explicit_false"),
						Number: proto.Int32(4),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						Options: func() *descriptorpb.FieldOptions {
							opts := &descriptorpb.FieldOptions{}
							proto.SetExtension(opts, protogo_values.E_ValueSlice, false)
							return opts
						}(),
					},
					{
						Name:   proto.String("tags"),
						Number: proto.Int32(5),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), // Primitive, should be ignored
						Options: func() *descriptorpb.FieldOptions {
							opts := &descriptorpb.FieldOptions{}
							proto.SetExtension(opts, protogo_values.E_ValueSlice, true)
							return opts
						}(),
					},
					{
						Name:   proto.String("single_user"),
						Number: proto.Int32(6),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), // Not repeated
						Type:   descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						Options: func() *descriptorpb.FieldOptions {
							opts := &descriptorpb.FieldOptions{}
							proto.SetExtension(opts, protogo_values.E_ValueSlice, true)
							return opts
						}(),
					},
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

	// Should only find 2 fields: users_with_option and products_with_struct_option
	// Other fields are excluded because they are:
	// - users_without_option: no field option
	// - products_explicit_false: explicitly set to false
	// - tags: primitive type (ignored)
	// - single_user: not repeated (ignored)
	if fields.Count() != 2 {
		t.Errorf("Expected 2 annotated fields, got %d", fields.Count())
	}

	if !fields.Contains("UsersWithOption") {
		t.Error("Expected UsersWithOption field to be annotated")
	}

	if !fields.Contains("ProductsWithStructOption") {
		t.Error("Expected ProductsWithStructOption field to be annotated")
	}

	if fields.Contains("UsersWithoutOption") {
		t.Error("UsersWithoutOption field should not be annotated")
	}

	if fields.Contains("ProductsExplicitFalse") {
		t.Error("ProductsExplicitFalse field should not be annotated")
	}

	if fields.Contains("Tags") {
		t.Error("Tags field should not be annotated (primitive type)")
	}

	if fields.Contains("SingleUser") {
		t.Error("SingleUser field should not be annotated (not repeated)")
	}
}

// Test error handling and edge cases
func TestFindAnnotatedFieldsEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		request *pluginpb.CodeGeneratorRequest
		wantErr bool
	}{
		{
			name: "nil request",
			request: nil,
			wantErr: true,
		},
		{
			name: "empty request",
			request: &pluginpb.CodeGeneratorRequest{
				ProtoFile: []*descriptorpb.FileDescriptorProto{},
			},
			wantErr: false,
		},
		{
			name: "nil proto file slice",
			request: &pluginpb.CodeGeneratorRequest{
				ProtoFile: nil,
			},
			wantErr: false,
		},
		{
			name: "proto file with nil message slice",
			request: &pluginpb.CodeGeneratorRequest{
				ProtoFile: []*descriptorpb.FileDescriptorProto{
					{
						Name:        proto.String("test.proto"),
						MessageType: nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "message with nil field slice",
			request: &pluginpb.CodeGeneratorRequest{
				ProtoFile: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("test.proto"),
						MessageType: []*descriptorpb.DescriptorProto{
							{
								Name:  proto.String("TestMessage"),
								Field: nil,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "field with nil options",
			request: &pluginpb.CodeGeneratorRequest{
				ProtoFile: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("test.proto"),
						MessageType: []*descriptorpb.DescriptorProto{
							{
								Name: proto.String("TestMessage"),
								Field: []*descriptorpb.FieldDescriptorProto{
									{
										Name:    proto.String("test_field"),
										Number:  proto.Int32(1),
										Label:   descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
										Type:    descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
										Options: nil,
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields, err := FindAnnotatedFields(tt.request)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("FindAnnotatedFields() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("FindAnnotatedFields() unexpected error: %v", err)
				return
			}
			
			if fields == nil {
				t.Errorf("FindAnnotatedFields() returned nil fields")
			}
		})
	}
}

// Test field option combinations and precedence
func TestShouldUseValueSliceAdvanced(t *testing.T) {
	tests := []struct {
		name     string
		field    *descriptorpb.FieldDescriptorProto
		expected bool
	}{
		{
			name: "both extensions set - simple extension takes precedence",
			field: &descriptorpb.FieldDescriptorProto{
				Name: proto.String("test_field"),
				Options: func() *descriptorpb.FieldOptions {
					opts := &descriptorpb.FieldOptions{}
					// Set simple extension to true
					proto.SetExtension(opts, protogo_values.E_ValueSlice, true)
					// Set structured extension to false
					fieldOpts := &protogo_values.FieldOptions{
						ValueSlice: proto.Bool(false),
					}
					proto.SetExtension(opts, protogo_values.E_FieldOpts, fieldOpts)
					return opts
				}(),
			},
			expected: true, // Simple extension should win
		},
		{
			name: "structured extension with empty FieldOptions",
			field: &descriptorpb.FieldDescriptorProto{
				Name: proto.String("test_field"),
				Options: func() *descriptorpb.FieldOptions {
					opts := &descriptorpb.FieldOptions{}
					fieldOpts := &protogo_values.FieldOptions{
						// No fields set
					}
					proto.SetExtension(opts, protogo_values.E_FieldOpts, fieldOpts)
					return opts
				}(),
			},
			expected: false,
		},
		{
			name: "field with empty options object",
			field: &descriptorpb.FieldDescriptorProto{
				Name:    proto.String("test_field"),
				Options: &descriptorpb.FieldOptions{}, // Empty but not nil
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldUseValueSlice(tt.field)
			if result != tt.expected {
				t.Errorf("shouldUseValueSlice() = %t, expected %t", result, tt.expected)
			}
		})
	}
}

// Test nested message handling
func TestProcessMessageNested(t *testing.T) {
	fields := types.NewAnnotatedFields()
	
	// Create a message with nested messages
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("OuterMessage"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   proto.String("outer_field"),
				Number: proto.Int32(1),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				Options: func() *descriptorpb.FieldOptions {
					opts := &descriptorpb.FieldOptions{}
					proto.SetExtension(opts, protogo_values.E_ValueSlice, true)
					return opts
				}(),
			},
		},
		NestedType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("InnerMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("inner_field"),
						Number: proto.Int32(1),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						Options: func() *descriptorpb.FieldOptions {
							opts := &descriptorpb.FieldOptions{}
							proto.SetExtension(opts, protogo_values.E_ValueSlice, true)
							return opts
						}(),
					},
				},
			},
		},
	}

	err := processMessage(msg, fields)
	if err != nil {
		t.Errorf("processMessage() unexpected error: %v", err)
	}

	// Should find the outer field
	if !fields.Contains("OuterField") {
		t.Error("Expected OuterField to be found")
	}

	// Note: Nested message processing doesn't currently recurse into nested types
	// This is by design - only top-level message fields are processed
	if fields.Contains("InnerField") {
		t.Error("InnerField should not be found (nested type processing not implemented)")
	}
}

// Test malformed field names
func TestToGoFieldNameEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		protoName string
		expected  string
	}{
		{"empty string", "", ""},
		{"single underscore", "_", ""},
		{"multiple underscores", "___", ""},
		{"numbers only", "123", "123"},
		{"mixed numbers and underscores", "1_2_3", "123"},
		{"special characters remain", "field-name", "Field-name"},
		{"unicode characters", "field_测试", "Field测试"},
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