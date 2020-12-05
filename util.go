package recycle

import (
	"encoding/base64"
	"fmt"

	thrifter "github.com/thrift-iterator/go"
	"github.com/thrift-iterator/go/general"
	"github.com/thrift-iterator/go/protocol"
	"go.uber.org/thriftrw/compile"
)

func ParseTypeSpec(filePath, typeName string) (compile.TypeSpec, error) {
	module, err := compile.Compile(filePath, compile.NonStrict())
	if err != nil {
		return nil, err
	}

	for name, typeSpec := range module.Types {
		if name == typeName {
			return typeSpec, nil
		}
	}

	return nil, fmt.Errorf("type name %s not found", typeName)
}

func DecodeThrift(raw string) (*general.Message, error) {
	rawBytes, ok := isBase64(raw)
	if !ok {
		rawBytes = []byte(raw)
	}

	var msg general.Message
	err := thrifter.Unmarshal(rawBytes, &msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func IsCall(msg *general.Message) bool {
	return msg.MessageType == protocol.MessageTypeCall
}

func IsReply(msg *general.Message) bool {
	return msg.MessageType == protocol.MessageTypeReply
}

func isBase64(s string) ([]byte, bool) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, false
	}
	return bytes, true
}

func typeName(ts compile.TypeSpec) string {
	switch ts.(type) {
	case *compile.StructSpec:
		return "struct"
	case *compile.MapSpec:
		return "map"
	case *compile.ListSpec:
		return "list"
	case *compile.SetSpec:
		return "set"
	case *compile.EnumSpec:
		// enum always should be int32
		return "int32"
	case *compile.BoolSpec:
		return "bool"
	case *compile.I8Spec:
		return "byte"
	case *compile.I16Spec:
		return "int16"
	case *compile.I32Spec:
		return "int32"
	case *compile.I64Spec:
		return "int64"
	case *compile.DoubleSpec:
		return "float64"
	case *compile.StringSpec:
		return "string"
	case *compile.BinarySpec:
		return "[]byte"
	default:
		return "unknown"
	}
}
