# Protobuf Field Options Implementation Specification

**Feature**: Migrate from comment-based annotations to proper protobuf field options  
**Status**: Planning  
**Priority**: High  
**Created**: 2025-08-29  
**Owner**: Development Team  

## Overview

Replace the current comment-based annotation system (`@nullable=false`, `@valueslice`) with proper protobuf field options to provide type safety, IDE support, and better integration with the protobuf ecosystem.

## Current State Analysis

### Existing Implementation
- **Comment Parsing**: Uses `SourceCodeInfo` to find comments above field declarations
- **Annotation Patterns**: Supports `@nullable=false` and `@valueslice` in leading comments
- **Transformation Logic**: Post-processes generated Go code using string replacement
- **Integration**: Works as wrapper around `protoc-gen-go`

### Limitations of Current Approach
1. **No IDE Support**: Comments don't provide autocomplete or validation
2. **Fragile Parsing**: Relies on comment positioning and string matching
3. **No Type Safety**: Comments are unvalidated text
4. **Poor Tooling Integration**: Can't be used with other protobuf tools
5. **Limited Discoverability**: Options not visible in protobuf schema introspection

## Requirements (EARS Format)

### Core Functionality Requirements

**R1**: **WHEN** a field has the `(protogo.value_slice) = true` option, **THEN** the plugin SHALL generate a value slice `[]Type` instead of pointer slice `[]*Type`.

**R2**: **WHEN** a field has no protobuf field options related to slice handling, **THEN** the plugin SHALL maintain default protoc-gen-go behavior (pointer slices for message types).

**R3**: **WHEN** processing a repeated primitive field (string, int32, etc.), **THEN** the plugin SHALL maintain existing behavior regardless of options (primitives are already value slices).

### Backward Compatibility Requirements

**R4**: **WHEN** a proto file uses the legacy comment-based annotations, **THEN** the plugin SHALL continue to support them with a deprecation warning.

**R5**: **WHEN** both field options and comment annotations are present on the same field, **THEN** field options SHALL take precedence over comments.

**R6**: **WHEN** migrating from comments to field options, **THEN** the generated Go code SHALL remain functionally identical.

### Schema Requirements  

**R7**: **WHEN** defining the protobuf extension, **THEN** it SHALL use a unique extension number > 50000 to avoid conflicts.

**R8**: **WHEN** defining the extension package, **THEN** it SHALL use the module's Go package path for consistency.

**R9**: **WHEN** the extension proto file is created, **THEN** it SHALL be importable and reusable across multiple proto files.

### Developer Experience Requirements

**R10**: **WHEN** a developer uses the field option, **THEN** their IDE SHALL provide autocomplete and validation support.

**R11**: **WHEN** a developer imports the options proto, **THEN** they SHALL have access to clear documentation about each option.

**R12**: **WHEN** building proto files with the options, **THEN** protoc SHALL validate option usage and provide clear error messages for misuse.

### Integration Requirements

**R13**: **WHEN** using buf or other protobuf tooling, **THEN** the custom field options SHALL be recognized and processed correctly.

**R14**: **WHEN** generating code with other protoc plugins, **THEN** the field options SHALL not interfere with their operation.

**R15**: **WHEN** using the plugin in CI/CD pipelines, **THEN** the options proto SHALL be easily distributable and versionable.

## Technical Design

### Protobuf Extension Schema

```proto
// proto/protogo/options.proto
syntax = "proto3";

package protogo;

import "google/protobuf/descriptor.proto";

option go_package = "github.com/benjamin-rood/protogo-values/proto/protogo";

// Primary extension for explicit value slice control
extend google.protobuf.FieldOptions {
  // value_slice controls whether to generate value slices for repeated fields.
  // When true, generates []Type instead of []*Type.
  // Only applies to repeated message fields.
  bool value_slice = 50001;
}

// Future extensibility: structured options
message FieldOptions {
  // value_slice controls value vs pointer slice generation
  optional bool value_slice = 1;
  
  // Future options can be added here
  // optional bool omit_empty = 2;
  // optional string validation_rule = 3;
}

extend google.protobuf.FieldOptions {
  // Structured field options for future extensibility
  optional FieldOptions field_opts = 50002;
}
```

### Usage Examples

```proto
// service.proto
syntax = "proto3";

import "protogo/options.proto";
import "google/api/field_behavior.proto";
import "buf/validate/validate.proto";

message BatchRequest {
  // Using explicit value_slice option
  repeated RequestItem items = 1 [
    (google.api.field_behavior) = REQUIRED,
    (buf.validate.field).repeated.min_items = 1,
    (protogo.value_slice) = true
  ];
  
  // Another field using value slices
  repeated ResponseItem responses = 2 [
    (protogo.value_slice) = true
  ];
  
  // Using structured options (future extensibility)
  repeated MetadataItem metadata = 3 [
    (protogo.field_opts).value_slice = true
  ];
  
  // Default behavior (pointer slices)
  repeated AdminItem admin_items = 4;
}
```

### Implementation Architecture

#### 1. Plugin Infrastructure Changes
- **Replace** comment parsing with protobuf option reading
- **Enhance** field analysis to check for custom extensions
- **Maintain** string transformation logic for generated code modification
- **Add** option validation to ensure proper usage

#### 2. Code Generation Flow
```
1. protoc calls our plugin with CodeGeneratorRequest
2. Plugin calls protoc-gen-go to generate standard code
3. Plugin analyzes each field for custom options:
   - Check for (protogo.nullable) = false
   - Check for (protogo.value_slice) = true  
   - Check for structured (protogo.field_opts)
   - Fall back to comment parsing (deprecated)
4. Plugin applies transformations based on options
5. Plugin returns modified CodeGeneratorResponse
```

#### 3. Option Processing Logic
```go
func shouldUseValueSlice(field *protogen.Field) (bool, error) {
    if !field.Desc.IsList() {
        return false, nil // Only repeated fields
    }
    
    if field.Desc.Kind() != protoreflect.MessageKind {
        return false, nil // Only message types (primitives already value slices)
    }
    
    options := field.Desc.Options().(*descriptorpb.FieldOptions)
    
    // Check protogo.value_slice option (primary approach)
    if proto.HasExtension(options, protogo.E_ValueSlice) {
        valueSlice := proto.GetExtension(options, protogo.E_ValueSlice).(bool)
        return valueSlice, nil
    }
    
    // Check structured options (future extensibility)
    if proto.HasExtension(options, protogo.E_FieldOpts) {
        opts := proto.GetExtension(options, protogo.E_FieldOpts).(*protogo.FieldOptions)
        if opts != nil && opts.ValueSlice != nil {
            return *opts.ValueSlice, nil  
        }
    }
    
    // Fall back to comment parsing (deprecated)
    return shouldUseValueSliceFromComments(field)
}
```

## Implementation Plan

### Phase 1: Extension Definition
1. Create `proto/protogo/options.proto`
2. Generate Go code for options
3. Update module dependencies
4. Create usage examples

### Phase 2: Plugin Enhancement  
1. Update field analysis to read protobuf options
2. Implement option validation logic
3. Add deprecation warnings for comment usage
4. Update error handling and reporting

### Phase 3: Testing & Documentation
1. Create comprehensive test cases for all option combinations
2. Update README with new usage examples
3. Create migration guide from comments to options
4. Add integration tests with buf and other tools

### Phase 4: Backward Compatibility
1. Implement dual parsing (options + comments)
2. Add precedence rules (options > comments)
3. Create automated migration tools
4. Plan deprecation timeline for comments

## Testing Strategy

### Unit Tests
- Option parsing logic for all extension types
- Validation of option combinations and conflicts
- Error handling for malformed options
- Backward compatibility with existing comment parsing

### Integration Tests
- End-to-end code generation with field options
- Compatibility with protoc-gen-go versions
- Integration with buf generate workflow
- Cross-platform testing (Linux, macOS, Windows)

### Migration Tests
- Side-by-side comparison: comments vs options
- Generated code equivalence verification
- Performance impact measurement
- Real-world proto file conversion

## Migration Path

### For Users
1. **Phase 1**: Add options proto to your project
2. **Phase 2**: Update proto imports to include `protogo/options.proto`
3. **Phase 3**: Replace comment annotations with field options
4. **Phase 4**: Remove comment-based annotations after verification

### Example Migration
```proto
// Before (comments)
message UserList {
  // @valueslice
  repeated User users = 1;
  
  // @nullable=false  
  repeated Item items = 2;
}

// After (field options)
message UserList {
  repeated User users = 1 [(protogo.value_slice) = true];
  
  repeated Item items = 2 [(protogo.value_slice) = true];
}
```

## Risks and Mitigations

### Risk: Breaking Changes
- **Mitigation**: Maintain full backward compatibility during transition
- **Detection**: Comprehensive test suite comparing old vs new behavior  

### Risk: Protobuf Ecosystem Conflicts
- **Mitigation**: Use high extension numbers (50000+) and unique package names
- **Detection**: Test with popular protobuf tools and plugins

### Risk: Performance Impact  
- **Mitigation**: Benchmark option parsing vs comment parsing
- **Detection**: Performance tests in CI pipeline

### Risk: Developer Adoption
- **Mitigation**: Provide clear migration guide and tooling
- **Detection**: User feedback and adoption metrics

## Success Criteria

1. **Functional**: All existing comment-based behavior works with field options
2. **Performance**: No significant performance regression in code generation  
3. **Compatibility**: Works with major protobuf tools (buf, grpc, etc.)
4. **Usability**: IDE provides autocomplete and validation for options
5. **Migration**: Clear path from comments to options with tooling support

## Future Enhancements

1. **Additional Options**: `omit_empty`, `validation_rules`, etc.
2. **Message-Level Options**: Apply transformations to entire messages
3. **Service Options**: Extend to gRPC service definitions
4. **Code Generation Options**: Control other aspects of Go code generation

---

*This specification follows EARS format for requirements and provides comprehensive coverage of the protobuf field options implementation feature.*