# kinda-bot-proxy

Simple proxy for Telegram Bot API with key-based protection. Designed for small bots (e.g., [GyverLibs/FastBot2](https://github.com/GyverLibs/FastBot2)).

## How it works

The proxy adds simple authorization via a key embedded in the bot token. Token format: `KEY_ORIGINAL_TOKEN`

- `KEY` — your secret key (set via environment variable)
- `_` — separator (underscore is not used in original Telegram tokens)
- `ORIGINAL_TOKEN` — your actual bot token

The proxy validates the key and, if correct, forwards the request to the real Telegram API without the key.  
If key is incorrect - silently close connection. Behave like port is closed.  

## Quick start with Docker

```bash
docker run -d \
  -p 8080:8080 \
  -e PROXY_KEY=your_secret_key \
  --name telegram-proxy \
  kondr1/kinda-bot-proxy:latest
```

## Usage

### Example for FastBot2

```cpp
#include <FastBot2.h>

FastBot2 bot;

void setup() {
  // Instead of regular token use KEY_TOKEN
  // If your key is: mysecret
  // And bot token is: 123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
  // Then use: mysecret_123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11

  bot.setToken("mysecret_123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11");

  // Set your proxy address
  bot.setProxy("http://your-proxy-server.com:8080");

  fb::Result res = bot.sendCommand(tg_cmd::getMe);
  Serial.println(res[tg_apih::first_name]);
  Serial.println(res[tg_apih::username]);
  Serial.println(res.getRaw());
}
```

## Deployment

### Docker Compose

```yaml
version: '3.8'

services:
  telegram-proxy:
    image: kondr1/kinda-bot-proxy:latest
    ports:
      - "8080:8080"
    environment:
      - PROXY_KEY=your_secret_key_here
      - PORT=8080
    restart: unless-stopped
```
