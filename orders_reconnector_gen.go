// Code generated by genny. DO NOT EDIT.
// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package exchanges

import (
	"context"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// Should be one of three
// In case of first connection no reconnection event should be sent
type _ struct {
	DisconnectedWithErr error
	Reconnected         *struct{}
	Payload             *OrderEventPayload
}

// TODO: move to config
var defaultOrderEventReconnectOptions []retry.Option = []retry.Option{
	// 2 sec * (2^(3-1)-1) = 6 sec
	// M[U(0sec, 1sec)] * (3-1) = 1 sec / 2 * 2 = 1 sec
	retry.Attempts(3),
	retry.Delay(time.Second * 2),
	retry.DelayType(retry.BackOffDelay),
	retry.MaxJitter(time.Second),
	retry.LastErrorOnly(true),
}

type OrderEventReconnectorFn func(context.Context) (<-chan OrderEvent, error)
type OrderEventReconnector struct {
	connect          OrderEventReconnectorFn
	reconnectOptions []retry.Option // optional
	logger           *zap.Logger
}

func NewOrderEventReconnector(
	connect OrderEventReconnectorFn, reconnectOpts []retry.Option, l *zap.Logger,
) *OrderEventReconnector {
	return &OrderEventReconnector{
		connect:          connect,
		reconnectOptions: reconnectOpts,
		logger:           l.Named("OrderEventReconnector"),
	}
}

func (r *OrderEventReconnector) getReconnectOptions(
	ctx context.Context,
) []retry.Option {
	result := []retry.Option{retry.Context(ctx)}
	if r.reconnectOptions != nil {
		result = append(result, r.reconnectOptions...)
	} else {
		result = append(result, defaultOrderEventReconnectOptions...)
	}
	return result
}

func (r *OrderEventReconnector) chanShifter(in <-chan OrderEvent, out chan<- OrderEvent) {
	for ev := range in {
		out <- ev
	}
}

func (r *OrderEventReconnector) Watch(
	ctx context.Context,
) (<-chan OrderEvent, error) {
	in, err := r.connect(ctx)
	if err != nil {
		return nil, err
	}

	intermediate := make(chan OrderEvent, 100)
	out := make(chan OrderEvent, 100) // TODO: extract to config

	// Shifting required b/c we changing in channel
	go r.chanShifter(in, intermediate)

	reconnectAttempt := atomic.NewUint32(0)
	reconnect := func() bool {
		reconnectAttempt.Inc()
		serialAttempt := 1

		r.logger.Info("Reconnecting...", zap.Uint32("reconnectAttempt", reconnectAttempt.Load()))
		e := retry.Do(func() error {
			defer func() { serialAttempt++ }()
			lg := r.logger.With(
				zap.Uint32("attempt", reconnectAttempt.Load()),
				zap.Int("serialAttempt", serialAttempt))

			in, err := r.connect(ctx)
			if err == nil {
				lg.Info("Reconnected successfully")
				out <- OrderEvent{Reconnected: &struct{}{}}
				go r.chanShifter(in, intermediate)
				return nil
			}

			lg.Warn("Reconnect error", zap.Error(err))
			return err
		}, r.getReconnectOptions(ctx)...)
		return e == nil
	}

	go func() {
		defer close(out)
		for ev := range intermediate {
			switch {
			case ev.DisconnectedWithErr != nil:
				if !reconnect() {
					out <- OrderEvent{
						DisconnectedWithErr: errors.Wrap(ev.DisconnectedWithErr, "all reconnects failed"),
					}
					return
				}
			case ev.Reconnected != nil:
				out <- ev
			case ev.Payload != nil:
				out <- ev
			default:
				r.logger.Panic("unsupported event", zap.Any("ev", ev))
			}
		}
	}()

	return out, nil
}