![Hermes](img/logo.png)

[![Linter](https://github.com/ksdfg/hermes-backend/actions/workflows/linter.yml/badge.svg?branch=master)](https://github.com/ksdfg/hermes-backend/actions/workflows/linter.yml)
[![Security](https://github.com/ksdfg/hermes-backend/actions/workflows/security.yml/badge.svg?branch=master)](https://github.com/ksdfg/hermes-backend/actions/workflows/security.yml)
[![Deploy to Heroku](https://github.com/ksdfg/hermes-backend/actions/workflows/deploy.yml/badge.svg?branch=master)](https://github.com/ksdfg/hermes-backend/actions/workflows/deploy.yml)

Web API for WhatsApp messaging automation.

Built with [Fiber](https://github.com/gofiber/fiber) and [go-whatsapp](https://github.com/Rhymen/go-whatsapp), this is a
small web API that allows a user to automate sending a message to multiple people (whether they are or aren't in your
contacts list) with your number. You can also use placeholders in the message body that will be replaced with values
from the CSV file.

This application is meant primarily for event management - to contact people who have signed up for events for relaying
information and reminders. I cannot condone any usage of this application outside this usecase.

## Environment variables

The application uses [viper](https://github.com/spf13/viper) to read environment variables from a `.env` in the working
directory, or set in the environment. If a key is present in the `.env` and the environment, then the value set in the
environment will overwrite the one in the `.env`.

### List of all environment variables to set

| Key | Description |
| --- | --- |
| WHATSAPP_VERSION_MAJOR | Major version of the WhatsApp client |
| WHATSAPP_VERSION_MINOR | Minor version of the WhatsApp client |
| WHATSAPP_VERSION_PATCH | Patch version of the WhatsApp client |
| CLIENT_LONG | Long name of the application (will show up in users' WhatsApp) |
| CLIENT_SHORT | Short name of the application (for WhatsApp internal logs) |
| CLIENT_VERSION | Version of the application |
| QR_SIZE | Number of pixels each side of the QR code should be |
| CONCURRENCY | Peak concurrency while sending messages |
| ALLOW_ORIGINS | Origins to be allowed by CORS |

## Start the application

### Development

```bash
go run app.go
```

### Production

```bash
docker build -t hermes .
docker run -d -p 3000:3000 --env-file .env hermes
```
