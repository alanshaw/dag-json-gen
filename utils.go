package typegen

import (
	"fmt"
	"io"
	"reflect"
	"sort"
)

var errMaxLength = fmt.Errorf("length beyond maximum allowed")

type DagJsonUnmarshaler interface {
	UnmarshalDagJSON(io.Reader) error
}

type DagJsonMarshaler interface {
	MarshalDagJSON(io.Writer) error
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
