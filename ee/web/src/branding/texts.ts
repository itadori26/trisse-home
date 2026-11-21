import type { Catalogue } from '../lib/texts';

/** Texts of the branding section (#305). */
export const texts: Catalogue = {
  label: { en: 'Branding', de: 'Branding' },
  description: {
    en: "Your organisation's name, logo and colours on the sign-in page, in the header and in the browser tab.",
    de: 'Name, Logo und Farben eurer Organisation auf der Anmeldeseite, in der Kopfzeile und im Browser-Tab.',
  },
  name: { en: 'Name', de: 'Name' },
  step_title: { en: 'Name of this Waylight', de: 'Name dieses Waylight' },
  step_name_label: { en: 'Name', de: 'Name' },
  step_name_hint: {
    en: 'Shown on the sign-in page, in the header and in the browser tab. You can change it later under Settings → Branding.',
    de: 'Erscheint auf der Anmeldeseite, in der Kopfzeile und im Browser-Tab. Ändern lässt er sich später unter Einstellungen → Branding.',
  },
  accent: { en: 'Accent colour', de: 'Akzentfarbe' },
  secondary: { en: 'Second colour', de: 'Zweite Farbe' },
  none: { en: 'Default', de: 'Standard' },
  hide_powered_by: {
    en: 'Hide "Powered by waylight" on the sign-in page',
    de: '„Läuft mit waylight“ auf der Anmeldeseite ausblenden',
  },
  default_theme: {
    en: 'Theme for the sign-in page and new members',
    de: 'Design für die Anmeldeseite und neue Mitglieder',
  },
  theme_none: { en: 'No default', de: 'Keine Vorgabe' },
  theme_system: { en: 'Follow the device', de: 'Wie das Gerät' },
  theme_light: { en: 'Light', de: 'Hell' },
  theme_dark: { en: 'Dark', de: 'Dunkel' },
  lock_theme: {
    en: 'Keep members on this theme and the accent colour',
    de: 'Mitglieder bei diesem Design und der Akzentfarbe halten',
  },
  save: { en: 'Save', de: 'Speichern' },
  saved: {
    en: 'Saved. Reload the page to see the change everywhere.',
    de: 'Gespeichert. Lade die Seite neu, um die Änderung überall zu sehen.',
  },
  not_applied: {
    en: 'Saved, but not shown: the licence does not include branding.',
    de: 'Gespeichert, aber nicht angezeigt: Die Lizenz umfasst kein Branding.',
  },
  images: { en: 'Images', de: 'Bilder' },
  'kind.logo': { en: 'Logo', de: 'Logo' },
  'kind.logo-dark': {
    en: 'Logo for the dark theme (optional)',
    de: 'Logo für das dunkle Design (optional)',
  },
  'kind.favicon': { en: 'Browser tab icon', de: 'Symbol im Browser-Tab' },
  upload: { en: 'Upload', de: 'Hochladen' },
  remove: { en: 'Remove', de: 'Entfernen' },
  uploaded: { en: 'Uploaded {date}', de: 'Hochgeladen am {date}' },
  none_uploaded: {
    en: "None, the product's own is used",
    de: 'Keins, das des Produkts wird verwendet',
  },
  image_hint: {
    en: 'PNG, JPEG or GIF, at most 2 MB. SVG is not accepted.',
    de: 'PNG, JPEG oder GIF, höchstens 2 MB. SVG wird nicht angenommen.',
  },
  too_big: {
    en: 'The image is larger than 2 MB.',
    de: 'Das Bild ist größer als 2 MB.',
  },
  svg: {
    en: 'SVG images are not accepted; use PNG instead.',
    de: 'SVG-Bilder werden nicht angenommen; nimm stattdessen PNG.',
  },
};
