# as-webhook

A Matrix Application Server written in Go to route messages as webhooks. This server receives messages from Matrix homeservers and forwards them to HTTP endpoints based on configurable routing rules.

## Features

- **Matrix Application Server Protocol**: Implements the Matrix AS API endpoints
- **Message Routing**: Route messages to different webhooks based on message content patterns
- **Configurable**: TOML-based configuration for routing rules
- **Lightweight**: Simple, focused implementation in Go

## Installation

### Using Go

```bash
go build -o as-webhook ./cmd/go-as-webhook
```

### Using Docker

```bash
# Build the image
docker build -t as-webhook .

# Run the container
docker run -p 8080:8080 -v $(pwd)/config.toml:/app/config.toml as-webhook -config /app/config.toml
```

## Usage

```bash
# With positional config argument
./as-webhook config.toml

# Or with -config flag
./as-webhook -config config.toml -port 8080
```

### Command-line Options

- `<config>`: Config file path as positional argument (overrides `-config` flag)
- `-config`: Path to configuration file (default: `config.toml`)
- `-port`: Port to listen on (default: `8080`)

## Configuration

Create a `config.toml` file to define your routing rules (CEL selectors):

```toml
[[routes]]
name = "alerts"
selector = "event.type == 'm.room.message' && event.content.body.contains('alert')"
webhook_url = "http://localhost:9000/alerts"
method = "POST"

[[routes]]
name = "notifications"
selector = "event.content.body.contains('notification')"
webhook_url = "http://localhost:9000/notifications"
method = "POST"

[[routes]]
name = "default"
selector = "true"
webhook_url = "http://localhost:9000/default"
method = "POST"
```

### Configuration Options

- `selector`: CEL expression evaluated against the Matrix event as `event`. Return `true` to match (e.g., `event.content.body.contains('alert')`).
- `webhook_url`: The HTTP endpoint to send matched messages to
- `method`: HTTP method to use (default: `POST`)

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
