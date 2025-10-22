package main

import (
	jsg "github.com/alanshaw/dag-json-gen"
	types "github.com/alanshaw/dag-json-gen/testing"
)

func main() {
	if err := jsg.WriteTupleEncodersToFile("testing/dag_json_gen.go", "testing",
		types.SignedArray{},
		types.SimpleTypeOne{},
		types.SimpleTypeTwo{},
		types.DeferredContainer{},
		types.FixedArrays{},
		types.ThingWithSomeTime{},
		types.BigField{},
		types.IntArray{},
		types.IntAliasArray{},
		types.TupleIntArray{},
		types.TupleIntArrayOptionals{},
		types.IntArrayNewType{},
		types.IntArrayAliasNewType{},
		types.MapTransparentType{},
		types.BigIntContainer{},
		types.TupleWithOptionalFields{},
	); err != nil {
		panic(err)
	}

	if err := jsg.WriteMapEncodersToFile("testing/dag_json_map_gen.go", "testing",
		types.SimpleTypeTree{},
		types.NeedScratchForMap{},
		types.SimpleStructV1{},
		types.SimpleStructV2{},
		types.RenamedFields{},
		types.TestEmpty{},
		types.TestConstField{},
		types.TestCanonicalFieldOrder{},
		types.MapStringString{},
		types.TestSliceNilPreserve{},
		types.StringPtrSlices{},
		types.FieldNameOverlap{},
	); err != nil {
		panic(err)
	}

	err := jsg.Gen{
		MaxArrayLength:  10,
		MaxByteLength:   9,
		MaxStringLength: 8,
	}.WriteTupleEncodersToFile("testing/dag_json_options_gen.go", "testing",
		types.LimitedStruct{},
	)
	if err != nil {
		panic(err)
	}

	err = jsg.Gen{
		MaxArrayLength:  10,
		MaxByteLength:   9,
		MaxStringLength: 10000,
	}.WriteTupleEncodersToFile("testing/dag_json_options_gen2.go", "testing",
		types.LongString{},
	)
	if err != nil {
		panic(err)
	}
}
