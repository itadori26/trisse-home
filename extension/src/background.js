import { captureUrl } from './url.js';
import { api, loadBase } from './storage.js';

// "Save page to trisse-home": toolbar button and context menu open the quick
// capture page (#66) with the page in a small window. activeTab grants the
// URL and title of the clicked tab only, so no host permissions are needed.
const MENU = 'save-to-trisse-home';

api.runtime.onInstalled.addListener(() => {
  api.contextMenus.create({ id: MENU, title: 'Save page to trisse-home', contexts: ['page', 'link', 'selection'] });
});

async function save(tab, info = {}) {
  const base = await loadBase();
  if (!base) {
    api.runtime.openOptionsPage();
    return;
  }
  const page = info.linkUrl ? { url: info.linkUrl, title: info.linkText || '' } : { url: tab?.url, title: tab?.title };
  const url = captureUrl(base, { ...page, selection: info.selectionText });
  api.windows.create({ url, type: 'popup', width: 520, height: 640 });
}

api.action.onClicked.addListener((tab) => save(tab));
api.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId === MENU) save(tab, info);
});
