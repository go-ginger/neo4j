package neo4j

import (
	"encoding/json"
	"github.com/go-ginger/models"
	"strings"
)

func (handler *DbHandler) NormalizeFilter(filters *models.Filters) (err error) {
	if filters == nil {
		return
	}
	if id, ok := (*filters)["_id"]; ok {
		delete(*filters, "_id")
		filters.Add("id", id)
	}
	return
}

func (handler *DbHandler) getFilterStr(request models.IRequest) (filterStr string, params map[string]interface{}, err error) {
	req := request.GetBaseRequest()
	var filters map[string]interface{}
	if req.Filters != nil {
		filters = *req.Filters
	}
	if filters == nil {
		filters = map[string]interface{}{}
	}
	filterMap, params := structToMapParam(filters, "f_")
	marshalledFilter, err := json.Marshal(filterMap)
	if err != nil {
		return
	}
	filterStr = string(marshalledFilter)
	filterStr = strings.Replace(filterStr, "\"", "", -1)
	return
}
