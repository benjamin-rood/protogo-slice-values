# Plugin Transformation Bug Specification

**Issue ID**: plugin-transformation-bug  
**Status**: ⚠️ Architecture Issue - Value Slices Incompatible with Protobuf  
**Severity**: High  
**Discovered**: During validation platform development  
**Reporter**: Validation Platform Integration Testing

## Summary - INVESTIGATION COMPLETE

**PLUGIN TECHNICALLY WORKING**: Investigation shows the `protoc-gen-go-values` plugin correctly transforms only fields with explicit options.

**FUNDAMENTAL ISSUE DISCOVERED**: Value slices for protobuf message types cause runtime panics during marshaling. The protobuf reflection system expects pointer slices (`[]*Type`) for message types, not value slices (`[]Type`).

## Problem Description

### Expected Behavior
The plugin should ONLY transform repeated message fields that have explicit field options:
- `repeated Type field = 1 [(protogo_values.value_slice) = true];`
- `repeated Type field = 2 [(protogo_values.field_opts).value_slice = true];`

Fields without these options should remain as standard pointer slices `[]*Type`.

### Actual Behavior
The plugin transforms ALL repeated message fields from `[]*Type` to `[]Type`, even those without field options.

### Impact
This causes runtime panics during protobuf marshaling with the error:
```
panic: reflect: Elem of invalid type v1.MessageType
```

The protobuf reflection system expects pointer slices for message types, but the plugin creates value slices, causing a type mismatch during serialization.

## Reproduction Case

### Proto Definition (No Field Options)
```protobuf
syntax = "proto3";

package test.v1;

message TestResponse {
  // This field has NO field options - should remain []*TestResult
  repeated TestResult results = 1;
}

message TestResult {
  string name = 1;
  bool passed = 2;
}
```

### Expected Generated Code
```go
type TestResponse struct {
    Results []*TestResult `protobuf:"bytes,1,rep,name=results,proto3" json:"results,omitempty"`
}
```

### Actual Generated Code (Bug)
```go
type TestResponse struct {
    Results []TestResult `protobuf:"bytes,1,rep,name=results,proto3" json:"results,omitempty"`
}
```

### Runtime Error
When attempting to marshal `TestResponse`:
```go
response := &TestResponse{
    Results: []TestResult{{Name: "test", Passed: true}},
}
data, err := proto.Marshal(response) // PANIC: reflect: Elem of invalid type
```

## Discovery Context

This bug was discovered during development of the validation platform (`protogo-values-validation-demo`) when implementing integration tests for a gRPC validation service. The service would panic when trying to marshal response messages containing `ValidationResult` slices.

### Affected Files in Validation Platform
- `api/validation/v1/validation.proto` - Contains fields without options that were incorrectly transformed
- Generated: `gen/api/validation/v1/validation.pb.go` - Had incorrect value slice types
- `internal/server/validation_service.go` - Service implementation affected by type mismatches

## Root Cause Analysis

The plugin's transformation logic appears to be missing validation for field options. Instead of checking:
1. Does this field have `(protogo_values.value_slice) = true`?
2. Does this field have `(protogo_values.field_opts).value_slice = true`?

The plugin is applying transformations to all repeated message fields.

## Suggested Fix Areas

1. **Field Option Validation**: The plugin should validate field options before applying transformations
2. **Parser Logic**: The parser in `internal/parser/` should only identify fields WITH options
3. **Transform Logic**: The transform in `internal/transform/` should only process validated fields
4. **Integration Tests**: Add test cases for fields WITHOUT options to ensure they remain untransformed

## Workaround

For projects affected by this bug, generate problematic proto files using standard `protoc-gen-go` instead of the custom plugin:

### buf.validation.yaml
```yaml
version: v1
plugins:
  # Use standard protoc-gen-go for files without field options
  - plugin: go
    out: gen
    opt:
      - paths=source_relative
  - plugin: go-grpc
    out: gen
    opt:
      - paths=source_relative
```

Generate specific files:
```bash
buf generate --template buf.validation.yaml api/validation/v1/validation.proto
```

## Test Cases Required

1. **Negative Test**: Fields without options should remain `[]*Type`
2. **Positive Test**: Fields with options should become `[]Type`
3. **Mixed Test**: Proto files with both transformed and untransformed fields
4. **gRPC Integration**: Generated types should work with gRPC marshaling

## Related Files to Investigate

- `internal/parser/field_options.go` - Field option parsing logic
- `internal/transform/transform.go` - Type transformation logic  
- `internal/plugin/plugin.go` - Main plugin orchestration
- Integration tests with fields that lack options

## Investigation Findings

### Plugin Behavior Verification ✅
- **Parser Logic**: Correctly checks for field options using `shouldUseValueSlice(field *descriptorpb.FieldDescriptorProto)`
- **Transformation Logic**: Only processes fields present in `AnnotatedFields` collection  
- **Generated Code**: Fields without options remain as `[]*Type`, fields with options become `[]Type`

### Evidence from Generated Code
```go
// ValidationTestMessage - CORRECT behavior
type ValidationTestMessage struct {
    ValueSliceData   []DataPoint   // HAS field option: [(protogo_values.value_slice) = true]
    PointerSliceData []*DataPoint  // NO field option: remains pointer slice ✅
    Metrics          []MetricPoint // HAS field option: [(protogo_values.field_opts).value_slice = true]
}

// ValidateTypesResponse - CORRECT behavior  
type ValidateTypesResponse struct {
    Results []*ValidationResult  // NO field option: remains pointer slice ✅
}

// BenchmarkResponse - CORRECT behavior
type BenchmarkResponse struct {
    Results []*BenchmarkResult  // NO field option: remains pointer slice ✅
}
```

### Actual Root Cause
The compilation error occurs in `internal/server/validation_service.go` where:
```go
// Line 35-36: Type mismatch in service implementation
performanceResults := s.validatePerformanceTestMessageTypes() // returns []ValidationResult
results = append(results, performanceResults...) // expects []*ValidationResult
```

### Protobuf Marshaling Panic (CRITICAL FINDING)
```
panic: reflect: Elem of invalid type v1.DataPoint

goroutine 15 [running]:
reflect.elem(0x100a97220?)
	/usr/local/go/src/reflect/type.go:733 +0xc4
reflect.(*rtype).Elem(0x0?)
	/usr/local/go/src/reflect/type.go:737 +0x20
google.golang.org/protobuf/internal/impl.sizeMessageSlice({0x100a49e40?}, {0x100a97220, 0x100a512c0}, 0x1, {0x50?})
	/Users/br/go/pkg/mod/google.golang.org/protobuf@v1.36.8/internal/impl/codec_field.go:473 +0x84
```

**Root Cause**: The protobuf marshaler's internal reflection code calls `.Elem()` on message slice types, expecting pointer types that can be dereferenced. Value slices don't support this operation.

**Impact**: ANY attempt to marshal protobuf messages containing value slices will panic at runtime.

## Resolution Options

**ARCHITECTURE DECISION REQUIRED**: The plugin concept may be fundamentally incompatible with protobuf for message types.

**Options**:
1. **Restrict to primitives only**: Only transform primitive slices ([]string, []int32, etc.) which are already value slices  
2. **Deprecate plugin**: Acknowledge that protobuf marshaling requires pointer slices for message types
3. **Custom marshaling**: Implement custom proto marshal/unmarshal for transformed types (complex)

## Priority

**High** - Plugin currently breaks protobuf marshaling for any transformed message fields.