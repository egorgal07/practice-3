const list = document.getElementById('items');
const errorBox = document.getElementById('error');

async function load() {
  try {
    const res = await fetch('/api/items');
    if (!res.ok) throw new Error(await res.text());
    const items = await res.json();
    list.innerHTML = '';
    for (const it of items) {
      const li = document.createElement('li');
      li.textContent = it.name;
      list.appendChild(li);
    }
    errorBox.textContent = '';
  } catch (e) {
    errorBox.textContent = 'Ошибка: ' + e.message;
  }
}

load();