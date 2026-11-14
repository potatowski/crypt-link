// --- Encoding helpers ---

const BASE64_CHUNK = 0x8000;
const FRAGMENT_RE = /^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\.([A-Za-z0-9_+/-]{43}=?)$/i;

function bytesToBase64(bytes) {
  let binary = '';
  for (let i = 0; i < bytes.length; i += BASE64_CHUNK) {
    binary += String.fromCharCode.apply(null, bytes.subarray(i, i + BASE64_CHUNK));
  }
  return btoa(binary);
}

// Accepts both standard base64 (legacy links) and base64url, padded or not.
function base64ToBytes(b64) {
  let std = b64.replace(/-/g, '+').replace(/_/g, '/');
  while (std.length % 4 !== 0) std += '=';
  const binary = atob(std);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
  return bytes;
}

function bytesToBase64Url(bytes) {
  return bytesToBase64(bytes).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

// --- Services ---

const CryptoService = {
  async generateKeyAndIV() {
    const key = await crypto.subtle.generateKey({ name: 'AES-GCM', length: 256 }, true, ['encrypt', 'decrypt']);
    const iv = crypto.getRandomValues(new Uint8Array(12));
    return { key, iv };
  },

  async encrypt(message, key, iv) {
    const encodedMsg = new TextEncoder().encode(message);
    const encrypted = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, key, encodedMsg);
    return bytesToBase64(new Uint8Array(encrypted));
  },

  async exportKeyAndIV(key, iv) {
    const rawKey = await crypto.subtle.exportKey('raw', key);
    const keyBase64 = bytesToBase64Url(new Uint8Array(rawKey));
    const ivBase64 = bytesToBase64(iv);
    return { keyBase64, ivBase64 };
  },

  async decrypt(ciphertext, keyBase64, ivBase64) {
    const rawKey = base64ToBytes(keyBase64);
    const cryptoKey = await crypto.subtle.importKey('raw', rawKey, { name: 'AES-GCM' }, false, ['decrypt']);
    
    const cipherArray = base64ToBytes(ciphertext);
    const ivArray = base64ToBytes(ivBase64);
    
    const decrypted = await crypto.subtle.decrypt({ name: 'AES-GCM', iv: ivArray }, cryptoKey, cipherArray);
    return new TextDecoder().decode(decrypted);
  }
};

// Server rejects encrypted payloads larger than this (api/core/domain/validation.go).
const MAX_ENCRYPTED_LENGTH = 48 * 1024;

const SAVE_ERRORS = {
  413: 'Message is too large.',
  429: 'Too many requests. Please wait a moment and try again.',
};

const ApiService = {
  async saveMessage(id, encryptedContent, expiresIn) {
    const res = await fetch('/api/message', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id, encrypted: encryptedContent, expiresIn })
    });
    if (!res.ok) {
      const err = new Error(`save failed with status ${res.status}`);
      err.userMessage = SAVE_ERRORS[res.status];
      throw err;
    }
    return res.json();
  },

  // Errors carry `retryable` when the server was not reached or failed before
  // deleting the message (network error, 429, 5xx): the link is still valid.
  async fetchMessage(id) {
    let res;
    try {
      res = await fetch(`/api/message/${id}`);
    } catch (e) {
      const err = new Error('network error');
      err.retryable = true;
      throw err;
    }
    if (res.status === 404) {
      const err = new Error('Message not found or already read');
      err.notFound = true;
      throw err;
    }
    if (!res.ok) {
      const err = new Error(`fetch failed with status ${res.status}`);
      err.retryable = true;
      err.userMessage = res.status === 429
        ? 'Too many requests. Please wait a moment and try again.'
        : undefined;
      throw err;
    }
    const data = await res.json();
    return JSON.parse(data.encrypted);
  }
};

// --- DOM & UI Controllers ---

const UI = {
  showToast(message, duration = 3000) {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.innerText = message;
    container.appendChild(toast);
    
    setTimeout(() => {
      toast.style.animation = 'toastIn 0.3s cubic-bezier(0.16, 1, 0.3, 1) reverse forwards';
      setTimeout(() => toast.remove(), 300);
    }, duration);
  },

  showModalText(text) {
    this._resetModal();
    const modal = document.getElementById('modal');
    const msgElement = document.getElementById('modal-message');
    msgElement.innerText = text;
    msgElement.classList.remove('hidden');
    modal.classList.remove('hidden');
  },

  showModalDecrypted(text) {
    this._resetModal();
    const modal = document.getElementById('modal');
    const alert = document.getElementById('modal-alert');
    const wrapper = document.getElementById('decrypted-wrapper');
    const textarea = document.getElementById('decrypted-message');
    
    alert.classList.remove('hidden');
    wrapper.classList.remove('hidden');
    textarea.value = text;
    modal.classList.remove('hidden');
  },

  showModalReveal() {
    this._resetModal();
    document.getElementById('reveal-wrapper').classList.remove('hidden');
    document.getElementById('modal').classList.remove('hidden');
  },

  hideModal() {
    document.getElementById('modal').classList.add('hidden');
  },

  _resetModal() {
    document.getElementById('modal-message').classList.add('hidden');
    document.getElementById('modal-alert').classList.add('hidden');
    document.getElementById('reveal-wrapper').classList.add('hidden');
    document.getElementById('decrypted-wrapper').classList.add('hidden');
    document.getElementById('decrypted-message').value = '';
  },

  copyToClipboard(elementId, btnElement) {
    const el = document.getElementById(elementId);
    el.select();
    navigator.clipboard.writeText(el.value).then(() => {
      const originalText = btnElement.innerText;
      btnElement.innerText = 'Copied!';
      setTimeout(() => btnElement.innerText = originalText, 1500);
      this.showToast('Copied to clipboard');
    }).catch(() => this.showToast('Failed to copy'));
  }
};

// --- Main Interactions ---

async function handleEncryptAndSend() {
  const msgInput = document.getElementById('message');
  const message = msgInput.value.trim();
  
  if (!message) {
    UI.showToast('Please type a message first.');
    return;
  }

  if (!isCryptoAvailable()) {
    UI.showModalText('This page requires HTTPS (a secure context) to encrypt messages. Please open the site over https://.');
    return;
  }

  try {
    const { key, iv } = await CryptoService.generateKeyAndIV();
    const encryptedBase64 = await CryptoService.encrypt(message, key, iv);
    const { keyBase64, ivBase64 } = await CryptoService.exportKeyAndIV(key, iv);
    
    const id = crypto.randomUUID();
    const payload = JSON.stringify({ ciphertext: encryptedBase64, iv: ivBase64 });
    if (payload.length > MAX_ENCRYPTED_LENGTH) {
      UI.showToast('Message is too large.');
      return;
    }

    await ApiService.saveMessage(id, payload, Number(document.getElementById('expires-in').value));
    
    const link = `${location.origin}/#${id}.${keyBase64}`;
    document.getElementById('link-input').value = link;
    document.getElementById('result').classList.remove('hidden');
    msgInput.value = ''; // clear input
    
    UI.showToast('Secure link generated!');
  } catch (err) {
    console.error(err);
    UI.showToast(err.userMessage || 'Failed to save message.');
  }
}

function isCryptoAvailable() {
  return window.isSecureContext && !!(window.crypto && crypto.subtle);
}

function handleOnLoadDecrypt() {
  const rawHash = location.hash.slice(1);
  if (!rawHash) return;

  // Be lenient with common link mangling: %3D for '=' and one trailing
  // punctuation character added by chat apps.
  const cleaned = rawHash.replace(/%3D/gi, '=').replace(/[.,)\]>]$/, '');
  const match = FRAGMENT_RE.exec(cleaned);
  if (!match) {
    UI.showModalText('This link is incomplete or malformed.');
    return;
  }
  const [, id, keyBase64] = match;

  if (!isCryptoAvailable()) {
    UI.showModalText('This page requires HTTPS (a secure context) to decrypt messages. Please open the link over https://.');
    return;
  }

  // Do not fetch yet: the API deletes the message on read, and link-preview
  // bots / scanners that run JS must not be able to burn it.
  UI.showModalReveal();
  const btn = document.getElementById('reveal-btn');
  btn.addEventListener('click', () => revealMessage(id, keyBase64, btn));
}

async function revealMessage(id, keyBase64, btn) {
  if (btn.disabled) return;
  btn.disabled = true;
  let clearFragment = true;
  try {
    const payload = await ApiService.fetchMessage(id);
    let decryptedText;
    try {
      decryptedText = await CryptoService.decrypt(payload.ciphertext, keyBase64, payload.iv);
    } catch (e) {
      console.error(e);
      UI.showModalText('Message is corrupted or the key is wrong.');
      return;
    }
    UI.showModalDecrypted(decryptedText);
  } catch (err) {
    console.error(err);
    if (err.retryable) {
      // Message still exists on the server: keep the fragment and allow retry.
      clearFragment = false;
      btn.disabled = false;
      document.getElementById('reveal-prompt').innerText =
        err.userMessage || 'Could not reach the server, try again.';
    } else if (err.notFound) {
      UI.showModalText('Message not found or already read.');
    } else {
      UI.showModalText('Message is corrupted or the key is wrong.');
    }
  } finally {
    // Clear the fragment so reloading doesn't show confusing errors
    if (clearFragment) history.replaceState(null, null, ' ');
  }
}

// --- Theme Management ---

function setupTheme() {
  const toggleBtn = document.getElementById('toggle-theme-btn');
  const iconSun = document.getElementById('theme-icon-sun');
  const iconMoon = document.getElementById('theme-icon-moon');
  
  const updateIcons = (isDark) => {
    if (isDark) {
      iconSun.classList.remove('hidden');
      iconMoon.classList.add('hidden');
    } else {
      iconMoon.classList.remove('hidden');
      iconSun.classList.add('hidden');
    }
  };

  const savedTheme = localStorage.getItem('theme') || 'dark'; // dark default
  if (savedTheme === 'light') {
    document.body.classList.remove('dark-mode');
    updateIcons(false);
  } else {
    document.body.classList.add('dark-mode');
    updateIcons(true);
  }

  toggleBtn.addEventListener('click', () => {
    const isDark = document.body.classList.toggle('dark-mode');
    localStorage.setItem('theme', isDark ? 'dark' : 'light');
    updateIcons(isDark);
  });
}

// --- Initialization ---

document.addEventListener('DOMContentLoaded', () => {
  setupTheme();
  
  document.getElementById('encrypt-btn').addEventListener('click', handleEncryptAndSend);
  document.getElementById('modal-close').addEventListener('click', () => UI.hideModal());
  document.getElementById('link-input').addEventListener('click', (e) => e.target.select());
  document.getElementById('copy-link-btn').addEventListener('click', function() {
    UI.copyToClipboard('link-input', this);
  });
  document.getElementById('decrypted-copy-btn').addEventListener('click', function() {
    UI.copyToClipboard('decrypted-message', this);
  });

  handleOnLoadDecrypt();
});