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
    return btoa(String.fromCharCode(...new Uint8Array(encrypted)));
  },

  async exportKeyAndIV(key, iv) {
    const rawKey = await crypto.subtle.exportKey('raw', key);
    const keyBase64 = btoa(String.fromCharCode(...new Uint8Array(rawKey)));
    const ivBase64 = btoa(String.fromCharCode(...iv));
    return { keyBase64, ivBase64 };
  },

  async decrypt(ciphertext, keyBase64, ivBase64) {
    const rawKey = Uint8Array.from(atob(keyBase64), c => c.charCodeAt(0));
    const cryptoKey = await crypto.subtle.importKey('raw', rawKey, { name: 'AES-GCM' }, false, ['decrypt']);
    
    const cipherArray = Uint8Array.from(atob(ciphertext), c => c.charCodeAt(0));
    const ivArray = Uint8Array.from(atob(ivBase64), c => c.charCodeAt(0));
    
    const decrypted = await crypto.subtle.decrypt({ name: 'AES-GCM', iv: ivArray }, cryptoKey, cipherArray);
    return new TextDecoder().decode(decrypted);
  }
};

const ApiService = {
  async saveMessage(id, encryptedContent) {
    const res = await fetch('/api/message', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id, encrypted: encryptedContent })
    });
    if (!res.ok) throw new Error('API failed to save message');
    return res.json();
  },

  async fetchMessage(id) {
    const res = await fetch(`/api/message/${id}`);
    if (!res.ok) throw new Error('Message not found or already read');
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

  hideModal() {
    document.getElementById('modal').classList.add('hidden');
  },

  _resetModal() {
    document.getElementById('modal-message').classList.add('hidden');
    document.getElementById('modal-alert').classList.add('hidden');
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

  try {
    const { key, iv } = await CryptoService.generateKeyAndIV();
    const encryptedBase64 = await CryptoService.encrypt(message, key, iv);
    const { keyBase64, ivBase64 } = await CryptoService.exportKeyAndIV(key, iv);
    
    const id = crypto.randomUUID();
    const payload = JSON.stringify({ ciphertext: encryptedBase64, iv: ivBase64 });
    
    await ApiService.saveMessage(id, payload);
    
    const link = `${location.origin}/#${id}.${keyBase64}`;
    document.getElementById('link-input').value = link;
    document.getElementById('result').classList.remove('hidden');
    msgInput.value = ''; // clear input
    
    UI.showToast('Secure link generated!');
  } catch (err) {
    console.error(err);
    UI.showToast('Failed to save message.');
  }
}

async function handleOnLoadDecrypt() {
  if (!location.hash.includes('#')) return;
  
  const [id, keyBase64] = location.hash.slice(1).split('.', 2);
  if (!id || !keyBase64) return;

  try {
    const payload = await ApiService.fetchMessage(id);
    const expiresAt = new Date(payload.expiresAt); // Note: server currently doesn't send expiresAt inside encrypted string explicitly anymore unless configured that way? Wait, earlier code packed expiresAt in JSON, let's just assume we decrypt what we have or rely on backend to 404 if expired.
    
    const decryptedText = await CryptoService.decrypt(payload.ciphertext, keyBase64, payload.iv);
    UI.showModalDecrypted(decryptedText);
    
  } catch (err) {
    console.error(err);
    UI.showModalText('Message not found, already read, or corrupted.');
  } finally {
    // Clear the fragment so reloading doesn't show confusing errors
    history.replaceState(null, null, ' ');
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
  document.getElementById('copy-link-btn').addEventListener('click', function() {
    UI.copyToClipboard('link-input', this);
  });
  document.getElementById('decrypted-copy-btn').addEventListener('click', function() {
    UI.copyToClipboard('decrypted-message', this);
  });

  handleOnLoadDecrypt();
});