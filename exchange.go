package exchanges

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
)

// type OrderExecutionType int

// // Defenition taken from Binance.
// const (
// 	UnknownOET                     = 0
// 	NewOET      OrderExecutionType = 1 // NEW - The order has been accepted into the engine.
// 	CanceledOET                    = 2 // CANCELED - The order has been canceled by the user.
// 	// REPLACED (currently unused at Binance)
// 	RejectedOET = 3 // REJECTED - The order has been rejected and was not processed. (This is never pushed into the User Data Stream)
// 	TradeOET    = 4 // TRADE - Part of the order or all of the order's quantity has filled.
// 	ExpiredOET  = 5 // EXPIRED - The order was canceled according to the order type's rules (e.g. LIMIT FOK orders with no fill, LIMIT IOC or MARKET orders that partially fill) or by the exchange, (e.g. orders canceled during liquidation, orders canceled during maintenance)
// )

type OrderStatusType int

// Defenition from Binance
const (
	UnknownOST         OrderStatusType = 0
	NewOST             OrderStatusType = 1 // NEW - The order has been accepted by the engine.
	PartiallyFilledOST OrderStatusType = 2 // PARTIALLY_FILLED - A part of the order has been filled.
	FilledOST          OrderStatusType = 3 // FILLED - The order has been completely filled.
	CanceledOST        OrderStatusType = 4 // CANCELED - The order has been canceled by the user.
	// PENDING_CANCEL (currently unused)
	RejectedOST OrderStatusType = 5 // REJECTED - The order was not accepted by the engine and not processed.
	ExpiredOST  OrderStatusType = 6 // EXPIRED - The order was canceled according to the order type's rules (e.g. LIMIT FOK orders with no fill, LIMIT IOC or MARKET orders that partially fill) or by the exchange, (e.g. orders canceled during liquidation, orders canceled during maintenance)
	OpenOST     OrderStatusType = 7 // OPEN status for FTX exchange
	ClosedOST   OrderStatusType = 8 // Closed status for FTX exchange
)

func (ost OrderStatusType) String() string {
	switch ost {
	case UnknownOST:
		return "Unknown"
	case NewOST:
		return "New"
	case PartiallyFilledOST:
		return "PartiallyFilled"
	case FilledOST:
		return "Filled"
	case CanceledOST:
		return "Canceled"
	case RejectedOST:
		return "Rejected"
	case ExpiredOST:
		return "Expired"
	case OpenOST:
		return "Open"
	case ClosedOST:
		return "Closed"
	default:
		return "(ERROR: unexpected status)"
	}
}

func (ost OrderStatusType) IsFinalStatus() bool {
	switch ost {
	case UnknownOST:
		return false
	case NewOST, PartiallyFilledOST:
		return false
	case FilledOST, CanceledOST, RejectedOST, ExpiredOST:
		return true
	default:
		log.Printf("ERROR: invalid status = %v", ost)
		return false
	}
}

var OrderExecutedError = errors.New("order was executed")
var OrderNotFoundError = errors.New("order not found")
var NewOrderRejectedError = errors.New("new order was rejected")

type OrderEventPayload struct {
	OrderID     string
	OrderStatus OrderStatusType
	Symbol      *string // Now it is optional. But it's better to fill it if possible
}

func (oep *OrderEventPayload) String() string {
	symbol := ""
	if oep.Symbol != nil {
		symbol = ", Symbol: " + *oep.Symbol
	}
	return fmt.Sprintf("{OrderID: %s, OrderStatus: %v%s",
		oep.OrderID, oep.OrderStatus, symbol)
}

// Should be one of three
// In case of first connection no reconnection event should be sent
type OrderEvent struct {
	DisconnectedWithErr error
	Reconnected         *struct{}
	Payload             *OrderEventPayload
}

func (ev OrderEvent) String() string {
	switch {
	case ev.DisconnectedWithErr != nil:
		return fmt.Sprintf("{DisconnectedWithError = (%v)}", ev.DisconnectedWithErr)
	case ev.Reconnected != nil:
		return "{Reconnected}"
	case ev.Payload != nil:
		return fmt.Sprintf("{Payload = %v}", ev.Payload)
	}
	return "(ERROR: invalid state)"
}

// Should be one of three
// In case of first connection no reconnection event should be sent
type PriceEvent struct {
	DisconnectedWithErr error
	Reconnected         *struct{}
	Payload             *apd.Decimal
}

func (ev PriceEvent) String() string {
	switch {
	case ev.DisconnectedWithErr != nil:
		return fmt.Sprintf("{DisconnectedWithError = (%v)}", ev.DisconnectedWithErr)
	case ev.Reconnected != nil:
		return "{Reconnected}"
	case ev.Payload != nil:
		return fmt.Sprintf("{Payload = %v}", ev.Payload)
	}
	return "(ERROR: invalid state)"
}

type OrderInfo struct {
	ID            string
	ClientOrderID *string // optional
	Status        OrderStatusType
}

func (oi OrderInfo) String() string {
	result := fmt.Sprintf("{ID='%s', Status='%v', ClientOrderID=", oi.ID, oi.Status)
	if oi.ClientOrderID != nil {
		result = result + "'" + *oi.ClientOrderID + "'}"
	} else {
		result += "nil}"
	}
	return result
}

type SymbolInfo struct {
	DisplayName    string // "BN-LTCBTC (LTC margin)"
	OriginalSymbol string // "LTC_BTC"
	Symbol         string // "BN-LTCBTC"
	Filters        []map[string]interface{}
}

type OrderType string

const (
	LIMIT             OrderType = "LIMIT"
	MARKET            OrderType = "MARKET"
	LIMIT_MAKER       OrderType = "LIMIT_MAKER"
	STOP_LOSS         OrderType = "STOP_LOSS"
	STOP_LOSS_LIMIT   OrderType = "STOP_LOSS_LIMIT"
	TAKE_PROFIT       OrderType = "TAKE_PROFIT"
	TAKE_PROFIT_LIMIT OrderType = "TAKE_PROFIT_LIMIT"
	MARKET_IF_TOUCHED OrderType = "MARKET_IF_TOUCHED"
	LIMIT_IF_TOUCHED  OrderType = "LIMIT_IF_TOUCHED"
	PEGGED            OrderType = "PEGGED"
)

type OrderSide string

const (
	BUY                OrderSide = "BUY"
	SELL               OrderSide = "SELL"
	UNKNOWN_ORDER_SIDE OrderSide = "UNKNOWN_ORDER_SIDE"
)

type OrderTimeInForce string

const (
	GTC_TIME_IN_FORCE                 OrderTimeInForce = "GTC"
	IOC_TIME_IN_FORCE                 OrderTimeInForce = "IOC"
	FOK_TIME_IN_FORCE                 OrderTimeInForce = "FOK"
	DAY_TIME_IN_FORCE                 OrderTimeInForce = "DAY"
	GOOD_TILL_CANCEL_TIME_IN_FORCE    OrderTimeInForce = "GOOD_TILL_CANCEL"
	IMMEDIATE_OR_CANCEL_TIME_IN_FORCE OrderTimeInForce = "IMMEDIATE_OR_CANCEL"
	FILL_OR_KILL_TIME_IN_FORCE        OrderTimeInForce = "FILL_OR_KILL"
)

type OrderDetailInfo struct {
	Symbol        string
	ID            string
	ClientOrderID *string
	Price         *apd.Decimal
	Quantity      *apd.Decimal
	ExecutedQty   *apd.Decimal
	Status        OrderStatusType
	OrderType     *OrderType
	Time          int64
	OrderSide     OrderSide
	TimeInForce   *OrderTimeInForce
	StopPrice     *apd.Decimal
	QuoteQuantity *apd.Decimal
}

type AccountPosition struct {
	Symbol           string
	UnrealizedProfit *apd.Decimal
	Leverage         *apd.Decimal
	EntryPrice       *apd.Decimal
	Size             *apd.Decimal
	MarkPrice        *apd.Decimal
	PositionValue    *apd.Decimal
	Side             string
	CumRealisedPnl   *apd.Decimal
	LiqPrice         *apd.Decimal
	Category         string
}

type AccountBalance struct {
	Coin   string
	Free   *apd.Decimal
	Locked *apd.Decimal
}

type Account struct {
	AccountBalances  []AccountBalance
	AccountPositions []AccountPosition
}

type OrderFilter struct {
	Symbol        *string
	OrderID       *string
	ClientOrderID *string
}

// `id` of orders placing result should be consistent (accept/return) with other methods.
//
// The Exchange is responsible for all tryings to reconnect to WebSockets,
// the implementation should try the best to make connection reliable, but if
// for several retries it can't connect it should send error without closing the channel.
// Call an Unwatch method is not required in this case.
type Exchange interface {
	GetPrefix() string
	GetName() string

	RoundPrice(_ context.Context, symbol string, price *apd.Decimal, tickSize *string) (*apd.Decimal, error)
	RoundQuantity(_ context.Context, symbol string, qty *apd.Decimal) (*apd.Decimal, error)

	// This method should use `clientOrderID` if it's possible
	PlaceBuyOrder(
		_ context.Context, isRetry bool, symbol string, price, qty *apd.Decimal, clientOrderID string,
	) (id string, e error)

	// This method should use `clientOrderID` if it's possible
	PlaceSellOrder(
		_ context.Context, isRetry bool, symbol string, price, qty *apd.Decimal, clientOrderID string,
	) (id string, e error)

	// can return `OrderExecutedError` in case of executed order.
	// can return `OrderNotFoundError` in case of not found order.
	// It's important to have Canceled state if no error returned
	// !Always release order after cancellation
	CancelOrder(_ context.Context, symbol, id string) error

	// Order should be released after cancellation
	ReleaseOrder(_ context.Context, symbol, id string) error

	GetPrice(_ context.Context, symbol string) (*apd.Decimal, error)

	// `createdAt` can be omitted. But some exchanges can require it
	GetOrderInfo(_ context.Context, symbol, id string, createdAt *time.Time) (OrderInfo, error)
	GetOrderInfoByClientOrderID(_ context.Context, symbol, clientOrderID string, createdAt *time.Time) (OrderInfo, error)

	GetTradableSymbols(context.Context) ([]SymbolInfo, error)

	// Returns control immediately
	// If error is sent then channel will be closed automatically
	// Channel will be closed in two ways: by context and by disconnection
	WatchOrdersStatuses(context.Context) (<-chan OrderEvent, error)

	// Returns control immediately
	// If error is sent then channel will be closed automatically
	// Channel will be closed in two ways: by context and by disconnection
	WatchSymbolPrice(_ context.Context, symbol string) (<-chan PriceEvent, error)

	PlaceBuyOrderV2(
		_ context.Context, isRetry bool, symbol string, price, qty *apd.Decimal, clientOrderID string, orderType string,
	) (id string, e error)

	PlaceSellOrderV2(
		_ context.Context, isRetry bool, symbol string, price, qty *apd.Decimal, clientOrderID string, orderType string,
	) (id string, e error)

	WatchAccountPositions(context.Context) (<-chan PositionEvent, error)

	GetOpenOrders(context.Context) ([]OrderDetailInfo, error)

	GetOrders(ctx context.Context, filter OrderFilter) ([]OrderDetailInfo, error)

	GetAccount(context.Context) (Account, error)

	GenerateClientOrderID(ctx context.Context, identifierID string) (string, error)
}

type BulkCancelResult struct {
	ID  string
	Err error
}

type BulkCancelExchange interface {
	// can return `OrderExecutedError` in case of executed order.
	// can return `OrderNotFoundError` in case of not found order.
	BulkCancelOrder(symbol string, ids []string) ([]BulkCancelResult, error)
}

type PositionEvent struct {
	DisconnectedWithErr error
	Reconnected         *struct{}
	Payload             []*PositionPayload
}

type PositionPayload struct {
	Value  *apd.Decimal
	Symbol string
}
