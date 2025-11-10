package typegen

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"

	cid "github.com/ipfs/go-cid"
	"pitr.ca/jsontokenizer"
)

var _ io.Writer = (*DagJsonWriter)(nil)

type DagJsonWriter struct {
	w io.Writer
}

func (d *DagJsonWriter) Write(p []byte) (n int, err error) {
	return d.w.Write(p)
}

func NewDagJsonWriter(w io.Writer) *DagJsonWriter {
	if jw, ok := w.(*DagJsonWriter); ok {
		return jw
	}
	return &DagJsonWriter{w}
}

func (d *DagJsonWriter) WriteArrayClose() error {
	_, err := fmt.Fprintf(d.w, "]")
	return err
}

func (d *DagJsonWriter) WriteArrayOpen() error {
	_, err := fmt.Fprintf(d.w, "[")
	return err
}

func (d *DagJsonWriter) WriteBigInt(n *big.Int) error {
	_, err := fmt.Fprintf(d.w, `%s`, n.String())
	return err
}

func (d *DagJsonWriter) WriteBool(b bool) error {
	_, err := fmt.Fprintf(d.w, `%t`, b)
	return err
}

func (d *DagJsonWriter) WriteBytes(b []byte) error {
	_, err := fmt.Fprintf(d.w, `{"/":{"bytes":"%s"}}`, base64.RawStdEncoding.EncodeToString(b))
	return err
}

func (d *DagJsonWriter) WriteCid(c cid.Cid) error {
	_, err := fmt.Fprintf(d.w, `{"/":"%s"}`, c)
	return err
}

func (d *DagJsonWriter) WriteComma() error {
	_, err := fmt.Fprintf(d.w, ",")
	return err
}

func (d *DagJsonWriter) WriteInt64(n int64) error {
	_, err := fmt.Fprintf(d.w, "%d", n)
	return err
}

func (d *DagJsonWriter) WriteNull() error {
	_, err := fmt.Fprintf(d.w, "null")
	return err
}

func (d *DagJsonWriter) WriteObjectClose() error {
	_, err := fmt.Fprintf(d.w, "}")
	return err
}

func (d *DagJsonWriter) WriteObjectColon() error {
	_, err := fmt.Fprintf(d.w, ":")
	return err
}

func (d *DagJsonWriter) WriteObjectOpen() error {
	_, err := fmt.Fprintf(d.w, "{")
	return err
}

func (d *DagJsonWriter) WriteString(s string) error {
	buf, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("writing JSON string: %w", err)
	}
	_, err = d.w.Write(buf)
	return err
}

func (d *DagJsonWriter) WriteUint8(n uint8) error {
	_, err := fmt.Fprintf(d.w, "%d", n)
	return err
}

func (d *DagJsonWriter) WriteUint64(n uint64) error {
	_, err := fmt.Fprintf(d.w, "%d", n)
	return err
}

var _ io.Reader = (*DagJsonReader)(nil)

type DagJsonReader struct {
	r    io.Reader
	tk   jsontokenizer.Tokenizer
	peek jsontokenizer.TokType
}

func NewDagJsonReader(r io.Reader) *DagJsonReader {
	if jr, ok := r.(*DagJsonReader); ok {
		return jr
	}
	return &DagJsonReader{r, jsontokenizer.New(r), -1}
}

func (d *DagJsonReader) token() (jsontokenizer.TokType, error) {
	if d.peek != -1 {
		tok := d.peek
		d.peek = -1
		return tok, nil
	}
	return d.tk.Token()
}

func (d *DagJsonReader) peekToken() (jsontokenizer.TokType, error) {
	if d.peek != -1 {
		return d.peek, nil
	}
	tok, err := d.tk.Token()
	if err != nil {
		return tok, err
	}
	d.peek = tok
	return tok, nil
}

// PeekType returns the name of the next JSON type in the stream. It errors if
// the next token is NOT the start of an object, array, number, string, boolean
// or null.
//
// It returns either "object", "array", "string", "number", "boolean" or "null".
func (d *DagJsonReader) PeekType() (string, error) {
	tok, err := d.peekToken()
	if err != nil {
		return "", err
	}
	switch tok {
	case jsontokenizer.TokNull:
		return "null", nil
	case jsontokenizer.TokTrue, jsontokenizer.TokFalse:
		return "boolean", nil
	case jsontokenizer.TokArrayOpen:
		return "array", nil
	case jsontokenizer.TokObjectOpen:
		return "object", nil
	case jsontokenizer.TokNumber:
		return "number", nil
	case jsontokenizer.TokString:
		return "string", nil
	default:
		return "", fmt.Errorf("unexpected state, wanted start of type but read: %s", tokenName(tok))
	}
}

// DiscardType reads the next JSON type in full and discards it.
func (d *DagJsonReader) DiscardType() error {
	typ, err := d.PeekType()
	if err != nil {
		return err
	}
	switch typ {
	case "object":
		if err := d.ReadObjectOpen(); err != nil {
			return err
		}
		for {
			close, err := d.PeekObjectClose()
			if err != nil {
				return err
			}
			if close {
				if err := d.ReadObjectClose(); err != nil {
					return err
				}
				break
			}
			if _, err := d.ReadString(MaxLength); err != nil {
				return err
			}
			if err := d.ReadObjectColon(); err != nil {
				return err
			}
			if err := d.DiscardType(); err != nil {
				return err
			}
			close, err = d.ReadObjectCloseOrComma()
			if err != nil {
				return err
			}
			if close {
				break
			}
		}
	case "array":
		if err := d.ReadArrayOpen(); err != nil {
			return err
		}
		for {
			close, err := d.PeekArrayClose()
			if err != nil {
				return err
			}
			if close {
				if err := d.ReadArrayClose(); err != nil {
					return err
				}
				break
			}
			if err := d.DiscardType(); err != nil {
				return err
			}
			close, err = d.ReadArrayCloseOrComma()
			if err != nil {
				return err
			}
			if close {
				break
			}
		}
	case "number":
		if _, err := d.ReadNumberAsString(MaxLength); err != nil {
			return err
		}
	case "string":
		if _, err := d.ReadString(MaxLength); err != nil {
			return err
		}
	case "boolean":
		if _, err := d.ReadBool(); err != nil {
			return err
		}
	case "null":
		if err := d.ReadNull(); err != nil {
			return err
		}
	default:
		panic(fmt.Errorf("unknown JSON type: %s", typ))
	}
	return nil
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
	if tok != jsontokenizer.TokNull {
		return fmt.Errorf("expected boolean but read %s", tokenName(tok))
	}
	return nil
}

// PeekNull returns true if the next token is null.
func (d *DagJsonReader) PeekNull() (bool, error) {
	tok, err := d.peekToken()
	if err != nil {
		return false, err
	}
	return tok == jsontokenizer.TokNull, nil
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
	case jsontokenizer.TokTrue:
		value = true
		return &value, nil
	case jsontokenizer.TokFalse:
		value = false
		return &value, nil
	case jsontokenizer.TokNull:
		return nil, nil
	default:
		return nil, fmt.Errorf("expected boolean but read %s", tokenName(tok))
	}
}

func (d *DagJsonReader) ReadBytes(maxLength int) ([]byte, error) {
	b, err := d.ReadBytesOrNull(maxLength)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, errors.New("expected bytes but read null")
	}
	return *b, nil
}

func (d *DagJsonReader) ReadBytesOrNull(maxLength int) (*[]byte, error) {
	tok, err := d.token()
	if err != nil {
		return nil, err
	}
	if tok == jsontokenizer.TokNull {
		return nil, nil
	}
	if tok != jsontokenizer.TokObjectOpen {
		return nil, fmt.Errorf("expected object open but read %s", tokenName(tok))
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
	if err := d.ReadObjectOpen(); err != nil {
		return nil, err
	}
	bytesKey, err := d.ReadString(5)
	if err != nil {
		return nil, err
	}
	if bytesKey != "bytes" {
		return nil, fmt.Errorf("expected \"bytes\" but read %s", bytesKey)
	}
	if err := d.ReadObjectColon(); err != nil {
		return nil, err
	}
	s, err := d.ReadString(base64.RawStdEncoding.EncodedLen(maxLength))
	if err != nil {
		return nil, err
	}
	decoded, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if err := d.ReadObjectClose(); err != nil {
		return nil, err
	}
	if err := d.ReadObjectClose(); err != nil {
		return nil, err
	}
	return &decoded, nil
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
	if tok == jsontokenizer.TokNull {
		return nil, nil
	}
	if tok != jsontokenizer.TokObjectOpen {
		return nil, fmt.Errorf("expected object open but read %s", tokenName(tok))
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

func (d *DagJsonReader) ReadNumberAsString(maxLength int) (string, error) {
	tok, err := d.token()
	if err != nil {
		return "", err
	}
	if tok != jsontokenizer.TokNumber {
		return "", fmt.Errorf("expected number but read %s", tokenName(tok))
	}
	var buf bytes.Buffer
	if _, err := d.tk.ReadNumber(NewLimitWriter(&buf, maxLength)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (d *DagJsonReader) ReadNumberAsUint8() (uint8, error) {
	tok, err := d.token()
	if err != nil {
		return 0, err
	}
	if tok != jsontokenizer.TokNumber {
		return 0, fmt.Errorf("expected number but read %s", tokenName(tok))
	}
	var buf bytes.Buffer
	if _, err := d.tk.ReadNumber(NewLimitWriter(&buf, 3)); err != nil {
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
	if tok == jsontokenizer.TokNull {
		return nil, nil
	}
	if tok != jsontokenizer.TokNumber {
		return nil, fmt.Errorf("expected number but read %s", tokenName(tok))
	}
	var buf bytes.Buffer
	if _, err := d.tk.ReadNumber(NewLimitWriter(&buf, 20)); err != nil {
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
	if tok == jsontokenizer.TokNull {
		return nil, nil
	}
	if tok != jsontokenizer.TokNumber {
		return nil, fmt.Errorf("expected number but read %s", tokenName(tok))
	}
	var buf bytes.Buffer
	if _, err := d.tk.ReadNumber(NewLimitWriter(&buf, 20)); err != nil {
		return nil, err
	}
	n, err := strconv.ParseUint(buf.String(), 10, 64)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (d *DagJsonReader) ReadNumberAsBigInt(maxLength int) (*big.Int, error) {
	tok, err := d.token()
	if err != nil {
		return nil, err
	}
	if tok != jsontokenizer.TokNumber {
		return nil, fmt.Errorf("expected number but read %s", tokenName(tok))
	}
	var buf bytes.Buffer
	if _, err := d.tk.ReadNumber(NewLimitWriter(&buf, maxLength)); err != nil {
		return nil, err
	}
	n, ok := big.NewInt(0).SetString(buf.String(), 10)
	if !ok {
		return nil, errors.New("failed to set big int value")
	}
	return n, nil
}

func (d *DagJsonReader) ReadString(maxLength int) (string, error) {
	value, err := d.ReadStringOrNull(maxLength)
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", errors.New("expected string but read null")
	}
	return *value, nil
}

func (d *DagJsonReader) ReadStringOrNull(maxLength int) (*string, error) {
	tok, err := d.token()
	if err != nil {
		return nil, err
	}
	if tok == jsontokenizer.TokNull {
		return nil, nil
	}
	if tok != jsontokenizer.TokString {
		return nil, fmt.Errorf("expected string but read %s", tokenName(tok))
	}

	var buf bytes.Buffer
	buf.Write([]byte(`"`))
	if _, err := d.tk.ReadString(NewLimitWriter(&buf, maxLength)); err != nil {
		return nil, err
	}
	buf.Write([]byte(`"`))
	var s string
	err = json.Unmarshal(buf.Bytes(), &s)
	if err != nil {
		return nil, fmt.Errorf("reading JSON string: %w", err)
	}
	return &s, nil
}

func (d *DagJsonReader) ReadObjectColon() error {
	tok, err := d.token()
	if err != nil {
		return err
	}
	if tok != jsontokenizer.TokObjectColon {
		return fmt.Errorf("expected object colon but read %s", tokenName(tok))
	}
	return nil
}

func (d *DagJsonReader) ReadObjectOpen() error {
	tok, err := d.token()
	if err != nil {
		return err
	}
	if tok != jsontokenizer.TokObjectOpen {
		return fmt.Errorf("expected object open but read %s", tokenName(tok))
	}
	return nil
}

// PeekArrayClose returns true if the next token is "}".
func (d *DagJsonReader) PeekObjectClose() (bool, error) {
	tok, err := d.peekToken()
	if err != nil {
		return false, err
	}
	return tok == jsontokenizer.TokObjectClose, nil
}

func (d *DagJsonReader) ReadObjectClose() error {
	tok, err := d.token()
	if err != nil {
		return err
	}
	if tok != jsontokenizer.TokObjectClose {
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
	case jsontokenizer.TokObjectClose:
		return true, nil
	case jsontokenizer.TokComma:
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
	if tok != jsontokenizer.TokArrayOpen {
		return fmt.Errorf("expected array open but read %s", tokenName(tok))
	}
	return nil
}

func (d *DagJsonReader) ReadArrayOpenOrNull() (bool, error) {
	tok, err := d.token()
	if err != nil {
		return false, err
	}
	if tok == jsontokenizer.TokNull {
		return false, nil
	}
	if tok != jsontokenizer.TokArrayOpen {
		return false, fmt.Errorf("expected array open but read %s", tokenName(tok))
	}
	return true, nil
}

// PeekArrayClose returns true if the next token is "]".
func (d *DagJsonReader) PeekArrayClose() (bool, error) {
	tok, err := d.peekToken()
	if err != nil {
		return false, err
	}
	return tok == jsontokenizer.TokArrayClose, nil
}

func (d *DagJsonReader) ReadArrayClose() error {
	tok, err := d.token()
	if err != nil {
		return err
	}
	if tok != jsontokenizer.TokArrayClose {
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
	case jsontokenizer.TokArrayClose:
		return true, nil
	case jsontokenizer.TokComma:
		return false, nil
	default:
		return false, fmt.Errorf("expected array close or comma but read %s", tokenName(tok))
	}
}

func tokenName(tok jsontokenizer.TokType) string {
	switch tok {
	case jsontokenizer.TokNull:
		return "null"
	case jsontokenizer.TokTrue, jsontokenizer.TokFalse:
		return "boolean"
	case jsontokenizer.TokArrayOpen:
		return "["
	case jsontokenizer.TokArrayClose:
		return "]"
	case jsontokenizer.TokObjectOpen:
		return "{"
	case jsontokenizer.TokObjectClose:
		return "}"
	case jsontokenizer.TokObjectColon:
		return ":"
	case jsontokenizer.TokComma:
		return ","
	case jsontokenizer.TokNumber:
		return "number"
	case jsontokenizer.TokString:
		return "string"
	default:
		panic(fmt.Errorf("unknown token: %d", tok))
	}
}

var ErrLimitExceeded = errors.New("limit exceeded")

func NewLimitWriter(w io.Writer, n int) io.Writer {
	return &LimitedWriter{w, n}
}

// A LimitedWriter allows up to N bytes of data to be written to W. If a
// written chunk would cause the limit to be exceeded, [ErrLimitExceeded] is
// returned and the written chunk is not written to W.
type LimitedWriter struct {
	W io.Writer // underlying writer
	N int       // max bytes remaining
}

func (l *LimitedWriter) Write(p []byte) (n int, err error) {
	if len(p) > l.N {
		return 0, ErrLimitExceeded
	}
	n, err = l.W.Write(p)
	l.N -= n
	return
}
