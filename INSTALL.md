# Installation Guide

## Prerequisites

Before installing protoc-gen-go-values, make sure you have:

1. **Go 1.19 or later**
   ```bash
   go version
   ```

2. **Protocol Buffers compiler (protoc)**
   ```bash
   # macOS with Homebrew
   brew install protobuf
   
   # Ubuntu/Debian
   sudo apt-get install protobuf-compiler
   
   # Or download from: https://github.com/protocolbuffers/protobuf/releases
   ```

3. **protoc-gen-go plugin**
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   ```

## Installation Methods

### Method 1: Direct Installation from Source

```bash
go install github.com/benjamin-rood/protoc-gen-go-values@latest
```

### Method 2: Clone and Build

```bash
git clone https://github.com/benjamin-rood/protoc-gen-go-values.git
cd protoc-gen-go-values
make dev-setup
```

### Method 3: Manual Build

```bash
git clone https://github.com/benjamin-rood/protoc-gen-go-values.git
cd protoc-gen-go-values
go build -o protoc-gen-go-values .
cp protoc-gen-go-values $HOME/go/bin/  # or anywhere in your PATH
```

## Verify Installation

Check that the plugin is correctly installed:

```bash
which protoc-gen-go-values
```

Make sure all dependencies are available:

```bash
make check-deps
```

## Quick Start

1. Create a proto file with annotated fields:
   ```protobuf
   syntax = "proto3";
   
   message Example {
     // @valueslice
     repeated User users = 1;
   }
   ```

2. Generate code using buf:
   ```bash
   # Create buf.gen.yaml (see examples/)
   buf generate
   ```

3. Or use protoc directly:
   ```bash
   protoc --go-value-slices_out=. --go-value-slices_opt=paths=source_relative example.proto
   ```

## Troubleshooting

### Plugin not found
- Ensure the binary is in your PATH
- Check that it's named exactly `protoc-gen-go-values`
- Verify Go's bin directory is in your PATH: `echo $PATH | grep $(go env GOPATH)/bin`

### protoc-gen-go not found
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

### Permission denied
```bash
chmod +x $(which protoc-gen-go-values)
```

For more help, see the README.md or create an issue on GitHub.