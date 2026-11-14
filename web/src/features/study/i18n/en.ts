/** English UI texts of the study progress feature (#57). */
export const en = {
  'widget.study.title': 'Study progress',
  'widget.study.description':
    'Credits passed as a progress bar, the weighted grade average and the modules in progress.',
  'settings.section.study': 'Study progress',
  'study.intro':
    'List your modules with credits, semester and status. Grades use the German scale (1.0 best, 4.0 passed); passed modules without a grade count for the credits only.',
  'study.total_ects': 'ECTS of the degree',
  'study.modules': 'Modules',
  'study.name': 'Module',
  'study.ects': 'ECTS',
  'study.semester': 'Semester',
  'study.semester_none': '–',
  'study.status': 'Status',
  'study.status_open': 'open',
  'study.status_in_progress': 'in progress',
  'study.status_passed': 'passed',
  'study.grade': 'Grade',
  'study.grade_none': 'no grade',
  'study.add': 'Add module',
  'study.save': 'Save',
  'study.cancel': 'Cancel',
  'study.edit': 'Edit {name}',
  'study.delete': 'Delete {name}',
  'study.delete_confirm': 'Delete the module “{name}”?',
  'study.saved': 'Saved.',
  'study.no_modules': 'No modules yet.',
  'study.set_up': 'Add your modules under Settings → Study progress.',
  'study.of_ects': '{passed} of {total} ECTS',
  'study.percent': '{n} %',
  'study.average': 'Average {grade}',
  'study.average_hint': 'ECTS-weighted, based on {ects} ECTS',
  'study.passed_modules': '{n} of {total} modules passed',
  'study.in_progress': 'In progress',
  'study.name_required': 'Enter a module name.',
  'study.ects_invalid': 'ECTS must be between 0.5 and 60 in steps of 0.5.',
  'study.grade_only_passed': 'Only passed modules have a grade.',
  'study.grade_invalid': 'A passing grade lies between 1.0 and 4.0.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
