package utils

import (
	"context"
	"log"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	schemes "github.com/max-messenger/max-bot-api-client-go/schemes"
)

func Send_new_message(api *maxbot.Api, msg *schemes.Message, text string) error {

	nil_ctx := context.TODO()

	if err := api.Messages.Send(nil_ctx, maxbot.NewMessage().SetChat(msg.Recipient.ChatId).SetText(text)); err != nil {
		log.Printf("Failed to send message to %v: %v\n", msg.Recipient.ChatId, err)
		return err
	}

	return nil
}
