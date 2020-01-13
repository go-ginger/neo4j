package neo4j

import (
	"encoding/json"
	"fmt"
	"github.com/go-ginger/helpers"
	"github.com/go-ginger/models"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"math"
	"strings"
)

func (handler *DbHandler) countDocuments(query string, params map[string]interface{},
	done chan bool, count *uint64) {
	session, err := handler.DB.Driver.Session(neo4j.AccessModeRead)
	if err != nil {
		return
	}
	defer func() {
		done <- true
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
	for queryResult.Next() {
		totalCount := uint64(queryResult.Record().Values()[0].(int64))
		*count += totalCount
	}
}

func (handler *DbHandler) populateMap(nodeName string, source map[string]interface{}) (result map[string]interface{}, err error) {
	result = map[string]interface{}{}
	for key, value := range source {
		di := strings.Index(key, "$")
		if di > 0 {
			actualKey := key[:di]
			nestedSource := make(map[string]interface{}, 0)
			actualKeyLen := len(actualKey)
			for k, v := range source {
				if strings.Index(k, actualKey) == 0 {
					nestedSource[k[actualKeyLen+1:]] = v
					delete(source, k)
				}
			}
			nestedObj, e := handler.populateMap(nodeName, nestedSource)
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
			result[key] = value
		}
	}
	return
}

func (handler *DbHandler) Paginate(request models.IRequest) (result *models.PaginateResult, err error) {
	req := request.GetBaseRequest()
	model := handler.GetModelInstance()
	nodeName := handler.DB.Config.NodeNamer.GetName(model)
	iNodeKey := req.GetTemp("node_key")
	var nodeKey string
	if iNodeKey != nil {
		nodeKey = iNodeKey.(string)
	} else {
		nodeKey = "n"
	}
	keyPrefix := nodeKey + "."
	parseResult := handler.QueryParser.Parse(request, keyPrefix)
	var query string
	var countQuery string
	var params map[string]interface{}
	var countParams map[string]interface{}
	if req.ExtraQuery != nil {
		if iQuery, ok := req.ExtraQuery["query"]; ok {
			query = iQuery.(string)
		}
		if iQuery, ok := req.ExtraQuery["count_query"]; ok {
			countQuery = iQuery.(string)
		}
		if iParams, ok := req.ExtraQuery["params"]; ok {
			params = iParams.(map[string]interface{})
		}
		if iParams, ok := req.ExtraQuery["countParams"]; ok {
			countParams = iParams.(map[string]interface{})
		}
	}
	if query == "" {
		query = fmt.Sprintf("MATCH (%s:%s) "+
			"WHERE %v "+
			"RETURN %s",
			nodeKey, nodeName, parseResult.GetQuery(), nodeKey)
	}
	if countQuery == "" {
		countQuery = fmt.Sprintf("MATCH (%s:%s) "+
			"WHERE %v "+
			"RETURN COUNT(%s)",
			nodeKey, nodeName, parseResult.GetQuery(), nodeKey)
	}
	if params == nil {
		params = parseResult.GetParams().(map[string]interface{})
	}
	if countParams == nil {
		countParams = params
	}
	var totalCount uint64
	done := make(chan bool, 1)
	go handler.countDocuments(countQuery, countParams, done, &totalCount)
	query += fmt.Sprintf(" SKIP %v LIMIT %v",
		(req.Page-1)*req.PerPage, req.PerPage)
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
	queryResult, err := session.Run(query, params)
	if err != nil {
		return
	}
	items := handler.GetModelsInstance()
	var count uint64 = 0
	for queryResult.Next() {
		count++
		record := queryResult.Record()
		iNode, _ := record.Get(nodeKey)
		node := iNode.(neo4j.Node)
		properties := node.Props()
		obj, e := handler.populateMap(nodeKey, properties)
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
	<-done
	pageCount := uint64(math.Ceil(float64(totalCount) / float64(req.PerPage)))
	result = &models.PaginateResult{
		Items: items,
		Pagination: models.PaginationInfo{
			Page:       req.Page,
			PerPage:    req.PerPage,
			PageCount:  pageCount,
			TotalCount: totalCount,
			HasNext:    req.Page < pageCount,
		},
	}
	return
}
