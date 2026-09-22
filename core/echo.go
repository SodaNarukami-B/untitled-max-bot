package main

import (
	"bot/utils"
	"net/http"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	schemes "github.com/max-messenger/max-bot-api-client-go/schemes"
)

func handle_echo(resw http.ResponseWriter, req *http.Request, api *maxbot.Api, msg *schemes.Message) {
	resw.WriteHeader(http.StatusOK)
	go utils.Send_new_message(api, msg, msg.Body.Text)
}
