import type { en } from './en';

/** German UI texts of the proxmox feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.proxmox.title': 'Proxmox VE',
  'widget.proxmox.description': 'Auslastung der Knoten und Status von VMs und LXC-Containern.',
  'settings.section.proxmox': 'Proxmox',
  'proxmox.connect_under_settings': 'Verbinde Proxmox VE unter Einstellungen → Proxmox.',
  'proxmox.uptime': 'läuft seit {value}',
  'proxmox.offline': 'offline',
  'proxmox.cpu': 'CPU ({cpus} Kerne)',
  'proxmox.memory': 'Arbeitsspeicher',
  'proxmox.disk': 'Root-Laufwerk',
  'proxmox.no_guests': 'Keine VMs oder Container.',
  'proxmox.status.stopped': 'gestoppt',
  'proxmox.status.paused': 'pausiert',
  'proxmox.status.suspended': 'angehalten',
  'proxmox.status.prelaunch': 'startet',
  'proxmox.status.unknown': 'unbekannt',
  'proxmox.start': '{name} starten',
  'proxmox.shutdown': '{name} herunterfahren',
  'proxmox.confirm_start': '{name} starten?',
  'proxmox.confirm_shutdown': '{name} herunterfahren? Dienste darin werden beendet.',
  'proxmox.running_only': 'Nur laufende Gäste anzeigen',
  'proxmox.url': 'Proxmox-VE-URL',
  'proxmox.token_id': 'API-Token-ID',
  'proxmox.secret': 'Token-Secret {value}',
  'proxmox.fingerprint': 'Zertifikats-Fingerabdruck (SHA-256, optional)',
  'proxmox.fingerprint_hint':
    'Proxmox nutzt ein selbstsigniertes Zertifikat. Füge seinen SHA-256-Fingerabdruck ein (Knoten → System → Zertifikate), um genau diesem Zertifikat zu vertrauen, oder lass das Feld leer, wenn eine bekannte CA es signiert hat.',
  'proxmox.token_hint':
    'Lege auf einem Proxmox-Knoten ein schreibgeschütztes API-Token an (Rolle PVEAuditor) und trage Token-ID und Secret unten ein.',
  'proxmox.actions_hint':
    'Starten und Herunterfahren bleiben aus, solange PROXMOX_ACTIONS_GUESTS die Gast-IDs nicht auflistet und das Token sie nicht schalten darf (VM.PowerMgmt).',
};
