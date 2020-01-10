package neo4j

import (
	"fmt"
	"github.com/go-ginger/models"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"math"
	"strings"
)

func (handler *DbHandler) Paginate(request models.IRequest) (result *models.PaginateResult, err error) {
	req := request.GetBaseRequest()
	model := handler.GetModelInstance()
	nodeName := handler.DB.Config.NodeNamer.GetName(model)
	keys := getKeys("n", model)
	query := fmt.Sprintf("MATCH (n:%s)	"+
		"RETURN %s", nodeName, strings.Join(keys, ","))
	params := map[string]interface{}{}

	var totalCount uint64
	//done := make(chan bool, 1)
	//go handler.countDocuments(db, collection, filter, done, &totalCount)

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
	if err = queryResult.Err(); err != nil {
		return
	}
	pageCount := uint64(math.Ceil(float64(totalCount) / float64(req.PerPage)))
	result = &models.PaginateResult{
		Items: queryResult,
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
