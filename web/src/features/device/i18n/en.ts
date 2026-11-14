/** English UI texts of the device feature (#140). */
export const en = {
  'settings.section.device': 'Device',
  'device.question': 'How do you want to use this device?',
  'device.headless': 'Server only',
  'device.headless_hint': 'Open trisse-home on your phone, tablet or computer.',
  'device.kiosk': 'Server + display',
  'device.kiosk_hint':
    'Also show the dashboard full-screen on the screen connected to this Pi (installs a minimal browser, several hundred MB).',
  'device.display_connected': 'A display is connected to this device.',
  'device.display_disconnected': 'No display is connected right now.',
  'device.display_unknown': 'Whether a display is connected is unknown.',
  'device.apply': 'Apply',
  'device.pending': 'Requested – the device picks it up in a moment…',
  'device.installing': 'The device is setting up the display, this can take a few minutes…',
  'device.failed': 'The change failed: {error}',
  'device.active_kiosk': 'The dashboard is shown on the connected display.',
  'device.active_headless': 'This device runs as a server only.',
  'device.later': 'You can change this any time under Settings → Device.',
  'device.settings_hint': 'Choose whether this Raspberry Pi also shows the dashboard on its own display.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
