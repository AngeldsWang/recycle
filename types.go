package recycle

import "github.com/thrift-iterator/go/general"

type recycleType interface {
	Set(k, v interface{})
}

type recycleStruct map[string]interface{}
type recycleMap map[interface{}]interface{}
type recycleList []interface{}

func NewRecycleStruct() recycleStruct {
	return make(recycleStruct)
}

func NewRecycleMap() recycleMap {
	return make(recycleMap)
}

func NewRecycleList(len int) recycleList {
	return make(recycleList, len)
}

func (rs recycleStruct) Set(key, val interface{}) {
	field, _ := key.(string)
	rs[field] = val
}

func (rm recycleMap) Set(key, val interface{}) {
	rm[key] = val
}

func (rl recycleList) Set(key, val interface{}) {
	idx, _ := key.(int)
	rl[idx] = val
}

func NewRecycleType(v interface{}, len int) recycleType {
	switch v.(type) {
	case general.Struct:
		return NewRecycleStruct()
	case general.Map:
		return NewRecycleMap()
	case general.List:
		return NewRecycleList(len)
	default:
		return nil
	}
}
