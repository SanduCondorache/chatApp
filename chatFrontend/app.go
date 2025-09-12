package main

import (
	"context"
	"time"

	"github.com/SanduCondorache/chatApp/internal/client"
	"github.com/SanduCondorache/chatApp/internal/types"
)

type App struct {
	ctx    context.Context
	client *client.Client
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	client, err := client.NewClient()
	if err != nil {
		return
	}
	a.client = client
}

func (a *App) Login(username, email, password string) (string, error) {
	user := types.NewUser(username, email, password)

	err := a.client.SendMessage(user, types.Login)
	if err != nil {
		return "", err
	}

	return a.client.ReadMessage()
}

func (a *App) SearchUser(username string) (string, error) {
	msg := types.NewMessage(username)

	err := a.client.SendMessage(msg, types.Find)
	if err != nil {
		return "", err
	}

	return a.client.ReadMessage()
}

func (a *App) SendMsgBetweenUsers(user1 string, user2 string, msg string) (string, error) {
	temp := types.NewChatMessage(user1, user2, msg, time.Now())

	err := a.client.SendMessage(temp, types.Chat)
	if err != nil {
		return "", err
	}

	return a.client.ReadMessage()
}
