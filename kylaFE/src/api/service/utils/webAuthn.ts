import { PublicKeyCredentialCreationOptionsJSON, PublicKeyCredentialRequestOptionsJSON } from '@simplewebauthn/typescript-types';
import crypto from 'crypto';

export interface WebAuthnResponse {
  id: string;
  type: string;
  rawId: string;
  response: {
    clientDataJSON: string;
    authenticatorData: string;
    signature: string;
    userHandle?: string;
  };
}

export interface WebAuthnError extends Error {
  code?: string;
  name: string;
}

export class WebAuthnService {
  private static instance: WebAuthnService;
  private rpID: string;
  private rpName: string;

  private constructor() {
    this.rpID = process.env.RP_ID || 'localhost';
    this.rpName = process.env.RP_NAME || 'Resolv';
  }

  public static getInstance(): WebAuthnService {
    if (!WebAuthnService.instance) {
      WebAuthnService.instance = new WebAuthnService();
    }
    return WebAuthnService.instance;
  }

  public generateChallenge(): string {
    return crypto.randomBytes(32).toString('base64url');
  }

  public generateCredentialId(): string {
    return crypto.randomBytes(16).toString('base64url');
  }

  public async generateRegistrationOptions(userId: string, username: string): Promise<PublicKeyCredentialCreationOptionsJSON> {
    const challenge = this.generateChallenge();
    const credentialId = this.generateCredentialId();

    return {
      challenge,
      rp: {
        name: this.rpName,
        id: this.rpID,
      },
      user: {
        id: userId,
        name: username,
        displayName: username,
      },
      pubKeyCredParams: [
        {
          type: 'public-key',
          alg: -7, // ES256
        },
      ],
      timeout: 60000,
      attestation: 'none',
      authenticatorSelection: {
        authenticatorAttachment: 'platform',
        userVerification: 'required',
      },
    };
  }

  public async generateAuthenticationOptions(credentialId: string): Promise<PublicKeyCredentialRequestOptionsJSON> {
    const challenge = this.generateChallenge();

    return {
      challenge,
      rpId: this.rpID,
      allowCredentials: [
        {
          type: 'public-key',
          id: credentialId,
        },
      ],
      userVerification: 'required',
      timeout: 60000,
    };
  }

  public async verifyRegistrationResponse(
    response: WebAuthnResponse,
    expectedChallenge: string,
    expectedOrigin: string,
    expectedRPID: string
  ): Promise<boolean> {
    try {
      const clientData = JSON.parse(
        Buffer.from(response.response.clientDataJSON, 'base64url').toString()
      );

      // Verify challenge
      if (clientData.challenge !== expectedChallenge) {
        return false;
      }

      // Verify origin
      if (clientData.origin !== expectedOrigin) {
        return false;
      }

      // Verify type
      if (clientData.type !== 'webauthn.create') {
        return false;
      }

      // Verify authenticator data
      const authenticatorData = Buffer.from(response.response.authenticatorData, 'base64url');
      const rpIdHash = authenticatorData.slice(0, 32);
      const expectedRPIDHash = crypto
        .createHash('sha256')
        .update(expectedRPID)
        .digest();

      if (!rpIdHash.equals(expectedRPIDHash)) {
        return false;
      }

      // Verify signature
      const signature = Buffer.from(response.response.signature, 'base64url');
      const clientDataHash = crypto
        .createHash('sha256')
        .update(response.response.clientDataJSON)
        .digest();

      const dataToVerify = Buffer.concat([authenticatorData, clientDataHash]);

      // In a real implementation, you would verify the signature using the public key
      // This is just a placeholder
      return true;
    } catch (error) {
      console.error('Registration verification failed:', error);
      return false;
    }
  }

  public async verifyAuthenticationResponse(
    response: WebAuthnResponse,
    expectedChallenge: string,
    expectedOrigin: string,
    expectedRPID: string
  ): Promise<boolean> {
    try {
      const clientData = JSON.parse(
        Buffer.from(response.response.clientDataJSON, 'base64url').toString()
      );

      // Verify challenge
      if (clientData.challenge !== expectedChallenge) {
        return false;
      }

      // Verify origin
      if (clientData.origin !== expectedOrigin) {
        return false;
      }

      // Verify type
      if (clientData.type !== 'webauthn.get') {
        return false;
      }

      // Verify authenticator data
      const authenticatorData = Buffer.from(response.response.authenticatorData, 'base64url');
      const rpIdHash = authenticatorData.slice(0, 32);
      const expectedRPIDHash = crypto
        .createHash('sha256')
        .update(expectedRPID)
        .digest();

      if (!rpIdHash.equals(expectedRPIDHash)) {
        return false;
      }

      // Verify signature
      const signature = Buffer.from(response.response.signature, 'base64url');
      const clientDataHash = crypto
        .createHash('sha256')
        .update(response.response.clientDataJSON)
        .digest();

      const dataToVerify = Buffer.concat([authenticatorData, clientDataHash]);

      // In a real implementation, you would verify the signature using the public key
      // This is just a placeholder
      return true;
    } catch (error) {
      console.error('Authentication verification failed:', error);
      return false;
    }
  }

  private handleWebAuthnError(error: unknown): WebAuthnError {
    if (error instanceof Error) {
      const webAuthnError = error as WebAuthnError;
      webAuthnError.name = 'WebAuthnError';
      
      if (error.message.includes('NotAllowedError')) {
        webAuthnError.code = 'USER_CANCELLED';
      } else if (error.message.includes('NotSupportedError')) {
        webAuthnError.code = 'NOT_SUPPORTED';
      } else if (error.message.includes('InvalidStateError')) {
        webAuthnError.code = 'INVALID_STATE';
      }
      
      return webAuthnError;
    }
    
    return {
      name: 'WebAuthnError',
      message: 'An unknown error occurred',
      code: 'UNKNOWN_ERROR',
    };
  }
} 