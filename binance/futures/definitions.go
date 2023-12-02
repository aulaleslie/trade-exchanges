package futures

import exchanges "github.com/aulaleslie/trade-exchanges"

var orderStatusTypeMap map[string]exchanges.OrderStatusType = map[string]exchanges.OrderStatusType{
	"NEW":              exchanges.NewOST,             // NEW - The order has been accepted by the engine.
	"PARTIALLY_FILLED": exchanges.PartiallyFilledOST, // PARTIALLY_FILLED - A part of the order has been filled.
	"FILLED":           exchanges.FilledOST,          // FILLED - The order has been completely filled.
	"CANCELED":         exchanges.CanceledOST,        // CANCELED - The order has been canceled by the user.
	"REJECTED":         exchanges.RejectedOST,        // REJECTED - The order was not accepted by the engine and not processed.
	"EXPIRED":          exchanges.ExpiredOST,         // EXPIRED - The order was canceled according to the order type's rules (e.g. LIMIT FOK orders with no fill, LIMIT IOC or MARKET orders that partially fill) or by the exchange, (e.g. orders canceled during liquidation, orders canceled during maintenance)
}

func mapOrderStatusType(orderStatus string) exchanges.OrderStatusType {
	result, ok := orderStatusTypeMap[orderStatus]
	if ok {
		return result
	} else {
		return exchanges.UnknownOST
	}
}

// Mapping OrderType
var orderTypeMap map[string]exchanges.OrderType = map[string]exchanges.OrderType{
	"LIMIT":             exchanges.LIMIT,
	"MARKET":            exchanges.MARKET,
	"LIMIT_MAKER":       exchanges.LIMIT_MAKER,
	"STOP_LOSS":         exchanges.STOP_LOSS,
	"STOP_LOSS_LIMIT":   exchanges.STOP_LOSS_LIMIT,
	"TAKE_PROFIT":       exchanges.TAKE_PROFIT,
	"TAKE_PROFIT_LIMIT": exchanges.TAKE_PROFIT_LIMIT,
}

func mapOrderType(orderType string) *exchanges.OrderType {
	result, ok := orderTypeMap[orderType]
	if ok {
		return &result
	}
	return nil
}

// Mapping OrderSide
var orderSideMap map[string]exchanges.OrderSide = map[string]exchanges.OrderSide{
	"BUY":  exchanges.BUY,
	"SELL": exchanges.SELL,
}

func mapOrderSide(orderSide string) exchanges.OrderSide {
	result, ok := orderSideMap[orderSide]
	if ok {
		return result
	}
	return exchanges.UNKNOWN_ORDER_SIDE
}

// Mapping TimeInForce
var orderTimeInForceMap map[string]exchanges.OrderTimeInForce = map[string]exchanges.OrderTimeInForce{
	"GTC": exchanges.GTC_TIME_IN_FORCE,
	"IOC": exchanges.IOC_TIME_IN_FORCE,
	"FOK": exchanges.FOK_TIME_IN_FORCE,
}

func mapOrderTimeInForce(orderInForce string) *exchanges.OrderTimeInForce {
	result, ok := orderTimeInForceMap[orderInForce]
	if ok {
		return &result
	}
	return nil
}
