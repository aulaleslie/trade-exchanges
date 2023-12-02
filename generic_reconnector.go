package exchanges

import (
	"context"
	"time"

	"github.com/avast/retry-go/v3"
	"github.com/cheekybits/genny/generic"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type PayloadType generic.Type

// Should be one of three
// In case of first connection no reconnection event should be sent
type TypeEvent struct {
	DisconnectedWithErr error
	Reconnected         *struct{}
	Payload             *PayloadType
}

// TODO: move to config
var defaultTypeEventReconnectOptions []retry.Option = []retry.Option{
	// 2 sec * (2^(3-1)-1) = 6 sec
	// M[U(0sec, 1sec)] * (3-1) = 1 sec / 2 * 2 = 1 sec
	retry.Attempts(3),
	retry.Delay(time.Second * 2),
	retry.DelayType(retry.BackOffDelay),
	retry.MaxJitter(time.Second),
	retry.LastErrorOnly(true),
}

type TypeEventReconnectorFn func(context.Context) (<-chan TypeEvent, error)
type TypeEventReconnector struct {
	connect          TypeEventReconnectorFn
	reconnectOptions []retry.Option // optional
	logger           *zap.Logger
}

func NewTypeEventReconnector(
	connect TypeEventReconnectorFn, reconnectOpts []retry.Option, l *zap.Logger,
) *TypeEventReconnector {
	return &TypeEventReconnector{
		connect:          connect,
		reconnectOptions: reconnectOpts,
		logger:           l.Named("TypeEventReconnector"),
	}
}

func (r *TypeEventReconnector) getReconnectOptions(
	ctx context.Context,
) []retry.Option {
	result := []retry.Option{retry.Context(ctx)}
	if r.reconnectOptions != nil {
		result = append(result, r.reconnectOptions...)
	} else {
		result = append(result, defaultTypeEventReconnectOptions...)
	}
	return result
}

func (r *TypeEventReconnector) chanShifter(in <-chan TypeEvent, out chan<- TypeEvent) {
	for ev := range in {
		out <- ev
	}
}

func (r *TypeEventReconnector) Watch(
	ctx context.Context,
) (<-chan TypeEvent, error) {
	in, err := r.connect(ctx)
	if err != nil {
		return nil, err
	}

	intermediate := make(chan TypeEvent, 100)
	out := make(chan TypeEvent, 100) // TODO: extract to config

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
				out <- TypeEvent{Reconnected: &struct{}{}}
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
					out <- TypeEvent{
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
