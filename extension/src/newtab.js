import { dashboardUrl, healthUrl } from './url.js';
import { loadBase } from './storage.js';

// The new tab: go to the dashboard, or explain what is missing. The health
// check is a no-cors request, so the extension needs no host permission;
// an opaque answer is enough to know the server is there.
const message = document.getElementById('message');
const retry = document.getElementById('retry');

function show(text) {
  message.textContent = text;
  document.body.classList.add('problem');
}

async function go() {
  const base = await loadBase();
  if (!base) {
    show('Set the address of your trisse-home dashboard in the options.');
    return;
  }
  try {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), 5000);
    await fetch(healthUrl(base), { mode: 'no-cors', cache: 'no-store', signal: controller.signal });
    clearTimeout(timer);
    location.replace(dashboardUrl(base));
  } catch {
    show(`${base} is not reachable right now.`);
    retry.hidden = false;
  }
}

retry.addEventListener('click', () => location.reload());
document.getElementById('options').addEventListener('click', (e) => {
  e.preventDefault();
  (globalThis.browser ?? globalThis.chrome).runtime.openOptionsPage();
});
go();
