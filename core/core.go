package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	schemes "github.com/max-messenger/max-bot-api-client-go/schemes"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Command struct {
	name string
	args []string
}

func parse_command(text string) (*Command, error) {
	if text[0] != '/' {
		return &Command{}, errors.New("Received non-command string")
	}

	cmd := strings.Split(text, " ")[0]

	parts := strings.Fields(text)

	return &Command{
		name: strings.TrimPrefix(cmd, "/"),
		args: parts[:1],
	}, nil
}

const HOST = ""
const SECRET = ""
const TOKEN = ""

func main() {
	api, err := maxbot.New(os.Getenv(TOKEN))

	if err != nil {
		log.Printf("Failed to connect to API")
		return
	}

	// ----------------------- WEBHOOK SETUP --------------------------------------

	data := fmt.Sprintf(`{"url": "%v", "update_types": ["message_create"], "secret": "%v"}`, HOST, SECRET)

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, "https://platform-api2.max.ru/subscriptions", strings.NewReader(data))
	if err != nil {
		log.Fatalf("Failed to create new request for webhook setup")
		return
	}

	req.Header.Add("Authorization", TOKEN)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalf("Failed to setup webhook: %v", err)
		return
	}

	defer resp.Body.Close()

	// ------------------------- WEBHOOK ENDPOINT --------------------------------------

	start_time := time.Now()

	ctx := context.Background()
	http.HandleFunc(HOST+"/webhook", func(w http.ResponseWriter, r *http.Request) {

		// ----------- Webhook window ---------------------
		if time.Since(start_time) < 30*time.Second {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
			return
		}

		var upd schemes.MessageCreatedUpdate

		if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
			log.Printf("Failed to parse: %v", err)
			return
		}

		switch upd.UpdateType {
		case schemes.TypeMessageCreated:
			{
				raw_text := upd.Message.Body.Text

				command, err := parse_command(raw_text)

				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					err := api.Messages.Send(ctx, maxbot.NewMessage().SetChat(upd.Message.Recipient.ChatId).SetText("Invalid command"))

					if err != nil {
						log.Printf("Failed to send message. Chat id - %v", upd.Message.Recipient.ChatId)
						return
					}

				}

				switch command.name {
				case "echo":
					{
						err := api.Messages.Send(ctx, maxbot.NewMessage().SetChat(upd.Message.Recipient.ChatId).SetText(strings.Join(command.args, " ")))

						if err != nil {
							log.Printf("Failed to send message. Chat id - %v", upd.Message.Recipient.ChatId)
							return
						}

						w.WriteHeader(http.StatusOK)

						return
					}
				default:
					{
						err := api.Messages.Send(ctx, maxbot.NewMessage().SetChat(upd.Message.Recipient.ChatId).SetText("Invalid command"))

						w.WriteHeader(http.StatusNotFound)

						if err != nil {
							log.Printf("Failed to send message. Chat id - %v", upd.Message.Recipient.ChatId)
							return
						}

					}
				}
			}
		}

		err := api.Messages.Send(ctx, maxbot.NewMessage().SetChat(upd.Message.Recipient.ChatId).SetText("Hello via webhook"))

		if err != nil {
			log.Printf("Failed to send: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)

		// return
	})

	if err := http.ListenAndServe("0.0.0.0", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
