package typegen

import (
	"fmt"
	"io"
	"reflect"
	"sort"

	cid "github.com/ipfs/go-cid"
)

var errMaxLength = fmt.Errorf("length beyond maximum allowed")

type DagJsonUnmarshaler interface {
	UnmarshalDagJSON(io.Reader) error
}

type DagJsonMarshaler interface {
	MarshalDagJSON(io.Writer) error
}

func ReadCid(r io.Reader) (cid.Cid, error) {
	jr := NewDagJsonReader(r)
	if err := jr.ReadObjectOpen(); err != nil {
		return cid.Undef, err
	}
	slash, err := jr.ReadString(1)
	if err != nil {
		return cid.Undef, err
	}
	if slash != "/" {
		return cid.Undef, fmt.Errorf("expected / but read %s", slash)
	}
	s, err := jr.ReadString(59)
	if err != nil {
		return cid.Undef, err
	}
	// TODO: spec compliance:
	// decode multibase base32 then decode CIDv1
	// or
	// decode base58btc then decode CIDv0
	parsed, err := cid.Parse(s)
	if err != nil {
		return cid.Undef, err
	}
	if err := jr.ReadObjectClose(); err != nil {
		return cid.Undef, err
	}
	return parsed, nil
}

func WriteCid(w io.Writer, c cid.Cid) error {
	buf, err := cid.Cid(c).MarshalJSON()
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

// sort type example objects on name of type
func sortTypeNames(obs []any) []any {
	temp := make([]tnAny, len(obs))
	for i, ob := range obs {
		v := reflect.ValueOf(ob)
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		temp[i] = tnAny{v.Type().Name(), ob}
	}
	sortref := tnAnySorter(temp)
	sort.Sort(&sortref)
	out := make([]any, len(obs))
	for i, rec := range temp {
		out[i] = rec.ob
	}
	return out
}

// type-name and any
type tnAny struct {
	name string
	ob   any
}

type tnAnySorter []tnAny

// sort.Interface
func (tas *tnAnySorter) Len() int {
	return len(*tas)
}
func (tas *tnAnySorter) Less(i, j int) bool {
	return (*tas)[i].name < (*tas)[j].name
}
func (tas *tnAnySorter) Swap(i, j int) {
	t := (*tas)[i]
	(*tas)[i] = (*tas)[j]
	(*tas)[j] = t
}
