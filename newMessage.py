from __future__ import annotations
from conf import CONFIG

# CONFIG PULL
supported_commands = CONFIG.get("supported", {}).get("commands", []);

# --------------------- TEXT COMMAND CLASS ----------------------
class Text_Command:
    def __init__(self, command: str, params: list[str]):
        self.command = command
        self.params = params

    @classmethod
    def parse(cls, raw_text: str) -> Text_Command:
        separated = raw_text.strip().split()

        if not separated:
            return cls(command="", params=[])

        return cls(command=separated[0], params=separated[1:])

# ------------------- PARSER/VALIDATOR ------------------------

# TODO: Also you MUST verify that some fields like user id is not empty
def check_payload_fields(payload: dict) -> int:

    payload_req_fields = ["msgId", "chat", "from", "timestamp", "text", "parts"]
    chat_req_fields = ["chatId"];
    from_req_fields = ["userId", "firstName", "lastName"];

    # Payload check
    for req in payload_req_fields:
        field = payload.get(req);
        if req not in payload or payload[req] is None:
            return -1;

    # Types check
    chat = payload.get("chat") # 100% exists
    frm = payload.get("from")
    parts = payload.get("parts")

    if not isinstance(chat, dict) or not isinstance(frm, dict) or not isinstance(parts, list):
        return -1;

    # Chat check
    for req in chat_req_fields:
        field = chat.get(req);
        if req not in chat or not chat[req]:
            return -1

    for req in from_req_fields:
        field = frm.get(req);
        if req not in frm or not frm[req]:
            return -1

    return 0;


# -------------------- NEW MESSAGE HENDLER ----------------------------
def handle_newMessage(data: dict) -> dict | int:
    if not supported_commands:
        print("[newMessage/ERROR]: no field ['supported']['commands'] in configuration file");
        return 2; # server error

    
    # Paylaod
    payload = data.get("payload")

    if not isinstance(payload, dict):
        return 1  # User mistake / invalid structure

    if check_payload_fields(payload) == -1:
        return 1;

    # Requied fields
    chat_id = payload.get("chat", {}).get("chatId");
    user_id = payload.get("from", {}).get("userId");


    # Message text
    text = str(payload.get("text"));

    # Command parsing
    command = Text_Command.parse(text) # text is actualy exists, we checked that in the parser

    if not command.command:
        return 1

    if command.command not in supported_commands:
        return 1;
     
    # minimal logging
    log = f"[newMessage/INFO]: CID={chat_id} UID={user_id} TXT={(text[:10] + '...') if len(text) > 10 else text}"
    print(log)

    match command.command:
        case "/echo":  # Echo
            return {
                "type": "sendMessage",
                "payload": {
                        "chatId": chat_id,
                        "text": " ".join(command.params)
                }
            }
        case _:
            return 2


