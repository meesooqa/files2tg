# Video to Telegram

## Get started

1. Obtain `TELEGRAM_API_ID` and `TELEGRAM_API_HASH` from https://my.telegram.org/apps and `TELEGRAM_TOKEN` from https://core.telegram.org/bots/tutorial#obtain-your-bot-token.
2. Add Telegram Bot into Telegram Channel as admin.
3. Copy the `.env.example` file in the root directory of the project to the `.env` file and set vars.
4. `docker compose build`
5. `docker compose up`
6. Add files `var/files/*.mp4`
7. Customize `app/send/format.go`
8. Run `go run ./app/main.go`
9. Open https://localhost:8080