package settings

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
)

// mockRepo is an in-memory implementation of Repository for testing.
type mockRepo struct {
	userSettings   map[string]sqlc.Setting
	globalSettings map[string]sqlc.Setting
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		userSettings:   make(map[string]sqlc.Setting),
		globalSettings: make(map[string]sqlc.Setting),
	}
}

func userKey(userID uuid.UUID, key string) string {
	return userID.String() + ":" + key
}

func (m *mockRepo) GetByUserAndKey(_ context.Context, userID uuid.UUID, key string) (*sqlc.Setting, error) {
	if s, ok := m.userSettings[userKey(userID, key)]; ok {
		return &s, nil
	}
	return nil, nil
}

func (m *mockRepo) GetGlobalByKey(_ context.Context, key string) (*sqlc.Setting, error) {
	if s, ok := m.globalSettings[key]; ok {
		return &s, nil
	}
	return nil, nil
}

func (m *mockRepo) UpsertUserSetting(_ context.Context, params sqlc.UpsertUserSettingParams) (*sqlc.Setting, error) {
	uID, _ := uuid.FromBytes(params.UserId.Bytes[:])
	s := sqlc.Setting{
		ID:          params.ID,
		Key:         params.Key,
		Value:       params.Value,
		Description: params.Description,
		Category:    params.Category,
		IsEncrypted: params.IsEncrypted,
		UserId:      params.UserId,
	}
	m.userSettings[userKey(uID, params.Key)] = s
	return &s, nil
}

func (m *mockRepo) CreateGlobalSetting(_ context.Context, params sqlc.CreateGlobalSettingParams) (*sqlc.Setting, error) {
	s := sqlc.Setting{
		ID:          params.ID,
		Key:         params.Key,
		Value:       params.Value,
		Description: params.Description,
		Category:    params.Category,
		IsEncrypted: params.IsEncrypted,
	}
	m.globalSettings[params.Key] = s
	return &s, nil
}

func (m *mockRepo) UpdateGlobalSetting(_ context.Context, params sqlc.UpdateGlobalSettingByIDParams) (*sqlc.Setting, error) {
	for k, s := range m.globalSettings {
		if s.ID == params.ID {
			s.Value = params.Value
			s.Description = params.Description
			s.Category = params.Category
			s.IsEncrypted = params.IsEncrypted
			m.globalSettings[k] = s
			return &s, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) ListUserSettings(_ context.Context, userID uuid.UUID, category string) ([]sqlc.Setting, error) {
	var list []sqlc.Setting
	for _, s := range m.userSettings {
		uID, _ := uuid.FromBytes(s.UserId.Bytes[:])
		if uID == userID {
			if category == "" || s.Category == category {
				list = append(list, s)
			}
		}
	}
	return list, nil
}

func (m *mockRepo) ListGlobalSettings(_ context.Context, category string) ([]sqlc.Setting, error) {
	var list []sqlc.Setting
	for _, s := range m.globalSettings {
		if category == "" || s.Category == category {
			list = append(list, s)
		}
	}
	return list, nil
}

func (m *mockRepo) DeleteUserSetting(_ context.Context, userID uuid.UUID, key string) error {
	delete(m.userSettings, userKey(userID, key))
	return nil
}

func (m *mockRepo) DeleteGlobalSetting(_ context.Context, key string) error {
	delete(m.globalSettings, key)
	return nil
}

func (m *mockRepo) CountUserSetting(_ context.Context, userID uuid.UUID, key string) (int64, error) {
	if _, ok := m.userSettings[userKey(userID, key)]; ok {
		return 1, nil
	}
	return 0, nil
}

func (m *mockRepo) CountGlobalSetting(_ context.Context, key string) (int64, error) {
	if _, ok := m.globalSettings[key]; ok {
		return 1, nil
	}
	return 0, nil
}

func (m *mockRepo) WithTx(_ context.Context, fn func(repo Repository) error) error {
	return fn(m)
}

func TestSettingsService(t *testing.T) {
	fc, err := crypto.NewFieldCrypto("test-secret-key-for-orbito-settings")
	if err != nil {
		t.Fatalf("Failed to init FieldCrypto: %v", err)
	}

	repo := newMockRepo()
	svc := NewService(repo, fc)
	ctx := context.Background()

	// 1. Set global setting with encryption
	isEnc := true
	desc := "Dark or light theme"
	cat := "preferences"
	err = svc.Set(ctx, SetSettingDTO{
		Key:         "theme",
		Value:       "dark",
		Description: &desc,
		Category:    &cat,
		IsEncrypted: &isEnc,
	}, nil)
	if err != nil {
		t.Fatalf("Failed to set global setting: %v", err)
	}

	// Verify it's stored encrypted in mockRepo
	rawGlobal, ok := repo.globalSettings["theme"]
	if !ok {
		t.Fatal("Global setting 'theme' not found in repo")
	}
	if rawGlobal.Value.String == "dark" {
		t.Fatal("Expected encrypted value in raw database, but found plaintext")
	}

	// 2. Retrieve global setting (should be automatically decrypted)
	val, err := svc.Get(ctx, "theme", nil, nil)
	if err != nil {
		t.Fatalf("Failed to get setting: %v", err)
	}
	if val == nil || *val != "dark" {
		t.Fatalf("Expected decrypted 'dark', got %v", val)
	}

	// 3. User overrides global setting
	userID := uuid.New()
	err = svc.Set(ctx, SetSettingDTO{
		Key:      "theme",
		Value:    "light",
		Category: &cat,
	}, &userID)
	if err != nil {
		t.Fatalf("Failed to set user setting: %v", err)
	}

	// 4. User queries setting: user value should take precedence
	userVal, err := svc.Get(ctx, "theme", &userID, nil)
	if err != nil {
		t.Fatalf("Failed to get user setting: %v", err)
	}
	if userVal == nil || *userVal != "light" {
		t.Fatalf("Expected 'light', got %v", userVal)
	}

	// 5. Another user or global queries: should still get global setting
	otherUser := uuid.New()
	otherVal, err := svc.Get(ctx, "theme", &otherUser, nil)
	if err != nil {
		t.Fatalf("Failed to get other user setting: %v", err)
	}
	if otherVal == nil || *otherVal != "dark" {
		t.Fatalf("Expected 'dark', got %v", otherVal)
	}

	// 6. Test GetAll merging
	all, err := svc.GetAll(ctx, &userID, nil)
	if err != nil {
		t.Fatalf("Failed to get all settings: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("Expected 1 deduplicated setting, got %d", len(all))
	}
	if all[0].Value == nil || *all[0].Value != "light" || !all[0].IsUserSpecific {
		t.Fatalf("Expected user-specific 'light', got %+v", all[0])
	}

	// 7. Test RegistrationStatus
	regStatus, err := svc.GetRegistrationStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get registration status: %v", err)
	}
	if !regStatus.Enabled || !regStatus.AllowOrganizations {
		t.Fatalf("Expected default true, got %+v", regStatus)
	}

	// Disable registration via global setting
	regFalse := "false"
	_ = svc.Set(ctx, SetSettingDTO{
		Key:   "registration_enabled",
		Value: regFalse,
	}, nil)

	regStatus, err = svc.GetRegistrationStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get registration status after update: %v", err)
	}
	if regStatus.Enabled {
		t.Fatal("Expected registration to be disabled")
	}
}

func TestSettingsHTTPHandler(t *testing.T) {
	fc, _ := crypto.NewFieldCrypto("test-secret-key-for-orbito-settings")
	repo := newMockRepo()
	svc := NewService(repo, fc)
	h := NewHandler(svc)
	router := Routes(h, nil)

	userID := uuid.New()

	// POST / (Set setting)
	body, _ := json.Marshal(SetSettingDTO{
		Key:   "dashboard_layout",
		Value: "grid",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxhelper.WithUserID(req.Context(), userID.String()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /{key}
	req = httptest.NewRequest(http.MethodGet, "/dashboard_layout", nil)
	req = req.WithContext(ctxhelper.WithUserID(req.Context(), userID.String()))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
	var res KeyValResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if res.Value == nil || *res.Value != "grid" {
		t.Fatalf("Expected 'grid', got %v", res.Value)
	}

	// GET /registration-status
	req = httptest.NewRequest(http.MethodGet, "/registration-status", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
	var regRes RegistrationStatusResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &regRes)
	if !regRes.Enabled {
		t.Fatalf("Expected enabled true, got false")
	}
}
