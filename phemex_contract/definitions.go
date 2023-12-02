package phemex_contract

import exchanges "github.com/aulaleslie/trade-exchanges"

// Mapping OrderType
var orderTypeMap map[string]exchanges.OrderType = map[string]exchanges.OrderType{
	"Limit":           exchanges.LIMIT,
	"Market":          exchanges.MARKET,
	"Stop":            exchanges.STOP_LOSS,
	"StopLimit":       exchanges.STOP_LOSS_LIMIT,
	"MarketIfTouched": exchanges.MARKET_IF_TOUCHED,
	"LimitIfTouched":  exchanges.LIMIT_IF_TOUCHED,
	"Pegged":          exchanges.PEGGED,
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
	"Buy":  exchanges.BUY,
	"Sell": exchanges.SELL,
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
	"Day":               exchanges.DAY_TIME_IN_FORCE,
	"GoodTillCancel":    exchanges.GOOD_TILL_CANCEL_TIME_IN_FORCE,
	"ImmediateOrCancel": exchanges.IMMEDIATE_OR_CANCEL_TIME_IN_FORCE,
	"FillOrKill":        exchanges.FILL_OR_KILL_TIME_IN_FORCE,
}

func mapOrderTimeInForce(orderInForce string) *exchanges.OrderTimeInForce {
	result, ok := orderTimeInForceMap[orderInForce]
	if ok {
		return &result
	}
	return nil
}
