import { describe, expect, it } from 'vitest';
import { daysUntil, parseAssignees, splitDue, validateChore, type Chore } from './chores';

const chore = (id: string, dueNow: boolean): Chore => ({
  id,
  title: id,
  everyDays: 1,
  due: '2026-11-23',
  assignees: [],
  assignee: '',
  dueNow,
  overdue: false,
  lastDoneAt: null,
  lastDoneBy: '',
  visibility: 'everyone',
  mine: true,
});

describe('chores domain', () => {
  it('parses names from lines or commas', () => {
    expect(parseAssignees('Anna\n Ben , Mia\n\n')).toEqual(['Anna', 'Ben', 'Mia']);
  });

  it('splits due and later chores', () => {
    const { due, later } = splitDue([chore('a', true), chore('b', false)]);
    expect(due.map((c) => c.id)).toEqual(['a']);
    expect(later.map((c) => c.id)).toEqual(['b']);
  });

  it('counts days to a due day', () => {
    const today = new Date(2026, 10, 23, 22, 0);
    expect(daysUntil('2026-11-23', today)).toBe(0);
    expect(daysUntil('2026-11-25', today)).toBe(2);
    expect(daysUntil('2026-11-20', today)).toBe(-3);
  });

  it('validates the dialog', () => {
    const ok = { title: 'Dishes', everyDays: 7, due: '2026-11-23', assignees: '' };
    expect(validateChore(ok)).toBeNull();
    expect(validateChore({ ...ok, title: ' ' })).toBe('chores.error.title');
    expect(validateChore({ ...ok, everyDays: 0 })).toBe('chores.error.every');
    expect(validateChore({ ...ok, due: '' })).toBe('chores.error.due');
    expect(validateChore({ ...ok, assignees: 'a,b,c,d,e,f,g,h,i' })).toBe('chores.error.assignees');
  });
});
