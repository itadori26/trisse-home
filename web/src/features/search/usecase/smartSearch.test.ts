import { describe, expect, it, vi } from 'vitest';
import type { Answer } from '../domain/smart/answer';
import type { Match } from '../domain/smart/intent';
import { INTENTS } from '../domain/smart/intents';
import { type SearchSource, SmartSearch } from './smartSearch';

const metric = (m: Match, value: string): Answer => ({
  kind: 'metric',
  intent: m.intent.id,
  title: m.intent.id,
  value,
  at: null,
});

function setup(opts: { canAct?: boolean; sources?: string[] } = {}) {
  let clock = 0;
  const system = {
    id: 'system',
    resolve: vi.fn<(m: Match, s: AbortSignal) => Promise<Answer>>(async (m) => metric(m, '12 %')),
  } satisfies SearchSource;
  const docker = {
    id: 'docker',
    entities: vi.fn(async () => [{ kind: 'container', id: 'c1', name: 'jellyfin' }]),
    resolve: vi.fn(async (m: Match) => metric(m, 'running')),
    execute: vi.fn(async () => {}),
  } satisfies SearchSource;
  const all = { system, docker };
  const search = new SmartSearch({
    intents: INTENTS,
    sources: (opts.sources ?? ['system', 'docker']).map((id) => all[id as keyof typeof all]),
    canAct: opts.canAct ?? true,
    askBack: (m) => ({
      kind: 'text',
      intent: m.intent.id,
      title: 'which?',
      text: m.missing.join(),
      at: null,
    }),
    now: () => new Date(2026, 10, 12, 10, 0, 0, clock),
  });
  return { search, system, docker, tick: (ms: number) => (clock += ms) };
}

const signal = () => new AbortController().signal;

describe('SmartSearch', () => {
  it('answers through the matching source and caches per source', async () => {
    const { search, system, tick } = setup();
    expect((await search.answer('cpu', signal()))[0]).toMatchObject({ intent: 'system.cpu', value: '12 %' });
    const calls = vi.mocked(system.resolve).mock.calls.length; // "cpu" also asks for the temperature
    await search.answer('cpu', signal());
    expect(system.resolve).toHaveBeenCalledTimes(calls);
    tick(16_000);
    await search.answer('cpu', signal());
    expect(system.resolve).toHaveBeenCalledTimes(2 * calls);
  });

  it('loads entities once and uses them for names', async () => {
    const { search, docker } = setup();
    await search.answer('jellyfin', signal());
    await search.answer('jellyfin restart', signal());
    expect(docker.entities).toHaveBeenCalledTimes(1);
    expect(search.matches('jellyfin')[0]?.intent.id).toBe('docker.status');
  });

  it('registers no intents for disabled modules', async () => {
    const { search } = setup({ sources: ['system'] });
    expect(await search.answer('docker', signal())).toEqual([]);
  });

  it('asks back for missing slots without calling the source', async () => {
    const { search, docker } = setup();
    const [a] = await search.answer('neu starten', signal());
    expect(a).toMatchObject({ kind: 'text', title: 'which?', text: 'container' });
    expect(docker.resolve).not.toHaveBeenCalled();
  });

  it('hides actions from guests', async () => {
    const { search } = setup({ canAct: false });
    expect(search.matches('restart jellyfin').map((m) => m.intent.id)).not.toContain('docker.restart');
    await expect(
      search.execute({
        source: 'docker',
        action: 'restart',
        target: { kind: 'container', id: 'c1', name: 'jellyfin' },
        label: '',
        confirm: '',
      }),
    ).rejects.toThrow();
  });

  it('forgets cached answers of a source after an action', async () => {
    const { search, docker } = setup();
    await search.answer('jellyfin', signal());
    await search.execute({
      source: 'docker',
      action: 'restart',
      target: { kind: 'container', id: 'c1', name: 'jellyfin' },
      label: '',
      confirm: '',
    });
    await search.answer('jellyfin', signal());
    expect(docker.execute).toHaveBeenCalledTimes(1);
    expect(docker.resolve).toHaveBeenCalledTimes(2);
  });

  it('rejects an aborted query and does not share its result', async () => {
    const { search, system } = setup();
    let release: () => void = () => {};
    vi.mocked(system.resolve).mockImplementationOnce(
      (m, s) =>
        new Promise((resolve, reject) => {
          release = () => resolve(metric(m, 'old'));
          const abort = () => reject(new DOMException('aborted', 'AbortError'));
          if (s.aborted) abort();
          s.addEventListener('abort', abort);
        }),
    );
    const first = new AbortController();
    const pending = search.answer('cpu', first.signal);
    await vi.waitFor(() => expect(system.resolve).toHaveBeenCalled());
    first.abort();
    await expect(pending).rejects.toThrow();
    release();
    expect((await search.answer('cpu', signal()))[0]).toMatchObject({ value: '12 %' });
  });

  it('drops the answer of a failing source only', async () => {
    const { search, system } = setup();
    vi.mocked(system.resolve).mockImplementation(async (m) => {
      if (m.intent.id === 'system.cpu') throw new Error('down');
      return metric(m, '40 °C');
    });
    const answers = await search.answer('cpu', signal());
    expect(answers.map((a) => a.intent)).toEqual(['system.temperature']);
  });

  it('puts direct answers first and survives a throwing direct source', async () => {
    const calc: SearchSource = {
      id: 'calc',
      direct: (q) =>
        q === 'cpu' ? { kind: 'metric', intent: 'calc.expression', title: q, value: '1', at: null } : null,
    };
    const broken: SearchSource = {
      id: 'broken',
      direct: () => {
        throw new Error('boom');
      },
    };
    const search = new SmartSearch({
      intents: INTENTS,
      sources: [calc, broken, { id: 'system', resolve: async (m) => metric(m, '12 %') }],
      canAct: false,
      askBack: () => ({ kind: 'text', intent: '', title: '', text: '', at: null }),
    });
    const answers = await search.answer('cpu', signal());
    expect(answers[0]?.intent).toBe('calc.expression');
    expect(answers.length).toBeLessThanOrEqual(3);
    expect(search.direct('wetter')).toEqual([]);
  });
});
