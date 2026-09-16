import json

configPath = "./config.json";

with open(configPath, "r", encoding="utf-8") as f:
    CONFIG = json.load(f);


