package main

import (
	"context"

	"github.com/SanduCondorache/chatApp/internal/client"
	"github.com/SanduCondorache/chatApp/internal/types"
	"github.com/gorilla/websocket"
)

type App struct {
	ctx  context.Context
	conn *websocket.Conn
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	conn, err := client.CreateConn()
	if err != nil {
		return
	}
	a.conn = conn
}

func (a *App) Login(username, email, password string) (string, error) {
	user := types.NewUser(username, email, password)

	err := client.SendMessage(a.conn, user, types.Login)
	if err != nil {
		return "", err
	}

	return client.ReadMessage(a.conn)
}

func (a *App) SearchUser(username string) (string, error) {
	msg := types.NewMessage(username)

	err := client.SendMessage(a.conn, msg, types.Find)
	if err != nil {
		return "", err
	}

	return client.ReadMessage(a.conn)
}
