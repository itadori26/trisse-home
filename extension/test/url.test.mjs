import { test } from 'node:test';
import assert from 'node:assert/strict';
import { captureUrl, dashboardUrl, healthUrl, normaliseBase } from '../src/url.js';

test('normalises the dashboard address', () => {
  assert.equal(normaliseBase(' http://homeserver:8090/ '), 'http://homeserver:8090');
  assert.equal(normaliseBase('https://home.example/start/'), 'https://home.example/start');
  for (const bad of ['', 'homeserver:8090', 'ftp://x', 'javascript:alert(1)', 'http://u:p@x', 'http://x/?a=1', 'http://x/#y']) {
    assert.equal(normaliseBase(bad), null, bad);
  }
});

test('builds the dashboard, health and capture addresses', () => {
  assert.equal(dashboardUrl('http://h:8090'), 'http://h:8090/?focus=search');
  assert.equal(healthUrl('http://h:8090'), 'http://h:8090/healthz');
  const u = new URL(captureUrl('http://h:8090', { url: 'https://example.org/a?b=1', title: 'A page', selection: 'quote' }));
  assert.equal(u.pathname, '/capture');
  assert.equal(u.searchParams.get('url'), 'https://example.org/a?b=1');
  assert.equal(u.searchParams.get('title'), 'A page');
  assert.equal(u.searchParams.get('text'), 'quote');
});

test('does not hand over browser-internal pages', () => {
  const u = new URL(captureUrl('http://h', { url: 'chrome://settings', title: 'Settings' }));
  assert.equal(u.searchParams.has('url'), false);
  assert.equal(u.searchParams.get('title'), 'Settings');
});
