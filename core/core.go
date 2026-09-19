package main

/* INFO: About double connection

Max api uses double connect. First connection using for telling server that you received packet.
Second connection using for sending messages

Note: sending messages MUST be in goroutine, idk how works original max api package, so i think
that we better use that than not use
*/

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	schemes "github.com/max-messenger/max-bot-api-client-go/schemes"
)

// ----------------- Structures ------------------------
type Command struct {
	name string
	args []string
}

type MessageCU_handler func(http.ResponseWriter, *http.Request, *maxbot.Api, schemes.MessageCreatedUpdate)

// ------------------------ Configuraiton --------------------------
// TODO: move to config file or (recomended) flags
const HOST = ""
const SECRET = ""
const TOKEN = ""

// --------------------- Main ----------------------------
func main() {

	// INFO: tables contains function for supported commands.

	// Tables
	var MessageCU_command_table = map[string]MessageCU_handler{
		"echo": handle_echo,
	}

	// Get MAX api enviroment
	api, err := maxbot.New(os.Getenv(TOKEN))

	if err != nil {
		log.Printf("Failed to connect to API")
		return
	}

	// ----------------------- WEBHOOK SETUP --------------------------------------
	// ------------------------- WEBHOOK ENDPOINT --------------------------------------

	start_time := time.Now()

	http.HandleFunc("/webhook", func(resw http.ResponseWriter, req *http.Request) {

		// ----------- Webhook window for webhook setup ---------------------
		if time.Since(start_time) < 30*time.Second {
			resw.WriteHeader(http.StatusOK)
			return
		}

		// --------------- Main webhook logic ----------------------

		if req.Method != http.MethodPost {
			http.Error(resw, "Method not supported", http.StatusMethodNotAllowed)
			return
		}

		var upd schemes.MessageCreatedUpdate

		if err := json.NewDecoder(req.Body).Decode(&upd); err != nil {
			log.Printf("Failed to parse: %v", err)
			return
		}

		// UPDATE TYPE HANDLING
		switch upd.UpdateType {
		// ---------------------------- Message Created ----------------------------
		case schemes.TypeMessageCreated:
			raw_text := upd.Message.Body.Text
			command, err := parse_command(raw_text)

			if err != nil { // NOT COMMAND
				resw.WriteHeader(http.StatusBadRequest)
				go maxapi_send_MCU(api, upd, "Invalid request. Command expected.")

				return
			}

			// COMMAND HANDLING
			handler, ok := MessageCU_command_table[command.name]

			if !ok {
				resw.WriteHeader(http.StatusNotFound)
				go maxapi_send_MCU(api, upd, "Command not found")

				return
			}

			handler(resw, req, api, upd)

			return

		// --------------------------- Unknown mathod ----------------------------------
		default:
			resw.WriteHeader(http.StatusNotFound)
			go maxapi_send_MCU(api, upd, "Unsupported action")

			return
		}
	})

	// Starting server on background
	go func() {
		// starting server
		if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}

	}()

	time.Sleep(200 * time.Millisecond)

	// ---------------------------- WEBHOOK SETUP ----------------------------------------
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	data := fmt.Sprintf(`{"url": "%v", "update_types": ["message_create"], "secret": "%v"}`, HOST, SECRET)

	// Creating request structure
	req, err := http.NewRequest(http.MethodPost, "https://platform-api2.max.ru/subscriptions", strings.NewReader(data))
	if err != nil {
		log.Fatalf("Failed to create new request for webhook setup")
		return
	}

	// Setup headers
	req.Header.Add("Authorization", TOKEN)
	req.Header.Add("Content-Type", "application/json")

	// Sending
	resp, err := client.Do(req)

	if err != nil {
		log.Fatalf("Failed to setup webhook: %v", err)
		return
	}

	defer resp.Body.Close()

	// XXX: make stop channel via signal.Notify. Do not use that:
	select {}
}

// ------------------------ Command parsing function ------------------------------------
func parse_command(text string) (*Command, error) {
	if len(text) == 0 || text[0] != '/' { // Checks if text isn't a command
		return nil, errors.New("Received non-command string")
	}

	// "/command agr1 arg2" >> { "command": "command", "args": ["arg1", "arg2"] }

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil, errors.New("Empty command")
	}

	cmd := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	return &Command{name: cmd, args: args}, nil
}

// ----------------- Message Created Update -----------------------------------------

func maxapi_send_MCU(api *maxbot.Api, upd schemes.MessageCreatedUpdate, text string) int {

	nul_ctx := context.Background()

	err := api.Messages.Send(nul_ctx, maxbot.NewMessage().SetChat(upd.Message.Recipient.ChatId).SetText(text))

	if err != nil {
		log.Printf("Failed to call api.Messages.Send()")
		return -1
	}

	return 0
}

/*
 INFO: Small documentation for handler functions

Handler fucntions takes response counstructing after main webhook handler gets command name.
For example, when webhook handler see that command name is "echo", it calls special functtion like "handle echo".
That special fucntion parsing command argumets, counstucts http headers and sends message to user's chat.

For example "handle_echo" takes responseWriter, request, api and update message. It write OK status header, sends
message to user and return.

Note: handler function CANNOT finalize/send http response, so you MUST finalize it yourself

Note2: if handler function not using req or resw, you MUST still accept all of them
*/

func handle_echo(resw http.ResponseWriter, req *http.Request, api *maxbot.Api, upd schemes.MessageCreatedUpdate) {
	resw.WriteHeader(http.StatusOK)
	go maxapi_send_MCU(api, upd, upd.Message.Body.Text)
}
