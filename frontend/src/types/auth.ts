/**
 * Frontend-specific Authentication and Session Types
 */
export interface MessageResponse {
  success: boolean;
  message: string;
}

export interface VerifyResetTokenResponse {
  valid: boolean;
  message: string;
}

export interface RegistrationStatusResponse {
  enabled: boolean;
  allowOrgCreation: boolean;
  hasDefaultOrganization: boolean;
}
