package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"bot/utils"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	schemes "github.com/max-messenger/max-bot-api-client-go/schemes"
)

// TYPES
type Command struct {
	name string
	args []string
}

type text_command_handler func(http.ResponseWriter, *http.Request, *maxbot.Api, *schemes.Message)

func main() {
	HOST := flag.String("host", "", "")
	TOKEN := flag.String("token", "", "")
	SECRET := flag.String("secret", "", "")

	flag.Parse()

	text_command_table := map[string]text_command_handler{
		"echo": handle_echo,
	}

	// Api connection
	api, err := maxbot.New(*TOKEN)

	if err != nil {
		log.Fatalf("Filed to connect to api: %v", err)
		return
	}

	start_time := time.Now()

	http.HandleFunc("/webhook", func(resw http.ResponseWriter, req *http.Request) {

		if time.Since(start_time) < 30*time.Second {
			resw.WriteHeader(http.StatusOK)
			log.Print("+ webhook challenge")
			return
		}

		// At this moment we can parse only post requests
		if req.Method != http.MethodPost {
			resw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		secret := req.Header.Get("secret")

		if secret == "" || secret != *SECRET {
			resw.WriteHeader(http.StatusForbidden)
			return
		}

		// todo: search for max header structure in wireshark after you win first tour

		// Searching for method

		var msg schemes.MessageCreatedUpdate

		if json.NewDecoder(req.Body).Decode(&msg) != nil {
			resw.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parings received method
		switch msg.UpdateType {
		case schemes.TypeMessageCreated:
			{
				// MessageCreated can be used only for text command as analog for callback
				// Command parsing
				raw_text := msg.Message.Body.Text
				comm, err := parse_command(raw_text)

				if err != nil {
					resw.WriteHeader(http.StatusBadRequest)
					go utils.Send_new_message(api, &msg.Message, "I undestand commands only")
					return
				}

				// Handling command

				command_handler := text_command_table[comm.name]
				if command_handler == nil {
					resw.WriteHeader(http.StatusNotImplemented)
					go utils.Send_new_message(api, &msg.Message, "I don't know this command")
					return
				}

				command_handler(resw, req, api, &msg.Message)

				return
			}
		default:
			{
				resw.WriteHeader(http.StatusNotImplemented)
				go utils.Send_new_message(api, &msg.Message, "I can't answer on this action")
				return
			}
		}
	})

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	time.Sleep(1 * time.Second)

	// Webhook subscribe
	client := http.Client{
		Timeout: 250 * time.Millisecond,
	}

	data := fmt.Sprintf(`{"url": %v, "update_types": ["message_created"], "secret": %v"}`, HOST, SECRET)
	req, _ := http.NewRequest("POST", "https://platform-api2.max.ru/subscription", strings.NewReader(data))

	req.Header.Add("Authorization", *TOKEN)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to complete webhook subscription: %v\n", err)
		return
	}

	success := resp.Header.Get("success")

	if success == "" || success == "false" {
		log.Fatal("Failed to complete webhook subscription: subscription denied/error")
		return
	}

	log.Print("Subscripted to webhook")

	// Exit endpoint
	exit := make(chan os.Signal, 1)

	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit

	fmt.Print("Exiting...")
}

func parse_command(text string) (*Command, error) {
	if len(text) == 0 {
		return nil, errors.New("Empty command")
	}

	if text[0] != '/' {
		return nil, errors.New("Not a command")
	}

	parts := strings.Fields(text)

	comm := strings.TrimPrefix(parts[0], "/")

	var args []string

	if len(parts) > 1 {
		args = parts[1:]
	}

	return &Command{
		name: comm,
		args: args,
	}, nil
}
