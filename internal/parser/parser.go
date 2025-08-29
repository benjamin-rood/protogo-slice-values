package parser

import (
	"fmt"
	"strings"

	"github.com/benjamin-rood/protoc-gen-go-value-slices/internal/parser/types"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// FindAnnotatedFields parses proto files and finds fields marked with @nullable=false or @valueslice
func FindAnnotatedFields(req *pluginpb.CodeGeneratorRequest) (*types.AnnotatedFields, error) {
	fields := types.NewAnnotatedFields()

	for _, protoFile := range req.ProtoFile {
		if err := processProtoFile(protoFile, fields); err != nil {
			return nil, fmt.Errorf("failed to process proto file %s: %w", protoFile.GetName(), err)
		}
	}

	return fields, nil
}

func processProtoFile(protoFile *descriptorpb.FileDescriptorProto, fields *types.AnnotatedFields) error {
	// Build a map of source code locations
	locationMap := buildLocationMap(protoFile.GetSourceCodeInfo())

	// Process messages with their source locations
	for msgIdx, message := range protoFile.MessageType {
		if err := processMessage(
			message,
			locationMap,
			[]int32{4, int32(msgIdx)}, // Path to message
			fields,
		); err != nil {
			return fmt.Errorf("failed to process message %s: %w", message.GetName(), err)
		}
	}

	return nil
}

func buildLocationMap(sourceInfo *descriptorpb.SourceCodeInfo) map[string]*descriptorpb.SourceCodeInfo_Location {
	locationMap := make(map[string]*descriptorpb.SourceCodeInfo_Location)
	if sourceInfo == nil {
		return locationMap
	}

	for _, loc := range sourceInfo.GetLocation() {
		path := pathToString(loc.Path)
		locationMap[path] = loc
	}
	return locationMap
}

func processMessage(
	msg *descriptorpb.DescriptorProto,
	locations map[string]*descriptorpb.SourceCodeInfo_Location,
	path []int32,
	fields *types.AnnotatedFields,
) error {
	// Check each field
	for fieldIdx, field := range msg.Field {
		// Build path to this field: [4, msgIdx, 2, fieldIdx]
		fieldPath := append(append([]int32{}, path...), 2, int32(fieldIdx))

		// Look up source location for this field
		if loc, ok := locations[pathToString(fieldPath)]; ok {
			if isFieldAnnotated(loc.GetLeadingComments()) {
				if field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
					goFieldName := toGoFieldName(field.GetName())
					fields.Add(goFieldName)
				}
			}
		}
	}
	return nil
}

func isFieldAnnotated(comments string) bool {
	return strings.Contains(comments, "@nullable=false") ||
		strings.Contains(comments, "@valueslice")
}

func pathToString(path []int32) string {
	strs := make([]string, len(path))
	for i, p := range path {
		strs[i] = fmt.Sprintf("%d", p)
	}
	return strings.Join(strs, ".")
}

func toGoFieldName(protoName string) string {
	return toCamelCase(protoName)
}