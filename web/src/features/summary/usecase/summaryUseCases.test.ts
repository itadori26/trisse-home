import { describe, expect, it, vi } from 'vitest';
import { SummaryUseCases, type SummaryGateway } from './summaryUseCases';

const settings = { configured: true, url: 'http://ollama:11434', model: 'm', maxInputChars: 12000 };

describe('SummaryUseCases (#75)', () => {
  it('asks once whether Ollama is set up and treats errors as unavailable', async () => {
    const g = {
      settings: vi.fn().mockResolvedValue({ ...settings, configured: false }),
      saveSettings: vi.fn().mockResolvedValue(settings),
      disable: vi.fn(),
      models: vi.fn(),
      summarize: vi.fn().mockResolvedValue({ summary: '- a', truncated: false, cached: false }),
    } satisfies SummaryGateway;
    const uc = new SummaryUseCases(g);
    expect(await uc.available()).toBe(false);
    expect(await uc.available()).toBe(false);
    expect(g.settings).toHaveBeenCalledTimes(1);
    await uc.saveSettings(' http://ollama:11434 ', 'm', 12000);
    expect(g.saveSettings).toHaveBeenCalledWith('http://ollama:11434', 'm', 12000);
    expect(await uc.available()).toBe(true);
    await uc.link('https://news.example/a');
    expect(g.summarize).toHaveBeenCalledWith({ url: 'https://news.example/a', language: 'en' });
    await uc.disable();
    expect(await uc.available()).toBe(false);

    const failing = new SummaryUseCases({ ...g, settings: vi.fn().mockRejectedValue(new Error('404')) });
    expect(await failing.available()).toBe(false);
  });
});
