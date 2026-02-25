export interface LoginRequest {
  email?: string;
  username?: string;
  password?: string;
  mfaResponse?: {
    credentialId: string;
    clientDataJSON: string;
    authenticatorData: string;
    signature: string;
  };
  passkeyResponse?: {
    credentialId: string;
    clientDataJSON: string;
    authenticatorData: string;
    signature: string;
  };
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user?: {
    id: string;
    email: string;
    username: string;
    mfaEnabled: boolean;
    mfaChallenge?: string;
    mfaCredentialId?: string;
  };
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
  firstName: string;
  lastName: string;
  referralCode?: string;
}

export interface RegisterResponse {
  success: boolean;
  message: string;
  user?: {
    id: string;
    email: string;
    username: string;
  };
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ForgotPasswordResponse {
  success: boolean;
  message: string;
}

export interface ResetPasswordRequest {
  token: string;
  password: string;
}

export interface ResetPasswordResponse {
  success: boolean;
  message: string;
}

export interface VerifyEmailRequest {
  token: string;
}

export interface VerifyEmailResponse {
  success: boolean;
  message: string;
}

export interface RefreshTokenRequest {
  refreshToken: string;
}

export interface RefreshTokenResponse {
  accessToken: string;
  refreshToken: string;
}

export interface LogoutRequest {
  refreshToken: string;
}

export interface LogoutResponse {
  success: boolean;
  message: string;
}

export interface GetUserRequest {
  accessToken: string;
}

export interface GetUserResponse {
  user: {
    id: string;
    email: string;
    username: string;
    firstName: string;
    lastName: string;
    mfaEnabled: boolean;
    createdAt: string;
    updatedAt: string;
  };
}

export interface UpdateUserRequest {
  accessToken: string;
  firstName?: string;
  lastName?: string;
  email?: string;
  username?: string;
  currentPassword?: string;
  newPassword?: string;
}

export interface UpdateUserResponse {
  success: boolean;
  message: string;
  user?: {
    id: string;
    email: string;
    username: string;
    firstName: string;
    lastName: string;
    mfaEnabled: boolean;
    createdAt: string;
    updatedAt: string;
  };
}

export interface EnableMFARequest {
  accessToken: string;
}

export interface EnableMFAResponse {
  success: boolean;
  message: string;
  mfaChallenge: string;
  mfaCredentialId: string;
}

export interface DisableMFARequest {
  accessToken: string;
  mfaResponse: {
    credentialId: string;
    clientDataJSON: string;
    authenticatorData: string;
    signature: string;
  };
}

export interface DisableMFAResponse {
  success: boolean;
  message: string;
}

export interface RegisterPasskeyRequest {
  accessToken: string;
  challenge: string;
  user: {
    id: string;
    name: string;
    displayName: string;
  };
  excludeCredentials?: Array<{
    type: string;
    id: string;
  }>;
}

export interface RegisterPasskeyResponse {
  success: boolean;
  message: string;
  credentialId: string;
}

export interface DeletePasskeyRequest {
  accessToken: string;
  credentialId: string;
}

export interface DeletePasskeyResponse {
  success: boolean;
  message: string;
}

export interface ListPasskeysRequest {
  accessToken: string;
}

export interface ListPasskeysResponse {
  success: boolean;
  message: string;
  passkeys: Array<{
    id: string;
    type: string;
    createdAt: string;
    lastUsedAt?: string;
  }>;
} 