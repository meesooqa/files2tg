# Video to Telegram

## Get started

- Obtain `TELEGRAM_API_ID` and `TELEGRAM_API_HASH` from https://my.telegram.org/apps and `TELEGRAM_TOKEN` from https://core.telegram.org/bots/tutorial#obtain-your-bot-token).
- Add Telegram Bot into Telegram Channel as admin.
- Copy the `.env.example` file in the root directory of the project to the `.env` file and set vars.
- `docker compose build`
- `docker compose up`
- Add files `var/files/*.mp4`
- Run `go run ./app/main.go`
- Open https://localhost:8080