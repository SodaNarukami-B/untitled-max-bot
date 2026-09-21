package utils

import (
	"context"
	"log"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	schemes "github.com/max-messenger/max-bot-api-client-go/schemes"
)

func Maxapi_send_MCU(api *maxbot.Api, upd schemes.MessageCreatedUpdate, text string) int {

	nul_ctx := context.Background()

	err := api.Messages.Send(nul_ctx, maxbot.NewMessage().SetChat(upd.Message.Recipient.ChatId).SetText(text))

	if err != nil {
		log.Printf("Failed to call api.Messages.Send(): %v", err)
		return -1
	}

	return 0
}
