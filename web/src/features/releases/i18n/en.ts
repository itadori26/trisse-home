/** English UI texts of the releases feature (#193). */
export const en = {
  'widget.releases.title': 'New releases',
  'widget.releases.description': 'New albums, singles and EPs of the artists you follow (MusicBrainz).',
  'settings.section.releases': 'New releases',
  'releases.none': 'No new releases in the last {weeks} weeks. Follow artists under Settings → New releases.',
  'releases.weeks': 'Weeks',
  'releases.hint':
    'Follow artists to see their new albums, singles and EPs. Releases are checked every six hours.',
  'releases.search': 'Search an artist',
  'releases.follow': 'Follow',
  'releases.followed': 'Followed artists ({count})',
  'releases.unfollow': 'Unfollow {name}',
  'releases.no_artists': 'You do not follow any artist yet.',
  'releases.attribution': 'Data: MusicBrainz (CC0), covers: Cover Art Archive.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
