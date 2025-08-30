package parser

import (
	"fmt"

	"github.com/benjamin-rood/protogo-values/internal/parser/types"
	"github.com/benjamin-rood/protogo-values/proto/protogo_values"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// FindAnnotatedFields parses proto files and finds fields marked with protobuf field options
func FindAnnotatedFields(req *pluginpb.CodeGeneratorRequest) (*types.AnnotatedFields, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	
	fields := types.NewAnnotatedFields()

	for _, protoFile := range req.ProtoFile {
		if err := processProtoFile(protoFile, fields); err != nil {
			return nil, fmt.Errorf("failed to process proto file %s: %w", protoFile.GetName(), err)
		}
	}

	return fields, nil
}

func processProtoFile(protoFile *descriptorpb.FileDescriptorProto, fields *types.AnnotatedFields) error {
	// Process messages
	for _, message := range protoFile.MessageType {
		if err := processMessage(message, fields); err != nil {
			return fmt.Errorf("failed to process message %s: %w", message.GetName(), err)
		}
	}

	return nil
}

func processMessage(
	msg *descriptorpb.DescriptorProto,
	fields *types.AnnotatedFields,
) error {
	// Check each field
	for _, field := range msg.Field {
		// Only process repeated message fields
		if field.GetLabel() != descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
			continue
		}

		// Only process message types (not primitives)
		if field.GetType() != descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
			continue
		}

		// Check if field should use value slice
		shouldUse := shouldUseValueSlice(field)

		if shouldUse {
			goFieldName := toGoFieldName(field.GetName())
			fields.Add(goFieldName)
		}
	}
	return nil
}

// shouldUseValueSlice determines if a field should use value slices based on protobuf field options
func shouldUseValueSlice(field *descriptorpb.FieldDescriptorProto) bool {
	if field.Options == nil {
		return false
	}

	// Check simple value_slice extension
	if proto.HasExtension(field.Options, protogo_values.E_ValueSlice) {
		valueSlice := proto.GetExtension(field.Options, protogo_values.E_ValueSlice).(bool)
		return valueSlice
	}

	// Check structured field options
	if proto.HasExtension(field.Options, protogo_values.E_FieldOpts) {
		opts := proto.GetExtension(field.Options, protogo_values.E_FieldOpts).(*protogo_values.FieldOptions)
		if opts != nil && opts.ValueSlice != nil {
			return *opts.ValueSlice
		}
	}
	return false
}

func toGoFieldName(protoName string) string {
	return toCamelCase(protoName)
}
