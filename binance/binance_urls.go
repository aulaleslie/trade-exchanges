package binance

import (
	"fmt"
)

type BinanceURLs struct {
	APIURL                 string
	WebSocketBaseURL       string
	USAPIURL               string
	FutureAPIURL           string
	USWebSocketBaseURL     string
	FutureWebSocketBaseURL string
}

var OriginalBinanceURLs BinanceURLs = BinanceURLs{
	APIURL:                 "https://api.binance.com",
	WebSocketBaseURL:       "wss://stream.binance.com:9443/ws",
	USAPIURL:               "https://api.binance.us",
	FutureAPIURL:           "https://fapi.binance.com",
	USWebSocketBaseURL:     "wss://stream.binance.us:9443/ws",
	FutureWebSocketBaseURL: "wss://fstream.binance.com/ws",
}

var TestnetBinanceURLs BinanceURLs = BinanceURLs{
	APIURL:                 "https://testnet.binance.vision/api",
	WebSocketBaseURL:       "wss://testnet.binance.vision/ws",
	FutureWebSocketBaseURL: "wss://stream.binancefuture.com/ws",
	FutureAPIURL:           "https://testnet.binancefuture.com",
}

// WSUserDataServe serve user data handler with listen key
func (u BinanceURLs) WSUserDataURL(listenKey string) string {
	endpoint := fmt.Sprintf("%s/%s", u.WebSocketBaseURL, listenKey)
	return endpoint
}

// WSAllMiniMarketsStatServe serve websocket that push mini version of 24hr statistics for all market every second
func (u BinanceURLs) WSAllMiniMarketsStatURL() string {
	endpoint := fmt.Sprintf("%s/!miniTicker@arr", u.WebSocketBaseURL)
	return endpoint
}

func (u BinanceURLs) WSUSUserDataURL(listenKey string) string {
	endpoint := fmt.Sprintf("%s/%s", u.USWebSocketBaseURL, listenKey)
	return endpoint
}

func (u BinanceURLs) WSUSAllMiniMarketsStatURL() string {
	endpoint := fmt.Sprintf("%s/!ticker@arr", u.USWebSocketBaseURL)
	return endpoint
}

func (u BinanceURLs) WSFuturesUserDataURL(listenKey string) string {
	endpoint := fmt.Sprintf("%s/%s", u.FutureWebSocketBaseURL, listenKey)
	return endpoint
}

func (u BinanceURLs) WSFuturesAllMiniMarketStatsURL() string {
	endpoint := fmt.Sprintf("%s/!ticker@arr", u.FutureWebSocketBaseURL)
	return endpoint
}
