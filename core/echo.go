package main

import (
	"bot/utils"
	"net/http"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	schemes "github.com/max-messenger/max-bot-api-client-go/schemes"
)

func handle_echo(resw http.ResponseWriter, req *http.Request, api *maxbot.Api, upd schemes.MessageCreatedUpdate) {
	resw.WriteHeader(http.StatusOK)
	go utils.Maxapi_send_MCU(api, upd, upd.Message.Body.Text)
}
