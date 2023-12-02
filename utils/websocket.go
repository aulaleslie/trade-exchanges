package utils

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type WSMessage struct {
	Payload             []byte
	DisconnectedWithErr error
}

// WSConfig webservice configuration
type WSConfig struct {
	Endpoint           string
	InitialTextMessage []byte
	KeepAlive          bool
	Timeout            time.Duration
	HeartbeatInterval  time.Duration
}

func WSConnectAndWatch(ctx context.Context, cfg *WSConfig, lg *zap.Logger) (<-chan WSMessage, error) {
	connectCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	c, _, err := websocket.DefaultDialer.DialContext(connectCtx, cfg.Endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open WebSocket connection")
	}

	return WSWatch(ctx, c, cfg, lg)
}

func WSWatch(ctx context.Context, c *websocket.Conn, cfg *WSConfig, lg *zap.Logger) (<-chan WSMessage, error) {
	lg = lg.Named("WSWatch")
	if cfg.InitialTextMessage != nil {
		err := c.WriteMessage(websocket.TextMessage, cfg.InitialTextMessage)
		if err != nil {
			if cerr := c.Close(); cerr != nil {
				lg.Warn("Unable to close WS connection", zap.Error(cerr))
			}
			return nil, errors.Wrap(err, "unable to send initial text message")
		}
	}

	sendErrorToOut := atomic.NewBool(true)
	stopWatchingCtx := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			lg.Debug("Cancel event happened")
		case <-stopWatchingCtx:
			lg.Debug("stopWatchingCtx happened")
		}
		sendErrorToOut.Store(false)

		cerr := c.Close()
		if cerr != nil {
			lg.Debug("Can't close websocket", zap.Error(cerr))
		}
	}()

	out := make(chan WSMessage, 100) // TODO: move 100 to config
	go func() {
		defer close(out)
		defer close(stopWatchingCtx)

		if cfg.KeepAlive {
			keepAliveWS(c, cfg.Timeout, cfg.HeartbeatInterval, lg)
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

func keepAliveWS(c *websocket.Conn, timeout, heartbeatInterval time.Duration, lg *zap.Logger) {
	lg = lg.Named("KeepAlive")
	ticker := time.NewTicker(heartbeatInterval)

	lastResponse := atomic.NewInt64(time.Now().Unix())
	c.SetPongHandler(func(msg string) error {
		lastResponse.Store(time.Now().Unix())
		return nil
	})

	go func() {
		defer ticker.Stop()
		defer c.Close()
		for {
			deadline := time.Now().Add(timeout)
			err := c.WriteControl(websocket.PingMessage, []byte{}, deadline)
			if err != nil {
				lg.Warn("Can't write Ping message")
				return
			}
			<-ticker.C
			if time.Since(time.Unix(lastResponse.Load(), 0)) > timeout {
				lg.Warn("No Pong response came in timeout")
				return
			}
		}
	}()
}
