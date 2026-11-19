package users

// CreateUserDTO represents payload for admin creating a user.
type CreateUserDTO struct {
	Email        string  `json:"email" validate:"required,email"`
	Password     string  `json:"password" validate:"required,min=6"`
	FirstName    string  `json:"firstName" validate:"required"`
	LastName     string  `json:"lastName" validate:"required"`
	Username     *string `json:"username,omitempty"`
	MobileNumber *string `json:"mobileNumber,omitempty"`
	Avatar       *string `json:"avatar,omitempty"`
	Bio          *string `json:"bio,omitempty"`
	Role         *string `json:"role,omitempty"`
	Status       *string `json:"status,omitempty"`
	Timezone     *string `json:"timezone,omitempty"`
	Language     *string `json:"language,omitempty"`
}

// UpdateUserDTO represents fields that can be updated on a user profile.
type UpdateUserDTO struct {
	FirstName     *string `json:"firstName,omitempty"`
	LastName      *string `json:"lastName,omitempty"`
	Username      *string `json:"username,omitempty"`
	Avatar        *string `json:"avatar,omitempty"`
	Bio           *string `json:"bio,omitempty"`
	Timezone      *string `json:"timezone,omitempty"`
	Language      *string `json:"language,omitempty"`
	MobileNumber  *string `json:"mobileNumber,omitempty"`
	Role          *string `json:"role,omitempty"`
	Status        *string `json:"status,omitempty"`
	EmailVerified *bool   `json:"emailVerified,omitempty"`
}

// ChangePasswordDTO represents credentials for user changing their own password.
type ChangePasswordDTO struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" validate:"required"`
}

// UpdateStatusDTO represents payload for updating user account status.
type UpdateStatusDTO struct {
	Status string `json:"status" validate:"required,oneof=ACTIVE INACTIVE SUSPENDED PENDING"`
}

// PublicProfileDTO represents safe profile fields visible to shared organization members.
type PublicProfileDTO struct {
	ID           string  `json:"id"`
	FirstName    string  `json:"firstName"`
	LastName     string  `json:"lastName"`
	Email        string  `json:"email"`
	Avatar       *string `json:"avatar"`
	MobileNumber *string `json:"mobileNumber"`
	Bio          *string `json:"bio"`
	Timezone     string  `json:"timezone"`
	Role         string  `json:"role"`
}

// ResponseDTO represents public user details returned by users endpoints (omitting passwords).
type ResponseDTO struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	FirstName    string  `json:"firstName"`
	LastName     string  `json:"lastName"`
	Username     *string `json:"username,omitempty"`
	Role         string  `json:"role"`
	Status       string  `json:"status"`
	Avatar       *string `json:"avatar,omitempty"`
	Bio          *string `json:"bio,omitempty"`
	Timezone     string  `json:"timezone"`
	Language     string  `json:"language"`
	MobileNumber *string `json:"mobileNumber,omitempty"`
	LastLoginAt  *string `json:"lastLoginAt,omitempty"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

// OnlineStatusDTO represents real-time activity for a user.
type OnlineStatusDTO struct {
	IsOnline bool    `json:"isOnline"`
	LastSeen *string `json:"lastSeen,omitempty"`
}
