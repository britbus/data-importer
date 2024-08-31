package databaselookup

import (
	"context"
	"fmt"
	"time"

	"github.com/travigo/travigo/pkg/ctdf"
	"github.com/travigo/travigo/pkg/dataaggregator/query"
	"github.com/travigo/travigo/pkg/dataaggregator/source/cachedresults"
	"github.com/travigo/travigo/pkg/database"
	"github.com/travigo/travigo/pkg/transforms"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s Source) ServicesByStopQuery(q query.ServicesByStop) ([]*ctdf.Service, error) {
	var services []*ctdf.Service
	// Load from cache
	cacheItemPath := fmt.Sprintf("cachedresults/servicesbystopquery/%s", q.Stop.PrimaryIdentifier)
	services, err := cachedresults.Get[[]*ctdf.Service](s.CachedResults, cacheItemPath)
	if err == nil {
		return services, nil
	}

	// If not in cache then fallback to lookup
	servicesCollection := database.GetCollection("services")
	journeysCollection := database.GetCollection("journeys")

	// Contains the stops primary id and all platforms primary ids
	allStopIDs := q.Stop.GetAllStopIDs()
	filter := bson.M{
		"$or": bson.A{
			bson.M{"path.originstopref": bson.M{"$in": allStopIDs}},
			bson.M{"path.destinationstopref": bson.M{"$in": allStopIDs}},
		},
	}

	journeyOpts := options.Find().SetProjection(bson.D{
		bson.E{Key: "serviceref", Value: 1},
	})

	serviceOpts := options.FindOne().SetProjection(bson.D{
		bson.E{Key: "creationdatetime", Value: 0},
		bson.E{Key: "modificationdatetime", Value: 0},
		bson.E{Key: "otheridentifiers", Value: 0},
		bson.E{Key: "routes", Value: 0},
	})

	serviceFound := map[string]bool{}

	cursor, _ := journeysCollection.Find(context.Background(), filter, journeyOpts)
	for cursor.Next(context.Background()) {
		var journey struct {
			ServiceRef string
		}
		cursor.Decode(&journey)

		if !serviceFound[journey.ServiceRef] {
			serviceFound[journey.ServiceRef] = true

			var service *ctdf.Service
			servicesCollection.FindOne(context.Background(), bson.M{"primaryidentifier": journey.ServiceRef}, serviceOpts).Decode(&service)

			if service != nil {
				transforms.Transform(service, 1)
				services = append(services, service)
			}
		}
	}

	// Save into cache
	cachedresults.Set(s.CachedResults, cacheItemPath, services, 24*time.Hour)

	return services, nil
}
