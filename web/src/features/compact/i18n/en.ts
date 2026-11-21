/** English UI texts of compact widget views (#595). */
export const en = {
  'compact.loading': 'Loading…',
  'compact.unavailable': 'Not available',
  'compact.level_attention': 'Needs attention',
  'compact.level_problem': 'Problem',
  'compact.summary': 'System summary',
  'compact.show_all': 'Show all',
  'compact.collapse': 'Collapse',
  'compact.action_done': '{action}: done',
  'compact.example.todos': '3 open · 1 overdue',
  'compact.example.habits': '2 of 5 today',
  'compact.example.focus': '50 min focus today',
  'compact.example.services': '11 of 12 running',
  'compact.example.status-overview': '11 of 12 running',
  'compact.example.calendar': 'Dentist · Today at 14:00',
  'compact.example.backup-status': 'Backup OK',
  'compact.example.backup-health': 'Backups all good',
  'compact.example.diskhealth': '3 drives healthy',
  'compact.example.docker': '11 of 12 running · 1 stopped',
  'compact.example.parcels': '2 on the way · 1 arriving today',
  'compact.example.speedtest': '250 ↓ · 40 ↑ Mbit/s',
  'compact.example.proxmox': '3 of 4 guests running',
  'compact.example.adguard': '12 % blocked',
  'compact.example.wireguard': '2 of 3 peers connected',
  'compact.example.weather': '7° · Overcast',
  'compact.example.electricity': '28 ct now · Cheap from 14:00',
  'compact.example.smarthome': 'Coffee machine · Off',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
