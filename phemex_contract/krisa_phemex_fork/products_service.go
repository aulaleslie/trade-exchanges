package krisa_phemex_fork

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Krisa/go-phemex/common"
	"github.com/aulaleslie/trade-exchanges/utils"
)

type ProductsService struct {
	c *Client
}

// Do send request
func (s *ProductsService) Do(ctx context.Context, opts ...RequestOption) (*ProductsResponse, error) {
	r := &request{
		method:   "GET",
		endpoint: "/public/products",
		secType:  secTypeNone,
	}
	data, rateLimHeaders, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	_ = rateLimHeaders // They ain't supported yet

	resp := new(BaseResponse)
	resp.Data = &ProductsResponse{}
	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, err
	}
	if resp.Code > 0 {
		return nil, &common.APIError{
			Code:    resp.Code,
			Message: resp.Msg,
		}
	}

	if resp.Data == nil {
		return nil, errors.New("Null response")
	}

	rows := resp.Data.(*ProductsResponse)
	return rows, nil
}

var (
	PerpetualProductType = "Perpetual"
	SpotProductType      = "Spot"
)

var (
	ListedProductStatus   = "Listed"
	DelistedProductStatus = "Delisted"
)

type Currency struct {
	Currency   string `json:"currency"`   //  "BTC",
	Name       string `json:"name"`       //  "Bitcoin",
	ValueScale int64  `json:"valueScale"` //  8,
	// MinValueEv utils.APDJSON `json:"minValueEv"` //  1,
	// MaxValueEv utils.APDJSON `json:"maxValueEv"` //  5000000000000000000,
	// NeedAddrTag int64         `json:"needAddrTag"` //  0
}

type Product struct {
	Symbol                   string        `json:"symbol"`                   // "BTCUSD",
	Type                     string        `json:"type"`                     // "Perpetual",
	DisplaySymbol            string        `json:"displaySymbol"`            // "BTC / USD",
	IndexSymbol              string        `json:"indexSymbol"`              // ".BTC",
	MarkSymbol               string        `json:"markSymbol"`               // ".MBTC",
	FundingRateSymbol        string        `json:"fundingRateSymbol"`        // ".BTCFR",
	FundingRate8hSymbol      string        `json:"fundingRate8hSymbol"`      // ".BTCFR8H",
	ContractUnderlyingAssets string        `json:"contractUnderlyingAssets"` // "USD",
	SettleCurrency           string        `json:"settleCurrency"`           // "BTC",
	QuoteCurrency            string        `json:"quoteCurrency"`            // "USD",
	ContractSize             utils.APDJSON `json:"contractSize"`             // 1.0,
	LotSize                  utils.APDJSON `json:"lotSize"`                  // 1,
	TickSize                 utils.APDJSON `json:"tickSize"`                 // 0.5,
	PriceScale               int64         `json:"priceScale"`               // 4,
	RatioScale               int64         `json:"ratioScale"`               // 8,
	PricePrecision           utils.APDJSON `json:"pricePrecision"`           // 1,
	MinPriceEp               utils.APDJSON `json:"minPriceEp"`               // 5000,
	MaxPriceEp               utils.APDJSON `json:"maxPriceEp"`               // 10000000000,
	MaxOrderQty              utils.APDJSON `json:"maxOrderQty"`              // 1000000,
	Description              string        `json:"description"`              // "BTC/USD perpetual contracts are priced on the .BTC Index. Each contract is worth 1 USD. Funding fees are paid and received every 8 hours at UTC time: 00:00, 08:00 and 16:00.",
	Status                   string        `json:"status"`                   // "Listed",
	// TipOrderQty    utils.APDJSON `json:"tipOrderQty"`    // 1000000
}

// /public/products
type ProductsResponse struct {
	RatioScale int         `json:"ratioScale"`
	Currencies []*Currency `json:"currencies"`
	Products   []*Product  `json:"products"`
}
