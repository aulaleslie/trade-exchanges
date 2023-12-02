package bybit

import (
	"net/http"
	"time"

	"github.com/hirokisan/bybit/v2"
	"go.uber.org/zap"
)

func NewBybitRestClient(apiKey, secretKey string, l *zap.Logger) *bybit.Client {
	client := bybit.NewClient().WithAuth(apiKey, secretKey)

	return client
}

func NewBybitWSClient(apiKey, secretKey string, l *zap.Logger) *bybit.WebSocketClient {
	wsClient := bybit.NewWebsocketClient().WithBaseURL("wss://stream.bybit.com").WithAuth(apiKey, secretKey)

	return wsClient
}

func NewHTTPClient(timeout time.Duration) *http.Client {
	client := &http.Client{
		Timeout: timeout, // Example: set a timeout of 10 seconds
	}
	return client
}
