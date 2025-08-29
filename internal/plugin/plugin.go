package plugin

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/benjamin-rood/protoc-gen-go-value-slices/internal/parser"
	"github.com/benjamin-rood/protoc-gen-go-value-slices/internal/transform"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

// ProcessRequest handles the main plugin workflow
func ProcessRequest(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	// Pass the request to the standard protoc-gen-go
	resp, err := callProtocGenGo(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call protoc-gen-go: %w", err)
	}

	// Parse the proto files to find annotated fields
	annotatedFields, err := parser.FindAnnotatedFields(req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse annotated fields: %w", err)
	}

	// Transform the generated files
	if err := transform.ApplyTransformations(resp, annotatedFields); err != nil {
		return nil, fmt.Errorf("failed to apply transformations: %w", err)
	}

	return resp, nil
}

// callProtocGenGo calls the standard protoc-gen-go plugin
func callProtocGenGo(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	// Marshal the request
	input, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Execute protoc-gen-go
	cmd := exec.Command("protoc-gen-go")
	cmd.Stdin = bytes.NewReader(input)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute protoc-gen-go: %w", err)
	}

	// Parse the response
	var resp pluginpb.CodeGeneratorResponse
	if err := proto.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}