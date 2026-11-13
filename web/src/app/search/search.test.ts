import { describe, expect, it, vi } from 'vitest';
import type { Answer, ListAnswer } from '../../features/search/domain/smart/answer';
import { type Entity, interpret, type Match, rank } from '../../features/search/domain/smart/intent';
import type { SearchSource } from '../../features/search/usecase/smartSearch';
import type { Container } from '../container';
import { allIntents, contributions } from '../searchSources';
import { eventsOnDay } from './calendar';

const INTENTS = allIntents();
const entities: Entity[] = [
  { kind: 'container', id: 'c1', name: 'jellyfin' },
  { kind: 'service', id: 's2', name: 'NAS' },
  { kind: 'job', id: 'j1', name: 'restic-offsite' },
  { kind: 'job', id: 'j2', name: 'Paperless export' },
];
const today = new Date(2026, 10, 12, 10, 0); // Thursday
const top = (query: string) => rank(interpret(query, INTENTS, entities, today))[0];

describe('module search contributions (#161)', () => {
  it('are collected, with unique intent ids that point at their own source', () => {
    expect(contributions.map((c) => c.module)).toEqual([
      'backupmonitor',
      'calendar',
      'diskhealth',
      'parcels',
      'speedtest',
      'timeline',
      'transit',
      'waste',
    ]);
    const ids = INTENTS.map((i) => i.id);
    expect(new Set(ids).size).toBe(ids.length);
    for (const c of contributions) {
      const id = c.source(fakeContainer() as unknown as Container, { widgetConfig: () => null }).id;
      for (const i of c.intents) expect(i.source).toBe(id);
    }
  });
});

/** Regression table of the module intents (#161); the built-in ones live in intents.test.ts. */
const table: [string, string | null][] = [
  // waste
  ['müll', 'waste.next'],
  ['abfall', 'waste.next'],
  ['wann kommt die gelbe tonne', 'waste.next'],
  ['when is the trash collected', 'waste.next'],
  ['/müll', 'waste.next'],
  // calendar
  ['termine', 'calendar.agenda'],
  ['heute', 'calendar.agenda'],
  ['morgen', 'calendar.agenda'],
  ['termine morgen', 'calendar.agenda'],
  ['meine termine am freitag', 'calendar.agenda'],
  ['calendar tomorrow', 'calendar.agenda'],
  // parcels
  ['paket', 'parcels.status'],
  ['wo ist mein paket', 'parcels.status'],
  ['where is my package', 'parcels.status'],
  ['dhl', 'parcels.status'],
  // transit
  ['abfahrt', 'transit.departures'],
  ['nächster bus', 'transit.departures'],
  ['abfahrten hauptbahnhof', 'transit.departures'],
  ['wann fährt die nächste bahn', 'transit.departures'],
  ['/bus oldenburg hbf', 'transit.departures'],
  // backups and heartbeats
  ['backup', 'backup.overview'],
  ['ist das backup gelaufen', 'backup.overview'],
  ['heartbeats', 'backup.overview'],
  ['restic', 'backup.status'],
  ['/backup paperless', 'backup.status'],
  // speedtest
  ['speedtest', 'speedtest.latest'],
  ['wie schnell ist das internet', 'speedtest.latest'],
  ['internet speed', 'speedtest.latest'],
  // disk health; "platte voll" stays with the disk usage and its forecast
  ['smart', 'diskhealth.status'],
  ['festplatte kaputt', 'diskhealth.status'],
  ['ssd zustand', 'diskhealth.status'],
  ['platte voll', 'system.disk'],
  // timeline
  ['was ist passiert', 'timeline.recent'],
  ['vorfälle', 'timeline.recent'],
  ['what happened', 'timeline.recent'],
  ['timeline', 'timeline.recent'],
  // the built-in intents still win their questions
  ['wetter morgen', 'weather.forecast'],
  ['cpu', 'system.cpu'],
  ['jellyfin', 'docker.status'],
  ['nas', 'services.status'],
  ['hallo', null],
];

describe('regression table of module intents (#161)', () => {
  it.each(table)('%s → %s', (query, want) => {
    expect(top(query)?.intent.id ?? null).toBe(want);
  });

  it('takes a day alone as the agenda, but not next to other words', () => {
    expect(top('morgen')?.slots['day']).toEqual({ type: 'day', offset: 1 });
    const ids = rank(interpret('wetter morgen', INTENTS, entities, today)).map((m) => m.intent.id);
    expect(ids).not.toContain('calendar.agenda');
  });

  it('fills the stop of a departure query', () => {
    expect(top('abfahrten hauptbahnhof')?.slots['place']).toEqual({ type: 'place', text: 'hauptbahnhof' });
  });
});

// --- sources -------------------------------------------------------------

const HOUR = 3_600_000;
const now = new Date();
const at = (days: number, hour: number) =>
  new Date(now.getFullYear(), now.getMonth(), now.getDate() + days, hour);
const event = (summary: string, start: Date, end: Date, allDay = false) => ({
  uid: summary,
  summary,
  location: '',
  calendar: '',
  color: '',
  allDay,
  start,
  end,
});

function fakeContainer() {
  return {
    waste: {
      upcoming: vi.fn(async () => ({
        error: '',
        types: [
          { label: 'Papier', icon: '📦', color: '', dates: [], daysUntil: 5 },
          { label: 'Gelbe Tonne', icon: '♻️', color: '', dates: [], daysUntil: 1 },
        ],
      })),
    },
    calendar: {
      events: vi.fn(async () => ({
        error: '',
        events: [
          { ...event('Zahnarzt', at(0, 9), at(0, 10)), location: 'Praxis' },
          event('Urlaub', at(0, 0), at(3, 0), true),
          event('Morgen', at(1, 9), at(1, 10)),
        ],
      })),
    },
    parcels: {
      list: vi.fn(async () => [
        {
          id: 'p1',
          trackingNumber: '00340434',
          carrier: 'dhl',
          label: 'Schuhe',
          status: 'delivered',
          description: '',
          location: 'Packstation',
          statusAt: now,
          estimatedDelivery: null,
          error: '',
        },
        {
          id: 'p2',
          trackingNumber: '00340435',
          carrier: 'dhl',
          label: '',
          status: 'transit',
          description: '',
          location: '',
          statusAt: null,
          estimatedDelivery: new Date(2026, 10, 14),
          error: '',
        },
      ]),
    },
    transit: {
      searchStops: vi.fn(async (q: string) =>
        q === 'hauptbahnhof' ? [{ id: 'hbf', name: 'Hauptbahnhof', lat: 0, lon: 0 }] : [],
      ),
      departures: vi.fn(async () => [
        {
          line: '310',
          headsign: 'Uni',
          mode: 'bus',
          planned: new Date(now.getTime() + 5 * 60_000),
          expected: null,
          delayMinutes: 6,
          track: '',
          cancelled: false,
        },
        {
          line: 'RE1',
          headsign: 'Bremen',
          mode: 'rail',
          planned: now,
          expected: null,
          delayMinutes: 0,
          track: '2',
          cancelled: true,
        },
      ]),
    },
    backupMonitor: {
      list: vi.fn(async () => [
        {
          id: 'b1',
          name: 'restic-offsite',
          state: 'failed',
          lastFailureAt: new Date(now.getTime() - 2 * HOUR),
          lastSuccessAt: null,
          lastMessage: 'disk full',
        },
        {
          id: 'b2',
          name: 'Paperless export',
          state: 'ok',
          lastFailureAt: null,
          lastSuccessAt: new Date(now.getTime() - 3 * HOUR),
          lastMessage: '',
        },
      ]),
    },
    speedtest: {
      results: vi.fn(async () => ({
        error: '',
        history: [],
        latest: { at: now, downloadMbps: 42.25, uploadMbps: 9.5, pingMs: 12.4 },
      })),
      settings: vi.fn(async () => ({
        configured: true,
        url: '',
        hasToken: true,
        minDownMbps: 50,
        minUpMbps: 0,
      })),
    },
    diskhealth: {
      disks: vi.fn(async () => [
        {
          device: '/dev/sda',
          model: 'WD Red',
          status: 'warning',
          temperature: 41,
          checkedAt: now,
          findings: [],
        },
        { device: '/dev/nvme0', model: '', status: 'ok', temperature: null, checkedAt: now, findings: [] },
      ]),
    },
    timeline: {
      list: vi.fn(async () => [
        {
          id: 'e1',
          at: now,
          source: 'container',
          severity: 'critical',
          subject: 'jellyfin',
          message: 'stopped',
        },
      ]),
    },
  };
}

function setup(widgetConfig: (kind: string) => Record<string, unknown> | null = () => null) {
  const c = fakeContainer();
  const source = (module: string) => {
    const contribution = contributions.find((x) => x.module === module);
    if (!contribution) throw new Error(module);
    return contribution.source(c as unknown as Container, { widgetConfig }) as SearchSource &
      Required<Pick<SearchSource, 'resolve'>>;
  };
  return { c, source };
}

const match = (id: string, slots: Match['slots'] = {}): Match => {
  const intent = INTENTS.find((i) => i.id === id);
  if (!intent) throw new Error(id);
  return { intent, score: 3, slots, missing: [], suggestions: [] };
};
const signal = new AbortController().signal;
const list = (a: Answer) => {
  expect(a.kind).toBe('list');
  return a as ListAnswer;
};

describe('module search sources (#161)', () => {
  it('waste: next collections first, tomorrow flagged', async () => {
    const a = list(await setup().source('waste').resolve(match('waste.next'), signal));
    expect(a.items.map((i) => i.label)).toEqual(['♻️ Gelbe Tonne', '📦 Papier']);
    expect(a.items[0]?.state).toBe('warning');
  });

  it('calendar: events of the asked day, all-day first', async () => {
    const { source, c } = setup();
    const todays = list(await source('calendar').resolve(match('calendar.agenda'), signal));
    expect(todays.items.map((i) => i.label)).toEqual(['Urlaub', 'Zahnarzt']);
    const tomorrow = list(
      await source('calendar').resolve(match('calendar.agenda', { day: { type: 'day', offset: 1 } }), signal),
    );
    expect(tomorrow.items.map((i) => i.label)).toEqual(['Urlaub', 'Morgen']);
    expect(c.calendar.events).toHaveBeenLastCalledWith(2);
  });

  it('eventsOnDay keeps events that span the day, not those ending at midnight', () => {
    const d = new Date(2026, 10, 13);
    expect(eventsOnDay([event('a', new Date(2026, 10, 12, 22), new Date(2026, 10, 13, 1))], d)).toHaveLength(
      1,
    );
    expect(eventsOnDay([event('b', new Date(2026, 10, 12, 9), new Date(2026, 10, 13, 0))], d)).toHaveLength(
      0,
    );
  });

  it('parcels: status with location or expected date', async () => {
    const a = list(await setup().source('parcels').resolve(match('parcels.status'), signal));
    expect(a.items[0]).toMatchObject({ label: 'Schuhe', state: 'ok' });
    expect(a.items[0]?.value).toContain('Packstation');
    expect(a.items[1]?.label).toBe('00340435');
  });

  it('transit: named stop, widget stop, or asks for one', async () => {
    const place = (text: string) => match('transit.departures', { place: { type: 'place', text } });
    const named = list(await setup().source('transit').resolve(place('hauptbahnhof'), signal));
    expect(named.title).toBe('Hauptbahnhof');
    expect(named.items.map((i) => i.state)).toEqual(['warning', 'error']);
    const widget = setup((k) => (k === 'transit' ? { stopId: 'w1', stopName: 'Lappan', count: 4 } : null));
    expect(list(await widget.source('transit').resolve(match('transit.departures'), signal)).title).toBe(
      'Lappan',
    );
    expect(await setup().source('transit').resolve(match('transit.departures'), signal)).toMatchObject({
      kind: 'text',
    });
    expect(await setup().source('transit').resolve(place('atlantis'), signal)).toMatchObject({
      kind: 'text',
      title: 'atlantis',
    });
  });

  it('backups: overview with problems first, status of one job', async () => {
    const { source } = setup();
    expect(await source('backupmonitor').entities?.()).toContainEqual({
      kind: 'job',
      id: 'b1',
      name: 'restic-offsite',
    });
    const all = list(await source('backupmonitor').resolve(match('backup.overview'), signal));
    expect(all.items[0]).toMatchObject({ label: 'restic-offsite', state: 'error' });
    const job = { type: 'entity' as const, entity: { kind: 'job', id: 'b1', name: 'restic-offsite' } };
    const one = await source('backupmonitor').resolve(match('backup.status', { job }), signal);
    expect(one).toMatchObject({ kind: 'status', state: 'error' });
    expect(one.kind === 'status' && one.detail).toContain('disk full');
  });

  it('speedtest: latest result, warning below the threshold', async () => {
    const a = await setup().source('speedtest').resolve(match('speedtest.latest'), signal);
    expect(a).toMatchObject({ kind: 'metric', value: '↓ 42.3 · ↑ 9.5 Mbit/s', severity: 'warning' });
  });

  it('disk health and timeline map their states', async () => {
    const { source } = setup();
    const disks = list(await source('diskhealth').resolve(match('diskhealth.status'), signal));
    expect(disks.items.map((i) => [i.label, i.state])).toEqual([
      ['WD Red', 'warning'],
      ['/dev/nvme0', 'ok'],
    ]);
    const events = list(await source('timeline').resolve(match('timeline.recent'), signal));
    expect(events.items[0]).toMatchObject({ label: 'jellyfin', state: 'error' });
  });
});
