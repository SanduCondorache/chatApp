package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/SanduCondorache/chatApp/internal/client"
	"github.com/SanduCondorache/chatApp/internal/types"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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

	go func() {
		for m := range a.client.ChatCh {
			data, _ := json.Marshal(m)
			fmt.Println(m)
			runtime.EventsEmit(a.ctx, "chat:received", string(data))
		}
	}()
}

func (a *App) Register(username, email, password string) (string, error) {
	user := types.NewUser(username, email, password)

	err := a.client.SendMessage(user, types.Register)
	if err != nil {
		return "", err
	}

	return a.client.ReadMessage()
}

func (a *App) Login(username, password string) (string, error) {
	user := types.NewUser(username, "", password)

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

	str, err := a.client.ReadMessage()
	if err != nil {
		return "", err
	}

	return str, nil
}

func (a *App) CheckIsUserOnline(users []string) (map[string]bool, error) {
	data := map[string][]string{
		"users": users,
	}

	jsons, err := json.Marshal(data)

	if err != nil {
		return nil, err
	}

	temp := types.NewMessage(string(jsons))

	err = a.client.SendMessage(temp, types.GetConn)
	if err != nil {
		return nil, err
	}

	str, err := a.client.ReadMessage()

	var mp map[string]any

	var m types.Message

	if err = json.Unmarshal([]byte(str), &m); err != nil {
		slog.Error("unmarshal error", "err", err)
		return nil, err
	}

	if err = json.Unmarshal(m.Payload, &mp); err != nil {
		slog.Error("unmarshal error", "err", err)
		return nil, err
	}

	res := make(map[string]bool)
	for k, v := range mp {
		switch val := v.(type) {
		case bool:
			res[k] = val
		case string:
			res[k] = strings.ToLower(val) == "true"
		}
	}

	return res, nil
}

func (a *App) GetMessages(user1, user2 string) ([]types.MessageHist, error) {
	data := map[string]string{
		"user1": user1,
		"user2": user2,
	}

	jsons, err := json.Marshal(data)

	if err != nil {
		return nil, err
	}

	temp := types.NewMessage(string(jsons))

	err = a.client.SendMessage(temp, types.GetMsg)
	if err != nil {
		return nil, err
	}

	str, err := a.client.ReadMessage()
	if err != nil {
		return nil, err
	}
	var m types.Message

	if err := json.Unmarshal([]byte(str), &m); err != nil {
		return nil, err
	}

	var msgs []types.MessageHist
	err = json.Unmarshal(m.Payload, &msgs)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}
