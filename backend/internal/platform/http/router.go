package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"stratyx/backend/internal/config"
	"stratyx/backend/internal/domain/models"
	"stratyx/backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, services *service.Services) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestID())
	r.Use(StructuredLogger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.AllowedOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Organization-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "stratyx-api"})
	})
	r.Static("/uploads", cfg.UploadLocalPath)

	api := r.Group("/api/v1")
	authV1 := api.Group("/auth")
	authV1.Use(AuthRateLimit())
	RegisterAuthRoutes(authV1, services, cfg)

	protected := api.Group("/")
	protected.Use(AuthRequired(cfg.JWTSecret))
	RegisterProtectedRoutes(protected, services)

	apiLegacy := r.Group("/api")
	authLegacy := apiLegacy.Group("/auth")
	authLegacy.Use(AuthRateLimit())
	RegisterAuthRoutes(authLegacy, services, cfg)
	legacyProtected := apiLegacy.Group("/")
	legacyProtected.Use(AuthRequired(cfg.JWTSecret))
	RegisterProtectedRoutes(legacyProtected, services)

	return r
}

func RegisterProtectedRoutes(r *gin.RouterGroup, services *service.Services) {
	r.GET("/auth/me", meHandler(services))
	r.PUT("/auth/profile", func(c *gin.Context) {
		var body struct {
			Name              string                         `json:"name"`
			AvatarURL         string                         `json:"avatarUrl"`
			PrivacySettings   models.PrivacySettings         `json:"privacySettings"`
			NotificationPrefs models.NotificationPreferences `json:"notificationPrefs"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		user, err := services.Auth.UpdateProfile(c.Request.Context(), c.GetString("userId"), body.Name, body.AvatarURL, body.PrivacySettings, body.NotificationPrefs)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	})
	r.POST("/auth/avatar", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
			return
		}
		url, err := services.Auth.UploadAvatar(c.Request.Context(), c.GetString("userId"), file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"avatarUrl": url})
	})
	r.POST("/auth/change-password", func(c *gin.Context) {
		var body struct {
			CurrentPassword string `json:"currentPassword"`
			NewPassword     string `json:"newPassword"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if err := services.Auth.ChangePassword(c.Request.Context(), c.GetString("userId"), body.CurrentPassword, body.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "password changed"})
	})
	r.GET("/auth/connected-accounts", func(c *gin.Context) {
		out, err := services.Auth.ConnectedAccounts(c.Request.Context(), c.GetString("userId"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load connected accounts"})
			return
		}
		c.JSON(http.StatusOK, out)
	})
	r.DELETE("/auth/connected-accounts/:provider", func(c *gin.Context) {
		if err := services.Auth.DisconnectAccount(c.Request.Context(), c.GetString("userId"), strings.TrimSpace(c.Param("provider"))); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to disconnect account"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "account disconnected"})
	})
	r.GET("/auth/sessions", func(c *gin.Context) {
		sessions, err := services.Auth.ListSessions(c.Request.Context(), c.GetString("userId"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load sessions"})
			return
		}
		c.JSON(http.StatusOK, sessions)
	})
	r.POST("/auth/logout-all", func(c *gin.Context) {
		if err := services.Auth.LogoutAll(c.Request.Context(), c.GetString("userId")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout all sessions"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "logged out from all devices"})
	})
	r.GET("/dashboard/summary", func(c *gin.Context) {
		out, err := services.Items.DashboardSummary(c.Request.Context(), c.GetString("userId"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load summary"})
			return
		}
		c.JSON(http.StatusOK, out)
	})
	RegisterItemRoutes(r.Group("/items"), services)
	RegisterCollaborationRoutes(r, services)
	RegisterNotificationRoutes(r, services)
	RegisterAdminRoutes(r.Group("/admin"), services)
	RegisterFileRoutes(r, services)
}

func RegisterItemRoutes(r *gin.RouterGroup, services *service.Services) {
	r.POST("", func(c *gin.Context) {
		var body struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		item, err := services.Items.Create(c.Request.Context(), c.GetString("userId"), body.Title, body.Description)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, item)
	})

	r.GET("", func(c *gin.Context) {
		page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
		pageSize, _ := strconv.ParseInt(c.DefaultQuery("pageSize", "10"), 10, 64)
		items, err := services.Items.List(c.Request.Context(), c.GetString("userId"), service.ItemListQuery{
			Page:      page,
			PageSize:  pageSize,
			Search:    c.Query("search"),
			SortBy:    c.DefaultQuery("sortBy", "updatedAt"),
			SortOrder: c.DefaultQuery("sortOrder", "desc"),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load items"})
			return
		}
		c.JSON(http.StatusOK, items)
	})

	r.GET("/:id", func(c *gin.Context) {
		item, err := services.Items.GetByID(c.Request.Context(), c.GetString("userId"), strings.TrimSpace(c.Param("id")))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, item)
	})

	r.PUT("/:id", func(c *gin.Context) {
		var body struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		item, err := services.Items.Update(c.Request.Context(), c.GetString("userId"), strings.TrimSpace(c.Param("id")), body.Title, body.Description)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, item)
	})

	r.DELETE("/:id", func(c *gin.Context) {
		err := services.Items.Delete(c.Request.Context(), c.GetString("userId"), strings.TrimSpace(c.Param("id")))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})
}

func RegisterCollaborationRoutes(r *gin.RouterGroup, services *service.Services) {
	r.GET("/items/:id/comments", func(c *gin.Context) {
		comments, err := services.Collaboration.ListItemComments(c.Request.Context(), strings.TrimSpace(c.Param("id")))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load comments"})
			return
		}
		c.JSON(http.StatusOK, comments)
	})
	r.POST("/items/:id/comments", func(c *gin.Context) {
		var body struct {
			Content  string `json:"content"`
			ParentID string `json:"parentId"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		out, err := services.Collaboration.AddComment(c.Request.Context(), c.GetString("userId"), strings.TrimSpace(c.Param("id")), strings.TrimSpace(body.ParentID), body.Content)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, out)
	})
	r.GET("/activity", func(c *gin.Context) {
		out, err := services.Collaboration.ActivityFeed(c.Request.Context(), c.GetString("userId"), 30)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load activity"})
			return
		}
		c.JSON(http.StatusOK, out)
	})
}

func RegisterNotificationRoutes(r *gin.RouterGroup, services *service.Services) {
	r.GET("/notifications", func(c *gin.Context) {
		out, err := services.Notifications.List(c.Request.Context(), c.GetString("userId"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load notifications"})
			return
		}
		c.JSON(http.StatusOK, out)
	})
	r.PUT("/notifications/:id/read", func(c *gin.Context) {
		var body struct {
			Read bool `json:"read"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if err := services.Notifications.MarkRead(c.Request.Context(), c.GetString("userId"), strings.TrimSpace(c.Param("id")), body.Read); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "updated"})
	})
}

func RegisterAdminRoutes(r *gin.RouterGroup, services *service.Services) {
	r.Use(RequireRoles(services, models.RoleAdmin, models.RoleModerator))
	r.GET("/users", func(c *gin.Context) {
		users, err := services.Admin.ListUsers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load users"})
			return
		}
		c.JSON(http.StatusOK, users)
	})
	r.PUT("/users/:id/role", func(c *gin.Context) {
		var body struct {
			Role models.Role `json:"role"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if err := services.Admin.UpdateUserRole(c.Request.Context(), strings.TrimSpace(c.Param("id")), body.Role); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to update role"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "role updated"})
	})
	r.PUT("/users/:id/active", func(c *gin.Context) {
		var body struct {
			Active bool `json:"active"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if err := services.Admin.SetUserActive(c.Request.Context(), strings.TrimSpace(c.Param("id")), body.Active); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to update user status"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "status updated"})
	})
	r.GET("/metrics", func(c *gin.Context) {
		out, err := services.Admin.Metrics(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load metrics"})
			return
		}
		c.JSON(http.StatusOK, out)
	})
	r.GET("/moderation/reports", func(c *gin.Context) {
		out, err := services.Admin.ModerationQueue(c.Request.Context(), c.Query("status"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load moderation queue"})
			return
		}
		c.JSON(http.StatusOK, out)
	})
	r.PUT("/moderation/reports/:id/review", func(c *gin.Context) {
		var body struct {
			Status string `json:"status"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if err := services.Admin.ReviewModeration(c.Request.Context(), c.GetString("userId"), strings.TrimSpace(c.Param("id")), strings.ToUpper(strings.TrimSpace(body.Status))); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "report reviewed"})
	})
	r.GET("/audit-logs", func(c *gin.Context) {
		out, err := services.Admin.AuditLogs(c.Request.Context(), 100)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load audit logs"})
			return
		}
		c.JSON(http.StatusOK, out)
	})
	r.GET("/email-queue-health", func(c *gin.Context) {
		out, err := services.Admin.EmailQueueHealth(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load email queue health"})
			return
		}
		c.JSON(http.StatusOK, out)
	})
}

func RegisterFileRoutes(r *gin.RouterGroup, services *service.Services) {
	r.POST("/items/:id/attachments", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
			return
		}
		out, err := services.Items.UploadAttachment(c.Request.Context(), c.GetString("userId"), strings.TrimSpace(c.Param("id")), file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, out)
	})
	r.POST("/comments/:id/report", func(c *gin.Context) {
		var body struct {
			Reason string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		out, err := services.Collaboration.ReportComment(c.Request.Context(), c.GetString("userId"), strings.TrimSpace(c.Param("id")), body.Reason)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, out)
	})
	r.POST("/files/:id/access", func(c *gin.Context) {
		url, err := services.Items.CreateFileAccessURL(c.Request.Context(), c.GetString("userId"), strings.TrimSpace(c.Param("id")))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"url": url})
	})
	r.GET("/files/download/:id", func(c *gin.Context) {
		file, resolved, err := services.Items.ResolveFileForDownload(c.Request.Context(), strings.TrimSpace(c.Param("id")), c.Query("token"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if file.StorageKind == "s3" {
			c.Redirect(http.StatusTemporaryRedirect, resolved)
			return
		}
		c.FileAttachment(resolved, file.FileName)
	})
}

func meHandler(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := services.Auth.Me(c.Request.Context(), c.GetString("userId"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
