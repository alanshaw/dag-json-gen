package typegen

import (
	"io"

	cid "github.com/ipfs/go-cid"
)

type JsonCid cid.Cid

func (c JsonCid) MarshalDagJSON(w io.Writer) error {
	return WriteCid(w, cid.Cid(c))
}

func (c *JsonCid) UnmarshalDagJSON(r io.Reader) error {
	oc, err := ReadCid(r)
	if err != nil {
		return err
	}
	*c = JsonCid(oc)
	return nil
}
