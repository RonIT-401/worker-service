package eprocessor

import (
	"context"
	"sync"

	"github.com/RonIT-401/worker-service/internal/app/entity"
	ehandler "github.com/RonIT-401/worker-service/internal/app/handler/event"
	"github.com/RonIT-401/worker-service/internal/app/processor"
	"github.com/RonIT-401/worker-service/internal/pkg/broker"
	"github.com/rs/zerolog/log"
)

type orderCreatedProc struct {
	h   ehandler.OrderCreated
	bus broker.Bus[entity.EventOrderCreated]
}

func NewOrderCreatedEventsCatcher(
	h ehandler.OrderCreated,
	bus broker.Bus[entity.EventOrderCreated],
) processor.Processor {
	return &orderCreatedProc{h, bus}
}

func (p *orderCreatedProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	if err := p.bus.Subscribe(ctx, wg, p.h.CallbackOrderCreated); err != nil {
		log.Fatal().Err(err).Msg("failed to subscribe to order created topic")
	}

	log.Info().
		Str("topic_name", p.bus.QueueName()).
		Msg("subscription to order created topic requested")
}
