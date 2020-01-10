package neo4j

import (
	"github.com/go-ginger/dl"
)

type DbHandler struct {
	dl.BaseDbHandler

	DB *DB
}
