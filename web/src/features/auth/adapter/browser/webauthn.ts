import {
  type PasskeyAssertion,
  type PasskeyAttestation,
  type PasskeyRegistrationOptions,
  type SecondFactorChallenge,
  fromBase64Url,
  toBase64Url,
} from '../../domain/passkey';
import type { PasskeyAuthenticator } from '../../usecase/ports';

/** PasskeyAuthenticator backed by navigator.credentials (WebAuthn). */
export class BrowserPasskeys implements PasskeyAuthenticator {
  supported(): boolean {
    return typeof window !== 'undefined' && window.isSecureContext && 'PublicKeyCredential' in window;
  }

  async create(o: PasskeyRegistrationOptions): Promise<PasskeyAttestation> {
    const cred = (await navigator.credentials.create({
      publicKey: {
        challenge: fromBase64Url(o.challenge),
        rp: { id: o.rpId, name: 'trisse-home' },
        user: { id: fromBase64Url(o.userId), name: o.userName, displayName: o.userName },
        pubKeyCredParams: [
          { type: 'public-key', alg: -7 },
          { type: 'public-key', alg: -8 },
          { type: 'public-key', alg: -257 },
        ],
        excludeCredentials: o.excludeCredentials.map((id) => ({
          type: 'public-key' as const,
          id: fromBase64Url(id),
        })),
        attestation: 'none',
        authenticatorSelection: { residentKey: 'preferred', userVerification: 'preferred' },
        timeout: 120_000,
      },
    })) as PublicKeyCredential | null;
    if (!cred) throw new Error('cancelled');
    const res = cred.response as AuthenticatorAttestationResponse;
    return {
      clientDataJSON: toBase64Url(res.clientDataJSON),
      attestationObject: toBase64Url(res.attestationObject),
    };
  }

  async get(c: SecondFactorChallenge): Promise<PasskeyAssertion> {
    const cred = (await navigator.credentials.get({
      publicKey: {
        challenge: fromBase64Url(c.challenge),
        rpId: c.rpId,
        allowCredentials: c.allowCredentials.map((id) => ({
          type: 'public-key' as const,
          id: fromBase64Url(id),
        })),
        userVerification: 'preferred',
        timeout: 120_000,
      },
    })) as PublicKeyCredential | null;
    if (!cred) throw new Error('cancelled');
    const res = cred.response as AuthenticatorAssertionResponse;
    return {
      credentialId: toBase64Url(cred.rawId),
      clientDataJSON: toBase64Url(res.clientDataJSON),
      authenticatorData: toBase64Url(res.authenticatorData),
      signature: toBase64Url(res.signature),
    };
  }
}
