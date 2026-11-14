# Security Policy

CryptLink exists to keep messages confidential, so security reports are taken seriously.

## Supported versions

Only the latest release on `main` receives security fixes.

## Reporting a vulnerability

Please **do not open a public issue** for a security problem.

Report it privately through GitHub: open the repository's **Security** tab and choose **Report a vulnerability**. Include:

- what an attacker can do and what they need (for example, a link, network position or server access);
- steps or a proof of concept to reproduce it;
- the affected version or commit.

You should get a first answer within 7 days. Once a fix is released, the advisory is published and you are credited unless you ask otherwise.

## Scope

In scope:

- anything that lets the server, a proxy or a third party read message plaintext or the key in the URL fragment;
- reading a message more than once, after it expired, or without the reader clicking reveal;
- weaknesses in the browser encryption (AES-256-GCM, key and IV generation);
- bypasses of the CSP, security headers, rate limiting or request size limits;
- issues in the Docker, nginx, Caddy or MongoDB configuration shipped in this repository.

Out of scope:

- a recipient copying or forwarding a message they were sent;
- leaks caused by a compromised browser, device or browser extension;
- deployments without HTTPS, which the app warns about;
- denial of service that needs traffic volume rather than a flaw.
