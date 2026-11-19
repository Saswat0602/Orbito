package settings

// SetSettingDTO represents the payload for creating or updating a setting.
type SetSettingDTO struct {
	Key         string  `json:"key" validate:"required"`
	Value       string  `json:"value"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"`
	IsEncrypted *bool   `json:"isEncrypted,omitempty"`
}

// BulkSetSettingsDTO represents the payload for setting multiple settings at once.
type BulkSetSettingsDTO struct {
	Settings []SetSettingDTO `json:"settings" validate:"required,dive"`
}

// SettingResponse represents a setting returned to clients.
type SettingResponse struct {
	Key            string  `json:"key"`
	Value          *string `json:"value"`
	Description    *string `json:"description"`
	Category       string  `json:"category"`
	IsUserSpecific bool    `json:"isUserSpecific"`
	IsEncrypted    bool    `json:"isEncrypted"`
}

// KeyValResponse represents a simple key-value setting pair.
type KeyValResponse struct {
	Key   string  `json:"key"`
	Value *string `json:"value"`
}

// RegistrationStatusResponse matches the shape returned by the registration-status endpoint.
type RegistrationStatusResponse struct {
	Enabled            bool    `json:"enabled"`
	AllowOrganizations bool    `json:"allowOrganizations"`
	DefaultOrgID       *string `json:"defaultOrgId"`
}
