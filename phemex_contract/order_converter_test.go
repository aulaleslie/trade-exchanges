package phemex_contract

import (
	"encoding/json"
	"testing"

	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/phemex_contract/krisa_phemex_fork"
	"github.com/stretchr/testify/assert"
)

func TestConvertOrder(t *testing.T) {
	jsonOrder := `{
		"bizError": 11032,
		"orderID": "11112222-3333-4444-5555-666677778888",
		"clOrdID": "cl_ord_id",
		"symbol": "BTCUSD", "side": "Sell", "orderType": "Limit",
		"actionTimeNs": 1627000000000000000, "priceEp": 334800000, "price": null,
		"orderQty": 2, "displayQty": 0, "timeInForce": "GoodTillCancel",
		"reduceOnly": false, "takeProfitEp": 0, "takeProfit": null,
		"stopLossEp": 0, "closedPnlEv": 0, "closedPnl": null, "closedSize": 0,
		"cumQty": 0, "cumValueEv": 0, "cumValue": null, "leavesQty": 0,
		"leavesValueEv": 0, "leavesValue": null, "stopLoss": null,
		"stopDirection": "UNSPECIFIED", "ordStatus": "Rejected",
		"transactTimeNs": 1627000000000000000, "platform": 5
	}`
	order := &krisa_phemex_fork.OrderResponse{}
	err := json.Unmarshal([]byte(jsonOrder), &order)
	assert.NoError(t, err)

	oi, err := convertOrder(order)
	assert.NoError(t, err)

	assert.Equal(t,
		exchanges.OrderInfo{
			ID:            "11112222-3333-4444-5555-666677778888",
			ClientOrderID: &([]string{"cl_ord_id"}[0]),
			Status:        exchanges.RejectedOST,
		},
		oi)
}
