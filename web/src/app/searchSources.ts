import { canAct as dockerCanAct, indicator } from '../features/docker/domain/container';
import { calculate } from '../features/search/domain/calc/calc';
import type { Answer, State } from '../features/search/domain/smart/answer';
import type { Entity, Match } from '../features/search/domain/smart/intent';
import { INTENTS } from '../features/search/domain/smart/intents';
import { type SearchSource, SmartSearch } from '../features/search/usecase/smartSearch';
import type { ServiceState } from '../features/services/domain/service';
import { describe, type WeatherConfig } from '../features/weather/domain/weather';
import { intlLocale, t } from '../shared/i18n';
import { formatBytes, formatUptime } from '../shared/ui/format';
import type { Container } from './container';

/**
 * Answer sources of the smart search (#156): small adapters on top of the
 * existing feature use cases. They live in the composition root, so no
 * feature imports another one. Disabled modules register no source.
 */

const percent = (n: number) => `${Math.round(n)} %`;
const serviceState: Record<ServiceState, State> = { online: 'ok', offline: 'error', unknown: 'off' };
const entityOf = (m: Match, slot: string): Entity | null => {
  const v = m.slots[slot];
  return v?.type === 'entity' ? v.entity : null;
};

function systemSource(c: Container): SearchSource {
  return {
    id: 'system',
    cacheMs: 10_000,
    async resolve(m) {
      const s = await c.system.snapshot();
      const base = { intent: m.intent.id, widgetKind: 'system', at: new Date() };
      switch (m.intent.id) {
        case 'system.cpu':
          return {
            ...base,
            kind: 'metric',
            title: t('smart.cpu'),
            value: percent(s.cpuPercent),
            percent: s.cpuPercent,
            detail: t('smart.cpu_detail', { n: s.cpuCount, load: s.load[0].toFixed(2) }),
          };
        case 'system.memory':
          return {
            ...base,
            kind: 'metric',
            title: t('smart.memory'),
            value: percent(s.memory.percent),
            percent: s.memory.percent,
            detail: t('smart.of', { used: formatBytes(s.memory.used), total: formatBytes(s.memory.total) }),
          };
        case 'system.disk':
          return {
            ...base,
            kind: 'list',
            title: t('smart.disks'),
            empty: t('smart.no_disks'),
            items: s.disks.map((d) => ({
              label: d.label,
              value: `${percent(d.usage.percent)} · ${t('smart.free', { free: formatBytes(d.usage.total - d.usage.used) })}`,
              percent: d.usage.percent,
            })),
          };
        case 'system.load':
          return {
            ...base,
            kind: 'metric',
            title: t('smart.load'),
            value: s.load.map((l) => l.toFixed(2)).join(' · '),
            detail: t('smart.load_detail', { n: s.cpuCount }),
          };
        case 'system.uptime':
          return { ...base, kind: 'metric', title: t('smart.uptime'), value: formatUptime(s.uptimeSeconds) };
        default:
          return s.temperature > 0
            ? {
                ...base,
                kind: 'metric',
                title: t('smart.temperature'),
                value: `${Math.round(s.temperature)} °C`,
              }
            : { ...base, kind: 'text', title: t('smart.temperature'), text: t('smart.no_temperature') };
      }
    },
  };
}

function dockerSource(c: Container): SearchSource {
  return {
    id: 'docker',
    async entities() {
      const list = await c.docker.containers();
      return list.containers.map((x) => ({ kind: 'container', id: x.id, name: x.name }));
    },
    async resolve(m) {
      const base = { intent: m.intent.id, widgetKind: 'docker', at: new Date() };
      const list = await c.docker.containers();
      if (m.intent.id === 'docker.overview') {
        return {
          ...base,
          kind: 'list',
          title: t('smart.containers', { running: list.running, total: list.total }),
          empty: t('smart.no_containers'),
          items: list.containers
            .filter((x) => indicator(x) !== 'ok')
            .map((x) => ({ label: x.name, value: x.status, state: indicator(x) })),
        };
      }
      const target = entityOf(m, 'container');
      const ct = list.containers.find((x) => x.id === target?.id || x.name === target?.name);
      if (!target || !ct)
        return { ...base, kind: 'text', title: target?.name ?? '', text: t('smart.container_gone') };
      if (m.intent.id === 'docker.restart') {
        const action = m.slots['action']?.type === 'action' ? m.slots['action'].action : 'restart';
        const allowed = dockerCanAct(await c.docker.actions().catch(() => null), ct.name);
        if (action !== 'restart' || !allowed) {
          return {
            ...base,
            kind: 'text',
            title: ct.name,
            text: t(allowed ? 'smart.only_restart' : 'smart.actions_disabled'),
          };
        }
        return {
          ...base,
          kind: 'actions',
          title: ct.name,
          text: ct.status,
          actions: [
            {
              source: 'docker',
              action: 'restart',
              target,
              label: t('smart.restart'),
              confirm: t('smart.restart_confirm', { name: ct.name }),
            },
          ],
        };
      }
      const updates = await c.docker.updates().catch(() => null);
      const update = updates?.find((u) => u.containers.includes(ct.name) && u.status === 'update_available');
      return {
        ...base,
        kind: 'status',
        title: ct.name,
        state: indicator(ct),
        text: ct.status,
        detail: [ct.image, update ? t('smart.update_available', { tag: update.newerTag }) : '']
          .filter(Boolean)
          .join(' · '),
      };
    },
    async execute(item) {
      await c.docker.restart(item.target.name);
    },
  };
}

function servicesSource(c: Container): SearchSource {
  return {
    id: 'services',
    async entities() {
      return (await c.services.list()).map((s) => ({ kind: 'service', id: s.id, name: s.name }));
    },
    async resolve(m) {
      const list = await c.services.list();
      const base = { intent: m.intent.id, widgetKind: 'services' };
      const latest = list.reduce<Date | null>(
        (d, s) => (s.checkedAt && (!d || s.checkedAt > d) ? s.checkedAt : d),
        null,
      );
      if (m.intent.id === 'services.overview') {
        const down = list.filter((s) => s.state !== 'online');
        return {
          ...base,
          at: latest,
          kind: 'list',
          title: t('smart.services', { online: list.length - down.length, total: list.length }),
          empty: t('smart.all_online'),
          items: down.map((s) => ({
            label: s.name,
            value: s.error || t(`smart.state_${s.state}`),
            state: serviceState[s.state],
          })),
        };
      }
      const target = entityOf(m, 'service');
      const s = list.find((x) => x.id === target?.id);
      if (!s)
        return { ...base, at: null, kind: 'text', title: target?.name ?? '', text: t('smart.service_gone') };
      return {
        ...base,
        at: s.checkedAt,
        kind: 'status',
        title: s.name,
        state: serviceState[s.state],
        text: t(`smart.state_${s.state}`),
        detail: [s.state === 'online' ? `${s.latencyMs} ms` : s.error, s.url].filter(Boolean).join(' · '),
      };
    },
  };
}

function weatherSource(c: Container, home: () => WeatherConfig | null): SearchSource {
  return {
    id: 'weather',
    cacheMs: 60_000,
    async resolve(m) {
      const base = { intent: m.intent.id, widgetKind: 'weather' };
      let place = home();
      const named = m.slots['place']?.type === 'place' ? m.slots['place'].text : '';
      if (named) {
        const found = (await c.weather.searchPlaces(named))[0];
        if (!found)
          return {
            ...base,
            at: null,
            kind: 'text',
            title: t('smart.weather'),
            text: t('smart.place_unknown', { place: named }),
          };
        place = { lat: found.lat, lon: found.lon, label: found.name, days: 5, warnings: false };
      }
      if (!place)
        return { ...base, at: null, kind: 'text', title: t('smart.weather'), text: t('smart.no_location') };
      const f = await c.weather.forecast({ ...place, days: 5 });
      const offset = m.slots['day']?.type === 'day' ? m.slots['day'].offset : 0;
      const where = place.label || t('smart.weather');
      if (offset === 0) {
        const cond = describe(f.code, f.isDay);
        return {
          ...base,
          at: f.updatedAt,
          kind: 'metric',
          title: where,
          value: `${cond.icon} ${Math.round(f.temperature)} °C`,
          detail: `${cond.label} · ${t('smart.feels', { t: Math.round(f.apparentTemperature) })}`,
        };
      }
      const day = f.days[offset];
      if (!day) return { ...base, at: f.updatedAt, kind: 'text', title: where, text: t('smart.no_forecast') };
      const cond = describe(day.code);
      const date = day.date.toLocaleDateString(intlLocale(), { weekday: 'long' });
      return {
        ...base,
        at: f.updatedAt,
        kind: 'metric',
        title: `${where} · ${date}`,
        value: `${cond.icon} ${Math.round(day.max)}° / ${Math.round(day.min)}°`,
        detail: `${cond.label} · ${t('smart.rain_chance', { p: day.precipitationProbability })}`,
      };
    },
  };
}

/** Formats a calculated number, localised (10 significant digits, 6 for units). */
function formatNumber(n: number, grouping: boolean, digits = 10): string {
  const rounded = Number(n.toPrecision(digits));
  return new Intl.NumberFormat(intlLocale(), {
    maximumFractionDigits: 10,
    useGrouping: grouping,
  }).format(Object.is(rounded, -0) ? 0 : rounded);
}

/** Calculator and unit conversion (#160); in the browser, no module needed. */
export function calcSource(): SearchSource {
  return {
    id: 'calc',
    direct(query) {
      const r = calculate(query);
      if (!r) return null;
      if (r.kind === 'expression')
        return {
          kind: 'metric',
          intent: 'calc.expression',
          title: `${r.expression} =`,
          value: formatNumber(r.value, true),
          copy: formatNumber(r.value, false),
          at: null,
        };
      const unit = (x: { symbol: string }) => (x.symbol.startsWith('°') ? x.symbol : ` ${x.symbol}`);
      return {
        kind: 'metric',
        intent: 'calc.conversion',
        title: `${formatNumber(r.value, true)}${unit(r.from)} → ${r.to.symbol}`,
        value: `${formatNumber(r.result, true, 6)}${unit(r.to)}`,
        detail: t('smart.conversion'),
        copy: formatNumber(r.result, false, 6),
        at: null,
      };
    },
  };
}

/** Answer when a required slot is missing (#158): ask back with suggestions. */
function askBack(m: Match): Answer {
  const names = m.suggestions.map((e) => e.name);
  return {
    kind: 'text',
    intent: m.intent.id,
    title: t('smart.which_title'),
    text: t(m.missing.includes('service') ? 'smart.which_service' : 'smart.which_container'),
    suggestions: names,
    at: null,
  };
}

/** Builds the smart search with the sources of the enabled modules. */
export function createSmartSearch(
  c: Container,
  modules: Set<string>,
  opts: {
    canAct: boolean;
    weatherLocation: () => WeatherConfig | null;
    score?: (query: string, target: string, score: number) => number;
  },
): SmartSearch {
  const sources: SearchSource[] = [calcSource()];
  if (modules.has('system')) sources.push(systemSource(c));
  if (modules.has('docker')) sources.push(dockerSource(c));
  if (modules.has('services')) sources.push(servicesSource(c));
  if (modules.has('weather')) sources.push(weatherSource(c, opts.weatherLocation));
  return new SmartSearch({ intents: INTENTS, sources, canAct: opts.canAct, askBack, score: opts.score });
}
