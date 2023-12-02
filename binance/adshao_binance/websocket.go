package adshao_binance

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type WSMessage struct {
	Payload             []byte
	DisconnectedWithErr error
}

// WSConfig webservice configuration
type WSConfig struct {
	Endpoint  string
	KeepAlive bool
	Timeout   time.Duration
}

// WSServe TODO: look to utils package and take implementation from it
func WSServe(ctx context.Context, cfg *WSConfig, lg *zap.Logger) (<-chan WSMessage, error) {
	lg = lg.Named("BN-WSServe")

	c, _, err := websocket.DefaultDialer.Dial(cfg.Endpoint, nil)
	if err != nil {
		return nil, err
	}

	sendErrorToOut := atomic.NewBool(true)
	stopWatchingCtx := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			lg.Info("Cancel event happened")
		case <-stopWatchingCtx:
			lg.Info("stopWatchingCtx happened")
		}
		sendErrorToOut.Store(false)

		cerr := c.Close()
		if cerr != nil {
			lg.Error("Can't close websocket", zap.Error(cerr))
		}
	}()

	out := make(chan WSMessage, 100) // TODO: move 100 to config
	go func() {
		defer close(out)
		defer close(stopWatchingCtx)

		if cfg.KeepAlive {
			keepAlive(c, cfg.Timeout)
		}

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if sendErrorToOut.Load() {
					out <- WSMessage{DisconnectedWithErr: err}
				}
				return
			}
			out <- WSMessage{Payload: message}
		}
	}()
	return out, nil
}

func keepAlive(c *websocket.Conn, timeout time.Duration) {
	ticker := time.NewTicker(timeout)

	lastResponse := time.Now()
	c.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		defer ticker.Stop()
		for {
			deadline := time.Now().Add(10 * time.Second)
			err := c.WriteControl(websocket.PingMessage, []byte{}, deadline)
			if err != nil {
				return
			}
			<-ticker.C
			if time.Now().Sub(lastResponse) > timeout {
				c.Close()
				return
			}
		}
	}()
}
