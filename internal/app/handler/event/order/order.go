package eorder

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RonIT-401/worker-service/internal/app/entity"
	ehandler "github.com/RonIT-401/worker-service/internal/app/handler/event"
	"github.com/RonIT-401/worker-service/internal/app/service"
	"github.com/RonIT-401/worker-service/internal/pkg/broker"
	"github.com/rs/zerolog/log"
)

type handler struct {
	delivery service.Delivery
	bus      broker.Bus[entity.EventOrderDeliveryCalculated]
}

func NewHandler(
	delivery service.Delivery,
	bus broker.Bus[entity.EventOrderDeliveryCalculated],
) ehandler.OrderCreated {
	return &handler{delivery: delivery, bus: bus}
}

func (h *handler) CallbackOrderCreated(
	ctx context.Context, ev *entity.EventOrderCreated, _ []broker.Header,
) error {
	log.Info().
		Ctx(ctx).
		EmbedObject(ev).
		Msg("Получено событие order.created")

	if ev.OrderGUID == "" {
		return broker.NotCriticalError(fmt.Errorf("order.created: empty order_guid"))
	}

	deliveryPrice, err := h.delivery.Calculate(ctx, ev.Currency)
	if err != nil {
		if errors.Is(err, entity.ErrFixerCurrencyNotFound) {
			return broker.NotCriticalError(
				fmt.Errorf("order.created: unsupported currency %s: %w", ev.Currency, err),
			)
		}

		return fmt.Errorf("order.created: calculate delivery price: %w", err)
	}

	out := &entity.EventOrderDeliveryCalculated{
		OrderGUID:     ev.OrderGUID,
		Currency:      ev.Currency,
		DeliveryPrice: deliveryPrice,
		CalculatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	if err := h.bus.Send(
		ctx,
		out,
		entity.BrokerHeaderOrderDeliveryCalculatedType(),
		entity.BrokerHeaderOrderDeliveryCalculatedEventID(),
	); err != nil {
		return fmt.Errorf("order.created: send delivery calculated event: %w", err)
	}

	log.Info().
		Ctx(ctx).
		EmbedObject(out).
		Msg("Опубликовано событие order.delivery.calculated")

	return nil
}
