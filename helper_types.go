package typegen

import (
	"encoding/json"
	"io"
	"time"
)

type DagJsonTime time.Time

func (jt DagJsonTime) MarshalDagJSON(w io.Writer) error {
	nsecs := jt.Time().UnixNano()
	buf, err := json.Marshal(nsecs)
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

func (jt *DagJsonTime) UnmarshalDagJSON(r io.Reader) error {
	var nsecs int64
	jr := NewDagJsonReader(r)
	nsecs, err := jr.ReadNumberAsInt64()
	if err != nil {
		return nil
	}
	t := time.Unix(0, nsecs)
	*jt = (DagJsonTime)(t)
	return nil
}

func (jt DagJsonTime) Time() time.Time {
	return (time.Time)(jt)
}

func (jt DagJsonTime) MarshalJSON() ([]byte, error) {
	return jt.Time().MarshalJSON()
}

func (jt *DagJsonTime) UnmarshalJSON(b []byte) error {
	var t time.Time
	if err := t.UnmarshalJSON(b); err != nil {
		return err
	}
	*(*time.Time)(jt) = t
	return nil
}
