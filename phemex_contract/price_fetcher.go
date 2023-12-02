package phemex_contract

import (
	"context"
	"encoding/json"

	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
)

type PriceResponseWrapper struct {
	Error  json.RawMessage `json:"error"`  // "error": null,
	Id     *int64          `json:"id"`     // "id": 0,
	Result *PriceResponse  `json:"result"` // "result": {}
}

type PriceResponse struct {
	Symbol  string `json:"symbol"` // "symbol": "<symbol>",
	CloseEp int64  `json:"close"`  // "close": <close priceEp>,
	//   "open": <open priceEp>,
	//   "high": <high priceEp>,
	//   "low": <low priceEp>,
	//   "indexPrice": <index priceEp>,
	//   "markPrice": <mark priceEp>,
	//   "openInterest": <open interest>,
	//   "fundingRate": <funding rateEr>,
	//   "predFundingRate": <predicated funding rateEr>,
	//   "turnover": <turnoverEv>,
	//   "volume": <volume>,
	//   "timestamp": <timestamp>
}

func (pc *PhemexContract) GetPrice(ctx context.Context, symbol string) (*apd.Decimal, error) {
	symbol = ToPhemexSymbol(symbol)

	symbolScales, err := ScalesSubscriberInstance.GetLastSymbolScales(symbol)
	if err != nil {
		return nil, errors.Wrap(err, "can't take scales")
	}

	// This request is without rate limiter headers
	data, err := apiGetUnsigned(
		ctx, "https://api.phemex.com/md/ticker/24hr?symbol="+symbol)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch")
	}

	resp := PriceResponseWrapper{}
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshall JSON")
	}

	lastEp := apd.New(resp.Result.CloseEp, 0)
	return utils.Div(lastEp, symbolScales.PriceScaleDivider), nil
}
