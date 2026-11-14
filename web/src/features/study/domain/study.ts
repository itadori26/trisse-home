import { t } from '../../../shared/i18n';

export type ModuleStatus = 'open' | 'in_progress' | 'passed';
export const STATUSES: readonly ModuleStatus[] = ['open', 'in_progress', 'passed'];

/** A study module (#57). */
export interface StudyModule {
  id: string;
  name: string;
  ects: number;
  /** 0 = not planned. */
  semester: number;
  status: ModuleStatus;
  /** German scale 1.0–4.0; null = ungraded or not passed. */
  grade: number | null;
}

export type ModuleInput = Omit<StudyModule, 'id'>;

/** Progress of the studies. */
export interface Progress {
  totalEcts: number;
  passedEcts: number;
  inProgressEcts: number;
  openEcts: number;
  passedModules: number;
  modules: number;
  percent: number;
  average: number | null;
  gradedEcts: number;
  inProgress: StudyModule[];
}

export const emptyModule = (): ModuleInput => ({
  name: '',
  ects: 5,
  semester: 1,
  status: 'open',
  grade: null,
});

/** Checks the form before it is sent (the server checks again); an error message or null. */
export function validateModule(m: ModuleInput): string | null {
  if (m.name.trim() === '') return t('study.name_required');
  if (!(m.ects >= 0.5 && m.ects <= 60) || Math.round(m.ects * 2) !== m.ects * 2)
    return t('study.ects_invalid');
  if (m.grade !== null) {
    if (m.status !== 'passed') return t('study.grade_only_passed');
    if (!(m.grade >= 1 && m.grade <= 4)) return t('study.grade_invalid');
  }
  return null;
}

/** "1.5" / "1,5": a grade with one or two decimals in the locale. */
export const formatGrade = (g: number, locale: string): string =>
  g.toLocaleString(locale, { minimumFractionDigits: 1, maximumFractionDigits: 2 });

/** "7.5" / "7,5": credits without trailing zeros. */
export const formatEcts = (v: number, locale: string): string =>
  v.toLocaleString(locale, { maximumFractionDigits: 1 });
