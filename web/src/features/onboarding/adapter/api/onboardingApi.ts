import type { HttpClient } from '../../../../shared/api-client/http';
import type { StatusDTO } from '../../../../shared/api-client/generated/onboarding';
import type { UpdateDTO } from '../../../../shared/api-client/generated/onboarding-requests';
import type { OnboardingGateway } from '../../usecase/onboardingUseCases';

/** OnboardingGateway backed by the REST API. */
export class OnboardingApi implements OnboardingGateway {
  constructor(private readonly http: HttpClient) {}

  async completed(): Promise<boolean> {
    return (await this.http.get<StatusDTO>('/api/onboarding')).completed;
  }

  async setCompleted(completed: boolean): Promise<void> {
    const body: UpdateDTO = { completed };
    await this.http.put<StatusDTO>('/api/onboarding', body);
  }
}
