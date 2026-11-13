/**
 * Simple 24×24 stroke icons drawn for trisse-home (no icon library, font or
 * CDN, ADR 0010). Each entry is a list of SVG path data strings rendered with
 * stroke="currentColor", so they follow the theme in light and dark mode.
 */
export const iconPaths = {
  search: ['M11 4a7 7 0 1 0 0 14a7 7 0 1 0 0-14', 'M16.5 16.5L21 21'],
  clock: ['M12 3a9 9 0 1 0 0 18a9 9 0 1 0 0-18', 'M12 7v5l3 2'],
  sun: [
    'M12 8a4 4 0 1 0 0 8a4 4 0 1 0 0-8',
    'M12 2v2M12 20v2M2 12h2M20 12h2M5 5l1.5 1.5M17.5 17.5L19 19M5 19l1.5-1.5M17.5 6.5L19 5',
  ],
  rain: ['M7 15a4 4 0 0 1 0-8a5 5 0 0 1 9.6-1.5A4 4 0 0 1 17 15z', 'M8 18l-1 3M12 18l-1 3M16 18l-1 3'],
  server: ['M4 4h16v6H4z', 'M4 14h16v6H4z', 'M8 7h.01M8 17h.01'],
  activity: ['M3 12h4l3-8l4 16l3-8h4'],
  bookmark: ['M6 3h12v18l-6-4l-6 4z'],
  note: ['M5 3h10l4 4v14H5z', 'M15 3v4h4', 'M8 12h8M8 16h6'],
  check: ['M4 4h16v16H4z', 'M8 12l3 3l5-6'],
  rss: ['M5 19h.01', 'M5 12a7 7 0 0 1 7 7', 'M5 5a14 14 0 0 1 14 14'],
  calendar: ['M4 6h16v15H4z', 'M4 10h16', 'M8 3v4M16 3v4'],
  timer: ['M12 5a8 8 0 1 0 0 16a8 8 0 1 0 0-16', 'M12 9v4l2.5 1.5', 'M10 2h4M19 6l1.5-1.5'],
  hourglass: ['M6 3h12M6 21h12', 'M7 3c0 5 10 5 10 9s-10 4-10 9', 'M17 3c0 5-10 5-10 9s10 4 10 9'],
  cpu: ['M7 7h10v10H7z', 'M10 10h4v4h-4z', 'M9 3v4M15 3v4M9 17v4M15 17v4M3 9h4M3 15h4M17 9h4M17 15h4'],
  box: ['M12 3l8 4.5v9L12 21l-8-4.5v-9z', 'M4 7.5l8 4.5l8-4.5', 'M12 12v9'],
  shield: ['M12 3l8 3v6c0 5-3.5 8-8 9c-4.5-1-8-4-8-9V6z', 'M9 12l2 2l4-4'],
  cloud: ['M7 18a4.5 4.5 0 0 1-.5-9a6 6 0 0 1 11.5 1.5A4 4 0 0 1 17.5 18z'],
  router: ['M3 14h18v6H3z', 'M7 17h.01M11 17h.01', 'M15 10a4 4 0 0 1 4 0', 'M13 7a8 8 0 0 1 8 0'],
  key: ['M8 10a4 4 0 1 0 0 8a4 4 0 1 0 0-8', 'M11 12l9-9', 'M16 7l2 2M18 5l2 2'],
  trash: ['M4 7h16', 'M9 7V4h6v3', 'M6 7l1 14h10l1-14', 'M10 11v6M14 11v6'],
  cart: ['M3 4h3l2 11h10l2-8H7', 'M10 20h.01M17 20h.01'],
  bus: ['M5 4h14v13H5z', 'M5 11h14', 'M8 17v3M16 17v3', 'M8 14h.01M16 14h.01'],
  package: ['M4 7l8-4l8 4v10l-8 4l-8-4z', 'M4 7l8 4l8-4', 'M12 11v10', 'M8 5l8 4'],
  archive: ['M3 4h18v4H3z', 'M5 8v12h14V8', 'M10 12h4'],
  history: ['M4 12a8 8 0 1 0 2.3-5.7', 'M4 4v4h4', 'M12 8v4l3 2'],
  disk: ['M3 12h18v6H3z', 'M5 12l2-7h10l2 7', 'M7 15h.01M11 15h4'],
  gauge: ['M4 18a8 8 0 1 1 16 0', 'M12 18l4-6', 'M12 18h.01'],
} as const;

export type IconName = keyof typeof iconPaths;
