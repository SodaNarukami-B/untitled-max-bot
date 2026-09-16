from __future__ import annotations
from fastapi import FastAPI, Request, HTTPException
import newMessage as new_msg
from conf import CONFIG
import uvicorn

from pathlib import Path
import json

# ----------------- APP, WEBHOOK -------------------------
app = FastAPI()

# endpoint "/core"
@app.post("/core")
# Catches EVERY request, sended to this endpoint
async def handle_webhook(request: Request):

    # Parsing request as json
    try:
        data = await request.json()
        if not isinstance(data, dict):
            raise HTTPException(status_code=400, detail="Invalid data")

    except Exception:
        raise HTTPException(status_code=400, detail="Invalid data")

    # Parsing json as MAX message

    # Type check
    event_type = data.get("type")
    # We not need to check if event type not exists, because that will be checked in (match event_type)

    # Picking fucntions depending type
    match event_type:
        # New Message (just message)
        case "newMessage":
            response = new_msg.handle_newMessage(data)

            if response == 1:
                raise HTTPException(status_code=400, detail="Invalid syntax")
            if response == 2:
                raise HTTPException(status_code=500, detail="Server error")

            return response

        # Callback Query (button pressed or another actions)
        case "callbackQuery":
            return {"status": "ok"}

        case _:
            raise HTTPException(status_code=404, detail="Invalid type")


if __name__ == "__main__":
    uvicorn.run(app, host="127.0.0.1", port=8080)

