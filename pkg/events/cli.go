package events

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/travigo/travigo/pkg/consumer"
	"github.com/travigo/travigo/pkg/ctdf"
	dataaggregator "github.com/travigo/travigo/pkg/dataaggregator/global"
	"github.com/travigo/travigo/pkg/database"
	"github.com/travigo/travigo/pkg/redis_client"
	"github.com/urfave/cli/v2"
)

func RegisterCLI() *cli.Command {
	return &cli.Command{
		Name:  "events",
		Usage: "Provides the events runner",
		Subcommands: []*cli.Command{
			{
				Name:  "run",
				Usage: "run events server",
				Action: func(c *cli.Context) error {
					if err := database.Connect(); err != nil {
						return err
					}
					if err := redis_client.Connect(); err != nil {
						return err
					}

					dataaggregator.Setup()

					redisConsumer := consumer.RedisConsumer{
						QueueName:       "events-queue",
						NumberConsumers: 5,
						BatchSize:       20,
						Timeout:         2 * time.Second,
						Consumer:        NewEventsBatchConsumer(),
					}
					redisConsumer.Setup()

					signals := make(chan os.Signal, 1)
					signal.Notify(signals, syscall.SIGINT)
					defer signal.Stop(signals)

					<-signals // wait for signal
					go func() {
						<-signals // hard exit on second signal (in case shutdown gets stuck)
						os.Exit(1)
					}()

					<-redis_client.QueueConnection.StopAllConsuming() // wait for all Consume() calls to finish

					return nil
				},
			},
			{
				Name:  "test-event",
				Usage: "generate a test event",
				Action: func(c *cli.Context) error {
					if err := redis_client.Connect(); err != nil {
						return err
					}

					serviceAlert := ctdf.ServiceAlert{
						PrimaryIdentifier: "GB:SERVICEALERT:TEST",

						AlertType: ctdf.ServiceAlertTypeServiceSuspended,

						Title: "Line Suspended",
						Text:  "Northern Line has been suspended due to a fault on the line",

						MatchedIdentifiers: []string{"gb-noc-TFLO:1-NTN-_-y05-590847:1-NTN-_-y05-590847"},
					}

					eventsQueue, err := redis_client.QueueConnection.OpenQueue("events-queue")
					if err != nil {
						log.Fatal().Err(err).Msg("Failed to start event queue")
					}

					event := ctdf.Event{
						Type:      ctdf.EventTypeServiceAlertCreated,
						Timestamp: time.Now(),
						Body:      serviceAlert,
					}

					eventBytes, _ := json.Marshal(event)

					eventsQueue.PublishBytes(eventBytes)

					return nil
				},
			},
		},
	}
}
