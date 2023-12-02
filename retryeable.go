package exchanges

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go/v3"
	"github.com/cockroachdb/apd"
	"go.uber.org/zap"
)

// TODO: move to config
var defaultRetryOptions []retry.Option = []retry.Option{
	// 200 ms * (2^(3-1)-1) = 600 ms
	// M[U(0sec, 0.33sec)] * (3-1) = 0.33 sec / 2 * 2 = 0.33 sec
	retry.Attempts(3),
	retry.Delay(time.Millisecond * 200),
	retry.DelayType(retry.BackOffDelay),
	retry.MaxJitter(time.Second / 3),
	retry.LastErrorOnly(true),
}

// TODO: move to config
var cancelOrderDefaultRetryOptions []retry.Option = []retry.Option{
	// 200 ms*(2^(5-1)-1) = 3 sec
	// M[U(0sec, 1sec)] * (5-1) = 0.5 sec * 4 = 2 sec
	retry.Attempts(5),
	retry.Delay(time.Millisecond * 200),
	retry.DelayType(retry.BackOffDelay),
	retry.MaxJitter(time.Second),
	retry.LastErrorOnly(true),
}

// Timeline: [] 1<<0 [] 1<<1 [] 1<<2 []
// attempt = 1 => totalDelay = 0ms                     = delay*(2^0 - 1)
// attempt = 2 => totalDelay = delay*(2^0)             = delay*(2^1 - 1)
// attempt = 3 => totalDelay = delay*(2^0 + 2^1)       = delay*(2^2 - 1)
// attempt = 4 => totalDelay = delay*(2^0 + 2^1 + 2^2) = delay*(2^3 - 1)
// attempt = n => totalDelay =                         = delay*(2^(n-1) - 1)

// All methods are retryeable except of watch/unwatch methods
// TODO: all methods shouldn't think result of `NotFound` error (or similar erorrs) should be retryed
type RetryeableExchange struct {
	Target             Exchange
	RetryOptions       []retry.Option // optional
	CancelRetryOptions []retry.Option // optional
	Logger             *zap.Logger
}

var _ Exchange = (*RetryeableExchange)(nil)

func (re *RetryeableExchange) immutablyAddContext(opts []retry.Option, ctx context.Context) []retry.Option {
	result := []retry.Option{}
	result = append(result, opts...)
	result = append(result, retry.Context(ctx))
	return result
}

func (re *RetryeableExchange) getRetryOptions(ctx context.Context) []retry.Option {
	if re.RetryOptions != nil {
		return re.immutablyAddContext(re.RetryOptions, ctx)
	}
	return re.immutablyAddContext(defaultRetryOptions, ctx)
}

func (re *RetryeableExchange) getCancelRetryOptions(ctx context.Context) []retry.Option {
	if re.CancelRetryOptions != nil {
		return re.immutablyAddContext(re.CancelRetryOptions, ctx)
	}
	return re.immutablyAddContext(cancelOrderDefaultRetryOptions, ctx)
}

func (re *RetryeableExchange) GetPrefix() string {
	return re.Target.GetPrefix()
}

func (re *RetryeableExchange) GetName() string {
	return re.Target.GetName()
}

func (re *RetryeableExchange) RoundPrice(ctx context.Context, symbol string, price *apd.Decimal, tickSize *string) (result *apd.Decimal, e error) {
	e = retry.Do(
		func() error {
			var err error
			result, err = re.Target.RoundPrice(ctx, symbol, price, tickSize)
			return err
		},
		re.getRetryOptions(ctx)...,
	)
	return result, e
}

func (re *RetryeableExchange) RoundQuantity(ctx context.Context, symbol string, quantity *apd.Decimal) (result *apd.Decimal, e error) {
	e = retry.Do(
		func() error {
			var err error
			result, err = re.Target.RoundQuantity(ctx, symbol, quantity)
			return err
		},
		re.getRetryOptions(ctx)...,
	)
	return result, e
}

func (re *RetryeableExchange) PlaceBuyOrder(ctx context.Context,
	isRetry bool, symbol string, price, quantity *apd.Decimal, prefferedID string) (id string, e error) {

	e = retry.Do(
		func() error {
			var err error
			id, err = re.Target.PlaceBuyOrder(ctx, isRetry, symbol, price, quantity, prefferedID)
			isRetry = true
			return err
		},
		re.getRetryOptions(ctx)...,
	)
	return id, e
}

func (re *RetryeableExchange) PlaceSellOrder(ctx context.Context,
	isRetry bool, symbol string, price, quantity *apd.Decimal, prefferedID string,
) (id string, e error) {
	e = retry.Do(
		func() error {
			var err error
			id, err = re.Target.PlaceSellOrder(ctx, isRetry, symbol, price, quantity, prefferedID)
			isRetry = true
			return err
		},
		re.getRetryOptions(ctx)...,
	)
	return id, e
}

func (re *RetryeableExchange) CancelOrder(ctx context.Context, symbol, id string) error {
	opts := []retry.Option{
		retry.RetryIf(func(err error) bool {
			return !errors.Is(err, OrderExecutedError)
		}),
	}
	opts = append(opts, re.getCancelRetryOptions(ctx)...)
	return retry.Do(
		func() error {
			return re.Target.CancelOrder(ctx, symbol, id)
		},
		opts...,
	)
}

func (re *RetryeableExchange) ReleaseOrder(ctx context.Context, symbol, id string) error {
	return retry.Do(
		func() error {
			return re.Target.ReleaseOrder(ctx, symbol, id)
		},
		re.getRetryOptions(ctx)...,
	)
}

func (re *RetryeableExchange) GetPrice(ctx context.Context, symbol string) (price *apd.Decimal, e error) {
	e = retry.Do(
		func() error {
			var err error
			price, err = re.Target.GetPrice(ctx, symbol)
			return err
		},
		re.getRetryOptions(ctx)...,
	)
	return price, e
}

func (re *RetryeableExchange) GetOrderInfo(ctx context.Context, symbol, id string, createdAt *time.Time) (info OrderInfo, e error) {
	e = retry.Do(
		func() error {
			var err error
			info, err = re.Target.GetOrderInfo(ctx, symbol, id, createdAt)
			return err
		},
		re.getRetryOptions(ctx)...,
	)
	return info, e
}

func (re *RetryeableExchange) GetOpenOrders(ctx context.Context) (res []OrderDetailInfo, e error) {
	e = retry.Do(
		func() error {
			var err error
			res, err = re.Target.GetOpenOrders(ctx)
			return err
		},
		re.getRetryOptions(ctx)...,
	)
	return res, e
}

func (re *RetryeableExchange) GetOrders(ctx context.Context, filter OrderFilter) (res []OrderDetailInfo, e error) {
	e = retry.Do(
		func() error {
			var err error
			res, err = re.Target.GetOrders(ctx, filter)
			return err
		},
		re.getRetryOptions(ctx)...,
	)
	return res, e
}

func (re *RetryeableExchange) GetAccount(ctx context.Context) (res Account, e error) {
	e = retry.Do(
		func() error {
			var err error
			res, err = re.Target.GetAccount(ctx)
			return err
		},
		re.getRetryOptions(ctx)...,
	)
	return res, e
}

func (re *RetryeableExchange) GetOrderInfoByClientOrderID(ctx context.Context, symbol, clientOrderID string, createdAt *time.Time) (info OrderInfo, e error) {
	e = retry.Do(
		func() error {
			var err error
			info, err = re.Target.GetOrderInfoByClientOrderID(
				ctx, symbol, clientOrderID, createdAt)
			return err
		},
		re.getRetryOptions(ctx)...,
	)
	return info, e
}

func (re *RetryeableExchange) GetTradableSymbols(ctx context.Context) (info []SymbolInfo, e error) {
	e = retry.Do(
		func() error {
			var err error
			info, err = re.Target.GetTradableSymbols(ctx)
			return err
		},
		re.getRetryOptions(ctx)...,
	)
	return info, e
}

func (re *RetryeableExchange) WatchOrdersStatuses(ctx context.Context) (<-chan OrderEvent, error) {
	oer := NewOrderEventReconnector(re.Target.WatchOrdersStatuses, nil, re.Logger)
	return oer.Watch(ctx)
}

func (re *RetryeableExchange) WatchSymbolPrice(ctx context.Context, symbol string) (<-chan PriceEvent, error) {
	fn := func(ctx context.Context) (<-chan PriceEvent, error) {
		return re.Target.WatchSymbolPrice(ctx, symbol)
	}
	per := NewPriceEventReconnector(fn, nil, re.Logger)
	return per.Watch(ctx)
}

func (re *RetryeableExchange) PlaceBuyOrderV2(ctx context.Context, isRetry bool, symbol string, price, qty *apd.Decimal, clientOrderID string, orderType string) (id string, e error) {
	e = retry.Do(
		func() error {
			var err error
			id, err = re.Target.PlaceBuyOrderV2(ctx, isRetry, symbol, price, qty, clientOrderID, orderType)
			isRetry = true
			return err
		},
		re.getRetryOptions(ctx)...,
	)
	return id, e
}

func (re *RetryeableExchange) PlaceSellOrderV2(ctx context.Context, isRetry bool, symbol string, price, qty *apd.Decimal, clientOrderID string, orderType string) (id string, e error) {
	e = retry.Do(
		func() error {
			var err error
			id, err = re.Target.PlaceSellOrderV2(ctx, isRetry, symbol, price, qty, clientOrderID, orderType)
			isRetry = true
			return err
		},
		re.getRetryOptions(ctx)...,
	)
	return id, e
}

func (re *RetryeableExchange) WatchAccountPositions(ctx context.Context) (<-chan PositionEvent, error) {
	// TODO make it wrap by reconnector
	return re.Target.WatchAccountPositions(ctx)
}

func (re *RetryeableExchange) GenerateClientOrderID(ctx context.Context, identifierID string) (string, error) {
	// TODO make it wrap by reconnector
	return re.Target.GenerateClientOrderID(ctx, identifierID)
}
