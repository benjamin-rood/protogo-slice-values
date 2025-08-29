# protogo-values

A protoc plugin that converts pointer slices to value slices for fields marked with protobuf field options.

## Overview

By default, the Go protobuf generator creates pointer slices (`[]*Type`) for repeated message fields to handle potential nil values. This plugin allows you to specify which repeated fields should use value slices (`[]Type`) instead using protobuf field options.

## Field Options

Import the protogo_values options in your proto files:

```protobuf
import "proto/protogo_values/options.proto";
```

Then mark fields with the `value_slice` option:

```protobuf
repeated User users = 1 [(protogo_values.value_slice) = true];
```

## Example Usage

```protobuf
syntax = "proto3";

package example;

import "proto/protogo_values/options.proto";

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
  - plugin: go-values
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
  --go-values_out=. \
  --go-values_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  your_proto_file.proto
```

## How It Works

1. The plugin intercepts the protobuf compilation process
2. Calls the standard `protoc-gen-go` plugin to generate normal Go code
3. Parses the proto files for fields marked with protobuf field options
4. Post-processes the generated Go code to convert marked fields from `[]*Type` to `[]Type`
5. Updates both field declarations and getter methods

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

## Requirements

- Go 1.19+
- `protoc-gen-go` must be installed and available in PATH
- Protocol Buffers compiler (`protoc`)

## Limitations

- Only works with repeated message fields
- Requires `protoc-gen-go` to be available during compilation

## License

MIT License - see LICENSE file for details