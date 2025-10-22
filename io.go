package typegen

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"

	cid "github.com/ipfs/go-cid"
	json "pitr.ca/jsontokenizer"
)

var _ io.Reader = (*DagJsonReader)(nil)

type DagJsonReader struct {
	r    io.Reader
	tk   json.Tokenizer
	peek json.TokType
}

func NewDagJsonReader(r io.Reader) *DagJsonReader {
	if r, ok := r.(*DagJsonReader); ok {
		return r
	}
	return &DagJsonReader{r, json.New(r), -1}
}

func (d *DagJsonReader) token() (json.TokType, error) {
	if d.peek != -1 {
		tok := d.peek
		d.peek = -1
		return tok, nil
	}
	return d.tk.Token()
}

// Read from the underlying reader. You almost certainly don't want to do this.
func (d *DagJsonReader) Read(p []byte) (int, error) {
	return d.r.Read(p)
}

func (d *DagJsonReader) ReadNull() error {
	tok, err := d.token()
	if err != nil {
		return err
	}
	if tok != json.TokNull {
		return fmt.Errorf("expected boolean but read %s", tokenName(tok))
	}
	return nil
}

// PeekNull returns true if the next token is null.
func (d *DagJsonReader) PeekNull() (bool, error) {
	if d.peek > -1 {
		return false, errors.New("reader is already peeked")
	}
	tok, err := d.token()
	if err != nil {
		return false, err
	}
	d.peek = tok
	return tok == json.TokNull, nil
}

func (d *DagJsonReader) ReadBool() (bool, error) {
	value, err := d.ReadBoolOrNull()
	if err != nil {
		return false, err
	}
	if value == nil {
		return false, errors.New("expected boolean but read null")
	}
	return *value, nil
}

func (d *DagJsonReader) ReadBoolOrNull() (*bool, error) {
	var value bool
	tok, err := d.token()
	if err != nil {
		return nil, err
	}
	switch tok {
	case json.TokTrue:
		value = true
		return &value, nil
	case json.TokFalse:
		value = false
		return &value, nil
	case json.TokNull:
		return nil, nil
	default:
		return nil, fmt.Errorf("expected boolean but read %s", tokenName(tok))
	}
}

func (d *DagJsonReader) ReadCid() (cid.Cid, error) {
	value, err := d.ReadCidOrNull()
	if err != nil {
		return cid.Undef, err
	}
	if value == nil {
		return cid.Undef, errors.New("expected CID but read null")
	}
	return *value, nil
}

func (d *DagJsonReader) ReadCidOrNull() (*cid.Cid, error) {
	tok, err := d.token()
	if err != nil {
		return nil, err
	}
	if tok == json.TokNull {
		return nil, nil
	}
	if err := d.ReadObjectOpen(); err != nil {
		return nil, err
	}
	slash, err := d.ReadString(1)
	if err != nil {
		return nil, err
	}
	if slash != "/" {
		return nil, fmt.Errorf("expected / but read %s", slash)
	}
	if err := d.ReadObjectColon(); err != nil {
		return nil, err
	}
	s, err := d.ReadString(59)
	if err != nil {
		return nil, err
	}
	// TODO: spec compliance:
	// decode multibase base32 then decode CIDv1
	// or
	// decode base58btc then decode CIDv0
	parsed, err := cid.Parse(s)
	if err != nil {
		return nil, err
	}
	if err := d.ReadObjectClose(); err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (d *DagJsonReader) ReadNumberAsUint8() (uint8, error) {
	tok, err := d.token()
	if err != nil {
		return 0, err
	}
	if tok != json.TokNumber {
		return 0, fmt.Errorf("expected number but read %s", tokenName(tok))
	}
	var buf bytes.Buffer
	if _, err := d.tk.ReadNumber(LimitWriter(&buf, 3)); err != nil {
		return 0, err
	}
	n, err := strconv.ParseUint(buf.String(), 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(n), nil
}

func (d *DagJsonReader) ReadNumberAsInt64() (int64, error) {
	value, err := d.ReadNumberAsInt64OrNull()
	if err != nil {
		return 0, err
	}
	if value == nil {
		return 0, errors.New("expected number but read null")
	}
	return *value, nil
}

func (d *DagJsonReader) ReadNumberAsInt64OrNull() (*int64, error) {
	tok, err := d.token()
	if err != nil {
		return nil, err
	}
	if tok != json.TokNumber {
		return nil, fmt.Errorf("expected number but read %s", tokenName(tok))
	}
	var buf bytes.Buffer
	if _, err := d.tk.ReadNumber(LimitWriter(&buf, 20)); err != nil {
		return nil, err
	}
	n, err := strconv.ParseInt(buf.String(), 10, 64)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (d *DagJsonReader) ReadNumberAsUint64() (uint64, error) {
	value, err := d.ReadNumberAsUint64OrNull()
	if err != nil {
		return 0, err
	}
	if value == nil {
		return 0, errors.New("expected number but read null")
	}
	return *value, nil
}

func (d *DagJsonReader) ReadNumberAsUint64OrNull() (*uint64, error) {
	tok, err := d.token()
	if err != nil {
		return nil, err
	}
	if tok != json.TokNumber {
		return nil, fmt.Errorf("expected number but read %s", tokenName(tok))
	}
	var buf bytes.Buffer
	if _, err := d.tk.ReadNumber(LimitWriter(&buf, 20)); err != nil {
		return nil, err
	}
	n, err := strconv.ParseUint(buf.String(), 10, 64)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (d *DagJsonReader) ReadNumberAsBigInt(maxLength int64) (*big.Int, error) {
	tok, err := d.token()
	if err != nil {
		return nil, err
	}
	if tok != json.TokNumber {
		return nil, fmt.Errorf("expected number but read %s", tokenName(tok))
	}
	var buf bytes.Buffer
	if _, err := d.tk.ReadNumber(LimitWriter(&buf, maxLength)); err != nil {
		return nil, err
	}
	n, ok := big.NewInt(0).SetString(buf.String(), 10)
	if !ok {
		return nil, errors.New("failed to set big int value")
	}
	return n, nil
}

func (d *DagJsonReader) ReadString(maxLength int64) (string, error) {
	value, err := d.ReadStringOrNull(maxLength)
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", errors.New("expected string but read null")
	}
	return *value, nil
}

func (d *DagJsonReader) ReadStringOrNull(maxLength int64) (*string, error) {
	tok, err := d.token()
	if err != nil {
		return nil, err
	}
	if tok != json.TokString {
		return nil, fmt.Errorf("expected string but read %s", tokenName(tok))
	}
	var buf bytes.Buffer
	if _, err := d.tk.ReadString(LimitWriter(&buf, maxLength)); err != nil {
		return nil, err
	}
	s := buf.String()
	return &s, nil
}

func (d *DagJsonReader) ReadObjectColon() error {
	tok, err := d.token()
	if err != nil {
		return err
	}
	if tok != json.TokObjectColon {
		return fmt.Errorf("expected object colon but read %s", tokenName(tok))
	}
	return nil
}

func (d *DagJsonReader) ReadObjectOpen() error {
	tok, err := d.token()
	if err != nil {
		return err
	}
	if tok != json.TokObjectOpen {
		return fmt.Errorf("expected object open but read %s", tokenName(tok))
	}
	return nil
}

func (d *DagJsonReader) ReadObjectClose() error {
	tok, err := d.token()
	if err != nil {
		return err
	}
	if tok != json.TokObjectClose {
		return fmt.Errorf("expected object close but read %s", tokenName(tok))
	}
	return nil
}

// returns true for closed object
func (d *DagJsonReader) ReadObjectCloseOrComma() (bool, error) {
	tok, err := d.token()
	if err != nil {
		return false, err
	}
	switch tok {
	case json.TokObjectClose:
		return true, nil
	case json.TokComma:
		return false, nil
	default:
		return false, fmt.Errorf("expected object close or comma but read %s", tokenName(tok))
	}
}

func (d *DagJsonReader) ReadArrayOpen() error {
	tok, err := d.token()
	if err != nil {
		return err
	}
	if tok != json.TokArrayOpen {
		return fmt.Errorf("expected array open but read %s", tokenName(tok))
	}
	return nil
}

func (d *DagJsonReader) ReadArrayClose() error {
	tok, err := d.token()
	if err != nil {
		return err
	}
	if tok != json.TokArrayClose {
		return fmt.Errorf("expected array close but read %s", tokenName(tok))
	}
	return nil
}

// returns true for closed array
func (d *DagJsonReader) ReadArrayCloseOrComma() (bool, error) {
	tok, err := d.token()
	if err != nil {
		return false, err
	}
	switch tok {
	case json.TokArrayClose:
		return true, nil
	case json.TokComma:
		return false, nil
	default:
		return false, fmt.Errorf("expected array close or comma but read %s", tokenName(tok))
	}
}

func tokenName(tok json.TokType) string {
	switch tok {
	case json.TokNull:
		return "null"
	case json.TokTrue, json.TokFalse:
		return "boolean"
	case json.TokArrayOpen, json.TokArrayClose, json.TokObjectOpen, json.TokObjectClose, json.TokObjectColon, json.TokComma:
		return "delimiter"
	case json.TokNumber:
		return "number"
	case json.TokString:
		return "string"
	default:
		return "unknown"
	}
}

// LimitWriter returns a Writer that writes to w
// but stops with EOF after n bytes.
// The underlying implementation is a *LimitedWriter.
func LimitWriter(w io.Writer, n int64) io.Writer { return &LimitedWriter{w, n} }

// A LimitedWriter writes to W but limits the amount of
// data returned to just N bytes. Each call to Write
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying W returns EOF.
type LimitedWriter struct {
	W io.Writer // underlying writer
	N int64     // max bytes remaining
}

func (l *LimitedWriter) Write(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.W.Write(p)
	l.N -= int64(n)
	return
}
