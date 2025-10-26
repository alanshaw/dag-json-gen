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
	jr := NewDagJsonReader(r)
	typ, err := jr.PeekType()
	if err != nil {
		return err
	}
	switch typ {
	case "object":
		if err := jr.ReadObjectOpen(); err != nil {
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
				if _, err := fmt.Fprintf(w, `{"%s":`, k); err != nil {
					return err
				}
				if err := parse(jr, w); err != nil {
					return err
				}
				close, err = jr.ReadObjectCloseOrComma()
				if err != nil {
					return err
				}
			}
			if close {
				if _, err := fmt.Fprintf(w, `}`); err != nil {
					return err
				}
				break
			}
			if _, err := fmt.Fprintf(w, `,`); err != nil {
				return err
			}
		}
	case "array":
		if err := jr.ReadArrayOpen(); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, `[`); err != nil {
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
				if err := parse(jr, w); err != nil {
					return err
				}
				close, err = jr.ReadArrayCloseOrComma()
				if err != nil {
					return err
				}
			}
			if close {
				if _, err := fmt.Fprintf(w, `]`); err != nil {
					return err
				}
				break
			}
			if _, err := fmt.Fprintf(w, `,`); err != nil {
				return err
			}
		}
	case "number":
		n, err := jr.ReadNumberAsString(MaxLength)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%s", n); err != nil {
			return err
		}
	case "string":
		s, err := jr.ReadString(MaxLength)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, `"%s"`, s); err != nil {
			return err
		}
	case "boolean":
		b, err := jr.ReadBool()
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%t", b); err != nil {
			return err
		}
	case "null":
		if err := jr.ReadNull(); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "null"); err != nil {
			return err
		}
	default:
		panic(fmt.Errorf("unknown JSON type: %s", typ))
	}
	return nil
}
