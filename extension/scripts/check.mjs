// Lints the extension without dependencies: manifest shape, minimal
// permissions, every referenced file present, JavaScript syntax (#65).
import { readFileSync, existsSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import { files } from './files.mjs';

const fail = (msg) => {
  console.error(`✗ ${msg}`);
  process.exitCode = 1;
};

const manifest = JSON.parse(readFileSync('manifest.json', 'utf8'));
if (manifest.manifest_version !== 3) fail('manifest_version must be 3');
if (!/^\d+\.\d+\.\d+$/.test(manifest.version)) fail('version must be x.y.z');
const allowed = new Set(['storage', 'contextMenus', 'activeTab']);
for (const p of manifest.permissions ?? []) if (!allowed.has(p)) fail(`permission not allowed: ${p}`);
for (const key of ['host_permissions', 'optional_host_permissions', 'content_scripts', 'web_accessible_resources']) {
  if (manifest[key]) fail(`${key} is not allowed: the extension only opens the configured address`);
}
const referenced = [
  manifest.chrome_url_overrides?.newtab,
  manifest.options_ui?.page,
  manifest.background?.service_worker,
  ...Object.values(manifest.icons ?? {}),
].filter(Boolean);
for (const f of [...referenced, ...files]) if (!existsSync(f)) fail(`missing file: ${f}`);
for (const f of files.filter((f) => f.endsWith('.js'))) {
  try {
    execFileSync(process.execPath, ['--check', f], { stdio: 'pipe' });
  } catch (err) {
    fail(`${f}: ${String(err.stderr).trim()}`);
  }
}
// Remote code is not allowed in MV3 and not wanted here.
for (const f of files.filter((f) => f.endsWith('.html'))) {
  if (/<script[^>]+src=["']https?:/i.test(readFileSync(f, 'utf8'))) fail(`${f} loads a remote script`);
}
if (!process.exitCode) console.log(`✓ manifest and ${files.length} files OK`);
