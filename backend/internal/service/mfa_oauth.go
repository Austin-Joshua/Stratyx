package service

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"net/url"
	"os"
	"strings"
	"time"

	"stratyx/backend/internal/domain/models"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *AuthService) Setup2FA(ctx context.Context, userID string) (string, string, error) {
	var user models.User
	if err := s.repos.Users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		return "", "", errors.New("user not found")
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "STRATYX",
		AccountName: user.Email,
	})
	if err != nil {
		return "", "", err
	}
	_, err = s.repos.Users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{
		"$set": bson.M{"two_fa_secret": key.Secret(), "updated_at": time.Now()},
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

func (s *AuthService) Verify2FASetup(ctx context.Context, userID, code string) error {
	var user models.User
	if err := s.repos.Users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		return errors.New("user not found")
	}
	if user.TwoFASecret == "" {
		return errors.New("2fa not initialized")
	}
	if !totp.Validate(strings.TrimSpace(code), user.TwoFASecret) {
		return errors.New("invalid otp code")
	}
	_, err := s.repos.Users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{
		"$set": bson.M{"two_fa_enabled": true, "updated_at": time.Now()},
	})
	return err
}

func (s *AuthService) Disable2FA(ctx context.Context, userID, code string) error {
	var user models.User
	if err := s.repos.Users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		return errors.New("user not found")
	}
	if user.TwoFAEnabled && !totp.Validate(strings.TrimSpace(code), user.TwoFASecret) {
		return errors.New("invalid otp code")
	}
	_, err := s.repos.Users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{
		"$set": bson.M{"two_fa_enabled": false, "two_fa_secret": "", "updated_at": time.Now()},
	})
	return err
}

func (s *AuthService) CreateLoginChallenge(ctx context.Context, userID string) (string, error) {
	token := randomToken(20)
	_, err := s.repos.AuthChallenges.InsertOne(ctx, models.AuthChallenge{
		ID:        uuid.NewString(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	})
	return token, err
}

func (s *AuthService) Complete2FALogin(ctx context.Context, challengeToken, otp, device, ip string) (*models.User, string, string, error) {
	var challenge models.AuthChallenge
	if err := s.repos.AuthChallenges.FindOne(ctx, bson.M{"token": challengeToken}).Decode(&challenge); err != nil {
		return nil, "", "", errors.New("invalid challenge")
	}
	if challenge.ExpiresAt.Before(time.Now()) {
		return nil, "", "", errors.New("challenge expired")
	}
	var user models.User
	if err := s.repos.Users.FindOne(ctx, bson.M{"_id": challenge.UserID}).Decode(&user); err != nil {
		return nil, "", "", errors.New("user not found")
	}
	if !totp.Validate(strings.TrimSpace(otp), user.TwoFASecret) {
		return nil, "", "", errors.New("invalid otp code")
	}
	_, _ = s.repos.AuthChallenges.DeleteOne(ctx, bson.M{"_id": challenge.ID})
	return s.Login(ctx, user.Email, "__oauth_or_verified_2fa__", device, ip)
}

func (s *AuthService) OAuthLogin(ctx context.Context, provider, providerUserID, email, name, avatarURL, device, ip string) (*models.User, string, string, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	email = strings.ToLower(strings.TrimSpace(email))
	var oauth models.OAuthAccount
	if err := s.repos.OAuthAccounts.FindOne(ctx, bson.M{"provider": provider, "provider_user_id": providerUserID}).Decode(&oauth); err == nil {
		var existing models.User
		if err := s.repos.Users.FindOne(ctx, bson.M{"_id": oauth.UserID}).Decode(&existing); err == nil {
			return s.loginByUser(ctx, &existing, device, ip)
		}
	}

	var user models.User
	if err := s.repos.Users.FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		user = models.User{
			ID:              uuid.NewString(),
			Name:            name,
			Email:           email,
			PasswordHash:    "",
			Role:            models.RoleUser,
			IsActive:        true,
			IsEmailVerified: true,
			AvatarURL:       avatarURL,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			NotificationPrefs: models.NotificationPreferences{
				EmailMentions: true, InAppMentions: true, ProductNews: true,
			},
			PrivacySettings: models.PrivacySettings{ShowEmail: false, ShowProfile: true},
		}
		if _, err := s.repos.Users.InsertOne(ctx, user); err != nil {
			return nil, "", "", err
		}
	}
	_, _ = s.repos.OAuthAccounts.InsertOne(ctx, models.OAuthAccount{
		ID:             uuid.NewString(),
		UserID:         user.ID,
		Provider:       provider,
		ProviderUserID: providerUserID,
		Email:          email,
		CreatedAt:      time.Now(),
	})
	return s.loginByUser(ctx, &user, device, ip)
}

func (s *AuthService) ConnectedAccounts(ctx context.Context, userID string) ([]models.OAuthAccount, error) {
	cur, err := s.repos.OAuthAccounts.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]models.OAuthAccount, 0)
	for cur.Next(ctx) {
		var account models.OAuthAccount
		if err := cur.Decode(&account); err == nil {
			out = append(out, account)
		}
	}
	return out, nil
}

func (s *AuthService) DisconnectAccount(ctx context.Context, userID, provider string) error {
	_, err := s.repos.OAuthAccounts.DeleteOne(ctx, bson.M{"user_id": userID, "provider": strings.ToLower(strings.TrimSpace(provider))})
	return err
}

func (s *AuthService) loginByUser(ctx context.Context, user *models.User, device, ip string) (*models.User, string, string, error) {
	access, err := s.jwt.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}
	refresh, exp, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}
	_, err = s.repos.Sessions.InsertOne(ctx, models.Session{
		ID:           uuid.NewString(),
		UserID:       user.ID,
		RefreshToken: refresh,
		ExpiresAt:    exp,
		DeviceName:   device,
		IPAddress:    ip,
		LastSeenAt:   time.Now(),
		CreatedAt:    time.Now(),
	})
	if err != nil {
		return nil, "", "", err
	}
	return user, access, refresh, nil
}

func BuildOAuthRedirectURL(frontend, accessToken, refreshToken string) string {
	u, _ := url.Parse(strings.TrimRight(frontend, "/") + "/auth/callback")
	q := u.Query()
	q.Set("accessToken", accessToken)
	q.Set("refreshToken", refreshToken)
	u.RawQuery = q.Encode()
	return u.String()
}

func RandomOAuthState() string {
	return randomToken(12)
}

func (s *AuthService) SendEmail(to, subject, body string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USERNAME")
	pass := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM_EMAIL")
	if host == "" || port == "" || user == "" || pass == "" || from == "" {
		return errors.New("smtp is not configured")
	}
	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		body
	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", user, pass, host)
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}
