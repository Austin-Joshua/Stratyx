package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"stratyx/backend/internal/config"
	"stratyx/backend/internal/domain/models"
	"stratyx/backend/internal/platform/security"
	"stratyx/backend/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

func RegisterAuthRoutes(r *gin.RouterGroup, services *service.Services, cfg config.Config) {
	r.POST("/register", func(c *gin.Context) {
		var body struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user, err := services.Auth.Register(c.Request.Context(), body.Name, strings.ToLower(body.Email), body.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		verifyToken, _ := services.Auth.CreateEmailVerificationToken(c.Request.Context(), user.ID, user.Email)
		if verifyToken != "" {
			verifyLink := fmt.Sprintf("%s/settings?verifyToken=%s", strings.TrimRight(cfg.FrontendURL, "/"), verifyToken)
			_ = services.Auth.QueueEmail(c.Request.Context(), user.Email, "Verify your STRATYX email", "Open this link to verify your email:\n"+verifyLink)
		}
		c.JSON(http.StatusCreated, gin.H{"message": "registered", "user": user, "verificationToken": verifyToken})
	})

	r.POST("/login", func(c *gin.Context) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Remember bool   `json:"remember"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user, access, refresh, err := services.Auth.Login(
			c.Request.Context(),
			strings.ToLower(body.Email),
			body.Password,
			c.Request.UserAgent(),
			c.ClientIP(),
		)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if user.TwoFAEnabled {
			challengeToken, err := services.Auth.CreateLoginChallenge(c.Request.Context(), user.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize 2fa"})
				return
			}
			c.JSON(http.StatusAccepted, gin.H{
				"requires2FA":    true,
				"challengeToken": challengeToken,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"accessToken": access, "refreshToken": refresh, "user": user})
	})

	r.POST("/refresh", func(c *gin.Context) {
		var body struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		access, refresh, err := services.Auth.Refresh(c.Request.Context(), body.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"accessToken": access, "refreshToken": refresh})
	})

	r.POST("/logout", func(c *gin.Context) {
		var body struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := services.Auth.Logout(c.Request.Context(), body.RefreshToken); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
	})

	r.POST("/forgot-password", func(c *gin.Context) {
		var body struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		token, err := services.Auth.CreatePasswordResetToken(c.Request.Context(), body.Email)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"message": "If the account exists, reset instructions were generated."})
			return
		}
		resetLink := fmt.Sprintf("%s/forgot-password?token=%s", strings.TrimRight(cfg.FrontendURL, "/"), token)
		_ = services.Auth.QueueEmail(c.Request.Context(), body.Email, "Reset your STRATYX password", "Open this link to reset your password:\n"+resetLink)
		c.JSON(http.StatusOK, gin.H{"message": "password reset token generated", "resetToken": token})
	})

	r.POST("/reset-password", func(c *gin.Context) {
		var body struct {
			Token       string `json:"token"`
			NewPassword string `json:"newPassword"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := services.Auth.ResetPassword(c.Request.Context(), body.Token, body.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "password reset successful"})
	})

	r.POST("/verify-email", func(c *gin.Context) {
		var body struct {
			Token string `json:"token"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := services.Auth.VerifyEmail(c.Request.Context(), body.Token); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "email verified"})
	})

	r.GET("/oauth/google", func(c *gin.Context) {
		if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "google oauth not configured"})
			return
		}
		o := oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     google.Endpoint,
		}
		state, err := services.Auth.CreateOAuthState(c.Request.Context(), "google")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize oauth state"})
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, o.AuthCodeURL(state, oauth2.AccessTypeOnline))
	})

	r.GET("/oauth/github", func(c *gin.Context) {
		if cfg.GithubClientID == "" || cfg.GithubClientSecret == "" {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "github oauth not configured"})
			return
		}
		o := oauth2.Config{
			ClientID:     cfg.GithubClientID,
			ClientSecret: cfg.GithubClientSecret,
			RedirectURL:  cfg.GithubRedirectURL,
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     github.Endpoint,
		}
		state, err := services.Auth.CreateOAuthState(c.Request.Context(), "github")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize oauth state"})
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, o.AuthCodeURL(state, oauth2.AccessTypeOnline))
	})

	r.GET("/oauth/google/callback", func(c *gin.Context) {
		handleGoogleCallback(c, services, cfg)
	})
	r.GET("/oauth/github/callback", func(c *gin.Context) {
		handleGithubCallback(c, services, cfg)
	})

	r.POST("/2fa/setup", AuthRequired(cfg.JWTSecret), func(c *gin.Context) {
		secret, otpURL, err := services.Auth.Setup2FA(c.Request.Context(), c.GetString("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"secret": secret, "otpAuthURL": otpURL})
	})
	r.POST("/2fa/verify-setup", AuthRequired(cfg.JWTSecret), func(c *gin.Context) {
		var body struct {
			Code string `json:"code"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := services.Auth.Verify2FASetup(c.Request.Context(), c.GetString("userId"), body.Code); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "2fa enabled"})
	})
	r.POST("/2fa/disable", AuthRequired(cfg.JWTSecret), func(c *gin.Context) {
		var body struct {
			Code string `json:"code"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := services.Auth.Disable2FA(c.Request.Context(), c.GetString("userId"), body.Code); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "2fa disabled"})
	})
	r.POST("/2fa/complete-login", func(c *gin.Context) {
		var body struct {
			ChallengeToken string `json:"challengeToken"`
			Code           string `json:"code"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user, access, refresh, err := services.Auth.Complete2FALogin(c.Request.Context(), body.ChallengeToken, body.Code, c.Request.UserAgent(), c.ClientIP())
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": user, "accessToken": access, "refreshToken": refresh})
	})
}

func handleGoogleCallback(c *gin.Context, services *service.Services, cfg config.Config) {
	if err := services.Auth.ConsumeOAuthState(c.Request.Context(), "google", c.Query("state")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	o := oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}
	token, err := o.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "oauth exchange failed"})
		return
	}
	client := o.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to fetch user profile"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var profile struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider response"})
		return
	}
	user, access, refresh, err := services.Auth.OAuthLogin(c.Request.Context(), "google", profile.ID, profile.Email, profile.Name, profile.Picture, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	redirectURL := service.BuildOAuthRedirectURL(cfg.FrontendURL, access, refresh)
	_ = user
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func handleGithubCallback(c *gin.Context, services *service.Services, cfg config.Config) {
	if err := services.Auth.ConsumeOAuthState(c.Request.Context(), "github", c.Query("state")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	o := oauth2.Config{
		ClientID:     cfg.GithubClientID,
		ClientSecret: cfg.GithubClientSecret,
		RedirectURL:  cfg.GithubRedirectURL,
		Scopes:       []string{"read:user", "user:email"},
		Endpoint:     github.Endpoint,
	}
	token, err := o.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "oauth exchange failed"})
		return
	}
	client := o.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to fetch user profile"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var profile struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
		Email     string `json:"email"`
		Name      string `json:"name"`
	}
	if err := json.Unmarshal(body, &profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider response"})
		return
	}
	email := profile.Email
	if email == "" {
		resp2, _ := client.Get("https://api.github.com/user/emails")
		if resp2 != nil {
			defer resp2.Body.Close()
			raw, _ := io.ReadAll(resp2.Body)
			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			_ = json.Unmarshal(raw, &emails)
			for _, e := range emails {
				if e.Primary {
					email = e.Email
					break
				}
			}
			if email == "" && len(emails) > 0 {
				email = emails[0].Email
			}
		}
	}
	displayName := profile.Name
	if displayName == "" {
		displayName = profile.Login
	}
	user, access, refresh, err := services.Auth.OAuthLogin(c.Request.Context(), "github", fmt.Sprintf("%d", profile.ID), email, displayName, profile.AvatarURL, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	redirectURL := service.BuildOAuthRedirectURL(cfg.FrontendURL, access, refresh)
	_ = user
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func AuthRequired(secret string) gin.HandlerFunc {
	jwtManager := security.NewJWTManager(secret, 30, 14)
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		parts := strings.Split(raw, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		claims, err := jwtManager.Parse(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("userId", claims.UserID)
		c.Next()
	}
}

func RequireRoles(services *service.Services, roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := services.Auth.Me(c.Request.Context(), c.GetString("userId"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		for _, role := range roles {
			if user.Role == role || user.Role == models.RoleSuperAdmin {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}
