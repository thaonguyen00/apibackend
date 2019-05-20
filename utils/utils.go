package utils

import (
	"reflect"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"github.com/pkg/errors"
)

//

func GetField(v interface{}, field string) (string, error) {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	str:= fmt.Sprint(f)
	if str == "<invalid reflect.Value>" {
		return "",  errors.New("Field doesn't exist")
	}
	return str, nil
}

func CreateFilterFromString(filter interface{}) (map[string]interface{}, error) {
	var str string
	if reflect.TypeOf(filter) != reflect.TypeOf(str) {
		return nil, fmt.Errorf("Not compatible type")

	}
	// parse search query
	var searchQueryMongo map[string] interface{}
	err := bson.UnmarshalJSON([]byte(filter.(string)),&searchQueryMongo)
	if err != nil {
		return nil, fmt.Errorf("Invalid search query")
	}

	return searchQueryMongo, nil

}