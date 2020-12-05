package recycle

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/thrift-iterator/go/general"
	"github.com/thrift-iterator/go/protocol"
	"go.uber.org/thriftrw/compile"
)

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
		paths = append(paths, "struct")
		r.typeNameByPath[formatPath(paths)] = vv.Name
		for _, field := range vv.Fields {
			if typeName(field.Type) != "struct" {
				paths = append(paths, fmt.Sprintf("%d|%v", field.ID, typeName(field.Type)))
			} else {
				paths = append(paths, fmt.Sprintf("%d", field.ID))
			}
			r.typeNameByPath[formatPath(paths)] = field.Name
			r.collect(field.Type, paths)
			paths = paths[:len(paths)-1]
		}
		paths = paths[:len(paths)-1]
	case *compile.MapSpec:
		paths = append(paths, "key")
		r.typeNameByPath[formatPath(paths)] = vv.KeySpec.ThriftName()
		r.collect(vv.KeySpec, paths)
		paths[len(paths)-1] = "value"
		r.typeNameByPath[formatPath(paths)] = vv.ValueSpec.ThriftName()
		r.collect(vv.ValueSpec, paths)
	case *compile.ListSpec:
		paths = append(paths, "member")
		r.typeNameByPath[formatPath(paths)] = vv.ValueSpec.ThriftName()
		r.collect(vv.ValueSpec, paths)
		paths = paths[:len(paths)-1]
	case *compile.SetSpec:
		paths = append(paths, "elem")
		r.typeNameByPath[formatPath(paths)] = vv.ValueSpec.ThriftName()
		r.collect(vv.ValueSpec, paths)
		paths = paths[:len(paths)-1]
	case *compile.EnumSpec:
		for _, item := range vv.Items {
			paths = append(paths, fmt.Sprintf("value:%d", item.Value))
			r.typeNameByPath[formatPath(paths)] = fmt.Sprintf("%s_%s", vv.Name, item.Name)
			paths = paths[:len(paths)-1]
		}
	default:
		return
	}
}

func (r *Recycler) polish(v interface{}, paths []string, curr map[string]interface{}) (string, interface{}) {
	switch vv := v.(type) {
	case general.Struct:
		paths = append(paths, "struct")
		path := formatPath(paths)
		var ret interface{}
		if typeName, ok := r.typeNameByPath[path]; ok {
			curr[typeName] = make(map[string]interface{})
			for id, field := range vv {
				paths = append(paths, fmt.Sprintf("%d", id))
				path, val := r.polish(field, paths, curr)
				if name, ok := r.typeNameByPath[path]; ok {
					curr[typeName].(map[string]interface{})[name] = val
				}
				paths = paths[:len(paths)-1]
			}
			ret = curr[typeName]
		}
		paths = paths[:len(paths)-1]
		return path, ret
	case general.Map:
		paths = append(paths, "map")
		path := formatPath(paths)
		var ret interface{}
		if typeName, ok := r.typeNameByPath[path]; ok {
			curr[typeName] = make(map[interface{}]interface{})

			for key, value := range vv {
				paths = append(paths, "key")
				path, val := r.polish(key, paths, curr)
				if name, ok := r.typeNameByPath[path]; ok {
					curr[typeName].(map[interface{}]interface{})[name] = val
				}

				paths[len(paths)-1] = "value"
				path, val = r.polish(value, paths, curr)
				if name, ok := r.typeNameByPath[path]; ok {
					curr[typeName].(map[interface{}]interface{})[name] = val
				}
				paths = paths[:len(paths)-1]
			}
			ret = curr[typeName]
		}
		paths = paths[:len(paths)-1]
		return path, ret
	case general.List:
		paths = append(paths, "list")
		path := formatPath(paths)
		var ret interface{}
		if typeName, ok := r.typeNameByPath[path]; ok {
			curr[typeName] = make([]interface{}, len(vv))

			for i := 0; i < len(vv); i++ {
				paths = append(paths, "member")
				path, val := r.polish(vv[i], paths, curr)
				if _, ok := r.typeNameByPath[path]; ok {
					curr[typeName].([]interface{})[i] = val
				}
				paths = paths[:len(paths)-1]
			}
			ret = curr[typeName]
		}
		paths = paths[:len(paths)-1]
		return path, ret
	default:
		paths = append(paths, fmt.Sprintf("%v", reflect.TypeOf(v)))
		path := formatPath(paths)
		paths = paths[:len(paths)-1]
		return path, vv
	}
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
		_, shape := recycler.polish(raw, paths, curr)
		shapes = append(shapes, shape)
	}

	return shapes, nil
}
