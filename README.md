# protogo-values

⚠️ **PROJECT DISCONTINUED** ⚠️

**This plugin project has been discontinued due to fundamental architectural incompatibility with protobuf marshaling.**

A protoc plugin that attempted to convert pointer slices to value slices for fields marked with protobuf field options.

## Project Failure Analysis

### What We Attempted
By default, the Go protobuf generator creates pointer slices (`[]*Type`) for repeated message fields. This plugin attempted to allow specifying which repeated fields should use value slices (`[]Type`) instead, using protobuf field options.

### Why It Failed
**Fundamental Architectural Incompatibility**: The protobuf marshaling system is hardcoded to expect pointer slices for message types. Converting to value slices breaks protobuf's internal reflection system, causing runtime panics:

```
panic: reflect: Elem of invalid type v1.MessageType
```

**Root Cause**: Protobuf's marshaler calls `.Elem()` on slice types expecting pointers that can be dereferenced. Value slices don't support this operation.

### Technical Details
1. **Plugin Implementation**: ✅ Works correctly - only transforms fields with explicit options
2. **Code Generation**: ✅ Produces correct Go syntax  
3. **Type Safety**: ✅ Compiles without errors
4. **Runtime Marshaling**: ❌ **CRITICAL FAILURE** - Panics during `proto.Marshal()`

The plugin's wrapper approach (calling `protoc-gen-go` then post-processing) creates types that are syntactically correct but semantically incompatible with protobuf's runtime system.

## Field Options

Import the protogo_values options in your proto files:

```protobuf
import "protogo_values/options.proto";
```

Then mark fields with the `value_slice` option using one of two supported formats:

**Simple format (recommended for basic usage):**
```protobuf
repeated User users = 1 [(protogo_values.value_slice) = true];
```

**Structured format (for future extensibility):**
```protobuf
repeated User active_users = 2 [(protogo_values.field_opts).value_slice = true];
```

## Example Usage

```protobuf
syntax = "proto3";

package example;

import "protogo_values/options.proto";

message User {
  string id = 1;
  string name = 2;
}

message UserList {
  // Using field option - generates []User
  repeated User users = 1 [(protogo_values.value_slice) = true];
  
  // Using structured field option - generates []User  
  repeated User active_users = 2 [
    (protogo_values.field_opts).value_slice = true
  ];
  
  // No option - remains []*User (default)
  repeated User admins = 3;
  
  // Explicit false - remains []*User
  repeated User moderators = 4 [(protogo_values.value_slice) = false];
}
```

## Generated Go Code

Without this plugin:
```go
type UserList struct {
    Users  []*User   `protobuf:"bytes,1,rep,name=users"`
    Tags   []string  `protobuf:"bytes,2,rep,name=tags"`  
    Admins []*User   `protobuf:"bytes,3,rep,name=admins"`
}
```

With this plugin:
```go
type UserList struct {
    Users  []User    `protobuf:"bytes,1,rep,name=users"`  // Changed to value slice
    Tags   []string  `protobuf:"bytes,2,rep,name=tags"`   // Already value slice for primitives
    Admins []*User   `protobuf:"bytes,3,rep,name=admins"` // Unchanged
}
```

## Installation

### From Source

```bash
go install github.com/benjamin-rood/protogo-values/cmd/protoc-gen-go-values@latest
```

### From Repository

```bash
git clone https://github.com/benjamin-rood/protogo-values.git
cd protogo-values
make install
```

### Manual Build

```bash
git clone https://github.com/benjamin-rood/protogo-values.git
cd protogo-values
make build
cp protoc-gen-go-values $GOPATH/bin/  # or somewhere in your PATH
```

## Usage

### With Buf

Create a `buf.gen.yaml` file:

```yaml
version: v1
plugins:
  - plugin: protoc-gen-go-values
    out: gen
    opt:
      - paths=source_relative
  - plugin: go-grpc
    out: gen
    opt:
      - paths=source_relative
```

Then run:
```bash
buf generate
```

### With protoc directly

```bash
protoc \
  --protoc-gen-go-values_out=. \
  --protoc-gen-go-values_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  your_proto_file.proto
```

## How It Works

1. **Plugin Protocol**: The plugin follows the standard protoc plugin protocol, reading `CodeGeneratorRequest` from stdin
2. **Delegation**: Forwards the request to `protoc-gen-go` as a subprocess to generate normal Go code  
3. **Field Analysis**: Parses proto file descriptors to identify fields marked with `protogo_values` field options
4. **Code Transformation**: Applies pattern-based string replacements to convert `[]*Type` to `[]Type` for annotated fields
5. **Response Generation**: Returns the modified `CodeGeneratorResponse` with transformed field declarations and getter methods

## Alternative Solutions

Since this approach is fundamentally incompatible with protobuf, here are viable alternatives:

### 1. Accept Protobuf's Design
Protobuf uses pointer slices for message types by design to support:
- Nil value semantics
- Efficient marshaling/unmarshaling  
- Proper reflection support

### 2. Use Different Serialization
If value slices are critical for performance:
- **JSON**: Supports value slices natively
- **MessagePack**: Works well with value types
- **Custom binary formats**: Full control over serialization

### 3. Manual Conversion
Convert between pointer and value slices manually when needed:
```go
// Convert []*Message to []Message for processing
valueSlice := make([]Message, len(pointerSlice))
for i, ptr := range pointerSlice {
    if ptr != nil {
        valueSlice[i] = *ptr
    }
}
```

### 4. Code Generation Alternative
Write a completely separate code generator that:
- Parses `.proto` files independently  
- Generates Go code optimized for value slices
- Implements custom marshaling compatible with value types

**Note**: Option 4 would require rewriting significant portions of the protobuf ecosystem.

## Testing

The project includes comprehensive unit and integration tests:

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests (requires protoc and protoc-gen-go)
make test-integration

# Check test coverage
go test -cover ./internal/...
```

**Note**: Integration tests require both `protoc` and `protoc-gen-go` to be installed and available in your PATH. They use the `+build integration` tag and test the complete plugin protocol workflow with real protobuf compilation.

## Requirements

- Go 1.24+
- `protoc-gen-go` must be installed and available in PATH
- Protocol Buffers compiler (`protoc`)

## Lessons Learned

### Critical Limitation Discovered
**The plugin does NOT work with repeated message fields** due to protobuf marshaling incompatibility. Any attempt to marshal messages with value slices will result in runtime panics.

### What Worked
- Protobuf field options parsing and validation
- Code generation and string transformation  
- Integration with protoc plugin protocol
- Comprehensive testing infrastructure

### What Failed  
- **Runtime marshaling**: Protobuf's internal reflection system is incompatible with value slices for message types
- **Performance goals**: Custom marshaling workarounds would likely perform worse than standard pointer slices
- **Architectural approach**: Post-processing protoc-gen-go output creates types incompatible with protobuf runtime

### Key Insight
**Protobuf's architecture is tightly coupled**: You cannot change type representations without also replacing the entire marshaling/reflection system. Surface-level transformations create incompatible types.

## Repository Contents

This repository is preserved for educational purposes and contains:

- **Complete implementation** of a protoc plugin with field options
- **Comprehensive test suite** including integration tests  
- **Validation platform** demonstrating the marshaling failure
- **Bug analysis** in `specs/plugin-transformation-bug.spec.md`
- **Working examples** of protobuf field options implementation

### For Learning
- **Specifications**: `specs/protobuf-field-options/` contains detailed requirements in EARS format
- **Examples**: `examples/` directory shows protobuf field options usage
- **Validation Demo**: `../protogo-values-validation-demo/` demonstrates the runtime failures
- **Bug Documentation**: Complete analysis of why the approach fails

### Status
**DO NOT USE IN PRODUCTION** - This plugin will cause runtime panics when marshaling protobuf messages.

## License

MIT License - see LICENSE file for details

---

*This project serves as a case study in protobuf architecture limitations and the importance of understanding system constraints before implementation.*