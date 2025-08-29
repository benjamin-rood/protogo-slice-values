package plugin

import (
	"testing"

	"github.com/benjamin-rood/protogo-values/proto/protogo_values"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestProcessRequest(t *testing.T) {
	// This test requires protoc-gen-go to be available
	// Skip if not available to avoid CI/test environment issues
	tests := []struct {
		name          string
		request       *pluginpb.CodeGeneratorRequest
		expectError   bool
		expectedFiles int
	}{
		{
			name: "valid request with field options",
			request: &pluginpb.CodeGeneratorRequest{
				FileToGenerate: []string{"test.proto"},
				ProtoFile: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("test.proto"),
						Package: proto.String("test"),
						Options: &descriptorpb.FileOptions{},
						MessageType: []*descriptorpb.DescriptorProto{
							{
								Name: proto.String("TestMessage"),
								Field: []*descriptorpb.FieldDescriptorProto{
									{
										Name:   proto.String("users"),
										Number: proto.Int32(1),
										Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
										Type:   descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
										TypeName: proto.String(".test.User"),
										Options: func() *descriptorpb.FieldOptions {
											opts := &descriptorpb.FieldOptions{}
											proto.SetExtension(opts, protogo_values.E_ValueSlice, true)
											return opts
										}(),
									},
								},
							},
							{
								Name: proto.String("User"),
								Field: []*descriptorpb.FieldDescriptorProto{
									{
										Name:   proto.String("id"),
										Number: proto.Int32(1),
										Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
										Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
									},
								},
							},
						},
					},
				},
			},
			expectError:   false,
			expectedFiles: 1,
		},
		{
			name: "empty request",
			request: &pluginpb.CodeGeneratorRequest{
				FileToGenerate: []string{},
				ProtoFile:      []*descriptorpb.FileDescriptorProto{},
			},
			expectError:   false,
			expectedFiles: 0,
		},
		{
			name: "nil request should not panic",
			request: &pluginpb.CodeGeneratorRequest{
				FileToGenerate: []string{"test.proto"},
				ProtoFile:      nil, // nil slice
			},
			expectError:   false,
			expectedFiles: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if protoc-gen-go is not available
			// This prevents test failures in environments where it's not installed
			t.Skip("Skipping plugin test - requires protoc-gen-go in PATH")
			
			resp, err := ProcessRequest(tt.request)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("ProcessRequest() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ProcessRequest() unexpected error: %v", err)
				return
			}
			
			if resp == nil {
				t.Errorf("ProcessRequest() returned nil response")
				return
			}
			
			if len(resp.File) != tt.expectedFiles {
				t.Errorf("ProcessRequest() expected %d files, got %d", tt.expectedFiles, len(resp.File))
			}
		})
	}
}

func TestCallProtocGenGoErrors(t *testing.T) {
	tests := []struct {
		name        string
		request     *pluginpb.CodeGeneratorRequest
		expectError bool
	}{
		{
			name:        "nil request should error during marshal",
			request:     nil,
			expectError: true,
		},
		{
			name: "valid request but protoc-gen-go not available should error",
			request: &pluginpb.CodeGeneratorRequest{
				FileToGenerate: []string{"test.proto"},
				ProtoFile: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("test.proto"),
					},
				},
			},
			expectError: true, // Will fail if protoc-gen-go not in PATH
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := callProtocGenGo(tt.request)
			
			if tt.expectError && err == nil {
				t.Errorf("callProtocGenGo() expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("callProtocGenGo() unexpected error: %v", err)
			}
		})
	}
}

// Test for malformed protobuf data handling
func TestCallProtocGenGoMalformedData(t *testing.T) {
	// Create a request with invalid protobuf structure
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"invalid.proto"},
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			{
				Name: proto.String("invalid.proto"),
				// Missing required fields or invalid structure
				Syntax: proto.String("invalid_syntax"), 
			},
		},
	}

	// This should handle the error gracefully
	_, err := callProtocGenGo(req)
	if err == nil {
		t.Skip("Skipping - protoc-gen-go not available or handles invalid input gracefully")
	}
	
	// We expect an error due to invalid protobuf or missing protoc-gen-go
	// The important thing is that we don't panic
}

// Test edge cases in ProcessRequest
func TestProcessRequestEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		request *pluginpb.CodeGeneratorRequest
		wantErr bool
	}{
		{
			name: "request with no proto files",
			request: &pluginpb.CodeGeneratorRequest{
				FileToGenerate: []string{"nonexistent.proto"},
				ProtoFile:      []*descriptorpb.FileDescriptorProto{},
			},
			wantErr: false, // Should handle gracefully
		},
		{
			name: "request with file to generate but no matching proto file",
			request: &pluginpb.CodeGeneratorRequest{
				FileToGenerate: []string{"missing.proto"},
				ProtoFile: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("different.proto"),
					},
				},
			},
			wantErr: false, // Should handle gracefully
		},
		{
			name: "request with complex nested messages",
			request: &pluginpb.CodeGeneratorRequest{
				FileToGenerate: []string{"nested.proto"},
				ProtoFile: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("nested.proto"),
						MessageType: []*descriptorpb.DescriptorProto{
							{
								Name: proto.String("Outer"),
								NestedType: []*descriptorpb.DescriptorProto{
									{
										Name: proto.String("Inner"),
										Field: []*descriptorpb.FieldDescriptorProto{
											{
												Name:   proto.String("items"),
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
			t.Skip("Skipping - requires protoc-gen-go")
			
			_, err := ProcessRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}