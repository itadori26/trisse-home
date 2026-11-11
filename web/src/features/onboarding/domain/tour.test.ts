import { describe, expect, it } from 'vitest';
import { availableSteps, popoverPosition, tourSteps } from './tour';

describe('availableSteps', () => {
  it('drops steps whose target is missing but keeps centered ones', () => {
    const steps = availableSteps(tourSteps, (t) => t !== 'tab-Server');
    expect(steps.map((s) => s.id)).not.toContain('server');
    expect(steps[0]?.id).toBe('welcome');
    expect(steps.at(-1)?.id).toBe('done');
  });
});

describe('popoverPosition', () => {
  const pop = { width: 300, height: 150 };
  const vp = { width: 1200, height: 800 };
  it('centers without target', () => {
    expect(popoverPosition(null, pop, vp)).toEqual({ top: 325, left: 450 });
  });
  it('places below a target near the top', () => {
    expect(popoverPosition({ top: 10, left: 1100, width: 80, height: 30 }, pop, vp)).toEqual({
      top: 52,
      left: 888,
    });
  });
  it('places above a target near the bottom', () => {
    expect(popoverPosition({ top: 700, left: 100, width: 200, height: 50 }, pop, vp)).toEqual({
      top: 538,
      left: 50,
    });
  });
  it('stays inside small viewports', () => {
    const p = popoverPosition({ top: 5, left: 0, width: 40, height: 20 }, pop, { width: 320, height: 600 });
    expect(p.left).toBe(12);
  });
});
