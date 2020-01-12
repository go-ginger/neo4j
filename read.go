package neo4j

import (
	"encoding/json"
	"fmt"
	"github.com/go-ginger/helpers"
	"github.com/go-ginger/models"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"strings"
)

func (handler *DbHandler) getMapOfRecord(nodeName string, keys []string, values []interface{}) (result map[string]interface{}, err error) {
	result = map[string]interface{}{}
	keysLen := len(keys)
	for ind := 0; ind < keysLen; ind++ {
		key := keys[ind]
		di := strings.Index(key, "$")
		if di > 0 {
			actualKey := key[:di]
			objKeys := make([]string, 0)
			objValues := make([]interface{}, 0)
			actualKeyLen := len(actualKey)
			for ki := 0; ki < keysLen; ki++ {
				k := keys[ki]
				if strings.Index(k, actualKey) == 0 {
					objKeys = append(objKeys, k[actualKeyLen+1:])
					objValues = append(objValues, values[ki])
					keys = helpers.RemoveFromStringSlice(keys, ki)
					values = helpers.RemoveFromInterfaceSlice(values, ki)
					keysLen--
					ki--
				}
			}
			nestedObj, e := handler.getMapOfRecord(nodeName, objKeys, objValues)
			if e != nil {
				err = e
				return
			}
			nodeIndex := strings.Index(actualKey, nodeName+".")
			if nodeIndex == 0 {
				actualKey = actualKey[len(nodeName)+1:]
			}
			result[actualKey] = nestedObj
		} else {
			nodeIndex := strings.Index(key, nodeName+".")
			if nodeIndex == 0 {
				key = key[len(nodeName)+1:]
			}
			result[key] = values[ind]
		}
	}
	return
}

func (handler *DbHandler) Paginate(request models.IRequest) (result *models.PaginateResult, err error) {
	req := request.GetBaseRequest()
	model := handler.GetModelInstance()
	nodeName := handler.DB.Config.NodeNamer.GetName(model)
	nodeKey := "n"
	keys := getKeys(nodeKey, model, "$", keysOptions{})
	keyPrefix := "n."
	parseResult := handler.QueryParser.Parse(request, keyPrefix)
	query := fmt.Sprintf("MATCH (n:%s) "+
		"WHERE %v "+
		"RETURN %s "+
	//"ORDER BY n.id "+
		"SKIP %v LIMIT %v", nodeName, parseResult.GetQuery(), strings.Join(keys, ","),
		(req.Page-1)*req.PerPage, req.PerPage+1)
	err = handler.NormalizeFilter(req.Filters)
	if err != nil {
		return
	}
	session, err := handler.DB.Driver.Session(neo4j.AccessModeRead)
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
	queryResult, err := session.Run(query, parseResult.GetParams().(map[string]interface{}))
	if err != nil {
		return
	}
	items := handler.GetModelsInstance()
	var count uint64 = 0
	hasNext := false
	for queryResult.Next() {
		count++
		if count > req.PerPage {
			hasNext = true
			break
		}
		record := queryResult.Record()
		keys := record.Keys()
		values := record.Values()
		obj, e := handler.getMapOfRecord(nodeKey, keys, values)
		if e != nil {
			err = e
			return
		}
		bytes, e := json.Marshal(obj)
		if e != nil {
			err = e
			return
		}
		model := handler.GetModelInstance()
		err = json.Unmarshal(bytes, model)
		if err != nil {
			return
		}
		items = helpers.AppendToSlice(items, model)
	}
	if err = queryResult.Err(); err != nil {
		return
	}
	result = &models.PaginateResult{
		Items: items,
		Pagination: models.PaginationInfo{
			Page:    req.Page,
			PerPage: req.PerPage,
			//PageCount:  pageCount,
			//TotalCount: totalCount,
			HasNext: hasNext,
		},
	}
	return
}
