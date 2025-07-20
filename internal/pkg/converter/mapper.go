package converter

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type CustomConverterFunc func(srcValue reflect.Value, destType reflect.Type) (reflect.Value, error)

var (
	converterRegistry = make(map[string]CustomConverterFunc)
	mu                sync.RWMutex
)

func RegisterConverter(funcName string, converterFunc CustomConverterFunc) {
	mu.Lock()
	defer mu.Unlock()
	converterRegistry[funcName] = converterFunc
}

func Convert[S any, D any](src S, destPtr *D) error {
	if destPtr == nil {
		return fmt.Errorf("converter: destPtr is nil")
	}
	srcValue := reflect.ValueOf(src)
	destValue := reflect.ValueOf(destPtr).Elem()

	if srcValue.Kind() != reflect.Struct {
		return fmt.Errorf("srcValue must be a struct")
	}
	if destValue.Kind() != reflect.Struct {
		return fmt.Errorf("destPtr must be a struct")
	}

	for i := 0; i < destValue.NumField(); i++ {
		destField := destValue.Type().Field(i)
		destFieldValue := destValue.Field(i)
		if !destFieldValue.CanSet() {
			continue
		}
		tag := destField.Tag.Get("mapper")
		if tag == "-" || tag == "omitempty" {
			continue
		}
		srcFieldName := destField.Name
		customConverterFuncName := ""
		tagParts := strings.Split(tag, ",")
		for _, part := range tagParts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "convertFunc:") {
				customConverterFuncName = strings.TrimPrefix(part, "convertFunc:")
			} else if part != "" && part != "-" && part != "omitempty" {
				srcFieldName = part
			}
		}

		srcFieldValue := srcValue.FieldByName(srcFieldName)

		if !srcFieldValue.IsValid() {
			continue
		}
		if strings.Contains(tag, "omitempty") && srcFieldValue.IsZero() {
			continue
		}
		if customConverterFuncName != "" {
			mu.RLock()
			converterFunc, ok := converterRegistry[customConverterFuncName]
			mu.RUnlock()

			if !ok {
				return fmt.Errorf("converter: custom converter func %s not found", customConverterFuncName)
			}
			convertedValue, err := converterFunc(srcFieldValue, destFieldValue.Type())
			if err != nil {
				return fmt.Errorf("converter: custom converter func %s failed", customConverterFuncName)
			}
			if convertedValue.IsValid() && convertedValue.Type().AssignableTo(destFieldValue.Type()) {
				destFieldValue.Set(convertedValue)
				continue
			} else if convertedValue.IsValid() {
				return fmt.Errorf("coverterFunc %s convert %s to type %s", customConverterFuncName, destField.Name, destFieldValue.Type())
			}
		}

		if srcFieldValue.Type().AssignableTo(destFieldValue.Type()) {
			destFieldValue.Set(srcFieldValue)
		} else if srcFieldValue.Kind() == reflect.Ptr && srcFieldValue.Elem().Type().AssignableTo(destFieldValue.Type()) {
			if !srcFieldValue.IsNil() {
				destFieldValue.Set(srcFieldValue.Elem())
			}
		} else if destFieldValue.Kind() == reflect.Ptr && srcFieldValue.Type().AssignableTo(destFieldValue.Type().Elem()) {
			ptr := reflect.New(destFieldValue.Type().Elem())
			ptr.Elem().Set(srcFieldValue)
			destFieldValue.Set(ptr)
		}
	}
	return nil
}
