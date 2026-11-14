import { normaliseBase } from './url.js';
import { loadBase, saveBase } from './storage.js';

const input = document.getElementById('base');
const status = document.getElementById('status');

loadBase().then((base) => (input.value = base));

document.getElementById('form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const base = normaliseBase(input.value);
  if (!base) {
    status.textContent = 'Enter an http(s) address such as http://homeserver:8090 (no user name, query or #).';
    return;
  }
  await saveBase(base);
  input.value = base;
  status.textContent = 'Saved. New tabs now open your dashboard.';
});
