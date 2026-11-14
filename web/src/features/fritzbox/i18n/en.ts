/** English UI texts of the fritzbox feature added with the calls widget (#168, #193). */
export const en = {
  'widget.fritzbox_calls.title': 'FRITZ!Box calls',
  'widget.fritzbox_calls.description': 'Recent calls of the FRITZ!Box: all, missed, incoming or outgoing.',
  'fritzbox.show_uptime': 'Show the uptime',
  'fritzbox.show_ip': 'Show the external IP address',
  'fritzbox.show_rates': 'Show the current download and upload',
  'fritzbox.calls_filter': 'Calls',
  'fritzbox.filter.all': 'All calls',
  'fritzbox.filter.missed': 'Missed',
  'fritzbox.filter.incoming': 'Incoming',
  'fritzbox.filter.outgoing': 'Outgoing',
  'fritzbox.calls_count': 'Number of calls (5–50)',
  'fritzbox.show_numbers': 'Show numbers of callers without a name',
  'fritzbox.save': 'Save',
  'fritzbox.no_calls': 'No calls.',
  'fritzbox.call.incoming': 'incoming',
  'fritzbox.call.missed': 'missed',
  'fritzbox.call.outgoing': 'outgoing',
  'fritzbox.call.rejected': 'rejected',
  'fritzbox.call.active': 'active',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
