package helpers

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// https://stackoverflow.com/questions/67900919/check-if-a-map-is-subset-of-another-map
func IsMapSubset(mapSet interface{}, mapSubset interface{}) bool {

	mapSetValue := reflect.ValueOf(mapSet)
	mapSubsetValue := reflect.ValueOf(mapSubset)

	if fmt.Sprintf("%T", mapSet) != fmt.Sprintf("%T", mapSubset) {
		return false
	}

	if len(mapSetValue.MapKeys()) < len(mapSubsetValue.MapKeys()) {
		return false
	}

	if len(mapSubsetValue.MapKeys()) == 0 {
		return true
	}

	iterMapSubset := mapSubsetValue.MapRange()

	for iterMapSubset.Next() {
		k := iterMapSubset.Key()
		v := iterMapSubset.Value()

		value := mapSetValue.MapIndex(k)

		if !value.IsValid() || v.Interface() != value.Interface() {
			return false
		}
	}

	return true
}

//https://github.com/Myafq/limit-operator/blob/master/pkg/controller/clusterlimit/clusterlimit_controller.go

func areTheSame(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, ae := range a {
		if !includes(ae, b) {
			return false
		}
	}
	for _, be := range b {
		if !includes(be, a) {
			return false
		}
	}
	return true
}

func includes(a string, b []string) bool {
	for _, be := range b {
		if a == be {
			return true
		}
	}
	return false
}

//https://gosamples.dev/compare-slices/
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func AreEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}
