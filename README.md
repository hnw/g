# gaproxy: Google Assistant Command Proxy

`gaproxy` is a lightweight HTTP proxy server that acts as a gateway to the [Google Assistant Service (gRPC) API](https://developers.google.com/assistant/sdk/reference/rpc/google.assistant.embedded.v1alpha2).

It accepts HTTP POST requests containing text commands (e.g., "Turn on the light"), forwards them to Google Assistant via gRPC, and returns the result text.

> [!WARNING]
> **Legacy API / Credential Requirement**
> This project relies on the **Google Assistant SDK (Embedded Assistant API)**.
> New registrations for this API are deprecated and significantly restricted by Google.
> **You must already possess valid OAuth2 credentials (`client_id`, `client_secret`, `refresh_token`) and a registered `device_model_id` to use this tool.**
> We cannot provide support for obtaining new credentials.

## Use Case

This tool is designed for **Headless Smart Home Control**.
It is not a conversational chatbot. It is intended to trigger Google Assistant actions (like controlling lights, checking weather, or broadcasting messages) from systems that cannot run the heavy Assistant SDK natively.

## Usage

### 1. Prerequisites

You need the following credentials obtained from your Google Cloud / Actions Console project:

* **OAuth2 Client ID**
* **OAuth2 Client Secret**
* **OAuth2 Refresh Token** (authorized for `https://www.googleapis.com/auth/assistant-sdk-prototype`)
* **Device Model ID** (registered in the Actions Console)
* **Device ID** (any unique string)

### 2. Running with Docker

The application is configured entirely via environment variables.

```bash
docker run -d -p 8080:8080 \
  -e GAPROXY_CLIENT_ID="your-client-id" \
  -e GAPROXY_CLIENT_SECRET="your-client-secret" \
  -e GAPROXY_REFRESH_TOKEN="your-refresh-token" \
  -e GAPROXY_DEVICE_ID="default" \
  -e GAPROXY_DEVICE_MODEL_ID="default" \
  -e GAPROXY_LANGUAGE_CODE="en-US" \
  ghcr.io/hnw/gaproxy:latest
```

Example `docker-compose.yml`:

```yaml
services:
  gaproxy:
    image: ghcr.io/hnw/gaproxy:latest
    ports:
      - "8080:8080"
    environment:
      - GAPROXY_CLIENT_ID=${GAPROXY_CLIENT_ID}
      - GAPROXY_CLIENT_SECRET=${GAPROXY_CLIENT_SECRET}
      - GAPROXY_REFRESH_TOKEN=${GAPROXY_REFRESH_TOKEN}
      - GAPROXY_DEVICE_ID=default
      - GAPROXY_DEVICE_MODEL_ID=default
      - GAPROXY_LANGUAGE_CODE=en-US
    restart: always
```

### 3. Sending Commands

Send a text command via HTTP POST. The command is treated as a "New Conversation" every time (context is not preserved).

**Example: Control devices**

```bash
curl -X POST -d "Turn on the light" http://localhost:8080
```

**Example: Broadcast**

```bash
curl -X POST -d "Broadcast breakfast is ready" http://localhost:8080
```

## Configuration

| Variable | Description | Required | Default |
| --- | --- | --- | --- |
| `GAPROXY_CLIENT_ID` | OAuth2 Client ID | ✅ | - |
| `GAPROXY_CLIENT_SECRET` | OAuth2 Client Secret | ✅ | - |
| `GAPROXY_REFRESH_TOKEN` | Authorized Refresh Token | ✅ | - |
| `GAPROXY_DEVICE_MODEL_ID` | Registered Device Model ID | ❌ | `default` |
| `GAPROXY_DEVICE_ID` | Unique ID for this instance | ❌ | `default` |
| `GAPROXY_LANGUAGE_CODE` | Language code (e.g., `en-US`, `ja-JP`) | ❌ | `en-US` |
| `PORT` | HTTP Server listening port | ❌ | `8080` |

## Troubleshooting

* **Broadcast not working:** If the "broadcast" command doesn't work, ensure your device (represented by the Model ID) is added to a "Home" in the Google Home app.
* **Changing Language:** The language setting is fixed at container startup. If you want to send English commands, you must set `GAPROXY_LANGUAGE_CODE=en-US`.

## License

Apache License 2.0
