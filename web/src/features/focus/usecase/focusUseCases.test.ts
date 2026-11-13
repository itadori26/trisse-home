import { describe, expect, it } from 'vitest';
import type { FocusDay } from '../domain/focus';
import { type FocusGateway, FocusUseCases, type Notifier } from './focusUseCases';

class FakeGateway implements FocusGateway {
  recorded: [number, Date][] = [];
  async stats(): Promise<FocusDay[]> {
    return [];
  }
  async record(minutes: number, endedAt: Date): Promise<FocusDay[]> {
    this.recorded.push([minutes, endedAt]);
    return [{ date: '2026-11-12', sessions: 1, minutes }];
  }
}

class FakeNotifier implements Notifier {
  shown: string[] = [];
  constructor(private readonly perm: NotificationPermission | 'unsupported') {}
  permission() {
    return this.perm;
  }
  async request() {
    return this.perm;
  }
  show(title: string) {
    this.shown.push(title);
  }
}

describe('FocusUseCases', () => {
  it('records sessions', async () => {
    const gw = new FakeGateway();
    const days = await new FocusUseCases(gw, new FakeNotifier('default')).record(25, new Date(0));
    expect(gw.recorded).toEqual([[25, new Date(0)]]);
    expect(days[0]?.minutes).toBe(25);
  });

  it('notifies only when enabled and allowed', () => {
    const granted = new FakeNotifier('granted');
    const uc = new FocusUseCases(new FakeGateway(), granted);
    uc.notify(false, 'a', '');
    uc.notify(true, 'b', '');
    const denied = new FakeNotifier('denied');
    new FocusUseCases(new FakeGateway(), denied).notify(true, 'c', '');
    expect(granted.shown).toEqual(['b']);
    expect(denied.shown).toEqual([]);
  });
});
