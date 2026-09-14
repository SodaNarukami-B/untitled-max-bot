from __future__ import annotations
from fastapi import FastAPI, Request, HTTPException
import uvicorn


app = FastAPI();

class Command:
    def __init__(self, command: str, params: list[str]):
        self.command = command;
        self.params = params;

    @classmethod
    def parse(cls, raw_text: str) -> Command:
        separated = raw_text.strip().split();

        if not separated:
            return cls(command="", params=[]);

        return cls(command=separated[0], params=separated[1:])



@app.post("/core")
async def handle_webhook(request: Request):
    data = await request.json();

    uid = data.get("user_id");
    cid = data.get("chat_id");
    text = data.get("text");

    if uid is None or cid is None or text is None:
        raise HTTPException(
            status_code=400,
            detail="Unexpected message"
        );

    command = Command.parse(text);
    if command.command == "" and command.params == []:
        raise HTTPException(
            status_code=400,
            detail="Invalid command"
        );
    print(f"Command received:\n\tcommand='{command.command}'\n\tparams={command.params}");

    match command.command:
        case "/find":
            return {
                    "method": "sendMessage",
                    "chat_id": cid,
                    "text": f"[Find command] called: text={command.params[0]}"
            };
        case _:
            return {
                    "method": "sendMessage",
                    "chat_id": cid,
                    "text": "[No command] called"
            };


    print("Received:", data);

    return {"code": 1};

if __name__ == "__main__":
    uvicorn.run(app, host="127.0.0.1", port=8080);

