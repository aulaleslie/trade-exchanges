package binance

import (
	"net/http"

	api "github.com/adshao/go-binance/v2"
	apiFutures "github.com/adshao/go-binance/v2/futures"
	"go.uber.org/zap"
)

func NewBinanceClient(baseURL, apiKey, secretKey string, l *zap.Logger) *api.Client {
	return &api.Client{
		APIKey:     apiKey,
		SecretKey:  secretKey,
		BaseURL:    baseURL,
		UserAgent:  "Binance/golang",
		HTTPClient: http.DefaultClient,
		Logger:     zap.NewStdLog(l.Named("adshao-binance")),
	}
}

func NewBinanceFuturesClient(baseURL, apiKey, secretKey string, l *zap.Logger) *apiFutures.Client {
	return &apiFutures.Client{
		APIKey:     apiKey,
		SecretKey:  secretKey,
		BaseURL:    baseURL,
		UserAgent:  "Binance/golang",
		HTTPClient: http.DefaultClient,
		Logger:     zap.NewStdLog(l.Named("adshao-binance")),
	}
}
