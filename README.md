# dag-json-gen

`dag-json-gen` is a code generation tool that automatically creates DAG-JSON marshaling and unmarshaling functions for Go types. It is a fork of [github.com/whyrusleeping/cbor-gen](https://github.com/whyrusleeping/cbor-gen).

## Overview

The library generates `MarshalDagJSON` and `UnmarshalDagJSON` methods for your Go structs and types, allowing them to be encoded to and decoded from DAG-JSON format. This is particularly useful when you need efficient binary serialization with DAG-JSON support.

## Usage

1. Define your types in a .go file:

```go
type MyStruct struct {
	Field1 string
	Field2 uint64
	Field3 []byte
	Field4 *CustomType
}
```

2. Create a generator file (e.g., `gen/main.go`):

```go
package main

import (
	jsg "github.com/alanshaw/dag-json-gen"
)

func main() {
	// Generate tuple-style encoders (list of fields)...
	if err := cbg.WriteTupleEncodersToFile("dag_json_gen.go", "mypackage",
		MyStruct{},
		// Add other types here...
	); err != nil {
		panic(err)
	}

	// Or generate map-style encoders (map of field names to field values)...
	if err := cbg.WriteMapEncodersToFile("dag_json_map_gen.go", "mypackage",
		MyStruct{},
		// Add other types here...
	); err != nil {
		panic(err)
	}
}
```

3. Run the generator:
```bash
go run gen/main.go
```

## Features

### Encoding Styles

- **Tuple encoding** (`WriteTupleEncodersToFile`): Encodes structs as JSON arrays, more space-efficient but order-dependent.
- **Map encoding** (`WriteMapEncodersToFile`): Encodes structs as JSON maps with field names, more verbose but order-independent.

### Field Tags

The library supports several struct tags to customize encoding:

```go
type Example struct {
	// Rename field in CBOR output
	Field1 string `dagjsongen:"renamed_field"`

	// Preserve nil slices (encode nil as null, not as an empty cbor list)
	Field2 []uint64 `dagjsongen:"preservenil"`

	// Skip encoding this field when it matches its zero value. This is only applicable
	// when map-encoding; when tuple-encoding, all fields are always written.
	Field3 string `dagjsongen:"omitempty"`

	// Set constant value XXX WTF IS THIS?
	Field4 string `dagjsongen:"const=somevalue"`

	// Maximum length for slices/strings, defaults to 8192 for lists,
	// 2MiB for strings/bytes unless otherwise specified in cbg.Gen.
	Field5 []byte `dagjsongen:"maxlen=1000000"`

	// Allow this field to be missing when decoding a tuple-style struct. Optional fields
	// may only appear at the end of of structs and cannot be followed by mandatory fields.
	//
	// This tag is ignored when decoding map-style structs as all fields are optional in
	// map-style structs.
	Field6 string `dagjsongen:"optional"`
}
```

Additionally, you can specify that struct with a single field should be encoded as that field alone as if it didn't appear in a struct by specifying the `"transparent"` tag. For example, the following `TransparentExample` will encode as a string (regardless of whether tuple or map encoding was specified when generating the `UnmarshalDagJSON` and `MarshalDagJSON` functions):

```go
type TransparentExample struct {
	TheOneAndOnlyField string `dagjsongen:"transparent"`
}
```

### Upgrading Type Schemas

When working with map-encoded types, all fields are optional when decoding (they default to the field's zero-value) and unknown fields are skipped.

When working with tuple-encoded types, fields are mandatory by default and unknown (additional) fields will be rejected. However, it's possible to add additional optional fields to the end of the tuple-encoded struct by tagging them as `dagjsongen:"optional"`; this makes it possible to decode a tuple-encoded struct that omits a suffix of the expected fields. On encoding, optional fields will always be included.

### Customizing Generation

You can customize generation parameters using the `Gen` type:

```go
err := cbg.Gen{
	MaxArrayLength:  8192,	// Maximum length for arrays
	MaxByteLength:   2<<20,   // Maximum length for byte slices
	MaxStringLength: 2<<20,   // Maximum length for strings
}.WriteTupleEncodersToFile("dag_json_gen.go", "mypackage",
	MyType{},
)
```

## Supported Types

The library can generate encoders/decoders for:

- Basic Go types (integers, strings, etc.)
- Slices, arrays, and maps.
- Custom types that implement `UnmarshalDagJSON` and `MarshalDagJSON` (usually generated with this library).
- `cid.Cid` from [github.com/ipfs/go-cid](https://github.com/ipfs/go-cid)
- `big.Int`

## Generated Code

The tool will generate `MarshalDagJSON(w io.Writer) error` and `UnmarshalDagJSON(r io.Reader) error` methods for each type. These methods handle the DAG-JSON encoding and decoding respectively.

## License

MIT
