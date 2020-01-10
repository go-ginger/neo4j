package neo4j

import (
	"github.com/go-ginger/helpers/namer"
	"github.com/go-ginger/models"
	neo "github.com/neo4j/neo4j-go-driver/neo4j"
)

type Config struct {
	models.IConfig

	Target        string
	AuthToken     *neo.AuthToken
	NodeNamer     namer.INamer
	RelationNamer namer.INamer

	DB *DB
}
