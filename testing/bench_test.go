package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	jsg "github.com/alanshaw/dag-json-gen"
)

func BenchmarkMarshaling(b *testing.B) {
	r := rand.New(rand.NewSource(56887))
	val, ok := quick.Value(reflect.TypeFor[SimpleTypeTwo](), r)
	if !ok {
		b.Fatal("failed to construct type")
	}

	tt := val.Interface().(SimpleTypeTwo)

	b.ReportAllocs()

	for b.Loop() {
		if err := tt.MarshalDagJSON(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshaling(b *testing.B) {
	r := rand.New(rand.NewSource(123456))
	val, ok := quick.Value(reflect.TypeFor[SimpleTypeTwo](), r)
	if !ok {
		b.Fatal("failed to construct type")
	}

	tt := val.Interface().(SimpleTypeTwo)

	buf := new(bytes.Buffer)
	if err := tt.MarshalDagJSON(buf); err != nil {
		b.Fatal(err)
	}

	reader := bytes.NewReader(buf.Bytes())

	b.ReportAllocs()

	for b.Loop() {
		reader.Seek(0, io.SeekStart)
		var tt SimpleTypeTwo
		if err := tt.UnmarshalDagJSON(reader); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDeferred(b *testing.B) {
	r := rand.New(rand.NewSource(123456))
	val, ok := quick.Value(reflect.TypeFor[SimpleTypeTwo](), r)
	if !ok {
		b.Fatal("failed to construct type")
	}

	tt := val.Interface().(SimpleTypeTwo)

	buf := new(bytes.Buffer)
	if err := tt.MarshalDagJSON(buf); err != nil {
		b.Fatal(err)
	}

	var (
		deferred jsg.Deferred
		reader   = bytes.NewReader(buf.Bytes())
	)

	b.ReportAllocs()

	for b.Loop() {
		reader.Seek(0, io.SeekStart)
		if err := deferred.UnmarshalDagJSON(reader); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdJSONMarshaling(b *testing.B) {
	r := rand.New(rand.NewSource(56887))
	val, ok := quick.Value(reflect.TypeFor[SimpleTypeTwo](), r)
	if !ok {
		b.Fatal("failed to construct type")
	}

	tt := val.Interface().(SimpleTypeTwo)

	b.ReportAllocs()

	for b.Loop() {
		if err := json.NewEncoder(io.Discard).Encode(&tt); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdJSONUnmarshaling(b *testing.B) {
	r := rand.New(rand.NewSource(123456))
	val, ok := quick.Value(reflect.TypeFor[SimpleTypeTwo](), r)
	if !ok {
		b.Fatal("failed to construct type")
	}

	tt := val.Interface().(SimpleTypeTwo)

	data, err := json.Marshal(&tt)
	if err != nil {
		b.Fatal(err)
	}

	reader := bytes.NewReader(data)

	b.ReportAllocs()

	for b.Loop() {
		reader.Seek(0, io.SeekStart)
		var tt SimpleTypeTwo
		if err := json.NewDecoder(reader).Decode(&tt); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMapMarshaling(b *testing.B) {
	r := rand.New(rand.NewSource(56887))
	val, ok := quick.Value(reflect.TypeFor[SimpleTypeTree](), r)
	if !ok {
		b.Fatal("failed to construct type")
	}

	tt := val.Interface().(SimpleTypeTree)

	b.ReportAllocs()

	for b.Loop() {
		if err := tt.MarshalDagJSON(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMapUnmarshaling(b *testing.B) {
	r := rand.New(rand.NewSource(123456))
	val, ok := quick.Value(reflect.TypeFor[SimpleTypeTree](), r)
	if !ok {
		b.Fatal("failed to construct type")
	}

	tt := val.Interface().(SimpleTypeTree)

	buf := new(bytes.Buffer)
	if err := tt.MarshalDagJSON(buf); err != nil {
		b.Fatal(err)
	}

	reader := bytes.NewReader(buf.Bytes())

	b.ReportAllocs()

	for b.Loop() {
		reader.Seek(0, io.SeekStart)
		var tt SimpleTypeTree
		if err := tt.UnmarshalDagJSON(reader); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdJSONMapMarshaling(b *testing.B) {
	r := rand.New(rand.NewSource(56887))
	val, ok := quick.Value(reflect.TypeFor[SimpleTypeTree](), r)
	if !ok {
		b.Fatal("failed to construct type")
	}

	tt := val.Interface().(SimpleTypeTree)

	b.ReportAllocs()

	for b.Loop() {
		if err := json.NewEncoder(io.Discard).Encode(&tt); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdJSONMapUnmarshaling(b *testing.B) {
	r := rand.New(rand.NewSource(123456))
	val, ok := quick.Value(reflect.TypeFor[SimpleTypeTree](), r)
	if !ok {
		b.Fatal("failed to construct type")
	}

	tt := val.Interface().(SimpleTypeTree)

	data, err := json.Marshal(&tt)
	if err != nil {
		b.Fatal(err)
	}

	reader := bytes.NewReader(data)

	b.ReportAllocs()

	for b.Loop() {
		reader.Seek(0, io.SeekStart)
		var tt SimpleTypeTree
		if err := json.NewDecoder(reader).Decode(&tt); err != nil {
			b.Fatal(err)
		}
	}
}
