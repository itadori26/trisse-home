import type { ThemeMode } from '../../../shared/theme/theme';

/**
 * The first-start flow (#671): the essential settings, one per screen,
 * before the template choice and the tour.
 */

/** The personal settings the flow asks; they live in the appearance settings. */
export interface Personal {
  greetingName: string;
  language: '' | 'en' | 'de';
  theme: ThemeMode;
  accentColor: string;
  /** The organisation keeps members on its look (#305); the look step is left out. */
  themeLocked?: boolean;
}

/** A saved place (#591) as the flow and the template choice need it. */
export interface Place {
  id: string;
  name: string;
  label: string;
  lat: number;
  lon: number;
}

/** A step of the flow; plugin steps are "extra:<id>". */
export type FirstStartStep = 'name' | 'place' | 'language' | 'look' | `extra:${string}`;

/**
 * The steps in order. The owner gets all of them; invited members only the
 * personal ones (name, language, look), since the place and plugin steps
 * such as the instance name concern the whole installation. A step whose
 * setting is not available (module off, look locked) is left out.
 */
export function firstStartSteps(o: {
  owner: boolean;
  personal: boolean;
  look: boolean;
  places: boolean;
  extra: readonly string[];
}): FirstStartStep[] {
  const steps: FirstStartStep[] = [];
  if (o.personal) steps.push('name');
  if (o.owner && o.places) steps.push('place');
  if (o.personal) steps.push('language');
  if (o.look) steps.push('look');
  if (o.owner) steps.push(...o.extra.map((id) => `extra:${id}` as const));
  return steps;
}
