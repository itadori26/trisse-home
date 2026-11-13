import { t } from '../../../shared/i18n';

/** One value picked out of the JSON response (#48). */
export interface Extraction {
  label: string;
  path: string;
  unit: string;
  warn: number | null;
  crit: number | null;
  /** Thresholds are reached at or below (e.g. free space) instead of at or above. */
  below: boolean;
}

/** A configured endpoint; the header value never comes back from the server. */
export interface Endpoint {
  id: string;
  name: string;
  url: string;
  headerName: string;
  hasHeader: boolean;
  intervalSeconds: number;
  allowPublic: boolean;
  extractions: Extraction[];
}

/** The editor form. An empty headerValue keeps the stored one. */
export interface EndpointInput {
  name: string;
  url: string;
  headerName: string;
  headerValue: string;
  intervalSeconds: number;
  allowPublic: boolean;
  extractions: Extraction[];
}

export type ValueStatus = 'ok' | 'warning' | 'critical' | 'unknown';

/** One extracted value. */
export interface Value {
  label: string;
  unit: string;
  text: string;
  number: number | null;
  status: ValueStatus;
  /** Why the path matched nothing. */
  error: string;
}

/** The values of one fetch; error is set (translated) when the request failed. */
export interface Values {
  name: string;
  values: Value[];
  fetchedAt: Date;
  error: string;
}

/** A test request from the settings. */
export interface Preview extends Values {
  body: string;
}

/** Interval choices of the editor, in seconds. */
export const INTERVALS = [30, 60, 300, 900, 3600, 21_600, 86_400] as const;
export const MAX_EXTRACTIONS = 10;

export const emptyExtraction = (): Extraction => ({
  label: '',
  path: '$.',
  unit: '',
  warn: null,
  crit: null,
  below: false,
});

export const emptyInput = (): EndpointInput => ({
  name: '',
  url: '',
  headerName: '',
  headerValue: '',
  intervalSeconds: 300,
  allowPublic: false,
  extractions: [emptyExtraction()],
});

/** The editor form of a stored endpoint. */
export const toInput = (e: Endpoint): EndpointInput => ({
  name: e.name,
  url: e.url,
  headerName: e.headerName,
  headerValue: '',
  intervalSeconds: e.intervalSeconds,
  allowPublic: e.allowPublic,
  extractions: e.extractions.map((x) => ({ ...x })),
});

/** Checks the form before it is sent (the server checks again); an error message or null. */
export function validateInput(i: EndpointInput, hasStoredHeader: boolean): string | null {
  if (i.name.trim() === '') return t('apiwidget.name_required');
  if (!/^https?:\/\/\S+$/i.test(i.url.trim())) return t('apiwidget.url_invalid');
  if (i.headerName.trim() !== '' && i.headerValue === '' && !hasStoredHeader)
    return t('apiwidget.header_value_required');
  if (i.extractions.length === 0) return t('apiwidget.values_required');
  for (const x of i.extractions) {
    if (x.label.trim() === '' || x.path.trim() === '') return t('apiwidget.value_incomplete');
  }
  return null;
}

/** "21.5 °C": numbers in the user's locale (at most 2 decimals), else the text. */
export function formatValue(v: Pick<Value, 'text' | 'number' | 'unit'>, locale: string): string {
  const text = v.number !== null ? v.number.toLocaleString(locale, { maximumFractionDigits: 2 }) : v.text;
  return v.unit ? `${text} ${v.unit}` : text;
}

/** Configuration of the widget: which endpoint it shows. */
export function parseWidgetConfig(raw: Record<string, unknown>): { endpointId: string } {
  return { endpointId: typeof raw['endpointId'] === 'string' ? raw['endpointId'] : '' };
}

/** Human interval label: 30 s, 5 min, 1 h. */
export function intervalLabel(seconds: number): string {
  if (seconds < 60) return t('apiwidget.seconds', { n: seconds });
  if (seconds < 3600) return t('apiwidget.minutes', { n: seconds / 60 });
  return t('apiwidget.hours', { n: seconds / 3600 });
}
