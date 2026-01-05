# as-webhook

A Matrix Application Server written in Go to route messages as webhooks. This server receives messages from Matrix homeservers and forwards them to HTTP endpoints based on configurable routing rules.

This can be helpful if you have secondary services, such as a bot, or a function, that you don't want to run continuously, listening for messages that only appear occasionally.

This AS can be set up as a single service, and can be configured with [complex logic](https://cel.dev/) to call out or forward messages to a simple HTTP endpoint.

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

### Add to channel

1. Invite the bot to the channel
1. Run the following to add the AS to the channel:

```sh
curl -X POST \
    "https://homeserver.org/_matrix/client/v3/join/!<channel id>?user_id=%40<user id>
    -H "Authorization: Bearer <AS token>" \
    -H "Content-Type: application/json" \
    -d '{}'
```

### Command-line Options

- `<config>`: Config file path as positional argument (overrides `-config` flag)
- `-config`: Path to configuration file (default: `config.toml`)
- `-port`: Port to listen on (default: `8080`)
- `-generate-registration <path>`: Generate a Matrix AS `registration.yaml` at the given path and exit
- `-server <url>`: Public URL where this AS is reachable (used in `registration.yaml`)
- `-as-token <token>`: Optional AS token to include in the registration (auto-generated if omitted)

### Generate registration.yaml

Homeservers (HS) require an Application Service registration file to know how to talk to this AS. You can have the AS generate this file for you:

```bash
# Minimal: writes registration.yaml with generated tokens
./as-webhook -generate-registration registration.yaml -server http://localhost:8080

# With a custom AS token
./as-webhook -generate-registration registration.yaml -server http://app.local:8080 -as-token my-custom-token-12345
```

The generated `registration.yaml` includes:
- **id**: `matrix-as-webhook`
- **url**: The public URL you pass via `-server`
- **as_token**: Provided via `-as-token`, or securely generated if omitted
- **hs_token**: Securely generated token for the HS to authenticate to the AS
- **rate_limited**: `false` by default
- **namespaces**: Empty by default (this AS does not require reserved user namespaces)

Provide the generated file to your homeserver according to its AS registration process (e.g., placing it in the HS config and restarting).

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

The `AS_TOKEN` can be set via the `AS_TOKEN` environment variable.

### Configuration Options

- `selector`: CEL expression evaluated against the Matrix event as `event`. Return `true` to match (e.g., `event.content.body.contains('alert')`).
- `webhook_url`: The HTTP endpoint to send matched messages to
- `method`: HTTP method to use (default: `POST`)
- `stop_on_match`: If `true`, prevents further routes from being evaluated after this route matches (default: `false`)
- `send_body`: If `false`, excludes the `message` field from the webhook payload (default: `true`)
- `shared_secret`: Optional secret key for signing webhook requests with HMAC-SHA256

## Webhook Authentication

To verify that webhook requests come from as-webhook, configure a `shared_secret` for each route:

```toml
[[routes]]
name = "secure-endpoint"
selector = "true"
webhook_url = "https://myserver.com/webhook"
shared_secret = "my-secret-key-12345"
```

When a `shared_secret` is configured, as-webhook signs each webhook request using HMAC-SHA256 and includes the signature in the `X-Webhook-Signature` header:

```
X-Webhook-Signature: sha256=abcdef0123456789...
```

### Verifying Signatures

Your webhook endpoint should verify the signature using the shared secret:

**Python example:**
```python
import hmac
import hashlib
import json

def verify_webhook(request, shared_secret):
    # Get the signature from the header
    signature = request.headers.get('X-Webhook-Signature', '')

    # Get the raw body
    body = request.get_data()

    # Compute the expected signature
    expected_sig = 'sha256=' + hmac.new(
        shared_secret.encode(),
        body,
        hashlib.sha256
    ).hexdigest()

    # Verify using constant-time comparison
    return hmac.compare_digest(signature, expected_sig)
```

**Go example:**
```go
import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

func verifyWebhook(r *http.Request, sharedSecret string) bool {
	signature := r.Header.Get("X-Webhook-Signature")
	if signature == "" {
		return false
	}

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	h := hmac.New(sha256.New, []byte(sharedSecret))
	h.Write(body)
	expectedSig := "sha256=" + hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}
```

**Node.js example:**
```javascript
const crypto = require('crypto');

function verifyWebhook(req, sharedSecret) {
	const signature = req.headers['x-webhook-signature'];
	if (!signature) return false;

	const hmac = crypto.createHmac('sha256', sharedSecret);
	hmac.update(req.rawBody); // Make sure you capture the raw body
	const expectedSig = 'sha256=' + hmac.digest('hex');

	return crypto.timingSafeEqual(
		Buffer.from(signature),
		Buffer.from(expectedSig)
	);
}
```

**Important:** Always use constant-time comparison functions (like `hmac.compare_digest`, `crypto.timingSafeEqual`, etc.) to prevent timing attacks.

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
