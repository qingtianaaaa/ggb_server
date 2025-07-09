package aiModule

import (
	"log"
	"reflect"
	"testing"
)

func Test_reflect(t *testing.T) {
	client := &ChatCompletionClient{}

	typeC := reflect.TypeOf(client)
	valueC := reflect.ValueOf(client)

	structFieldType := typeC
	structFieldValue := valueC
	if reflect.TypeOf(client).Kind() == reflect.Pointer {
		structFieldType = structFieldType.Elem()
		structFieldValue = structFieldValue.Elem()
	}

	for i := 0; i < structFieldType.NumField(); i++ {
		fieldValue := structFieldValue.Field(i)
		fieldType := structFieldType.Field(i)
		log.Println("filedType.Type.Name(): ", fieldType.Type.Name())
		log.Println("filedType.Type.Kind(): ", fieldType.Type.Kind())
		log.Println("fieldType.Name: ", fieldType.Name)
		log.Println("fieldValue.CanSet(): ", fieldValue.CanSet())
		if fieldType.Type.Kind() == reflect.String && fieldValue.CanSet() {
			fieldValue.SetString("hello")
			log.Println("fieldValue.Interface(): ", fieldValue.Interface())
		}
		log.Println("----")
	}
}
