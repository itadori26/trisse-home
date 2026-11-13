/** English UI texts of the proxmox feature (#193). */
export const en = {
  'widget.proxmox.title': 'Proxmox VE',
  'widget.proxmox.description': 'Node load and the status of VMs and LXC containers.',
  'settings.section.proxmox': 'Proxmox',
  'proxmox.connect_under_settings': 'Connect Proxmox VE under Settings → Proxmox.',
  'proxmox.uptime': 'up {value}',
  'proxmox.offline': 'offline',
  'proxmox.cpu': 'CPU ({cpus} cores)',
  'proxmox.memory': 'Memory',
  'proxmox.disk': 'Root disk',
  'proxmox.no_guests': 'No VMs or containers.',
  'proxmox.status.stopped': 'stopped',
  'proxmox.status.paused': 'paused',
  'proxmox.status.suspended': 'suspended',
  'proxmox.status.prelaunch': 'starting',
  'proxmox.status.unknown': 'unknown',
  'proxmox.start': 'Start {name}',
  'proxmox.shutdown': 'Shut down {name}',
  'proxmox.confirm_start': 'Start {name}?',
  'proxmox.confirm_shutdown': 'Shut down {name}? Running services in it will stop.',
  'proxmox.running_only': 'Only show running guests',
  'proxmox.url': 'Proxmox VE URL',
  'proxmox.token_id': 'API token ID',
  'proxmox.secret': 'Token secret {value}',
  'proxmox.fingerprint': 'Certificate fingerprint (SHA-256, optional)',
  'proxmox.fingerprint_hint':
    'Proxmox uses a self-signed certificate. Paste its SHA-256 fingerprint (node → System → Certificates) to trust exactly this certificate, or leave it empty if the certificate is signed by a known CA.',
  'proxmox.token_hint':
    'Create a read-only API token (role PVEAuditor) on a Proxmox node, then enter the token ID and secret below.',
  'proxmox.actions_hint':
    'Start and shutdown stay off unless PROXMOX_ACTIONS_GUESTS lists the guest IDs and the token may power them (VM.PowerMgmt).',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
