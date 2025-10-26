package typegen

import (
	"io"

	"github.com/ipfs/go-cid"
)

type JsonCid cid.Cid

func (c JsonCid) MarshalDagJSON(w io.Writer) error {
	jw := NewDagJsonWriter(w)
	return jw.WriteCid(cid.Cid(c))
}

func (c *JsonCid) UnmarshalDagJSON(r io.Reader) error {
	jr := NewDagJsonReader(r)
	oc, err := jr.ReadCid()
	if err != nil {
		return err
	}
	*c = JsonCid(oc)
	return nil
}
