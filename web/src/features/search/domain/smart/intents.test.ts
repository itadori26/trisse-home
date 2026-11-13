import { describe, expect, it } from 'vitest';
import { type Entity, interpret, rank } from './intent';
import { INTENTS } from './intents';

const entities: Entity[] = [
  { kind: 'container', id: 'c1', name: 'jellyfin' },
  { kind: 'container', id: 'c2', name: 'nextcloud-app' },
  { kind: 'container', id: 'c3', name: 'home-assistant' },
  { kind: 'container', id: 'c4', name: 'pihole' },
  { kind: 'container', id: 'c5', name: 'postgres' },
  { kind: 'service', id: 's1', name: 'Router' },
  { kind: 'service', id: 's2', name: 'NAS' },
  { kind: 'service', id: 's3', name: 'Nextcloud' },
  { kind: 'service', id: 's4', name: 'Grafana' },
];

// Thursday, 12 November 2026
const today = new Date(2026, 10, 12, 10, 0);

const top = (query: string) => rank(interpret(query, INTENTS, entities, today))[0];

/** Regression table (#156): real queries → expected best intent; null = no answer. */
const table: [string, string | null][] = [
  // system – CPU, memory, disk, load, uptime, temperature
  ['cpu', 'system.cpu'],
  ['CPU Auslastung', 'system.cpu'],
  ['wie hoch ist die cpu', 'system.cpu'],
  ['prozessor', 'system.cpu'],
  ['processor usage', 'system.cpu'],
  ['/cpu', 'system.cpu'],
  ['ram', 'system.memory'],
  ['arbeitsspeicher', 'system.memory'],
  ['wie viel ram ist frei', 'system.memory'],
  ['memory usage', 'system.memory'],
  ['/ram', 'system.memory'],
  ['wie voll ist die platte', 'system.disk'],
  ['festplatte', 'system.disk'],
  ['speicherplatz', 'system.disk'],
  ['disk', 'system.disk'],
  ['is the disk full', 'system.disk'],
  ['how much storage is free', 'system.disk'],
  ['plate voll', 'system.disk'], // typo
  ['platte voll', 'system.disk'],
  ['/disk', 'system.disk'],
  ['load', 'system.load'],
  ['systemlast', 'system.load'],
  ['load average', 'system.load'],
  ['uptime', 'system.uptime'],
  ['laufzeit', 'system.uptime'],
  ['seit wann läuft der server', 'system.uptime'],
  ['wann war der letzte neustart', 'system.uptime'],
  ['temperatur server', 'system.temperature'],
  ['wie heiß ist die cpu', 'system.cpu'],
  // docker
  ['docker', 'docker.overview'],
  ['welche container laufen', 'docker.overview'],
  ['containers', 'docker.overview'],
  ['jellyfin', 'docker.status'],
  ['jelyfin', 'docker.status'], // typo
  ['jel', 'docker.status'], // prefix while typing
  ['läuft jellyfin', 'docker.status'],
  ['pihole status', 'docker.status'],
  ['home assistant', 'docker.status'],
  ['/docker postgres', 'docker.status'],
  ['jellyfin neu starten', 'docker.restart'],
  ['restart jellyfin', 'docker.restart'],
  ['starte pihole neu', 'docker.restart'],
  ['stop postgres', 'docker.restart'],
  ['/docker restart pihole', 'docker.restart'],
  ['neu starten', 'docker.restart'], // asks which container
  // services
  ['dienste', 'services.overview'],
  ['welche dienste sind offline', 'services.overview'],
  ['what is down', 'services.overview'],
  ['router', 'services.status'],
  ['ist das nas erreichbar', 'services.status'],
  ['grafana ping', 'services.status'],
  ['/service router', 'services.status'],
  // weather
  ['wetter', 'weather.forecast'],
  ['weather', 'weather.forecast'],
  ['wtter', 'weather.forecast'], // typo
  ['we', 'weather.forecast'], // prefix while typing
  ['wetter morgen', 'weather.forecast'],
  ['wetter berlin morgen', 'weather.forecast'],
  ['regnet es morgen', 'weather.forecast'],
  ['brauche ich einen schirm', 'weather.forecast'],
  ['weather in london tomorrow', 'weather.forecast'],
  ['/wetter hamburg', 'weather.forecast'],
  ['/weather', 'weather.forecast'],
  // no answer
  ['', null],
  ['hallo', null],
  ['xyz', null],
  ['der die das', null],
];

describe('regression table (#156)', () => {
  it('has at least 50 real queries', () => expect(table.length).toBeGreaterThanOrEqual(50));
  it.each(table)('%s → %s', (query, want) => {
    expect(top(query)?.intent.id ?? null).toBe(want);
  });
});

describe('slots (#158)', () => {
  it('fills place and day for the weather', () => {
    const m = top('wetter berlin morgen');
    expect(m?.slots['place']).toEqual({ type: 'place', text: 'berlin' });
    expect(m?.slots['day']).toEqual({ type: 'day', offset: 1 });
    expect(top('weather in london on saturday')?.slots['day']).toEqual({ type: 'day', offset: 2 });
  });
  it('fills container and action', () => {
    const m = top('jellyfin neu starten');
    expect(m?.slots['action']).toEqual({ type: 'action', action: 'restart' });
    expect(m?.slots['container']).toMatchObject({ entity: { name: 'jellyfin' } });
    expect(top('stop postgres')?.slots['action']).toEqual({ type: 'action', action: 'stop' });
  });
  it('asks back for a missing container with suggestions', () => {
    const m = top('neu starten');
    expect(m?.missing).toEqual(['container']);
    expect(m?.suggestions.map((e) => e.name)).toContain('jellyfin');
  });
  it('does not offer an action for a bare name', () => {
    const ids = rank(interpret('jellyfin', INTENTS, entities, today)).map((m) => m.intent.id);
    expect(ids).not.toContain('docker.restart');
  });
  it('resolves service names, not containers, for service intents', () => {
    expect(top('nas')?.slots['service']).toMatchObject({ entity: { name: 'NAS' } });
  });
});
