import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  EnrollmentDTO,
  RecoveryCodesDTO,
  RevokedDTO,
  SessionDTO,
  SessionInfoDTO,
  SSOTicketDTO,
  StatusDTO,
  TokenDTO,
  TwoFactorStatusDTO,
  CreatedTokenDTO,
  GuestDTO,
  SecurityEventDTO,
} from '../../../../shared/api-client/generated/auth';
import type {
  ChangePasswordRequestDTO,
  DisableTwoFactorDTO,
  GuestRequestDTO,
  LoginRequestDTO,
  SetupRequestDTO,
  SSOSecondFactorDTO,
  TokenRequestDTO,
  TwoFactorCodeDTO,
} from '../../../../shared/api-client/generated/auth-requests';
import type { ApiToken } from '../../domain/apiToken';
import type { AuthStatus, Guest, Session } from '../../domain/auth';
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
      guest: dto.role === 'guest',
      sso: dto.sso ?? { oidc: false, header: false, localLogin: true },
    };
  }

  async demoLogin(): Promise<Session> {
    return toSession(await this.http.post<SessionDTO>('/api/auth/demo'));
  }

  async setup(setupToken: string, password: string): Promise<Session> {
    const body: SetupRequestDTO = { setupToken, password };
    return toSession(await this.http.post<SessionDTO>('/api/auth/setup', body));
  }

  async login(password: string, code?: string, username?: string): Promise<Session> {
    const body: LoginRequestDTO = { password, ...(code ? { code } : {}), ...(username ? { username } : {}) };
    return toSession(await this.http.post<SessionDTO>('/api/auth/login', body));
  }

  async logout(): Promise<void> {
    await this.http.post('/api/auth/logout');
  }

  async changePassword(currentPassword: string, newPassword: string): Promise<Session> {
    const body: ChangePasswordRequestDTO = { currentPassword, newPassword };
    return toSession(await this.http.put<SessionDTO>('/api/auth/password', body));
  }

  async guests(): Promise<Guest[]> {
    return (await this.http.get<GuestDTO[]>('/api/auth/guests')).map(toGuest);
  }

  async createGuest(name: string, password: string, tabs: string[]): Promise<Guest> {
    const body: GuestRequestDTO = { name, password, tabs, disabled: false };
    return toGuest(await this.http.post<GuestDTO>('/api/auth/guests', body));
  }

  async updateGuest(id: string, tabs: string[], disabled: boolean, password: string): Promise<Guest> {
    const body: GuestRequestDTO = { tabs, disabled, ...(password ? { password } : {}) };
    return toGuest(await this.http.put<GuestDTO>(`/api/auth/guests/${encodeURIComponent(id)}`, body));
  }

  deleteGuest(id: string): Promise<void> {
    return this.http.delete(`/api/auth/guests/${encodeURIComponent(id)}`);
  }

  async apiTokens(): Promise<ApiToken[]> {
    return (await this.http.get<TokenDTO[]>('/api/auth/tokens')).map(toApiToken);
  }

  async createApiToken(name: string, scopes: string[], expiresInDays: number): Promise<string> {
    const body: TokenRequestDTO = { name, scopes, expiresInDays };
    return (await this.http.post<CreatedTokenDTO>('/api/auth/tokens', body)).token;
  }

  revokeApiToken(id: string): Promise<void> {
    return this.http.delete(`/api/auth/tokens/${encodeURIComponent(id)}`);
  }

  async headerLogin(): Promise<Session | { ticket: string }> {
    const res = await this.http.post<SessionDTO | SSOTicketDTO>('/api/auth/sso/header');
    return 'ticket' in res ? { ticket: res.ticket } : toSession(res);
  }

  async ssoSecondFactor(ticket: string, code: string): Promise<Session> {
    const body: SSOSecondFactorDTO = { ticket, code };
    return toSession(await this.http.post<SessionDTO>('/api/auth/sso/2fa', body));
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

const toGuest = (d: GuestDTO): Guest => ({ ...d, tabs: d.tabs ?? [], createdAt: new Date(d.createdAt) });

const toApiToken = (d: TokenDTO): ApiToken => ({
  ...d,
  scopes: d.scopes ?? [],
  createdAt: new Date(d.createdAt),
  expiresAt: d.expiresAt ? new Date(d.expiresAt) : null,
  lastUsedAt: d.lastUsedAt ? new Date(d.lastUsedAt) : null,
});
