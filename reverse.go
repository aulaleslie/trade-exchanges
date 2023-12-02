package exchanges

import (
	"context"
	"time"

	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"go.uber.org/zap"
)

func RevPrice(p *apd.Decimal) *apd.Decimal {
	return utils.Div(utils.One, p)
}

type ReversedExchange struct {
	target Exchange
	lgs    *zap.SugaredLogger
}

func NewReversedExchange(target Exchange, lg *zap.Logger) *ReversedExchange {
	return &ReversedExchange{
		target: target,
		lgs:    lg.Named("ReversedExchange").Sugar(),
	}
}

// var _ Exchange = (*ReversedExchange)(nil)

func (re *ReversedExchange) PlaceBuyOrder(
	ctx context.Context, isRetry bool, symbol string, price, quantity *apd.Decimal, prefferedID string,
) (id string, e error) {
	rev := RevPrice(price)
	re.lgs.Debugf("Forward = %v; Reverse = %v", price, rev)
	return re.target.PlaceSellOrder(ctx, isRetry, symbol, rev, quantity, prefferedID)
}

func (re *ReversedExchange) PlaceSellOrder(
	ctx context.Context, isRetry bool, symbol string, price, quantity *apd.Decimal, prefferedID string,
) (id string, e error) {
	rev := RevPrice(price)
	re.lgs.Debugf("Forward = %v; Reverse = %v", price, rev)
	return re.target.PlaceBuyOrder(ctx, isRetry, symbol, rev, quantity, prefferedID)
}
func (re *ReversedExchange) CancelOrder(ctx context.Context, symbol, id string) error {
	return re.target.CancelOrder(ctx, symbol, id)
}
func (re *ReversedExchange) GetPrice(ctx context.Context, symbol string) (*apd.Decimal, error) {
	forward, err := re.target.GetPrice(ctx, symbol)
	if err != nil {
		return nil, err
	}
	rev := RevPrice(forward)
	re.lgs.Debugf("Forward = %v; Reverse = %v", forward, rev)
	return rev, nil
}

func (re *ReversedExchange) GetOrderInfo(ctx context.Context, symbol, id string, createdAt *time.Time) (OrderInfo, error) {
	return re.target.GetOrderInfo(ctx, symbol, id, createdAt)
}

func (re *ReversedExchange) GetOrderInfoByClientOrderID(
	ctx context.Context, symbol, clientOrderID string, createdAt *time.Time,
) (OrderInfo, error) {
	return re.target.GetOrderInfoByClientOrderID(ctx, symbol, clientOrderID, createdAt)
}

func (re *ReversedExchange) GetTradableSymbols(ctx context.Context) ([]SymbolInfo, error) {
	return re.target.GetTradableSymbols(ctx)
}

func (re *ReversedExchange) WatchOrdersStatuses(
	ctx context.Context,
) (<-chan OrderEvent, error) {
	return re.target.WatchOrdersStatuses(ctx)
}

func (re *ReversedExchange) WatchSymbolPrice(
	ctx context.Context, symbol string,
) (<-chan PriceEvent, error) {
	in, err := re.target.WatchSymbolPrice(ctx, symbol)
	if err != nil {
		return nil, err
	}

	out := make(chan PriceEvent, 100)
	go func() {
		for ev := range in {
			switch {
			case ev.Reconnected != nil:
				out <- ev
			case ev.DisconnectedWithErr != nil:
				out <- ev
				close(out)
			case ev.Payload != nil:
				rev := RevPrice(ev.Payload)
				re.lgs.Debugf("Forward = %v; Reverse = %v", ev.Payload, rev)
				out <- PriceEvent{Payload: rev}
			}
		}
	}()

	return out, nil
}
