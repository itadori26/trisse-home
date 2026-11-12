/** A registered passkey. */
export interface Passkey {
  id: string;
  name: string;
  createdAt: Date;
  lastUsedAt: Date | null;
}

/** Second login step after a correct password (base64url values). */
export interface SecondFactorChallenge {
  ticket: string;
  challenge: string;
  rpId: string;
  allowCredentials: string[];
  /** TOTP is set up as well, so a 6-digit code also works. */
  totp: boolean;
}

/** Parameters for creating a passkey (base64url values). */
export interface PasskeyRegistrationOptions {
  challenge: string;
  rpId: string;
  userId: string;
  userName: string;
  excludeCredentials: string[];
}

/** Browser response of navigator.credentials.create() (base64url). */
export interface PasskeyAttestation {
  clientDataJSON: string;
  attestationObject: string;
}

/** Browser response of navigator.credentials.get() (base64url). */
export interface PasskeyAssertion {
  credentialId: string;
  clientDataJSON: string;
  authenticatorData: string;
  signature: string;
}

/** Encodes bytes as unpadded base64url. */
export function toBase64Url(bytes: ArrayBuffer | Uint8Array): string {
  const b = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes);
  let s = '';
  for (const x of b) s += String.fromCharCode(x);
  return btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

/** Decodes unpadded base64url. */
export function fromBase64Url(s: string): Uint8Array<ArrayBuffer> {
  const bin = atob(s.replace(/-/g, '+').replace(/_/g, '/') + '='.repeat((4 - (s.length % 4)) % 4));
  const out = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i);
  return out;
}
