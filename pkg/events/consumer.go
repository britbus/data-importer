package events

import (
	"github.com/rs/zerolog/log"

	"github.com/adjust/rmq/v5"
	"github.com/kr/pretty"
)

type EventsBatchConsumer struct {
}

func NewEventsBatchConsumer() *EventsBatchConsumer {
	return &EventsBatchConsumer{}
}

func (c *EventsBatchConsumer) Consume(batch rmq.Deliveries) {
	payloads := batch.Payloads()

	for _, payload := range payloads {
		pretty.Println(string(payload))
	}

	if ackErrors := batch.Ack(); len(ackErrors) > 0 {
		for _, err := range ackErrors {
			log.Fatal().Err(err).Msg("Failed to consume from queue")
		}
	}
}
