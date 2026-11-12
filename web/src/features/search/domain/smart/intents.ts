import type { IntentSpec } from './intent';

/**
 * The intents of the first sources (#156). Terms are folded (no umlauts,
 * lower case); weights say how strongly a word points at the intent.
 * Adding a question means adding words here – the regression table in
 * intents.test.ts shows the effect.
 */
const t = (weight: number, ...terms: string[]) => terms.map((term) => ({ term, weight }));

export const INTENTS: IntentSpec[] = [
  {
    id: 'system.cpu',
    source: 'system',
    prefixes: ['cpu'],
    answer: 'metric',
    widgetKind: 'system',
    terms: [...t(3, 'cpu', 'prozessor', 'processor'), ...t(1, 'auslastung', 'usage', 'utilization')],
  },
  {
    id: 'system.memory',
    source: 'system',
    prefixes: ['ram'],
    answer: 'metric',
    widgetKind: 'system',
    terms: [...t(3, 'ram', 'arbeitsspeicher', 'memory', 'mem'), ...t(1, 'speicher', 'frei', 'free')],
  },
  {
    id: 'system.disk',
    source: 'system',
    prefixes: ['disk'],
    answer: 'list',
    widgetKind: 'system',
    terms: [
      ...t(3, 'disk', 'disks', 'festplatte', 'platte', 'speicherplatz', 'storage', 'laufwerk'),
      ...t(1.5, 'voll', 'full', 'space', 'platz'),
      ...t(1, 'frei', 'free', 'speicher'),
    ],
  },
  {
    id: 'system.load',
    source: 'system',
    prefixes: ['load'],
    answer: 'metric',
    widgetKind: 'system',
    terms: [...t(3, 'load', 'loadavg', 'systemlast'), ...t(1.5, 'last')],
  },
  {
    id: 'system.uptime',
    source: 'system',
    prefixes: ['uptime'],
    answer: 'metric',
    widgetKind: 'system',
    terms: [...t(3, 'uptime', 'laufzeit', 'betriebszeit'), ...t(1.5, 'neustart', 'reboot', 'seit', 'since')],
  },
  {
    id: 'system.temperature',
    source: 'system',
    answer: 'metric',
    widgetKind: 'system',
    terms: [...t(1.5, 'temperatur', 'temperature', 'temp'), ...t(1, 'heiss', 'hot', 'cpu', 'server')],
  },
  {
    id: 'docker.overview',
    source: 'docker',
    answer: 'list',
    widgetKind: 'docker',
    terms: [
      ...t(3, 'docker', 'container', 'containers'),
      ...t(1, 'laufen', 'running', 'gestoppt', 'stopped'),
    ],
  },
  {
    id: 'docker.status',
    source: 'docker',
    prefixes: ['docker'],
    answer: 'status',
    widgetKind: 'docker',
    slots: [{ name: 'container', type: 'entity', entityKind: 'container', required: true }],
    terms: t(1, 'status', 'lauft', 'laeuft', 'running', 'image', 'update'),
  },
  {
    id: 'docker.restart',
    source: 'docker',
    prefixes: ['docker'],
    answer: 'actions',
    widgetKind: 'docker',
    slots: [
      { name: 'action', type: 'action', required: true },
      { name: 'container', type: 'entity', entityKind: 'container', required: true },
    ],
    terms: [],
  },
  {
    id: 'services.overview',
    source: 'services',
    answer: 'list',
    widgetKind: 'services',
    terms: [
      ...t(3, 'dienste', 'services', 'dienst'),
      ...t(2, 'offline', 'down', 'erreichbar', 'reachable', 'ausgefallen'),
    ],
  },
  {
    id: 'services.status',
    source: 'services',
    prefixes: ['service'],
    answer: 'status',
    widgetKind: 'services',
    slots: [{ name: 'service', type: 'entity', entityKind: 'service', required: true }],
    terms: t(1, 'status', 'online', 'erreichbar', 'antwortzeit', 'latency', 'ping'),
  },
  {
    id: 'weather.forecast',
    source: 'weather',
    prefixes: ['wetter', 'weather'],
    answer: 'metric',
    widgetKind: 'weather',
    slots: [
      { name: 'place', type: 'place' },
      { name: 'day', type: 'day' },
    ],
    terms: [
      ...t(3, 'wetter', 'weather', 'vorhersage', 'forecast'),
      ...t(2, 'regen', 'regnet', 'rain', 'schnee', 'snow', 'sonne', 'sunny'),
      ...t(1.5, 'temperatur', 'temperature', 'draussen', 'outside'),
      ...t(1, 'schirm', 'umbrella', 'kalt', 'cold', 'warm', 'jacke'),
    ],
  },
];
