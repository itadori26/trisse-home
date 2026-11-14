import type { PollenBurden, UVLevel } from './weather';

// Only the widget needs these; kept apart from weather.ts, which is part of
// the start-up bundle (greeting line).

/** Colours of the pollen burden (none … high). */
export const pollenColors: Record<PollenBurden, string> = {
  none: 'var(--color-border)',
  low: '#4caf50',
  moderate: '#f5a623',
  high: '#d33a3a',
  unknown: 'transparent',
};

/** WHO colour scale of the UV index. */
export const uvColors: Record<UVLevel, string> = {
  low: '#4caf50',
  moderate: '#f2c200',
  high: '#f57c00',
  very_high: '#d32f2f',
  extreme: '#8e24aa',
  unknown: 'transparent',
};
