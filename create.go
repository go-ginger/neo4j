package neo4j

import (
	"encoding/json"
	"fmt"
	"github.com/go-ginger/models"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"strings"
)

func (handler *DbHandler) Insert(request models.IRequest) (result models.IBaseModel, err error) {
	req := request.GetBaseRequest()
	nodeName := handler.DB.Config.NodeNamer.GetName(req.Body)
	bodyMap, params := structToMapParam(req.Body)
	marshalledBody, err := json.Marshal(bodyMap)
	if err != nil {
		return
	}
	body := string(marshalledBody)
	body = strings.Replace(body, "\"", "", -1)
	query := fmt.Sprintf("CREATE (n:%s %s) RETURN n", nodeName, body)
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
	for queryResult.Next() {
		result = req.Body
		return
	}
	if err = queryResult.Err(); err != nil {
		return
	}
	_, err = handler.BaseDbHandler.Insert(req)
	return
}
