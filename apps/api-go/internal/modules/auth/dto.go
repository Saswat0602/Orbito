package auth

// LoginDTO represents the credentials for user login.
type LoginDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterDTO represents the payload for user registration.
type RegisterDTO struct {
	Email           string  `json:"email" validate:"required,email"`
	Password        string  `json:"password" validate:"required,min=6"`
	FirstName       string  `json:"firstName" validate:"required"`
	LastName        string  `json:"lastName" validate:"required"`
	Username        *string `json:"username,omitempty"`
	InvitationToken *string `json:"invitationToken,omitempty"`
}

// UserDTO represents public user details in auth responses.
type UserDTO struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	FirstName    string  `json:"firstName"`
	LastName     string  `json:"lastName"`
	Username     *string `json:"username,omitempty"`
	Role         string  `json:"role"`
	Avatar       *string `json:"avatar,omitempty"`
	Bio          *string `json:"bio,omitempty"`
	MobileNumber *string `json:"mobileNumber,omitempty"`
	Language     *string `json:"language,omitempty"`
	Timezone     *string `json:"timezone,omitempty"`
}

// ResponseDTO represents the standard auth payload containing JWTs and user profile.
type ResponseDTO struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	User         UserDTO `json:"user"`
}

// RefreshTokenDTO represents the payload for explicit token refresh.
type RefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

// ForgotPasswordDTO represents the payload for initiating a password reset.
type ForgotPasswordDTO struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordDTO represents the payload for executing a password reset.
type ResetPasswordDTO struct {
	Token           string `json:"token" validate:"required"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,min=8"`
}

// VerifyResetTokenResponseDTO represents the token verification result.
type VerifyResetTokenResponseDTO struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// SetupRequiredResponseDTO represents the system setup status.
type SetupRequiredResponseDTO struct {
	Required bool   `json:"required"`
	CanSetup bool   `json:"canSetup"`
	Message  string `json:"message"`
}

// SetupAdminDTO represents the initial super admin bootstrap payload.
type SetupAdminDTO struct {
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required,min=8"`
	FirstName string  `json:"firstName" validate:"required"`
	LastName  string  `json:"lastName" validate:"required"`
	Username  *string `json:"username,omitempty"`
}

// OIDCConfigResponseDTO represents SSO configuration status.
type OIDCConfigResponseDTO struct {
	Configured   bool   `json:"configured"`
	ProviderName string `json:"providerName,omitempty"`
}
