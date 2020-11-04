package neo4j

import (
	neo "github.com/neo4j/neo4j-go-driver/neo4j"
	"log"
)

type DB struct {
	Driver neo.Driver
	Config *Config
}

func (db *DB) Connect() {
	if db.Config.DB != nil {
		db.Driver = db.Config.DB.Driver
		return
	}
	auth := db.Config.AuthToken
	if auth == nil {
		noAuth := neo.NoAuth()
		auth = &noAuth
	}
	driver, err := neo.NewDriver(db.Config.Target, *auth, func(config *neo.Config) {
		if db.Config.NeoConfig != nil {
			*config = *db.Config.NeoConfig
		}
	})
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	db.Driver = driver
	db.Config.DB = db
	return
}
