package neo4j

import (
	"encoding/json"
	"reflect"
	"strings"
)

var cachedKeys map[string][]string

func init() {
	cachedKeys = make(map[string][]string)
}

func mapToMapParam(item map[string]interface{}) (result, params map[string]interface{}) {
	result = map[string]interface{}{}
	params = map[string]interface{}{}
	for k, v := range item {
		if mapItem, ok := v.(map[string]interface{}); ok {
			nestedResult, nestedParams := mapToMapParam(mapItem)
			result[k] = nestedResult
			params[k] = nestedParams
		} else {
			result[k] = "$" + k
			params[k] = v
		}
	}
	return
}

func structToMapParam(item interface{}) (result, params map[string]interface{}) {
	result = map[string]interface{}{}
	params = map[string]interface{}{}
	if item == nil {
		return
	}
	bytes, err := json.Marshal(item)
	if err != nil {
		return
	}
	var data map[string]interface{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return
	}
	result, params = mapToMapParam(data)
	return
}

func getKeys(nodeName string, model interface{}) (keys []string) {
	if keys, ok := cachedKeys[nodeName]; ok {
		return keys
	}
	keys = make([]string, 0)
	s := reflect.ValueOf(model).Elem()
	sType := s.Type()
	for i := 0; i < s.NumField(); i++ {
		ff := sType.Field(i)
		tag, ok := ff.Tag.Lookup("json")
		if ok && tag != "-"{
			tp := strings.Split(tag, ",")
			for _, part := range tp {
				if part != "-" && part != "" {
					keys = append(keys, nodeName+"."+part)
					break
				}
			}
		}
	}
	cachedKeys[nodeName] = keys
	return
}
