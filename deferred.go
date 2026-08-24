package typegen

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type Deferred struct {
	Raw []byte
}

func (d *Deferred) MarshalDagJSON(w io.Writer) error {
	if d == nil {
		_, err := w.Write([]byte("null"))
		return err
	}
	if d.Raw == nil {
		return errors.New("cannot marshal Deferred with nil value for Raw (will not unmarshal)")
	}
	_, err := w.Write(d.Raw)
	return err
}

func (d *Deferred) UnmarshalDagJSON(r io.Reader) error {
	var buf bytes.Buffer
	err := parse(r, NewLimitWriter(&buf, ByteArrayMaxLen))
	if err != nil {
		return err
	}
	d.Raw = buf.Bytes()
	return nil
}

func parse(r io.Reader, w io.Writer) error {
	// no-ops when r/w are already wrapped, so recursion reuses one instance
	jr := NewDagJsonReader(r)
	jw := NewDagJsonWriter(w)
	typ, err := jr.PeekType()
	if err != nil {
		return err
	}
	switch typ {
	case "object":
		if err := jr.ReadObjectOpen(); err != nil {
			return err
		}
		if err := jw.WriteObjectOpen(); err != nil {
			return err
		}
		for {
			close, err := jr.PeekObjectClose()
			if err != nil {
				return err
			}
			if close {
				if err := jr.ReadObjectClose(); err != nil {
					return err
				}
			} else {
				k, err := jr.ReadString(MaxLength)
				if err != nil {
					return err
				}
				if err := jr.ReadObjectColon(); err != nil {
					return err
				}
				if err := jw.WriteString(k); err != nil {
					return err
				}
				if err := jw.WriteObjectColon(); err != nil {
					return err
				}
				if err := parse(jr, jw); err != nil {
					return err
				}
				close, err = jr.ReadObjectCloseOrComma()
				if err != nil {
					return err
				}
			}
			if close {
				if err := jw.WriteObjectClose(); err != nil {
					return err
				}
				break
			}
			if err := jw.WriteComma(); err != nil {
				return err
			}
		}
	case "array":
		if err := jr.ReadArrayOpen(); err != nil {
			return err
		}
		if err := jw.WriteArrayOpen(); err != nil {
			return err
		}
		for {
			close, err := jr.PeekArrayClose()
			if err != nil {
				return err
			}
			if close {
				if err := jr.ReadArrayClose(); err != nil {
					return err
				}
			} else {
				if err := parse(jr, jw); err != nil {
					return err
				}
				close, err = jr.ReadArrayCloseOrComma()
				if err != nil {
					return err
				}
			}
			if close {
				if err := jw.WriteArrayClose(); err != nil {
					return err
				}
				break
			}
			if err := jw.WriteComma(); err != nil {
				return err
			}
		}
	case "number":
		n, err := jr.ReadNumberAsString(MaxLength)
		if err != nil {
			return err
		}
		if _, err := io.WriteString(jw, n); err != nil {
			return err
		}
	case "string":
		s, err := jr.ReadString(MaxLength)
		if err != nil {
			return err
		}
		if err := jw.WriteString(s); err != nil {
			return err
		}
	case "boolean":
		b, err := jr.ReadBool()
		if err != nil {
			return err
		}
		if err := jw.WriteBool(b); err != nil {
			return err
		}
	case "null":
		if err := jr.ReadNull(); err != nil {
			return err
		}
		if err := jw.WriteNull(); err != nil {
			return err
		}
	default:
		panic(fmt.Errorf("unknown JSON type: %s", typ))
	}
	return nil
}
