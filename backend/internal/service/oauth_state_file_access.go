package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"stratyx/backend/internal/domain/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *AuthService) CreateOAuthState(ctx context.Context, provider string) (string, error) {
	state := RandomOAuthState()
	_, err := s.repos.OAuthStates.InsertOne(ctx, models.OAuthState{
		ID:        uuid.NewString(),
		Provider:  provider,
		State:     state,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
	})
	return state, err
}

func (s *AuthService) ConsumeOAuthState(ctx context.Context, provider, state string) error {
	var row models.OAuthState
	if err := s.repos.OAuthStates.FindOne(ctx, bson.M{"provider": provider, "state": state}).Decode(&row); err != nil {
		return errors.New("invalid oauth state")
	}
	if !row.UsedAt.IsZero() || row.ExpiresAt.Before(time.Now()) {
		return errors.New("oauth state expired")
	}
	_, err := s.repos.OAuthStates.UpdateOne(ctx, bson.M{"_id": row.ID}, bson.M{"$set": bson.M{"used_at": time.Now()}})
	return err
}

func (s *ItemService) CreateFileAccessURL(ctx context.Context, userID, fileID string) (string, error) {
	var file models.FileAsset
	if err := s.repos.Files.FindOne(ctx, bson.M{"_id": fileID, "owner_id": userID}).Decode(&file); err != nil {
		return "", errors.New("file not found")
	}
	token := randomToken(24)
	_, err := s.repos.FileAccessTokens.InsertOne(ctx, models.FileAccessToken{
		ID:        uuid.NewString(),
		FileID:    fileID,
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
	})
	if err != nil {
		return "", err
	}
	return "/api/files/download/" + fileID + "?token=" + token, nil
}

func (s *ItemService) ResolveFileForDownload(ctx context.Context, fileID, token string) (*models.FileAsset, string, error) {
	var access models.FileAccessToken
	if err := s.repos.FileAccessTokens.FindOne(ctx, bson.M{"file_id": fileID, "token": token}).Decode(&access); err != nil {
		return nil, "", errors.New("invalid download token")
	}
	if !access.UsedAt.IsZero() || access.ExpiresAt.Before(time.Now()) {
		return nil, "", errors.New("download token expired")
	}
	var file models.FileAsset
	if err := s.repos.Files.FindOne(ctx, bson.M{"_id": fileID}).Decode(&file); err != nil {
		return nil, "", errors.New("file not found")
	}
	_, _ = s.repos.FileAccessTokens.UpdateOne(ctx, bson.M{"_id": access.ID}, bson.M{"$set": bson.M{"used_at": time.Now()}})
	if file.StorageKind == "local" {
		base := os.Getenv("UPLOAD_LOCAL_PATH")
		if base == "" {
			base = "./uploads"
		}
		relative := strings.TrimPrefix(file.URL, "/uploads/")
		if !strings.Contains(file.URL, "/uploads/") {
			relative = file.URL
		}
		return &file, filepath.Join(base, filepath.FromSlash(relative)), nil
	}
	return &file, file.URL, nil
}
