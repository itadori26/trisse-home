import { describe, expect, it } from 'vitest';
import type { Countdown, CountdownInput } from '../domain/countdown';
import { type CountdownGateway, CountdownUseCases } from './countdownUseCases';

class FakeGateway implements CountdownGateway {
  calls: string[] = [];
  last: CountdownInput | null = null;
  async list(grace: number): Promise<Countdown[]> {
    this.calls.push(`list:${grace}`);
    return [];
  }
  async create(input: CountdownInput): Promise<Countdown> {
    this.calls.push('create');
    this.last = input;
    return { id: '1', ...input };
  }
  async update(id: string, input: CountdownInput): Promise<Countdown> {
    this.calls.push(`update:${id}`);
    this.last = input;
    return { id, ...input };
  }
  async remove(id: string): Promise<void> {
    this.calls.push(`remove:${id}`);
  }
}

describe('CountdownUseCases', () => {
  it('creates without id, updates with id and trims the input', async () => {
    const gw = new FakeGateway();
    const uc = new CountdownUseCases(gw);
    await uc.save(null, { title: ' Exam ', emoji: ' 📚 ', date: '2026-12-24', time: '' });
    expect(gw.last).toEqual({ title: 'Exam', emoji: '📚', date: '2026-12-24', time: '' });
    await uc.save('7', { title: 'Exam', emoji: '', date: '2026-12-24', time: '09:00' });
    await uc.list(3);
    await uc.remove('7');
    expect(gw.calls).toEqual(['create', 'update:7', 'list:3', 'remove:7']);
  });
});
