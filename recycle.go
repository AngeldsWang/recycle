package recycle

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/thrift-iterator/go/general"
	"github.com/thrift-iterator/go/protocol"
	"go.uber.org/thriftrw/compile"
)

type collectFunc func(ts compile.TypeSpec, paths []string)
type polishFunc func(v interface{}, paths []string, curr map[string]interface{}) (string, []string, interface{})

type Recycler struct {
	thriftPath     string
	targetName     string
	typeNameByPath map[string]string
}

func NewRecycler(thriftPath, targetName string) *Recycler {
	recycle := &Recycler{
		thriftPath:     thriftPath,
		targetName:     targetName,
		typeNameByPath: make(map[string]string),
	}

	typeSpec, err := ParseTypeSpec(thriftPath, targetName)
	if err != nil {
		panic(err)
	}
	recycle.collect(typeSpec, []string{})

	return recycle
}

func formatPath(paths []string) string {
	return strings.Join(paths, "|")
}

func (r *Recycler) collect(ts compile.TypeSpec, paths []string) {
	switch vv := ts.(type) {
	case *compile.StructSpec:
		r.collectStruct(vv, paths, r.collect)
	case *compile.MapSpec:
		r.collectMapKey(vv.KeySpec, paths, r.collect)
		r.collectMapValue(vv.ValueSpec, paths, r.collect)
	case *compile.ListSpec:
		r.collectList(vv, paths, r.collect)
	case *compile.SetSpec:
		r.collectSet(vv, paths, r.collect)
	case *compile.EnumSpec:
		r.collectEnum(vv, paths, r.collect)
	default:
		return
	}
}

func (r *Recycler) collectStruct(v *compile.StructSpec, paths []string, cf collectFunc) []string {
	paths = append(paths, "struct")
	r.typeNameByPath[formatPath(paths)] = v.Name
	for _, field := range v.Fields {
		r.collectField(field, paths, r.collect)
	}
	paths = paths[:len(paths)-1]
	return paths
}

func (r *Recycler) collectField(v *compile.FieldSpec, paths []string, cf collectFunc) []string {
	var node string
	if typeName(v.Type) != "struct" {
		node = fmt.Sprintf("%d|%v", v.ID, typeName(v.Type))
	} else {
		node = fmt.Sprintf("%d", v.ID)
	}
	return r.collectType(v.Type, node, v.ThriftName(), paths, cf)
}

func (r *Recycler) collectMapKey(v compile.TypeSpec, paths []string, cf collectFunc) []string {
	return r.collectType(v, "key", "", paths, cf)
}

func (r *Recycler) collectMapValue(v compile.TypeSpec, paths []string, cf collectFunc) []string {
	return r.collectType(v, "value", "", paths, cf)
}

func (r *Recycler) collectList(v *compile.ListSpec, paths []string, cf collectFunc) []string {
	return r.collectType(v.ValueSpec, "member", "", paths, cf)
}

func (r *Recycler) collectSet(v *compile.SetSpec, paths []string, cf collectFunc) []string {
	return r.collectType(v.ValueSpec, "elem", "", paths, cf)
}

func (r *Recycler) collectType(v compile.TypeSpec, node, name string, paths []string, cf collectFunc) []string {
	paths = append(paths, node)
	if len(name) > 0 {
		r.typeNameByPath[formatPath(paths)] = name
	} else {
		r.typeNameByPath[formatPath(paths)] = v.ThriftName()
	}
	cf(v, paths)
	paths = paths[:len(paths)-1]
	return paths
}

func (r *Recycler) collectEnum(v *compile.EnumSpec, paths []string, cf collectFunc) []string {
	for _, item := range v.Items {
		paths = append(paths, fmt.Sprintf("value:%d", item.Value))
		r.typeNameByPath[formatPath(paths)] = fmt.Sprintf("%s_%s", v.ThriftName(), item.ThriftName())
		paths = paths[:len(paths)-1]
	}
	return paths
}

func (r *Recycler) polish(v interface{}, paths []string, curr map[string]interface{}) (string, []string, interface{}) {
	switch vv := v.(type) {
	case general.Struct:
		return r.polishStruct(vv, paths, curr, r.polish)
	case general.Map:
		return r.polishMap(vv, paths, curr, r.polish)
	case general.List:
		return r.polishList(vv, paths, curr, r.polish)
	default:
		paths = append(paths, fmt.Sprintf("%v", reflect.TypeOf(v)))
		path := formatPath(paths)
		paths = paths[:len(paths)-1]
		return path, paths, vv
	}
}

func (r *Recycler) findNameByPath(path string) (name string, ok bool) {
	name, ok = r.typeNameByPath[path]
	return
}

func (r *Recycler) polishStruct(v general.Struct, paths []string, curr map[string]interface{}, pf polishFunc) (string, []string, interface{}) {
	paths = append(paths, "struct")
	path := formatPath(paths)
	var ret interface{}
	if typeName, ok := r.typeNameByPath[path]; ok {
		curr[typeName] = NewRecycleType(v, 0)
		for id, field := range v {
			r.polishType(field, fmt.Sprintf("%d", id), typeName, -1, paths, curr, r.polish)
		}
		ret = curr[typeName]
	}
	paths = paths[:len(paths)-1]
	return path, paths, ret
}

func (r *Recycler) polishMap(v general.Map, paths []string, curr map[string]interface{}, pf polishFunc) (string, []string, interface{}) {
	paths = append(paths, "map")
	path := formatPath(paths)
	var ret interface{}
	if typeName, ok := r.typeNameByPath[path]; ok {
		curr[typeName] = NewRecycleType(v, 0)
		for key, value := range v {
			r.polishType(key, "key", typeName, -1, paths, curr, r.polish)
			r.polishType(value, "value", typeName, -1, paths, curr, r.polish)
		}
		ret = curr[typeName]
	}
	paths = paths[:len(paths)-1]
	return path, paths, ret
}

func (r *Recycler) polishList(v general.List, paths []string, curr map[string]interface{}, pf polishFunc) (string, []string, interface{}) {
	paths = append(paths, "list")
	path := formatPath(paths)
	var ret interface{}
	if typeName, ok := r.typeNameByPath[path]; ok {
		curr[typeName] = NewRecycleType(v, len(v))
		for i := 0; i < len(v); i++ {
			r.polishType(v[i], "member", typeName, i, paths, curr, r.polish)
		}
		ret = curr[typeName]
	}
	paths = paths[:len(paths)-1]
	return path, paths, ret
}

func (r *Recycler) polishType(v interface{}, node, typeName string, idx int, paths []string, curr map[string]interface{}, pf polishFunc) []string {
	paths = append(paths, node)
	path, _, val := pf(v, paths, curr)
	name, ok := r.findNameByPath(path)
	if ok {
		if idx >= 0 {
			curr[typeName].(recycleType).Set(idx, val)
		} else {
			curr[typeName].(recycleType).Set(name, val)
		}
	}
	paths = paths[:len(paths)-1]
	return paths
}

func Polish(thriftPath, targetName string, data []string) ([]interface{}, error) {
	recycler := NewRecycler(thriftPath, targetName)

	shapes := make([]interface{}, 0)
	for _, line := range data {
		msg, err := DecodeThrift(line)
		if err != nil {
			continue
		}

		// only parse canonical request and response type fields
		var raw interface{}
		if IsCall(msg) {
			raw = msg.Arguments.Get(protocol.FieldId(1))
		} else if IsReply(msg) {
			raw = msg.Arguments.Get(protocol.FieldId(0))
		} else {
			continue
		}

		curr := make(map[string]interface{})
		paths := make([]string, 0)
		_, _, shape := recycler.polish(raw, paths, curr)
		shapes = append(shapes, shape)
	}

	return shapes, nil
}
