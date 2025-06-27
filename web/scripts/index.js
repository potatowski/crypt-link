async function encryptAndSend() {
    const msg = document.getElementById('message').value.trim();
    const minutes = parseInt(document.getElementById('validity').value, 10) || 10;
    if (isNaN(minutes) || minutes <= 0 || minutes > 1440) {
        return showModal('Digite um valor válido para a validade (1 a 1440 minutos)');
    }

    if (!msg) {
        return showModal('Digite uma mensagem');
    }

    const key = await crypto.subtle.generateKey({ name: 'AES-GCM', length: 256 }, true, ['encrypt', 'decrypt']);
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const encodedMsg = new TextEncoder().encode(msg);
    const encrypted = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, key, encodedMsg);
    const encryptedBase64 = btoa(String.fromCharCode(...new Uint8Array(encrypted)));

    const rawKey = await crypto.subtle.exportKey('raw', key);
    const keyBase64 = btoa(String.fromCharCode(...new Uint8Array(rawKey)));
    const ivBase64 = btoa(String.fromCharCode(...iv));

    const id = crypto.randomUUID();
    const payload = {
        id,
        encrypted: JSON.stringify({ ciphertext: encryptedBase64, iv: ivBase64 }),
        validityTime: minutes
    };

    const res = await fetch('/api/message', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
    });

    if (!res.ok) {
        showModal('Erro ao salvar a mensagem');
        return;
    }

    const link = `${location.origin}/#${id}.${keyBase64}`;
    document.getElementById('link-input').value = link;

    document.getElementById('result').classList.remove('hidden');

}

function copyBtn(elementId) {
    const input = document.getElementById(elementId);
    const button = input.nextElementSibling;

    input.select();
    navigator.clipboard.writeText(input.value)
        .then(() => {
            const originalText = button.innerText;
            button.innerText = 'Copiado!';
            button.disabled = true;

            setTimeout(() => {
                button.innerText = originalText;
                button.disabled = false;
            }, 1000);
        })
        .catch(() => {
            showModal('Erro ao copiar texto');
        });
}

async function tryDecrypt() {
    if (!location.hash.includes('#')) {
        return;
    }

    const [id, keyBase64] = location.hash.slice(1).split('.', 2);
    if (!id || !keyBase64) return;

    const res = await fetch(`/api/message/${id}`);
    if (!res.ok) {
        showModal('Mensagem não encontrada ou já foi lida.');
        return;
    }

    const data = await res.json();
    const payload = JSON.parse(data.encrypted);
    const ciphertext = Uint8Array.from(atob(payload.ciphertext), c => c.charCodeAt(0));
    const iv = Uint8Array.from(atob(payload.iv), c => c.charCodeAt(0));
    const expiresAt = new Date(payload.expiresAt);

    if (new Date() > expiresAt) {
        showModal('Mensagem não encontrada ou já foi lida.')
        return;
    }

    const rawKey = Uint8Array.from(atob(keyBase64), c => c.charCodeAt(0));
    const cryptoKey = await crypto.subtle.importKey('raw', rawKey, { name: 'AES-GCM' }, false, ['decrypt']);

    try {
        const decrypted = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, cryptoKey, ciphertext);
        const msg = new TextDecoder().decode(decrypted);
        showModal(msg, true);
    } catch (e) {
        showModal('Falha ao descriptografar.');
    }
}

function showModal(message, sucess = false) {
    const modal = document.getElementById('modal');
    const content = document.getElementById('modal-message');
    const alert = document.getElementById('modal-alert');
    if (sucess) {
        const decryptedMessage = document.getElementById('decrypted-message');
        const copyButton = document.getElementById('decrypted-copy-btn');
        alert.classList.remove('hidden');
        alert.innerHTML = `<i class="fa-solid fa-check"></i> ⚠️ LEMBRE-SE: Essa mensagem foi deletada e não pode ser lida novamente.`;
        decryptedMessage.innerHTML = message;
        decryptedMessage.classList.remove('hidden');
        copyButton.classList.remove('hidden');
        decryptedMessage.focus();

        content.classList.add('hidden');
    } else {
        content.innerText = message;
        content.classList.remove('hidden');
    }
    modal.classList.remove('hidden');
}

function closeModal() {
    document.getElementById('modal').classList.add('hidden');
    const content = document.getElementById('modal-message');
    const alert = document.getElementById('modal-alert');
    const decryptedMessage = document.getElementById('decrypted-message');
    const copyButton = document.getElementById('decrypted-copy-btn');

    content.classList.add('hidden');
    alert.classList.add('hidden');
    decryptedMessage.classList.add('hidden');
    copyButton.classList.add('hidden');
    decryptedMessage.innerHTML = '';
    content.innerHTML = '';
    alert.innerHTML = '';
}

const toggleBtn = document.getElementById('toggle-theme-btn');

toggleBtn.addEventListener('click', () => {
    document.body.classList.toggle('dark-mode');

    if (document.body.classList.contains('dark-mode')) {
        localStorage.setItem('theme', 'dark');
        toggleBtn.textContent = '🌞';
    } else {
        localStorage.setItem('theme', 'light');
        toggleBtn.textContent = '🌙';
    }
});

window.addEventListener('DOMContentLoaded', () => {
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme === 'dark') {
        document.body.classList.add('dark-mode');
        toggleBtn.textContent = '🌞';
    } else {
        toggleBtn.textContent = '🌙';
    }
});

window.onload = tryDecrypt;