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
		// Handle any amount of whitespace between field name and slice declaration
		if strings.Contains(line, fieldName) && strings.Contains(line, "[]*") {
			// Find the field name and ensure it's followed by whitespace and []*
			fieldIndex := strings.Index(line, fieldName)
			if fieldIndex >= 0 {
				// Look for []*Type after the field name with any whitespace in between
				remaining := line[fieldIndex+len(fieldName):]
				if strings.Contains(remaining, "[]*") {
					// Use regex-like replacement: find the first []*Type pattern after the field name
					sliceIndex := strings.Index(remaining, "[]*")
					if sliceIndex >= 0 {
						// Check if there's only whitespace between field name and []*
						whitespace := remaining[:sliceIndex]
						if strings.TrimSpace(whitespace) == "" && len(whitespace) > 0 {
							// Replace []*Type with []Type
							beforeSlice := line[:fieldIndex+len(fieldName)+sliceIndex]
							afterSlice := line[fieldIndex+len(fieldName)+sliceIndex+3:] // +3 for "[]*"
							lines[i] = beforeSlice + "[]" + afterSlice
						}
					}
				}
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
