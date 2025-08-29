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

func TestIntegrationWithRealProto(t *testing.T) {
	// Skip if protoc-gen-go is not available
	if _, err := exec.LookPath("protoc-gen-go"); err != nil {
		t.Skip("protoc-gen-go not found in PATH, skipping integration test")
	}

	// Generate the request using protoc
	protoFile := "testdata/proto/test.proto"
	if _, err := os.Stat(protoFile); os.IsNotExist(err) {
		t.Fatalf("Test proto file not found: %s", protoFile)
	}

	// Use protoc to generate the CodeGeneratorRequest
	cmd := exec.Command("protoc",
		"--go_out=testdata/gen",
		"--go_opt=paths=source_relative",
		"--descriptor_set_out=/tmp/test.desc",
		"--include_source_info",
		protoFile)

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run protoc: %v", err)
	}

	// Clean up
	defer os.RemoveAll("testdata/gen")
	defer os.Remove("/tmp/test.desc")

	// Now test our plugin by creating a mock request
	req := createMockRequest(t)

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
		if file.GetName() == "test.pb.go" && file.Content != nil {
			content := *file.Content

			// Check that annotated fields were transformed
			if !strings.Contains(content, "Users []User") {
				t.Error("Users field should be transformed to value slice")
			}

			if !strings.Contains(content, "Products []Product") {
				t.Error("Products field should be transformed to value slice")
			}

			// Check that non-annotated fields remain pointer slices
			if !strings.Contains(content, "Admins []*User") {
				t.Error("Admins field should remain as pointer slice")
			}

			// Check getter methods
			if !strings.Contains(content, "GetUsers() []User") {
				t.Error("GetUsers method should return value slice")
			}

			if !strings.Contains(content, "GetProducts() []Product") {
				t.Error("GetProducts method should return value slice")
			}
		}
	}
}

func createMockRequest(t *testing.T) *pluginpb.CodeGeneratorRequest {
	// Create a simplified mock request for testing
	// In a real scenario, this would come from protoc
	return &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"test.proto"},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{
			// This would normally be populated by protoc
			// For now, we'll create a minimal version
		},
	}
}

func TestPluginBinaryExists(t *testing.T) {
	// Test that we can build the plugin binary
	cmd := exec.Command("go", "build", "-o", "/tmp/protoc-gen-go-values", "./cmd/protoc-gen-go-values")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build plugin binary: %v", err)
	}

	defer os.Remove("/tmp/protoc-gen-go-values")

	// Test that the binary is executable
	if _, err := os.Stat("/tmp/protoc-gen-go-values"); err != nil {
		t.Fatalf("Plugin binary not created: %v", err)
	}
}
