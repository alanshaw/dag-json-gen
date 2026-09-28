package testing

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestOptionalFields(t *testing.T) {
	ints := []int64{1, 2, 3, 4, 5}

	for count := 0; count <= 5; count++ {
		t.Run(fmt.Sprintf("length-%d", count), func(t *testing.T) {
			var buf bytes.Buffer
			obj := IntArray{Ints: ints[:count]}
			if err := obj.MarshalDagJSON(&buf); err != nil {
				t.Fatal(err)
			}

			// Pre-fill with garbage. We want optional fields to be reset to their
			// defaults.
			out := TupleWithOptionalFields{
				Int1:  0xf1,
				Uint2: 0xf2,
				Int3:  0xf3,
				Int4:  0xf4,
			}
			err := out.UnmarshalDagJSON(&buf)
			switch count {
			case 4:
				if out.Int4 != ints[3] {
					t.Errorf("field 4 should be %d, was %d", ints[3], out.Int4)
				}
				fallthrough
			case 3:
				if out.Int3 != ints[2] {
					t.Errorf("field 3 should be %d, was %d", ints[2], out.Int3)
				}
				fallthrough
			case 2:
				if out.Uint2 != uint64(ints[1]) {
					t.Errorf("field 2 should be %d, was %d", ints[1], out.Uint2)
				}
				if out.Int1 != ints[0] {
					t.Errorf("field 1 should be %d, was %d", ints[0], out.Int1)
				}
				switch count {
				case 2:
					if out.Int3 != 0 {
						t.Errorf("expected field 3 to be zero, was %d", out.Int3)
					}
					fallthrough
				case 3:
					if out.Int4 != 0 {
						t.Errorf("expected field 4 to be zero, was %d", out.Int4)
					}
				}
				if err != nil {
					t.Errorf("expected no error when unmarshaling, got: %s", err)
				}
			case 1, 0:
				if err == nil {
					t.Errorf("expected an error when unmarshaling with too few fields")
				}
			default:
				if err == nil {
					t.Errorf("expected an error when unmarshaling with too many fields")
				}
			}
		})
	}

}

func TestEmptyTuple(t *testing.T) {
	t.Run("rejected when fields are mandatory", func(t *testing.T) {
		out := TupleIntArray{Int1: 0xf1}
		err := out.UnmarshalDagJSON(strings.NewReader("[]"))
		if err == nil {
			t.Fatal("expected an error when unmarshaling an empty array into a tuple with mandatory fields")
		}
		if !strings.Contains(err.Error(), "too few fields 0 < 3") {
			t.Errorf("unexpected error: %s", err)
		}
	})

	t.Run("rejected when a mandatory prefix exists", func(t *testing.T) {
		var out TupleWithOptionalFields
		err := out.UnmarshalDagJSON(strings.NewReader("[]"))
		if err == nil {
			t.Fatal("expected an error when unmarshaling an empty array into a tuple with a mandatory prefix")
		}
		if !strings.Contains(err.Error(), "too few fields 0 < 2") {
			t.Errorf("unexpected error: %s", err)
		}
	})

	t.Run("accepted when all fields are optional", func(t *testing.T) {
		out := TupleAllOptionalFields{Int1: 0xf1, Int2: 0xf2}
		if err := out.UnmarshalDagJSON(strings.NewReader("[]")); err != nil {
			t.Fatalf("expected no error when unmarshaling an empty array into an all-optional tuple, got: %s", err)
		}
		if out.Int1 != 0 || out.Int2 != 0 {
			t.Errorf("expected optional fields to be reset to zero, got %+v", out)
		}
	})
}
