package converter

import (
	"fmt"
	"reflect"
)

type CustomTypeConverterFunc func(str interface{}, destType reflect.Type) (interface{}, error)

var customConverters = make(map[string]CustomTypeConverterFunc)

func RegisterCustomTypeConverter(srcType, destType reflect.Type, converterFunc CustomTypeConverterFunc) {
	key := fmt.Sprintf("%s->%s", srcType.String(), destType.String())
	customConverters[key] = converterFunc
}

func MustRegisterCustomTypeConverter(srcType, destType reflect.Type, converterFunc CustomTypeConverterFunc) {
	RegisterCustomTypeConverter(srcType, destType, converterFunc)
}

func ApplyCustomConverters() {
	//for key, converterFunc := range customConverters {
	//	parts := strings.Split(key, "->")
	//	if len(parts) != 2 {
	//		fmt.Printf("Warning: Invalid custom converter key format: %s\n", key)
	//		continue
	//	}
	//	srcTypeName := parts[0]
	//	destTypeName := parts[1]
	//
	//}
}
