package ttyserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"gotty/pkg/backend"
	"gotty/pkg/webtty"
)

// TtySession TtySession
type TtySession struct {
	factory backend.Factory
}

// NewTtySession NewTtySession
func NewTtySession(ctx context.Context, factory backend.Factory) *TtySession {
	return &TtySession{
		factory: factory,
	}
}

// WebSocket WebSocket
func (s *TtySession) WebSocket(c *gin.Context, conn *websocket.Conn) {
	closeReason := "unknown reason"

	defer func() {
		log.Println("close reason:", closeReason)
	}()

	err := s.processWSConn(c, conn)

	switch err {
	case c.Err():
		closeReason = "cancelation"
	case webtty.ErrSlaveClosed:
		closeReason = "slave closed"
	case webtty.ErrMasterClosed:
		closeReason = "master closed"
	default:
		closeReason = fmt.Sprintf("an error: %s", err)
	}
}

func (s *TtySession) processWSConn(ctx context.Context, conn *websocket.Conn) error {
	typ, initLine, err := conn.ReadMessage()
	if err != nil {
		return errors.Wrapf(err, "failed to authenticate websocket connection")
	}
	if typ != websocket.TextMessage {
		return errors.New("failed to authenticate websocket connection: invalid message type")
	}

	var init InitMessage
	err = json.Unmarshal(initLine, &init)
	if err != nil {
		return errors.Wrapf(err, "failed to authenticate websocket connection")
	}

	queryPath := "?"
	query, err := url.Parse(queryPath)
	if err != nil {
		return errors.Wrapf(err, "failed to parse arguments")
	}
	params := query.Query()
	var slave backend.Slave
	slave, err = s.factory.New(params)
	if err != nil {
		return errors.Wrapf(err, "failed to create backend")
	}
	defer slave.Close()

	opts := []webtty.Option{}
	opts = append(opts, webtty.WithPermitWrite())

	tty, err := webtty.New(&wsWrapper{conn}, slave, opts...)
	if err != nil {
		return errors.Wrapf(err, "failed to create webtty")
	}

	err = tty.Run(ctx)

	return err
}

// varUnits are name-keyed maps, whose names will be iterated using order.
func (s *TtySession) titleVariables(order []string, varUnits map[string]map[string]interface{}) map[string]interface{} {
	titleVars := map[string]interface{}{}

	for _, name := range order {
		vars, ok := varUnits[name]
		if !ok {
			panic("title variable name error")
		}
		for key, val := range vars {
			titleVars[key] = val
		}
	}

	// safe net for conflicted keys
	for _, name := range order {
		titleVars[name] = varUnits[name]
	}

	return titleVars
}
