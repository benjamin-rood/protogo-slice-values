// protoc-gen-go-value-slices - A protoc plugin that converts pointer slices to value slices for fields marked with @nullable=false or @valueslice
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/benjamin-rood/protoc-gen-go-value-slices/internal/plugin"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	if err := run(); err != nil {
		resp := &pluginpb.CodeGeneratorResponse{
			Error: proto.String(err.Error()),
		}
		data, _ := proto.Marshal(resp)
		os.Stdout.Write(data)
		os.Exit(1)
	}
}

func run() error {
	// Read the CodeGeneratorRequest from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	var req pluginpb.CodeGeneratorRequest
	if err := proto.Unmarshal(input, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Process the request using our plugin
	resp, err := plugin.ProcessRequest(&req)
	if err != nil {
		return fmt.Errorf("failed to process request: %w", err)
	}

	// Write the response
	output, err := proto.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	
	if _, err := os.Stdout.Write(output); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}