package typegen

import (
	"fmt"
	"strings"
	"testing"
)

func TestDeferredMaxLengthSingle(t *testing.T) {
	data := fmt.Sprintf("%q", strings.Repeat("0", ByteArrayMaxLen+1))
	var deferred Deferred
	err := deferred.UnmarshalDagJSON(strings.NewReader(data))
	if err != ErrLimitExceeded {
		t.Fatal("deferred: allowed more than the maximum allocation supported")
	}
}

// // TestReadEOFSemantics checks that our helper functions follow this rule when
// // dealing with EOF:
// // If the reader can't read a single byte because of EOF, it should return err == io.EOF.
// // If the reader could read _some_ of the bytes but not all because of EOF, it
// // should return err == io.ErrUnexpectedEOF.
// // Take a look at the io.EOF doc for  more info: https://pkg.go.dev/io#EOF
// func TestReadEOFSemantics(t *testing.T) {
// 	type testCase struct {
// 		name       string
// 		reader     io.Reader
// 		shouldFail bool
// 	}
// 	newTestCases := func() []testCase {
// 		return []testCase{
// 			{name: "Reader that returns EOF and n bytes read", reader: &testReader1Byte{b: 0x01}, shouldFail: false},
// 			{name: "Peeker with Reader that returns EOF and n bytes read", reader: GetPeeker(&testReader1Byte{b: 0x01}), shouldFail: false},
// 			{name: "Peeker with Exhausted Reader", reader: GetPeeker(&testReader1Byte{b: 0x01, emptied: true}), shouldFail: true},
// 			{name: "Exhausted reader", reader: &testReader1Byte{b: 0x01, emptied: true}, shouldFail: true},
// 			{name: "Byte buffer", reader: bytes.NewBuffer([]byte{0x01}), shouldFail: false},
// 			{name: "Empty Byte buffer", reader: bytes.NewBuffer([]byte{}), shouldFail: true},
// 			{name: "Byte Reader", reader: bytes.NewReader([]byte{0x01}), shouldFail: false},
// 			{name: "Empty Byte Reader", reader: bytes.NewReader([]byte{}), shouldFail: true},
// 			{name: "bufio Reader", reader: bufio.NewReader(bytes.NewReader([]byte{0x01})), shouldFail: false},
// 			{name: "bufio Reader with testReader", reader: bufio.NewReader(&testReader1Byte{b: 0x01}), shouldFail: false},
// 			{name: "bufio Reader with exhausted testReader", reader: bufio.NewReader(&testReader1Byte{b: 0x01, emptied: true}), shouldFail: true},
// 		}
// 	}

// 	utilFns := []func(io.Reader) (byte, error){
// 		func(r io.Reader) (byte, error) {
// 			return readByte(r)
// 		},
// 		func(r io.Reader) (byte, error) {
// 			return readByteBuf(r, []byte{0x00})
// 		},
// 		func(r io.Reader) (byte, error) {
// 			err := discard(r, 1)
// 			return 0x01, err
// 		},
// 	}

// 	for i, f := range utilFns {
// 		for _, tc := range newTestCases() {
// 			t.Run(fmt.Sprintf("util fn #%d against %s", i, tc.name), func(t *testing.T) {
// 				b, err := f(tc.reader)
// 				if tc.shouldFail && err == nil {
// 					t.Fatalf("Expected error. Got nil")
// 				} else if !tc.shouldFail && err != nil {
// 					t.Fatalf("Expected no error. Got %v", err)
// 				} else if tc.shouldFail && err != io.EOF {
// 					t.Fatalf("Expected io.EOF. Got %v", err)
// 				}

// 				// readByteBuf should return a nil error with the byte read.
// 				if err == nil {
// 					if b != 0x01 {
// 						t.Fatalf("Expected byte 0x01. Got %x", b)
// 					}
// 				}
// 			})
// 		}
// 	}

// }

// // Test that the `discard` helper returns ErrUnexpectedEOF when it discarded
// // some bytes not all and an EOF was encountered along the way.
// func TestDiscardReturnsErrUnexpectedEOF(t *testing.T) {
// 	type testCase struct {
// 		name   string
// 		reader io.Reader
// 	}
// 	newTestCases := func() []testCase {
// 		return []testCase{
// 			{name: "Reader that returns EOF and n bytes read", reader: &testReader1Byte{b: 0x01}},
// 			{name: "Byte buffer", reader: bytes.NewBuffer([]byte{0x01})},
// 			{name: "Byte Reader", reader: bytes.NewReader([]byte{0x01})},
// 			{name: "bufio Reader", reader: bufio.NewReader(bytes.NewReader([]byte{0x01}))},
// 			{name: "bufio Reader with testReader", reader: bufio.NewReader(&testReader1Byte{b: 0x01})},
// 		}
// 	}

// 	// Check that discard returns ErrUnexpectedEOF when it reads 1 but not all the bytes
// 	for _, tc := range newTestCases() {
// 		t.Run(fmt.Sprintf("discard many bytes against %s", tc.name), func(t *testing.T) {
// 			err := discard(tc.reader, 2)
// 			if err == nil {
// 				// All of these test cases will fail since we are discarding many bytes.
// 				t.Fatalf("Expected error. Got nil")
// 			} else if err != io.ErrUnexpectedEOF {
// 				t.Fatalf("Expected io.ErrUnexpectedEOF. Got %v", err)
// 			}
// 		})
// 	}
// }

// type testReader1Byte struct {
// 	emptied bool
// 	b       byte
// }

// func (tr *testReader1Byte) Read(p []byte) (n int, err error) {
// 	if tr.emptied {
// 		return 0, io.EOF
// 	}

// 	written, err := bytes.NewReader([]byte{tr.b}).Read(p)
// 	if written != 1 {
// 		panic("unreachable. testReader1Byte has a single byte" + err.Error())
// 	}
// 	tr.emptied = true
// 	return 1, io.EOF
// }

// type TTA struct{}
// type TTA_B struct{}
// type TTB struct{}

// func TestTypeSorter(t *testing.T) {
// 	sortedTypes := sortTypeNames([]any{
// 		TTA_B{},
// 		TTB{},
// 		TTA{},
// 	})
// 	_, ok := sortedTypes[0].(TTA)
// 	if !ok {
// 		t.Errorf("wanted [0]TTA, got %T", sortedTypes[0])
// 	}
// 	_, ok = sortedTypes[1].(TTA_B)
// 	if !ok {
// 		t.Errorf("wanted [0]TTA_B, got %T", sortedTypes[1])
// 	}
// 	_, ok = sortedTypes[2].(TTB)
// 	if !ok {
// 		t.Errorf("wanted [0]TTB, got %T", sortedTypes[2])
// 	}
// }
