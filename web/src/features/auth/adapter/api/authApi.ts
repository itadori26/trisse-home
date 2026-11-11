import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  EnrollmentDTO,
  RecoveryCodesDTO,
  RevokedDTO,
  SessionDTO,
  SessionInfoDTO,
  StatusDTO,
  TwoFactorStatusDTO,
  SecurityEventDTO,
} from '../../../../shared/api-client/generated/auth';
import type {
  ChangePasswordRequestDTO,
  DisableTwoFactorDTO,
  LoginRequestDTO,
  SetupRequestDTO,
  TwoFactorCodeDTO,
} from '../../../../shared/api-client/generated/auth-requests';
import type { AuthStatus, Session } from '../../domain/auth';
import type { TwoFactorEnrollment, TwoFactorStatus } from '../../domain/twoFactor';
import type { SecurityEvent } from '../../domain/securityEvent';
import type { SessionInfo } from '../../domain/session';
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
      demo: dto.demo,
    };
  }

  async demoLogin(): Promise<Session> {
    return toSession(await this.http.post<SessionDTO>('/api/auth/demo'));
  }

  async setup(setupToken: string, password: string): Promise<Session> {
    const body: SetupRequestDTO = { setupToken, password };
    return toSession(await this.http.post<SessionDTO>('/api/auth/setup', body));
  }

  async login(password: string, code?: string): Promise<Session> {
    const body: LoginRequestDTO = code ? { password, code } : { password };
    return toSession(await this.http.post<SessionDTO>('/api/auth/login', body));
  }

  async logout(): Promise<void> {
    await this.http.post('/api/auth/logout');
  }

  async changePassword(currentPassword: string, newPassword: string): Promise<Session> {
    const body: ChangePasswordRequestDTO = { currentPassword, newPassword };
    return toSession(await this.http.put<SessionDTO>('/api/auth/password', body));
  }

  async securityEvents(): Promise<SecurityEvent[]> {
    const list = await this.http.get<SecurityEventDTO[]>('/api/auth/events');
    return list.map((e) => ({
      id: e.id,
      at: new Date(e.at),
      type: e.type,
      ip: e.ip,
      userAgent: e.userAgent,
      detail: e.detail,
    }));
  }

  async sessions(): Promise<SessionInfo[]> {
    const dtos = await this.http.get<SessionInfoDTO[]>('/api/auth/sessions');
    return dtos.map((d) => ({
      id: d.id,
      userAgent: d.userAgent,
      clientIp: d.clientIp,
      createdAt: new Date(d.createdAt),
      lastSeenAt: new Date(d.lastSeenAt),
      current: d.current,
    }));
  }

  revokeSession(id: string): Promise<void> {
    return this.http.delete(`/api/auth/sessions/${encodeURIComponent(id)}`);
  }

  async revokeOtherSessions(): Promise<number> {
    return (await this.http.post<RevokedDTO>('/api/auth/sessions/revoke-others')).revoked;
  }

  async twoFactorStatus(): Promise<TwoFactorStatus> {
    const d = await this.http.get<TwoFactorStatusDTO>('/api/auth/totp');
    return { enabled: d.enabled, recoveryCodesLeft: d.recoveryCodesLeft };
  }

  async startTwoFactor(): Promise<TwoFactorEnrollment> {
    const d = await this.http.post<EnrollmentDTO>('/api/auth/totp/setup');
    return { secret: d.secret, uri: d.uri };
  }

  async confirmTwoFactor(code: string): Promise<string[]> {
    const body: TwoFactorCodeDTO = { code };
    return (await this.http.post<RecoveryCodesDTO>('/api/auth/totp/confirm', body)).codes;
  }

  async disableTwoFactor(password: string, code: string): Promise<void> {
    const body: DisableTwoFactorDTO = { password, code };
    await this.http.post('/api/auth/totp/disable', body);
  }
}
