package bybit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"time"

	exchanges "github.com/aulaleslie/trade-exchanges"
)

const (
	GetOrderHistoryPath = "/v5/order/history"
	BybitBaseURL        = "https://api.bybit.com"
)

var recWindow = "5000"

var priceFloorRE = regexp.MustCompile(`^[0-9]{1,20}(\.[0-9]{1,6})?`)

var qtyFloorRE = regexp.MustCompile(`^[0-9]{1,20}(\.[0-9]{1,5})?`)

type BybitResponse struct {
	RetCode int            `json:"retCode"`
	RetMsg  string         `json:"retMsg"`
	Result  ResultResponse `json:"result"`
}
type ResultResponse struct {
	NextPageCursor string                 `json:"nextPageCursor"`
	Category       string                 `json:"category"`
	List           []OrderHistoryResponse `json:"list"`
}
type OrderHistoryResponse struct {
	OrderId     string `json:"orderId"`
	OrderLinkId string `json:"orderLinkId"`
	Symbol      string `json:"symbol"`
	Side        string `json:"side"`
	OrderStatus string `json:"orderStatus"`
	CreatedTime string `json:"createdTime"`
	UpdatedTime string `json:"updatedTime"`
}

// type orderFields struct {
// 	ClOrdID  string
// 	OrderQty float64
// 	PriceEp  int64
// 	Symbol   string
// }

type SymbolFilter struct {
	LotSizeFilter struct {
		BasePrecision  string `json:"basePrecision"`
		QuotePrecision string `json:"quotePrecision"`
		MaxOrderQty    string `json:"maxOrderQty"`
		MinOrderQty    string `json:"minOrderQty"`
		MinOrderAmt    string `json:"minOrderAmt"`
		MaxOrderAmt    string `json:"maxOrderAmt"`
	} `json:"lotSizeFilter"`
	PriceFilter struct {
		TickSize string `json:"tickSize"`
	} `json:"priceFilter"`
}

type SymbolInverseFilter struct {
	LotSizeFilter struct {
		MaxOrderQty         string `json:"maxOrderQty"`
		MinOrderQty         string `json:"minOrderQty"`
		QtyStep             string `json:"qtyStep"`
		PostOnlyMaxOrderQty string `json:"postOnlyMaxOrderQty"`
	} `json:"lotSizeFilter"`
	PriceFilter struct {
		MinPrice string `json:"minPrice"`
		MaxPrice string `json:"maxPrice"`
		TickSize string `json:"tickSize"`
	} `json:"priceFilter"`
}

// Mapping OrderStatusType
var orderStatusTypeMap map[string]exchanges.OrderStatusType = map[string]exchanges.OrderStatusType{
	"New":             exchanges.NewOST,             // NEW - The order has been accepted by the engine.
	"PartiallyFilled": exchanges.PartiallyFilledOST, // PARTIALLY_FILLED - A part of the order has been filled.
	"Filled":          exchanges.FilledOST,          // FILLED - The order has been completely filled.
	"Cancelled":       exchanges.CanceledOST,        // CANCELED - The order has been canceled by the user.
	// PENDING_CANCEL (currently unused)
	"Rejected": exchanges.RejectedOST, // REJECTED - The order was not accepted by the engine and not processed.
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
	"Limit":  exchanges.LIMIT,
	"Market": exchanges.MARKET,
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

func bybitSignatureGenerator(apikey, apisecret, query string) (string, int64) {
	now := time.Now()
	unixNano := now.UnixNano()
	timeStamp := unixNano / 1000000

	// Generate bybit signature
	hmac256 := hmac.New(sha256.New, []byte(apisecret))
	hmac256.Write([]byte(strconv.FormatInt(timeStamp, 10) + apikey + recWindow + query))
	signature := hex.EncodeToString(hmac256.Sum(nil))
	return signature, timeStamp
}
