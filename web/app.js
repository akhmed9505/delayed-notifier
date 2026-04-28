const API_URL = window.API_URL || "http://localhost:8080";
const notifications = [];

document.addEventListener('DOMContentLoaded', function () {
  initializePage();
  setupFormListeners();
  loadNotificationsFromStorage();
  displayNotificationHistory();
});

function initializePage() {
  const sendAtInput = document.getElementById('send_at');
  const now = new Date();
  now.setMinutes(now.getMinutes() + 1);
  now.setSeconds(0);
  now.setMilliseconds(0);
  sendAtInput.value = now.toISOString().slice(0, 16);
}

function setupFormListeners() {
  document.getElementById('createForm')
    .addEventListener('submit', handleCreateNotification);

  document.getElementById('statusForm')
    .addEventListener('submit', handleGetStatus);
}

function showNotification(message, type = 'info') {
  const notification = document.getElementById('notification');
  notification.textContent = message;
  notification.className = `notification ${type}`;
  notification.classList.remove('hidden', 'hide-animation');

  setTimeout(() => {
    notification.classList.add('hide-animation');
    setTimeout(() => notification.classList.add('hidden'), 300);
  }, 3000);
}

async function handleCreateNotification(e) {
  e.preventDefault();

  const rawDate = document.getElementById('send_at').value;
  const channel = document.getElementById('channel').value;
  const recipient = document.getElementById('recipient').value;
  const message = document.getElementById('message').value;

  if (!rawDate || !recipient || !message) {
    showNotification('Заполните все поля', 'error');
    return;
  }

  const sendAtDate = new Date(rawDate);

  if (sendAtDate <= new Date()) {
    showNotification('Дата должна быть в будущем', 'error');
    return;
  }

  try {
    const res = await fetch(`${API_URL}/notify`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        channel,
        recipient,
        message,
        send_at: sendAtDate.toISOString()
      })
    });

    if (!res.ok) {
      showNotification(await res.text(), 'error');
      return;
    }

    const data = await res.json();

    notifications.unshift({
      id: data.id,
      channel,
      recipient,
      message,
      send_at: sendAtDate.toISOString(),
      status: 'pending'
    });

    displayNotificationHistory();

    e.target.reset();
    initializePage();

    showNotification('Уведомление создано', 'success');

  } catch {
    showNotification('Ошибка сети', 'error');
  }
}

async function handleGetStatus(e) {
  e.preventDefault();

  const id = document.getElementById('checkId').value;
  if (!id) {
    showNotification('Введите ID', 'error');
    return;
  }

  try {
    const res = await fetch(`${API_URL}/notify/${id}`);

    if (!res.ok) {
      showNotification(await res.text(), 'error');
      return;
    }

    const data = await res.json();

    const container = document.getElementById('result');

    const isPending = data.status === 'pending';

    container.innerHTML = `
      <div><b>ID:</b> ${data.id}</div>
      <div><b>Status:</b> ${data.status}</div>

      ${isPending ? `
        <button id="cancelBtn" data-id="${data.id}" class="cancel-btn">
          Отменить уведомление
        </button>
      ` : ''}
    `;

    document.getElementById('result-container').classList.remove('hidden');

    const btn = document.getElementById('cancelBtn');

    if (!btn) return;

    btn.addEventListener('click', async () => {
      const id = btn.dataset.id;

      const check = await fetch(`${API_URL}/notify/${id}`);
      const latest = await check.json();

      if (latest.status !== 'pending') {
        showNotification('Уведомление уже отправлено или отменено', 'error');

        container.innerHTML = `
          <div><b>ID:</b> ${latest.id}</div>
          <div><b>Status:</b> ${latest.status}</div>
        `;

        return;
      }

      try {
        const res = await fetch(`${API_URL}/notify/${id}`, {
          method: 'DELETE'
        });

        if (!res.ok && res.status !== 204) {
          showNotification(await res.text(), 'error');
          return;
        }

        showNotification('Уведомление отменено', 'success');

        handleGetStatus(e);

      } catch {
        showNotification('Ошибка сети', 'error');
      }
    });

  } catch {
    showNotification('Ошибка сети', 'error');
  }
}

function loadNotificationsFromStorage() {
  const stored = localStorage.getItem('notifications');
  if (stored) notifications.push(...JSON.parse(stored));
}

function displayNotificationHistory() {
  const container = document.getElementById('history-container');

  if (!notifications.length) {
    container.innerHTML = '<p>Нет уведомлений</p>';
    return;
  }

  container.innerHTML = notifications.map(n => {
    const t = formatSendAt(n.send_at);

    return `
      <div class="history-item">
        <div class="history-left">
          <div class="history-id">${n.id}</div>
          <div>${n.channel} → ${n.recipient}</div>
          <div>${n.message}</div>
        </div>

        <div>
          <div>${t.date}</div>
          <div>${t.time}</div>
        </div>
      </div>
    `;
  }).join('');
}

function formatSendAt(dateStr) {
  const d = new Date(dateStr);
  const pad = n => n.toString().padStart(2, '0');

  const day = pad(d.getDate());
  const month = pad(d.getMonth() + 1);
  const year = d.getFullYear();

  let h = d.getHours();
  const m = pad(d.getMinutes());

  const ampm = h >= 12 ? 'PM' : 'AM';
  h = h % 12 || 12;

  return {
    date: `${day}.${month}.${year}`,
    time: `${h}:${m} ${ampm}`
  };
}
