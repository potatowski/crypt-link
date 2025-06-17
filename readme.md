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

### 3. Create a `.env` file
Create a `.env` file in the root directory of the project with the following content:

```bash
cp api/config/.env.example api/config/.env
```

### 4. Configure the `.env` file
Edit the `.env` file to set your desired configuration options

### 5. Start the server with Docker Compose
```bash
docker compose up -d
```

This will start the CryptLink server on port 80. You can then access it at `http://localhost`.

## 🔐 What is it and how does it work?

CryptLink allows you to send sensitive messages securely and ephemerally. Here’s how it works:

1. **Browser-side encryption**: The message is encrypted in the sender’s browser before being sent to the server.
2. **Temporary storage**: The encrypted version is stored on our servers **without the decryption key**.
3. **Secret link**: A link is generated containing:

   * A unique identifier.
   * The encryption key.
     Both are stored in the URL fragment (after the `#`), which **is not sent to the server**.
4. **One-time read**: When the link is opened, the content is:

   * Retrieved from the server.
   * Permanently deleted.
   * Decrypted in the recipient’s browser using the key in the URL.

> If the link is opened a second time, the server returns a **404 error**, as the message has already been destroyed.

---

## 🛡 Security

* We use **AES-256** with randomly generated keys in the browser.
* Unique identifiers follow the **UUID v4** standard (RFC 4122).
* The combination of key + identifier makes leakage or brute-force attacks virtually impossible.
* Even with all the computational power in the world, breaking a message would take more than **a trillion years**.

---

## 💡 Use cases

* **Password sharing**: Just send the URL. Even if the communication channel is compromised later, the content will no longer be available.
* **Promotions or challenges**: Send a message with a single word to a group. The first person to open the link gets the info.
* **Sensitive temporary information**: Send temporary codes, private messages, and more.

---

## 🧪 How it works technically

1. In the sender's browser:

   * The message is encrypted with AES-256 using a random key.
   * The key is embedded in the URL fragment.
   * The encrypted version of the message is sent to the server along with a UUID.

2. In the recipient’s browser:

   * Upon opening the link, the identifier is used to fetch the message.
   * The server returns the message and immediately deletes it.
   * The key, present in the URL fragment, is used to decrypt the message locally.

> URL fragments (everything after `#`) **are not sent to the server**, ensuring that the key never leaves the browser.

---

## ⚠️ Considerations

While it’s technically possible for someone to access the link before the recipient and “steal” the message, the probability of a UUID v4 randomly colliding is approximately **8.64 × 10⁻⁷⁸** — virtually impossible.
