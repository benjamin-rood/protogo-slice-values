# protogo-values

A protoc plugin that converts pointer slices to value slices for fields marked with protobuf field options.

## Overview

By default, the Go protobuf generator creates pointer slices (`[]*Type`) for repeated message fields to handle potential nil values. This plugin allows you to specify which repeated fields should use value slices (`[]Type`) instead using protobuf field options.

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

## Benefits

- **Memory efficiency**: Value slices eliminate pointer indirection
- **Cleaner APIs**: No need to handle nil values in slices where they don't make sense
- **Better performance**: Direct value access without pointer dereferencing
- **Selective application**: Only affects specifically marked fields

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

## Limitations

- Only works with repeated message fields
- Requires `protoc-gen-go` to be available during compilation

## Development

For developers interested in contributing or understanding the implementation:

- **Specifications**: See `specs/protobuf-field-options/` for detailed requirements in EARS format
- **Examples**: The `examples/` directory contains working proto files and buf configuration
- **Validation Demo**: See `../protogo-values-validation-demo/` for comprehensive end-to-end testing with gRPC services and performance benchmarks
- **Architecture**: Plugin uses a two-phase approach - delegation to `protoc-gen-go` followed by selective post-processing

## License

MIT License - see LICENSE file for details