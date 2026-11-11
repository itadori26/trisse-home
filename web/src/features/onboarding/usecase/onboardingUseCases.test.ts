import { describe, expect, it, vi } from 'vitest';
import { OnboardingUseCases, type OnboardingGateway } from './onboardingUseCases';

describe('OnboardingUseCases', () => {
  it('shows the tour until it was completed', async () => {
    const gateway = {
      completed: vi.fn().mockResolvedValue(false),
      setCompleted: vi.fn(),
    } satisfies OnboardingGateway;
    const uc = new OnboardingUseCases(gateway);
    expect(await uc.shouldShowTour()).toBe(true);
    await uc.finish();
    expect(gateway.setCompleted).toHaveBeenCalledWith(true);
    await uc.restart();
    expect(gateway.setCompleted).toHaveBeenLastCalledWith(false);
  });
});
