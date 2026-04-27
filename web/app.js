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
  const createForm = document.getElementById('createForm');
  const statusForm = document.getElementById('statusForm');

  createForm.addEventListener('submit', handleCreateNotification);
  statusForm.addEventListener('submit', handleGetStatus);
}

function showNotification(message, type = 'info') {
  const notification = document.getElementById('notification');
  notification.textContent = message;
  notification.className = `notification ${type}`;
  notification.classList.remove('hidden', 'hide-animation');

  setTimeout(() => {
    notification.classList.add('hide-animation');
    setTimeout(() => {
      notification.classList.add('hidden');
    }, 300);
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
    const payload = {
      channel,
      recipient,
      message,
      send_at: sendAtDate.toISOString()
    };

    const res = await fetch(`${API_URL}/notify`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    if (!res.ok) {
      const errText = await res.text();
      showNotification(`Ошибка: ${errText}`, 'error');
      return;
    }

    const data = await res.json();

    notifications.unshift({
      id: data.id,
      channel,
      recipient,
      message,
      send_at: sendAtDate.toISOString()
    });

    displayNotificationHistory();

    document.getElementById('createForm').reset();
    initializePage();

    showNotification('Уведомление создано', 'success');

  } catch (err) {
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
      const errText = await res.text();
      showNotification(`Ошибка: ${errText}`, 'error');
      return;
    }

    const data = await res.json();

    document.getElementById('result').innerHTML = `
      <div><b>ID:</b> ${data.id}</div>
      <div><b>Status:</b> ${data.status}</div>
    `;

    document.getElementById('result-container').classList.remove('hidden');

  } catch (err) {
    showNotification('Ошибка сети', 'error');
  }
}

function loadNotificationsFromStorage() {
  const stored = localStorage.getItem('notifications');
  if (stored) {
    notifications.push(...JSON.parse(stored));
  }
}

function displayNotificationHistory() {
  const container = document.getElementById('history-container');

  if (notifications.length === 0) {
    container.innerHTML = '<p>Нет уведомлений</p>';
    return;
  }

  container.innerHTML = notifications.map((n) => {
    const t = formatSendAt(n.send_at);
    return `
  <div class="history-item">

    <div class="history-left">
      <div class="history-id">${n.id}</div>
      <div class="history-recipient">${n.channel} → ${n.recipient}</div>
      <div class="history-message collapsed">
        ${n.message}
      </div>
    </div>

    <div class="sendat-badge">
      <div class="sendat-label">Send at</div>
      <div class="sendat-date">${t.date}</div>
      <div class="sendat-time">${t.time}</div>
    </div>

  </div>
`;
  }).join('');

  document.querySelectorAll('.history-message').forEach(el => {
    el.addEventListener('click', () => {
      el.classList.toggle('collapsed');
    });
  });
}

function formatSendAt(dateStr) {
  const d = new Date(dateStr);

  const pad = (n) => n.toString().padStart(2, '0');

  const day = pad(d.getDate());
  const month = pad(d.getMonth() + 1);
  const year = d.getFullYear();

  let hours = d.getHours();
  const minutes = pad(d.getMinutes());

  const ampm = hours >= 12 ? 'PM' : 'AM';
  hours = hours % 12;
  hours = hours ? hours : 12;

  return {
    date: `${day}.${month}.${year}`,
    time: `${hours}:${minutes} ${ampm}`
  };
}
