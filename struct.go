package neo4j

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

var cachedKeys map[string][]string

func init() {
	cachedKeys = make(map[string][]string)
}

func mapToMapParam(item map[string]interface{}, existingParams map[string]interface{}, prefix ...string) (result, params map[string]interface{}) {
	var prf string
	if prefix != nil && len(prefix) > 0 {
		prf = prefix[0]
	}
	result = map[string]interface{}{}
	params = map[string]interface{}{}
	for k, v := range item {
		k2 := prf + k
		if existingParams != nil {
			tempK := k2
			ind := 1
			for {
				if _, keyExists := existingParams[tempK]; keyExists {
					ind++
					tempK = fmt.Sprintf("%s_%d", k2, ind)
				} else {
					break
				}
			}
			k2 = tempK
		}
		if mapItem, ok := v.(map[string]interface{}); ok {
			nestedResult, nestedParams := mapToMapParam(mapItem, params, prefix...)
			for rk, rv := range nestedResult {
				result[fmt.Sprintf("%s$%s", k, rk)] = rv
			}
			for pk, pv := range nestedParams {
				params[pk] = pv
			}
		} else {
			result[k] = "$" + k2
			params[k2] = v
		}
	}
	return
}

func structToMap(item interface{}) (result map[string]interface{}, err error) {
	if item == nil {
		return
	}
	bytes, err := json.Marshal(item)
	if err != nil {
		return
	}
	result = map[string]interface{}{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return
	}
	return
}

func structToMapParam(item interface{}, prefix ...string) (result, params map[string]interface{}) {
	params = map[string]interface{}{}
	if item == nil {
		return
	}
	result, err := structToMap(item)
	if err != nil {
		return
	}
	result, params = mapToMapParam(result, nil, prefix...)
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
		if ok && tag != "-" {
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
