import type { HttpClient } from '../../../../shared/api-client/http';
import type { SessionDTO, StatusDTO } from '../../../../shared/api-client/generated/auth';
import type {
  ChangePasswordRequestDTO,
  LoginRequestDTO,
  SetupRequestDTO,
} from '../../../../shared/api-client/generated/auth-requests';
import type { AuthStatus, Session } from '../../domain/auth';
import type { AuthGateway } from '../../usecase/ports';

const toSession = (dto: SessionDTO): Session => ({
  csrfToken: dto.csrfToken,
  expiresAt: new Date(dto.expiresAt),
});

/** AuthGateway implementation backed by the REST API. */
export class AuthApi implements AuthGateway {
  constructor(private readonly http: HttpClient) {}

  async status(): Promise<AuthStatus> {
    const dto = await this.http.get<StatusDTO>('/api/auth/status');
    return {
      setupRequired: dto.setupRequired,
      authenticated: dto.authenticated,
      csrfToken: dto.csrfToken ?? null,
    };
  }

  async setup(setupToken: string, password: string): Promise<Session> {
    const body: SetupRequestDTO = { setupToken, password };
    return toSession(await this.http.post<SessionDTO>('/api/auth/setup', body));
  }

  async login(password: string): Promise<Session> {
    const body: LoginRequestDTO = { password };
    return toSession(await this.http.post<SessionDTO>('/api/auth/login', body));
  }

  async logout(): Promise<void> {
    await this.http.post('/api/auth/logout');
  }

  async changePassword(currentPassword: string, newPassword: string): Promise<Session> {
    const body: ChangePasswordRequestDTO = { currentPassword, newPassword };
    return toSession(await this.http.put<SessionDTO>('/api/auth/password', body));
  }
}
