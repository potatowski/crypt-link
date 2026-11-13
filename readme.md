# CryptLink

**CryptLink** is an encrypted messaging system where the recipient can read the message **only once**. After it is read, the message is automatically deleted and cannot be recovered, ensuring maximum confidentiality.

---

## Quick Start

Assuming you have docker and docker compose installed, you can run CryptLink with the following command:

### 1. Clone the repository
```bash
git clone https://github.com/potatowski/crypt-link.git
```

### 2. Navigate to the project directory
```bash
cd crypt-link
```

### 3. Configure the environment
Copy the example file and set a MongoDB password (use URL-safe characters, e.g. `openssl rand -hex 24`):

```bash
cp .env.example .env
```

### 4. Start the server with Docker Compose
```bash
docker compose up -d
```

This starts CryptLink on port 80. You can then access it at `http://localhost`.

### ⚠️ HTTPS is required outside localhost

The browser encryption uses the WebCrypto API (`crypto.subtle`), which browsers only expose in a **secure context** (HTTPS or `localhost`). Over plain HTTP on any other host the app will not work.

To serve it publicly with automatic Let's Encrypt certificates (Caddy in front of nginx, ports 80 and 443 must be reachable and `DOMAIN` must point to the host):

```bash
DOMAIN=example.com docker compose -f docker-compose.yml -f docker-compose.tls.yml up -d
```

### Configuration

The API is configured through environment variables:

| Variable        | Default                     | Description                      |
|-----------------|-----------------------------|----------------------------------|
| `DATABASE_URL`  | `mongodb://localhost:27017` | MongoDB connection string        |
| `DATABASE_NAME` | `onetime`                   | MongoDB database name            |
| `PORT`          | `8080`                      | Port the API listens on          |

With Docker Compose, `DATABASE_URL` is built for you from `MONGO_USERNAME` and `MONGO_PASSWORD` in `.env`.

### Local development (without Docker)

Requires Go (see `api/go.mod`) and a running MongoDB:

```bash
cd api
DATABASE_URL=mongodb://localhost:27017 go run ./cmd/server
```

The API listens on `:8080` under `/api/`. Serve the `web/` directory with any static server that proxies `/api/` to it (the nginx config in `docker/nginx/default.conf` is a reference).

### Running tests

```bash
cd api
go test -race ./...
```

## 🔐 What is it and how does it work?

CryptLink allows you to send sensitive messages securely and ephemerally. Here’s how it works:

1. **Browser-side encryption**: The message is encrypted with **AES-256-GCM** in the sender’s browser before being sent to the server.
2. **Temporary storage**: The encrypted version is stored on our servers **without the decryption key**.
3. **Secret link**: A link is generated containing:

   * A unique identifier (in the path), used only to locate the ciphertext.
   * The encryption key, stored in the URL fragment (after the `#`), which **is not sent to the server**.
4. **One-time read**: When the link is opened, the content is:

   * Retrieved from the server.
   * Permanently deleted.
   * Decrypted in the recipient’s browser using the key in the URL.

> The recipient must **click to reveal** the message. Opening the link alone does not consume it, so link-preview bots (chat apps, email scanners) do not destroy the message.

> If the link is opened a second time, the server returns a **404 error**, as the message has already been destroyed.

---

## 🛡 Security

* Messages are encrypted with **AES-256-GCM** (authenticated encryption) using a random 256-bit key generated in the browser via WebCrypto.
* The key lives only in the URL fragment and is never sent to the server, so the server only ever stores ciphertext.
* The identifier is a UUID v4 that merely locates the ciphertext. It is **not** what protects the message.
* Without the key, brute-forcing a 256-bit AES key is computationally infeasible.
* Messages have a selectable expiry (**1 hour, 1 day or 7 days**, default 7 days). Messages that are never read are **automatically purged** after they expire.

---

## 💡 Use cases

* **Password sharing**: Just send the URL. Even if the communication channel is compromised later, the content will no longer be available.
* **Promotions or challenges**: Send a message with a single word to a group. The first person to open the link gets the info.
* **Sensitive temporary information**: Send temporary codes, private messages, and more.

---

## 🧪 How it works technically

1. In the sender's browser:

   * The message is encrypted with AES-256-GCM using a random key.
   * The key is embedded in the URL fragment.
   * The encrypted version of the message is sent to the server along with the chosen expiry; the server returns the identifier.

2. In the recipient’s browser:

   * Upon opening the link, the identifier is used to fetch the message.
   * The server returns the message and immediately deletes it.
   * The key, present in the URL fragment, is used to decrypt the message locally.

> URL fragments (everything after `#`) **are not sent to the server**, ensuring that the key never leaves the browser.

---

## ⚠️ Considerations

The security of a message comes from the **256-bit key in the URL fragment**, not from the identifier. The identifier only tells the server which ciphertext to return.

Anyone who gets the full link before the recipient can read (and thereby destroy) the message, so share the link over a channel you trust. If the link is opened by someone else first, the recipient will get a 404, which tells you the message may have been intercepted.

The server never sees the key, but a compromised server could serve malicious JavaScript. Host it yourself, over HTTPS, if that matters for your threat model.
