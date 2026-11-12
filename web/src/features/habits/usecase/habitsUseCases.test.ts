import { describe, expect, it } from 'vitest';
import type { Habit, HabitInput } from '../domain/habits';
import { type HabitsGateway, HabitsUseCases } from './habitsUseCases';

const habit = (id: string, input?: HabitInput): Habit => ({
  id,
  name: input?.name ?? 'Read',
  emoji: input?.emoji ?? '',
  days: input?.days ?? [],
  doneToday: false,
  current: 0,
  best: 0,
  heatmap: [],
});

class FakeGateway implements HabitsGateway {
  calls: string[] = [];
  last: HabitInput | null = null;
  async list() {
    return [habit('1')];
  }
  async create(input: HabitInput) {
    this.calls.push('create');
    this.last = input;
    return habit('2', input);
  }
  async update(id: string, input: HabitInput) {
    this.calls.push(`update:${id}`);
    this.last = input;
    return habit(id, input);
  }
  async remove(id: string) {
    this.calls.push(`remove:${id}`);
  }
  async toggle(id: string, date: string) {
    this.calls.push(`toggle:${id}:${date}`);
    return habit(id);
  }
}

describe('HabitsUseCases', () => {
  it('cleans the input and toggles today and yesterday', async () => {
    const gw = new FakeGateway();
    const uc = new HabitsUseCases(gw, () => new Date(2026, 10, 1, 9, 0));
    await uc.save(null, { name: ' Read ', emoji: ' 📖 ', days: [5, 1, 5] });
    expect(gw.last).toEqual({ name: 'Read', emoji: '📖', days: [1, 5] });
    await uc.save('1', { name: 'Gym', emoji: '', days: [] });
    await uc.toggleToday('1');
    await uc.toggleYesterday('1');
    await uc.remove('1');
    expect(gw.calls).toEqual([
      'create',
      'update:1',
      'toggle:1:2026-11-01',
      'toggle:1:2026-10-31',
      'remove:1',
    ]);
  });
});
