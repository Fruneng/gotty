package main

import (
	"log"
	"net/http"

	"gotty/pkg/backend/localcommand"
	ttyserver "gotty/pkg/tty-server"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func originChekcer(r *http.Request) bool {
	return true
}

var webttyUpgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{"webtty"},
	CheckOrigin:     originChekcer,
}

func main() {

	r := gin.Default()
	r.GET("/ws", func(c *gin.Context) {
		//Upgrade get request to webSocket protocol
		conn, err := webttyUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.AbortWithError(http.StatusUpgradeRequired, err)
			log.Fatalf("web_terminal: %v", err)
			return
		}
		defer conn.Close()

		command := "/bin/sh"

		factory, err := localcommand.NewFactory(command, []string{}, &localcommand.Options{})
		if err != nil {
			log.Fatalf("local command :%v", err)
			return
		}

		session := ttyserver.NewTtySession(c, factory)
		session.WebSocket(c, conn)
	})
	r.Static("/css", "./static/css")
	r.Static("/js", "./static/js")
	r.StaticFile("/", "./static/index.html")
	r.StaticFile("/index.html", "./static/index.html")
	r.StaticFile("/config.js", "./static/config.js")
	r.StaticFile("/auth_token.js", "./static/auth_token.js")
	r.StaticFile("/favicon.png", "./static/favicon.png")

	r.Run()
}
