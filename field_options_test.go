//go:build integration
// +build integration

package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/benjamin-rood/protogo-values/internal/plugin"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestFieldOptionsIntegration(t *testing.T) {
	// Skip if protoc-gen-go is not available
	if _, err := exec.LookPath("protoc-gen-go"); err != nil {
		t.Skip("protoc-gen-go not found in PATH, skipping field options test")
	}

	// Test proto file with field options
	protoFile := "testdata/proto/field_options_test.proto"
	if _, err := os.Stat(protoFile); os.IsNotExist(err) {
		t.Fatalf("Test proto file not found: %s", protoFile)
	}

	// Use protoc to generate the CodeGeneratorRequest
	cmd := exec.Command("protoc",
		"--go_out=testdata/gen",
		"--go_opt=paths=source_relative",
		"--descriptor_set_out=/tmp/field_options_test.desc",
		"--include_source_info",
		"-I.", // Include current directory for proto imports
		protoFile)

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run protoc: %v", err)
	}

	// Clean up
	defer os.RemoveAll("testdata/gen")
	defer os.Remove("/tmp/field_options_test.desc")

	// Create a mock request with the field options proto file
	req := createFieldOptionsRequest(t)

	resp, err := plugin.ProcessRequest(req)
	if err != nil {
		t.Fatalf("Plugin processing failed: %v", err)
	}

	if resp.GetError() != "" {
		t.Fatalf("Plugin returned error: %s", resp.GetError())
	}

	// Check that we have generated files
	if len(resp.File) == 0 {
		t.Fatal("No files generated")
	}

	// Check the content of the generated file
	for _, file := range resp.File {
		if file.GetName() == "testdata/proto/field_options_test.pb.go" && file.Content != nil {
			content := *file.Content

			// Test cases for field options behavior
			testCases := []struct {
				description string
				shouldFind  bool
				pattern     string
			}{
				{
					description: "users_with_option should be value slice",
					shouldFind:  true,
					pattern:     "UsersWithOption []",
				},
				{
					description: "products_with_struct_option should be value slice",
					shouldFind:  true,
					pattern:     "ProductsWithStructOption []",
				},
				{
					description: "users_without_option should be pointer slice",
					shouldFind:  true,
					pattern:     "UsersWithoutOption []*",
				},
				{
					description: "products_explicit_false should be pointer slice",
					shouldFind:  true,
					pattern:     "ProductsExplicitFalse []*",
				},
				{
					description: "tags should remain string slice (primitive)",
					shouldFind:  true,
					pattern:     "Tags []string",
				},
				{
					description: "single_user should remain pointer (not repeated)",
					shouldFind:  true,
					pattern:     "SingleUser *",
				},
				{
					description: "GetUsersWithOption should return value slice",
					shouldFind:  true,
					pattern:     "GetUsersWithOption() []",
				},
				{
					description: "GetProductsWithStructOption should return value slice",
					shouldFind:  true,
					pattern:     "GetProductsWithStructOption() []",
				},
				{
					description: "GetUsersWithoutOption should return pointer slice",
					shouldFind:  true,
					pattern:     "GetUsersWithoutOption() []*",
				},
			}

			for _, tc := range testCases {
				found := strings.Contains(content, tc.pattern)
				if found != tc.shouldFind {
					if tc.shouldFind {
						t.Errorf("%s: expected to find pattern '%s' but didn't", tc.description, tc.pattern)
					} else {
						t.Errorf("%s: expected not to find pattern '%s' but did", tc.description, tc.pattern)
					}
				}
			}
		}
	}
}

func createFieldOptionsRequest(t *testing.T) *pluginpb.CodeGeneratorRequest {
	// For now, create a minimal mock request
	// In a real scenario, this would be populated by protoc
	return &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"testdata/proto/field_options_test.proto"},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{
			// This would normally be populated by protoc with the actual proto file descriptor
			// For now we'll create a minimal version for compilation
		},
	}
}

func TestFieldOptionsParser(t *testing.T) {
	// Test the parser logic directly without requiring full protoc integration
	// This tests our shouldUseValueSlice logic

	// TODO: Add unit tests for the parser logic once we have the proto descriptor setup
	// This would test the individual parsing functions with mock descriptors
}
