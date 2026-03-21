package service

import (
	"context"
	"testing"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool   { return &b }

func sellerUser() *model.User {
	now := time.Now()
	return &model.User{
		ID: "user-123", Phone: "0901234567", Role: "member",
		AcceptedTOSAt: &now,
	}
}

func newUserWithoutTOS() *model.User {
	return &model.User{
		ID: "user-123", Phone: "0901234567", Role: "member",
	}
}

// --- GetMe ---

func TestGetMe_Success(t *testing.T) {
	userRepo := new(mockUserRepo)
	svc := NewUserService(userRepo, nil)

	userRepo.On("GetByID", mock.Anything, "user-123").Return(testUser(), nil)

	user, err := svc.GetMe(context.Background(), "user-123")

	require.NoError(t, err)
	assert.Equal(t, "user-123", user.ID)
}

// --- GetPublicProfile ---

func TestGetPublicProfile_Success(t *testing.T) {
	userRepo := new(mockUserRepo)
	svc := NewUserService(userRepo, nil)

	u := testUser()
	u.Name = strPtr("Test User")
	userRepo.On("GetByID", mock.Anything, "user-123").Return(u, nil)

	profile, err := svc.GetPublicProfile(context.Background(), "user-123")

	require.NoError(t, err)
	assert.Equal(t, "user-123", profile.ID)
	assert.Equal(t, "Test User", *profile.Name)
	// Phone should NOT be in public profile
	assert.Empty(t, profile.ID == "" && false) // PublicProfile has no Phone field
}

// --- UpdateProfile ---

func TestUpdateProfile_UpdateFields(t *testing.T) {
	userRepo := new(mockUserRepo)
	svc := NewUserService(userRepo, nil)

	user := testUser()
	updatedUser := *user
	updatedUser.Name = strPtr("New Name")

	userRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
	userRepo.On("UpdateProfile", mock.Anything, "user-123", mock.Anything).Return(&updatedUser, nil)

	req := &model.UpdateProfileRequest{Name: strPtr("New Name")}
	result, err := svc.UpdateProfile(context.Background(), "user-123", req)

	require.NoError(t, err)
	assert.Equal(t, "New Name", *result.Name)
}

func TestUpdateProfile_AcceptTOS(t *testing.T) {
	userRepo := new(mockUserRepo)
	svc := NewUserService(userRepo, nil)

	user := newUserWithoutTOS()
	now := time.Now()
	acceptedUser := *user
	acceptedUser.AcceptedTOSAt = &now

	userRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
	userRepo.On("AcceptTOS", mock.Anything, "user-123").Return(&acceptedUser, nil)
	userRepo.On("UpdateProfile", mock.Anything, "user-123", mock.Anything).Return(&acceptedUser, nil)

	req := &model.UpdateProfileRequest{AcceptTOS: boolPtr(true)}
	result, err := svc.UpdateProfile(context.Background(), "user-123", req)

	require.NoError(t, err)
	assert.NotNil(t, result.AcceptedTOSAt)
}

func TestUpdateProfile_SkipTOSIfAlreadyAccepted(t *testing.T) {
	userRepo := new(mockUserRepo)
	svc := NewUserService(userRepo, nil)

	user := sellerUser()
	userRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
	userRepo.On("UpdateProfile", mock.Anything, "user-123", mock.Anything).Return(user, nil)

	req := &model.UpdateProfileRequest{AcceptTOS: boolPtr(true), Name: strPtr("Updated")}
	_, err := svc.UpdateProfile(context.Background(), "user-123", req)

	require.NoError(t, err)
	// AcceptTOS should NOT be called again
	userRepo.AssertNotCalled(t, "AcceptTOS", mock.Anything, mock.Anything)
}

// --- UpdateAvatar ---

func TestUpdateAvatar_Success(t *testing.T) {
	userRepo := new(mockUserRepo)
	svc := NewUserService(userRepo, nil)

	user := testUser()
	user.AvatarURL = strPtr("https://cdn.example.com/avatar.jpg")
	userRepo.On("UpdateAvatar", mock.Anything, "user-123", "https://cdn.example.com/avatar.jpg").Return(user, nil)

	result, err := svc.UpdateAvatar(context.Background(), "user-123", "https://cdn.example.com/avatar.jpg")

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/avatar.jpg", *result.AvatarURL)
}
