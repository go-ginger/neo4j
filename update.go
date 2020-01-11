package neo4j

import (
	"fmt"
	"github.com/go-ginger/models"
	"github.com/go-ginger/models/errors"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func (handler *DbHandler) Update(request models.IRequest) (err error) {
	req := request.GetBaseRequest()
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
	nodeName := handler.DB.Config.NodeNamer.GetName(req.Body)
	bodyMap, params := structToMapParam(req.Body)
	if err != nil {
		return
	}
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
	query := fmt.Sprintf("MATCH (n:%s %s) ", nodeName, filterStr)
	if len(bodyMap) > 0 {
		sets := ""
		for k, v := range bodyMap {
			sets += fmt.Sprintf(",n.%v=%v", k, v)
		}
		query += fmt.Sprintf("SET %s ", sets[1:])
	}
	query += "RETURN n"
	queryResult, err := session.Run(query, params)
	if err != nil {
		return
	}
	if err = queryResult.Err(); err != nil {
		return
	}
	updated := false
	for queryResult.Next() {
		updated = true
	}
	if !updated {
		err = errors.GetError(errors.NotFoundError)
		return
	}
	_, err = handler.BaseDbHandler.Insert(req)
	return
}
