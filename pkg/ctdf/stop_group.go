package ctdf

import (
	"context"
	"log"
	"time"

	"github.com/britbus/britbus/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
)

const StopGroupIDFormat = "GB:STOPGRP:%s"

type StopGroup struct {
	Identifier string

	CreationDateTime     time.Time
	ModificationDateTime time.Time

	DataSource *DataSource

	Name   string
	Type   string
	Status string

	Stops []Stop `bson:"-"`
}

func (stopGroup *StopGroup) GetStops() {
	stopsCollection := database.GetCollection("stops")
	cursor, _ := stopsCollection.Find(context.Background(), bson.M{"associations.associatedidentifier": stopGroup.Identifier})

	for cursor.Next(context.TODO()) {
		//Create a value into which the single document can be decoded
		var stop *Stop
		err := cursor.Decode(&stop)
		if err != nil {
			log.Fatal(err)
		}

		stopGroup.Stops = append(stopGroup.Stops, *stop)
	}
}
