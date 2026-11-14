// The dashboard address is the only setting; it lives in storage.sync so it
// follows the browser profile. No other data is stored.
const KEY = 'baseUrl';
const api = globalThis.browser ?? globalThis.chrome;

export async function loadBase() {
  const got = await api.storage.sync.get(KEY);
  return typeof got[KEY] === 'string' ? got[KEY] : '';
}

export async function saveBase(base) {
  await api.storage.sync.set({ [KEY]: base });
}

export { api };
