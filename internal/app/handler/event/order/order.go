package eorder

import (
	"context"
	"fmt"

	"github.com/RonIT-401/worker-service/internal/app/entity"
	ehandler "github.com/RonIT-401/worker-service/internal/app/handler/event"
	"github.com/RonIT-401/worker-service/internal/pkg/broker"
	"github.com/rs/zerolog/log"
)

type handler struct{}

func NewHandler() ehandler.OrderCreated {
	return &handler{}
}

func (h *handler) CallbackOrderCreated(
	ctx context.Context, ev *entity.EventOrderCreated, _ []broker.Header,
) error {
	log.Info().Ctx(ctx).EmbedObject(ev).Msg("Получено событие order.created")

	if ev.OrderGUID == "" {
		return broker.NotCriticalError(
			fmt.Errorf("order_guid is empty"),
		)
	}

	return nil
}
