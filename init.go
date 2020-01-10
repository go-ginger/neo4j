package neo4j

import (
	"github.com/go-ginger/helpers/namer"
)

func (handler *DbHandler) InitializeConfig(config *Config) {
	if config.NodeNamer == nil {
		config.NodeNamer = &namer.Default{}
	}
	if config.RelationNamer == nil {
		config.RelationNamer = &namer.Default{}
	}
	config.NodeNamer.Initialize()
	config.RelationNamer.Initialize()
	handler.DB = &DB{
		Config: config,
	}
	handler.DB.Connect()
	handler.InsertInBackground = true
	handler.UpdateInBackground = true
	handler.DeleteInBackground = true
	handler.IsFullObjectOnUpdateRequired = true
	return
}
