package neo4j

import (
	"fmt"
	"github.com/go-ginger/models"
	"github.com/go-ginger/models/errors"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func (handler *DbHandler) Delete(request models.IRequest) (err error) {
	req := request.GetBaseRequest()
	model := handler.GetModelInstance()
	nodeName := handler.DB.Config.NodeNamer.GetName(model)
	// filter
	err = handler.NormalizeFilter(req.Filters)
	if err != nil {
		return
	}
	filterStr, params, err := handler.getFilterStr(request)
	if err != nil {
		return
	}
	//
	query := fmt.Sprintf("MATCH (n:%s %s) "+
		"DELETE n", nodeName, filterStr)
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
	summary, err := queryResult.Summary()
	if err != nil {
		return
	}
	if summary.Counters().NodesDeleted() == 0 {
		return errors.GetError(errors.NotFoundError)
	}
	if err = queryResult.Err(); err != nil {
		return
	}
	err = handler.BaseDbHandler.Delete(req)
	return
}
