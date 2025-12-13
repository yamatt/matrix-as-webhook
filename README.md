# go-as-webhook

A Matrix Application Server written in Go to route messages as webhooks. This server receives messages from Matrix homeservers and forwards them to HTTP endpoints based on configurable routing rules.

## Features

- **Matrix Application Server Protocol**: Implements the Matrix AS API endpoints
- **Message Routing**: Route messages to different webhooks based on message content patterns
- **Configurable**: JSON-based configuration for routing rules
- **Lightweight**: Simple, focused implementation in Go

## Installation

### Using Go

```bash
go build -o go-as-webhook
```

### Using Docker

```bash
# Build the image
docker build -t go-as-webhook .

# Run the container
docker run -p 8008:8008 -v $(pwd)/config.json:/app/config.json go-as-webhook -config /app/config.json
```

## Usage

```bash
./go-as-webhook -config config.json -port 8008
```

### Command-line Options

- `-config`: Path to configuration file (default: `config.json`)
- `-port`: Port to listen on (default: `8008`)

## Configuration

Create a `config.json` file to define your routing rules:

```json
{
  "routes": [
    {
      "pattern": "alert",
      "webhook_url": "http://localhost:9000/alerts",
      "method": "POST"
    },
    {
      "pattern": "notification",
      "webhook_url": "http://localhost:9000/notifications",
      "method": "POST"
    },
    {
      "pattern": "",
      "webhook_url": "http://localhost:9000/default",
      "method": "POST"
    }
  ]
}
```

### Configuration Options

- `pattern`: Text pattern to match in message body (uses substring matching). Empty string matches all messages.
- `webhook_url`: The HTTP endpoint to send matched messages to
- `method`: HTTP method to use (default: `POST`)

## Matrix Homeserver Configuration

To register this application service with your Matrix homeserver, create a registration file (e.g., `registration.yaml`):

```yaml
id: webhook-as
url: http://localhost:8008
as_token: your-application-service-token
hs_token: your-homeserver-token
sender_localpart: webhook_bot
namespaces:
  users:
    - exclusive: false
      regex: '@.*:yourdomain.com'
  rooms: []
  aliases: []
```

Then reference this file in your homeserver's configuration (e.g., for Synapse, add to `homeserver.yaml`):

```yaml
app_service_config_files:
  - /path/to/registration.yaml
```

## Webhook Payload

When a message is matched, the application server sends a JSON payload to the configured webhook:

```json
{
  "event_id": "!event_id",
  "room_id": "!room_id:domain.com",
  "sender": "@user:domain.com",
  "timestamp": 1234567890,
  "message": "The message body text",
  "content": {
    "body": "The message body text",
    "msgtype": "m.text"
  },
  "event_type": "m.room.message"
}
```

## API Endpoints

The server implements the Matrix Application Server Protocol:

- `PUT /_matrix/app/v1/transactions/{txnId}` - Receive events from the homeserver
- `GET /_matrix/app/v1/rooms/{roomAlias}` - Room alias queries (returns 404)
- `GET /_matrix/app/v1/users/{userId}` - User queries (returns 404)
- `GET /health` - Health check endpoint

## License

See LICENSE file for details.
