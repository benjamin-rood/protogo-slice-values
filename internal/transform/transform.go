package transform

import (
	"strings"

	"github.com/benjamin-rood/protoc-gen-go-value-slices/internal/parser/types"
	"google.golang.org/protobuf/types/pluginpb"
)

// ApplyTransformations modifies the generated Go code to convert pointer slices to value slices
func ApplyTransformations(resp *pluginpb.CodeGeneratorResponse, fields *types.AnnotatedFields) error {
	for _, file := range resp.File {
		if file.Content != nil {
			content := *file.Content
			content = transformPointerSlices(content, fields)
			file.Content = &content
		}
	}
	return nil
}

// transformPointerSlices converts []*Type to []Type for annotated fields
func transformPointerSlices(content string, fields *types.AnnotatedFields) string {
	for field := range fields.All() {
		content = transformField(content, field)
	}
	return content
}

// transformField transforms a specific field from pointer slice to value slice
func transformField(content, fieldName string) string {
	// Replace field declarations: FieldName []*Type -> FieldName []Type
	oldDecl := fieldName + " []*"
	newDecl := fieldName + " []"
	content = strings.ReplaceAll(content, oldDecl, newDecl)

	// Replace getter methods: GetFieldName() []*Type -> GetFieldName() []Type
	oldGetter := "Get" + fieldName + "() []*"
	newGetter := "Get" + fieldName + "() []"
	content = strings.ReplaceAll(content, oldGetter, newGetter)

	return content
}