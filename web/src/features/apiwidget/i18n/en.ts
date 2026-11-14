/** English UI texts of the JSON widget feature (#48). */
export const en = {
  'widget.apiwidget.title': 'JSON API',
  'widget.apiwidget.description':
    'Values from any JSON endpoint, picked with JSONPath, with unit and warning thresholds.',
  'settings.section.apiwidget': 'JSON APIs',
  'apiwidget.intro':
    'Show values from the JSON API of any service: the widget picks them out of the response with JSONPath, e.g. $.data.temperature or $.items[0].name.',
  'apiwidget.no_endpoints': 'No endpoints yet.',
  'apiwidget.add': 'Add endpoint',
  'apiwidget.edit': 'Edit {name}',
  'apiwidget.delete': 'Delete {name}',
  'apiwidget.delete_confirm': 'Delete the endpoint “{name}”? Widgets that show it will be empty.',
  'apiwidget.name': 'Name',
  'apiwidget.url': 'URL',
  'apiwidget.header_name': 'Header (optional)',
  'apiwidget.header_value': 'Header value',
  'apiwidget.header_stored': 'stored – leave empty to keep',
  'apiwidget.header_hint':
    'E.g. X-Api-Key or Authorization. The value is stored encrypted and never shown again.',
  'apiwidget.interval': 'Interval',
  'apiwidget.allow_public': 'Allow public addresses',
  'apiwidget.allow_public_hint':
    'By default only addresses in your network (loopback, 10.x, 172.16–31.x, 192.168.x, fd00::/8, 100.64.x) are requested.',
  'apiwidget.values': 'Values',
  'apiwidget.label': 'Label',
  'apiwidget.path': 'JSONPath',
  'apiwidget.unit': 'Unit',
  'apiwidget.warn': 'Warning at',
  'apiwidget.crit': 'Critical at',
  'apiwidget.below': 'lower is worse',
  'apiwidget.add_value': 'Add value',
  'apiwidget.remove_value': 'Remove value',
  'apiwidget.path_hint': "Supported: $.a.b, $['name with spaces'], $.list[0], $.list[-1], $.list.length()",
  'apiwidget.test': 'Test',
  'apiwidget.testing': 'Testing…',
  'apiwidget.response': 'Response (beginning)',
  'apiwidget.saved': 'Saved.',
  'apiwidget.cancel': 'Cancel',
  'apiwidget.save': 'Save',
  'apiwidget.name_required': 'Enter a name.',
  'apiwidget.url_invalid': 'Enter an http(s) URL.',
  'apiwidget.header_value_required': 'Enter the header value.',
  'apiwidget.values_required': 'Add at least one value.',
  'apiwidget.value_incomplete': 'Every value needs a label and a JSONPath.',
  'apiwidget.seconds': '{n} s',
  'apiwidget.minutes': '{n} min',
  'apiwidget.hours': '{n} h',
  'apiwidget.choose_endpoint': 'Endpoint',
  'apiwidget.set_up': 'Add a JSON endpoint under Settings → JSON APIs.',
  'apiwidget.choose_in_settings': 'Choose an endpoint in the widget settings.',
  'apiwidget.gone': 'The endpoint of this widget was deleted.',
  'apiwidget.status_warning': 'warning',
  'apiwidget.status_critical': 'critical',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
