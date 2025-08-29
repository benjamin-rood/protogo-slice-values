package transform

import (
	"fmt"
	"strings"

	"github.com/benjamin-rood/protogo-values/internal/parser/types"
	"google.golang.org/protobuf/types/pluginpb"
)

// ApplyTransformations modifies the generated Go code to convert pointer slices to value slices
func ApplyTransformations(resp *pluginpb.CodeGeneratorResponse, fields *types.AnnotatedFields) error {
	if resp == nil {
		return fmt.Errorf("response cannot be nil")
	}
	if fields == nil {
		return fmt.Errorf("fields cannot be nil")
	}
	
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
	if fieldName == "" {
		return content
	}
	
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Skip comment lines
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		
		// Replace field declarations: FieldName []*Type -> FieldName []Type
		// Handle various whitespace patterns (spaces, tabs)
		patterns := []string{
			fieldName + " []*",       // space separator
			fieldName + "\t[]*",      // tab separator  
			fieldName + "  []*",      // multiple spaces
			fieldName + "   []*",     // even more spaces
		}
		
		for _, pattern := range patterns {
			if strings.Contains(line, pattern) {
				replacement := strings.Replace(pattern, "[]*", "[]", 1)
				lines[i] = strings.ReplaceAll(line, pattern, replacement)
			}
		}
		
		// Replace getter methods: GetFieldName() []*Type -> GetFieldName() []Type
		oldGetter := "Get" + fieldName + "() []*"
		if strings.Contains(line, oldGetter) {
			lines[i] = strings.ReplaceAll(line, oldGetter, "Get"+fieldName+"() []")
		}
	}
	
	return strings.Join(lines, "\n")
}
