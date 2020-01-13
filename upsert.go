package neo4j

import (
	"fmt"
	"github.com/go-ginger/models"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"reflect"
	"strings"
)

func (handler *DbHandler) getInsertOnlyFields(model interface{}) []string {
	onlyInsertFields := make([]string, 0)
	value := reflect.ValueOf(model)
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		fType := valueType.Field(i)
		if fType.Type.Kind() == reflect.Struct {
			field := value.Field(i)
			nested := handler.getInsertOnlyFields(field.Interface())
			if len(nested) > 0 {
				onlyInsertFields = append(onlyInsertFields, nested...)
			}
		}
		tag, ok := fType.Tag.Lookup("neo")
		if ok {
			tagParts := strings.Split(tag, ",")
			for _, part := range tagParts {
				if part == "insert_only" {
					jsonTag, ok := fType.Tag.Lookup("json")
					if !ok {
						continue
					}
					jsonTagParts := strings.Split(jsonTag, ",")
					fieldName := jsonTagParts[0]
					onlyInsertFields = append(onlyInsertFields, fieldName)
					break
				}
			}
		}
	}
	return onlyInsertFields
}

func (handler *DbHandler) Upsert(request models.IRequest) (err error) {
	req := request.GetBaseRequest()
	nodeName := handler.DB.Config.NodeNamer.GetName(req.Body)
	bodyMap, params := structToMapParam(req.Body)
	// filter
	err = handler.NormalizeFilter(req.Filters)
	if err != nil {
		return
	}
	filterStr, filterParams, err := handler.getFilterStr(request)
	if err != nil {
		return
	}
	for k, v := range filterParams {
		params[k] = v
	}
	//
	insertOnlyFields := handler.getInsertOnlyFields(req.Body)
	insertData := map[string]interface{}{}
	for _, k := range insertOnlyFields {
		for key, value := range bodyMap {
			if k == key {
				insertData[key] = value
				delete(bodyMap, key)
			}
		}
	}
	//
	query := fmt.Sprintf("MERGE (n:%s %s) ", nodeName, filterStr)
	if len(insertData) > 0 {
		sets := ""
		for k, v := range insertData {
			sets += fmt.Sprintf(",n.%v=%v", k, v)
		}
		query += fmt.Sprintf("ON CREATE SET %s ", sets[1:])
	}
	if len(bodyMap) > 0 {
		sets := ""
		for k, v := range bodyMap {
			sets += fmt.Sprintf(",n.%v=%v", k, v)
		}
		query += fmt.Sprintf("ON CREATE SET %s ", sets[1:])
		query += fmt.Sprintf("ON MATCH SET %s ", sets[1:])
	}
	//query += "return n"
	session, err := handler.DB.Driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return
	}
	defer func() {
		e := session.Close()
		if e != nil && err == nil {
			err = e
			return
		}
	}()
	queryResult, err := session.Run(query, params)
	if err != nil {
		return
	}
	//for queryResult.Next() {
	//}
	if err = queryResult.Err(); err != nil {
		return
	}
	err = handler.BaseDbHandler.Upsert(req)
	return
}
