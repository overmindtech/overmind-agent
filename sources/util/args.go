package util

import (
	"reflect"
)

// Validate checks a value against a list of valid values and will panic with a
// supplied error if the values is invalid
func Validate(validValues interface{}, value interface{}, errorString string) {
	s := reflect.ValueOf(validValues)
	if s.Kind() != reflect.Slice && s.Kind() != reflect.Array {
		panic("Validate() first argument given a non-slice type")
	}

	sliceValues := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		sliceValues[i] = s.Index(i).Interface()
	}

	if !Contain(sliceValues, value) {
		panic(errorString)
	}
}

// Contain will return true if a slice contains a a value
func Contain(container []interface{}, i interface{}) bool {
	for _, val := range container {
		if val == i {
			return true
		}
	}

	return false
}
