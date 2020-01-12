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

type keysOptions struct {
	isType        bool
	keyPrefix     string
	isAnonymous   bool
	processedKeys map[string]bool
}

func getKeys(nodeName string, model interface{}, separator string, options ...keysOptions) (keys []string) {
	var option *keysOptions
	if options != nil && len(options) > 0 {
		option = &options[0]
	}
	if option != nil && option.processedKeys == nil {
		option.processedKeys = make(map[string]bool, 0)
	}
	keys = make([]string, 0)
	var sType reflect.Type
	if option != nil && option.isType {
		sType = model.(reflect.Type)
	} else {
		sType = reflect.TypeOf(model)
	}
	if sType.Kind() == reflect.Ptr {
		sType = sType.Elem()
	}
	sTypeName := sType.Name()
	if keys, ok := cachedKeys[sTypeName]; ok {
		return keys
	}
	for i := 0; i < sType.NumField(); i++ {
		tf := sType.Field(i)
		dlTag, ok := tf.Tag.Lookup("dl")
		if ok {
			parts := strings.Split(dlTag, ",")
			skip := false
			for _, part := range parts {
				if part == "read_only" {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}
		tag, ok := tf.Tag.Lookup("json")
		if ok && tag == "-" {
			continue
		}
		if (ok && tag != "-") || (option != nil && option.isAnonymous) || tf.Anonymous {
			var name string
			if !tf.Anonymous {
				if ok {
					tp := strings.Split(tag, ",")
					name = tp[0]
					if name == "-" {
						continue
					}
				}
				if name == "" {
					name = tf.Name
				}
			}

			ft := sType.Field(i).Type
			kind := ft.Kind()
			if kind == reflect.Ptr {
				ft = ft.Elem()
				kind = ft.Kind()
			}
			if kind == reflect.Struct {
				prefix := ""
				if option != nil && option.keyPrefix != "" {
					prefix += option.keyPrefix
				}
				prefix += name
				var processedKeys map[string]bool
				if option != nil {
					processedKeys = option.processedKeys
				}
				nestedKeys := getKeys(nodeName, ft, separator, keysOptions{
					isType:      true,
					keyPrefix:   prefix,
					isAnonymous: tf.Anonymous,
				})
				if processedKeys != nil {
					for _, k := range nestedKeys {
						if _, processed := processedKeys[k]; !processed {
							keys = append(keys, k)
							processedKeys[k] = true
						}
					}
				} else {
					keys = append(keys, nestedKeys...)
				}
				continue
			}
			if option != nil && option.keyPrefix != "" {
				name = option.keyPrefix + separator + name
			}
			if nodeName != "" {
				name = nodeName + "." + name
			}
			if option != nil {
				if _, processed := option.processedKeys[name]; processed {
					continue
				}
				option.processedKeys[name] = true
			}
			keys = append(keys, name)
		}
	}
	cachedKeys[sTypeName] = keys
	return
}
