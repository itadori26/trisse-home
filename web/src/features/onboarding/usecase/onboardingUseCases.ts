/** Output port: onboarding backend. */
export interface OnboardingGateway {
  completed(): Promise<boolean>;
  setCompleted(completed: boolean): Promise<void>;
}

/** Application logic of the guided tour. */
export class OnboardingUseCases {
  constructor(private readonly gateway: OnboardingGateway) {}

  /** Whether the tour should be shown automatically. */
  async shouldShowTour(): Promise<boolean> {
    return !(await this.gateway.completed());
  }

  /** Marks the tour as done (finished or skipped). */
  finish(): Promise<void> {
    return this.gateway.setCompleted(true);
  }

  /** Makes the tour appear again. */
  restart(): Promise<void> {
    return this.gateway.setCompleted(false);
  }
}
