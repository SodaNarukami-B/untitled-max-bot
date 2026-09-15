from __future__ import annotations
from fastapi import FastAPI, Request, HTTPException
import uvicorn

app = FastAPI()

supported_types = ["newMessage", "callbackQuery"]
supported_commands = ["/echo"];

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


def handle_newMessage(data: dict) -> dict | int:
    # Paylaod
    payload = data.get("payload")
    if not isinstance(payload, dict):
        return 1  # User mistake / invalid structure

    # Dicts with chat_id and user_id
    chat = payload.get("chat")
    frm = payload.get("from")

    # Message text
    text = payload.get("text")

    if not isinstance(chat, dict) or not isinstance(frm, dict) or not text:
        return 1

    # Requied fields
    chat_id = chat.get("chatId")
    user_id = frm.get("userId") # XXX: I don't think we need it

    if not chat_id or not user_id:
        return 1


    # Command parsing
    command = Text_Command.parse(text)
    if not command.command:
        return 1

    if command.command not in supported_commands:
        return 1;

    match command.command:
        case "/echo":  # Echo
            return {
                "chatId": chat_id,
                "text": " ".join(command.params)
            }
        case _:
            return 2

@app.post("/core")
async def handle_webhook(request: Request):
    try:
        data = await request.json()
        if not isinstance(data, dict):
            raise HTTPException(status_code=400, detail="Invalid data")

    except Exception:
        raise HTTPException(status_code=400, detail="Invalid data")

    # Type check
    event_type = data.get("type")

    # route by type
    match event_type:
        case "newMessage":
            response = handle_newMessage(data)

            if response == 1:
                raise HTTPException(status_code=400, detail="Invalid syntax")
            if response == 2:
                raise HTTPException(status_code=500, detail="Server error")

            return response

        case "callbackQuery":
            return {"status": "ok"}

        case _:
            raise HTTPException(status_code=404, detail="Invalid type")


if __name__ == "__main__":
    uvicorn.run(app, host="127.0.0.1", port=8080)
