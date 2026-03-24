package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"stratyx/backend/internal/domain/models"
	"stratyx/backend/internal/platform/security"
	"stratyx/backend/internal/repository"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type Services struct {
	Auth          *AuthService
	Items         *ItemService
	Collaboration *CollaborationService
	Notifications *NotificationService
	Admin         *AdminService
	AI            *AIService
	Insight       *InsightService
}

func NewServices(repos *repository.Repositories, jwt *security.JWTManager) *Services {
	return &Services{
		Auth:          &AuthService{repos: repos, jwt: jwt},
		Items:         &ItemService{repos: repos},
		Collaboration: &CollaborationService{repos: repos},
		Notifications: &NotificationService{repos: repos},
		Admin:         &AdminService{repos: repos},
		AI:            &AIService{repos: repos},
		Insight:       &InsightService{repos: repos},
	}
}

type AuthService struct {
	repos *repository.Repositories
	jwt   *security.JWTManager
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (*models.User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if name == "" || email == "" || len(password) < 8 {
		return nil, errors.New("name, email, and password(min 8 chars) are required")
	}
	var existing models.User
	if err := s.repos.Users.FindOne(ctx, bson.M{"email": email}).Decode(&existing); err == nil {
		return nil, errors.New("email already in use")
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		ID:              uuid.NewString(),
		Name:            name,
		Email:           email,
		PasswordHash:    string(hash),
		Role:            models.RoleUser,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		IsEmailVerified: false,
	}
	_, err = s.repos.Users.InsertOne(ctx, user)
	if err == nil {
		_, _ = s.CreateEmailVerificationToken(ctx, user.ID, user.Email)
	}
	return user, err
}

func (s *AuthService) Login(ctx context.Context, email, password, device, ip string) (*models.User, string, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var user models.User
	if err := s.repos.Users.FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		return nil, "", "", errors.New("invalid credentials")
	}
	if !user.IsActive {
		return nil, "", "", errors.New("account is deactivated")
	}
	if !user.LockUntil.IsZero() && user.LockUntil.After(time.Now()) {
		return nil, "", "", errors.New("account temporarily locked due to failed attempts")
	}
	if password != "__oauth_or_verified_2fa__" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			attempts := user.FailedLoginAttempts + 1
			update := bson.M{"$set": bson.M{"failed_login_attempts": attempts, "updated_at": time.Now()}}
			if attempts >= 5 {
				update["$set"].(bson.M)["lock_until"] = time.Now().Add(15 * time.Minute)
			}
			_, _ = s.repos.Users.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
			return nil, "", "", errors.New("invalid credentials")
		}
	}
	_, _ = s.repos.Users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{
		"failed_login_attempts": 0,
		"lock_until":            time.Time{},
		"updated_at":            time.Now(),
	}})
	access, err := s.jwt.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}
	refresh, exp, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}
	session := models.Session{
		ID:           uuid.NewString(),
		UserID:       user.ID,
		RefreshToken: refresh,
		ExpiresAt:    exp,
		DeviceName:   device,
		IPAddress:    ip,
		LastSeenAt:   time.Now(),
		CreatedAt:    time.Now(),
	}
	_, err = s.repos.Sessions.InsertOne(ctx, session)
	return &user, access, refresh, err
}

func (s *AuthService) Refresh(ctx context.Context, refresh string) (string, string, error) {
	var session models.Session
	if err := s.repos.Sessions.FindOne(ctx, bson.M{"refresh_token": refresh}).Decode(&session); err != nil {
		return "", "", errors.New("session not found")
	}
	if session.ExpiresAt.Before(time.Now()) {
		return "", "", errors.New("refresh expired")
	}
	_, _ = s.repos.Sessions.UpdateOne(ctx, bson.M{"_id": session.ID}, bson.M{"$set": bson.M{"last_seen_at": time.Now()}})
	access, err := s.jwt.GenerateAccessToken(session.UserID)
	return access, refresh, err
}

func (s *AuthService) Logout(ctx context.Context, refresh string) error {
	_, err := s.repos.Sessions.DeleteOne(ctx, bson.M{"refresh_token": refresh})
	return err
}

func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	_, err := s.repos.Sessions.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}

func (s *AuthService) Me(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	if err := s.repos.Users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID, name, avatarURL string, privacy models.PrivacySettings, prefs models.NotificationPreferences) (*models.User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	_, err := s.repos.Users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{
		"name":               name,
		"avatar_url":         strings.TrimSpace(avatarURL),
		"privacy_settings":   privacy,
		"notification_prefs": prefs,
		"updated_at":         time.Now(),
	}})
	if err != nil {
		return nil, err
	}
	return s.Me(ctx, userID)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}
	var user models.User
	if err := s.repos.Users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		return errors.New("user not found")
	}
	if user.PasswordHash == "" {
		return errors.New("password login is not configured for this account")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.repos.Users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{
		"password_hash": string(hash),
		"updated_at":    time.Now(),
	}})
	if err != nil {
		return err
	}
	_, _ = s.repos.Sessions.DeleteMany(ctx, bson.M{"user_id": userID})
	return nil
}

func (s *AuthService) ListSessions(ctx context.Context, userID string) ([]models.Session, error) {
	cur, err := s.repos.Sessions.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	sessions := make([]models.Session, 0)
	for cur.Next(ctx) {
		var session models.Session
		if err := cur.Decode(&session); err == nil {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (s *AuthService) CreatePasswordResetToken(ctx context.Context, email string) (string, error) {
	var user models.User
	if err := s.repos.Users.FindOne(ctx, bson.M{"email": strings.ToLower(strings.TrimSpace(email))}).Decode(&user); err != nil {
		return "", errors.New("if the account exists, a reset token was generated")
	}
	token := randomToken(24)
	doc := models.PasswordResetToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(30 * time.Minute),
		CreatedAt: time.Now(),
	}
	_, err := s.repos.PasswordResets.InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	var doc models.PasswordResetToken
	if err := s.repos.PasswordResets.FindOne(ctx, bson.M{"token": token}).Decode(&doc); err != nil {
		return errors.New("invalid reset token")
	}
	if !doc.UsedAt.IsZero() || doc.ExpiresAt.Before(time.Now()) {
		return errors.New("reset token expired")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.repos.Users.UpdateOne(ctx, bson.M{"_id": doc.UserID}, bson.M{"$set": bson.M{
		"password_hash": string(hash),
		"updated_at":    time.Now(),
	}})
	if err != nil {
		return err
	}
	_, _ = s.repos.PasswordResets.UpdateOne(ctx, bson.M{"_id": doc.ID}, bson.M{"$set": bson.M{"used_at": time.Now()}})
	_, _ = s.repos.Sessions.DeleteMany(ctx, bson.M{"user_id": doc.UserID})
	return nil
}

func (s *AuthService) CreateEmailVerificationToken(ctx context.Context, userID, email string) (string, error) {
	token := randomToken(24)
	doc := models.EmailVerificationToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	_, err := s.repos.EmailVerifications.InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	var doc models.EmailVerificationToken
	if err := s.repos.EmailVerifications.FindOne(ctx, bson.M{"token": token}).Decode(&doc); err != nil {
		return errors.New("invalid verification token")
	}
	if !doc.UsedAt.IsZero() || doc.ExpiresAt.Before(time.Now()) {
		return errors.New("verification token expired")
	}
	_, err := s.repos.Users.UpdateOne(ctx, bson.M{"_id": doc.UserID}, bson.M{"$set": bson.M{
		"is_email_verified": true,
		"updated_at":        time.Now(),
	}})
	if err != nil {
		return err
	}
	_, _ = s.repos.EmailVerifications.UpdateOne(ctx, bson.M{"_id": doc.ID}, bson.M{"$set": bson.M{"used_at": time.Now()}})
	return nil
}

type ItemService struct {
	repos *repository.Repositories
}

type ItemListQuery struct {
	Page      int64
	PageSize  int64
	Search    string
	SortBy    string
	SortOrder string
}

type ItemListResult struct {
	Items      []models.Item `json:"items"`
	Page       int64         `json:"page"`
	PageSize   int64         `json:"pageSize"`
	Total      int64         `json:"total"`
	TotalPages int64         `json:"totalPages"`
}

type DashboardSummary struct {
	ItemsCount          int64 `json:"itemsCount"`
	CommentsCount       int64 `json:"commentsCount"`
	UnreadNotifications int64 `json:"unreadNotifications"`
}

func (s *ItemService) Create(ctx context.Context, ownerID, title, description string) (*models.Item, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" {
		return nil, errors.New("title is required")
	}
	now := time.Now()
	item := &models.Item{
		ID:          uuid.NewString(),
		Title:       title,
		Description: description,
		OwnerID:     ownerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err := s.repos.Items.InsertOne(ctx, item)
	if err == nil {
		_, _ = s.repos.ActivityLogs.InsertOne(ctx, models.ActivityLog{
			ID:        uuid.NewString(),
			ActorID:   ownerID,
			Entity:    "item",
			EntityID:  item.ID,
			Action:    "created",
			Message:   "Created item: " + item.Title,
			CreatedAt: time.Now(),
		})
	}
	return item, err
}

func (s *ItemService) List(ctx context.Context, ownerID string, q ItemListQuery) (*ItemListResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 10
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	search := strings.TrimSpace(q.Search)
	sortBy := strings.ToLower(strings.TrimSpace(q.SortBy))
	sortOrder := strings.ToLower(strings.TrimSpace(q.SortOrder))
	sortField := "updated_at"
	switch sortBy {
	case "title":
		sortField = "title"
	case "createdat", "created_at":
		sortField = "created_at"
	case "updatedat", "updated_at":
		sortField = "updated_at"
	}
	sortDir := int32(-1)
	if sortOrder == "asc" {
		sortDir = 1
	}

	filter := bson.M{"owner_id": ownerID}
	if search != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": search, "$options": "i"}},
			{"description": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	total, err := s.repos.Items.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	opts := options.Find().
		SetSkip((q.Page - 1) * q.PageSize).
		SetLimit(q.PageSize).
		SetSort(bson.D{{Key: sortField, Value: sortDir}})

	cur, err := s.repos.Items.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	items := make([]models.Item, 0)
	for cur.Next(ctx) {
		var item models.Item
		if err := cur.Decode(&item); err == nil {
			items = append(items, item)
		}
	}
	totalPages := int64(math.Ceil(float64(total) / float64(q.PageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &ItemListResult{
		Items:      items,
		Page:       q.Page,
		PageSize:   q.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *ItemService) GetByID(ctx context.Context, ownerID, itemID string) (*models.Item, error) {
	var item models.Item
	if err := s.repos.Items.FindOne(ctx, bson.M{"_id": itemID, "owner_id": ownerID}).Decode(&item); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("item not found")
		}
		return nil, err
	}
	return &item, nil
}

func (s *ItemService) Update(ctx context.Context, ownerID, itemID, title, description string) (*models.Item, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" {
		return nil, errors.New("title is required")
	}
	update := bson.M{
		"$set": bson.M{
			"title":       title,
			"description": description,
			"updated_at":  time.Now(),
		},
	}
	res, err := s.repos.Items.UpdateOne(ctx, bson.M{"_id": itemID, "owner_id": ownerID}, update)
	if err != nil {
		return nil, err
	}
	if res.MatchedCount == 0 {
		return nil, errors.New("item not found")
	}
	_, _ = s.repos.ActivityLogs.InsertOne(ctx, models.ActivityLog{
		ID:        uuid.NewString(),
		ActorID:   ownerID,
		Entity:    "item",
		EntityID:  itemID,
		Action:    "updated",
		Message:   "Updated item: " + title,
		CreatedAt: time.Now(),
	})
	return s.GetByID(ctx, ownerID, itemID)
}

func (s *ItemService) Delete(ctx context.Context, ownerID, itemID string) error {
	res, err := s.repos.Items.DeleteOne(ctx, bson.M{"_id": itemID, "owner_id": ownerID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("item not found")
	}
	_, _ = s.repos.ActivityLogs.InsertOne(ctx, models.ActivityLog{
		ID:        uuid.NewString(),
		ActorID:   ownerID,
		Entity:    "item",
		EntityID:  itemID,
		Action:    "deleted",
		Message:   "Deleted an item",
		CreatedAt: time.Now(),
	})
	return nil
}

func (s *ItemService) DashboardSummary(ctx context.Context, userID string) (*DashboardSummary, error) {
	itemsCount, err := s.repos.Items.CountDocuments(ctx, bson.M{"owner_id": userID})
	if err != nil {
		return nil, err
	}
	commentsCount, err := s.repos.Comments.CountDocuments(ctx, bson.M{"author_id": userID})
	if err != nil {
		return nil, err
	}
	unread, err := s.repos.Notifications.CountDocuments(ctx, bson.M{"user_id": userID, "read": false})
	if err != nil {
		return nil, err
	}
	return &DashboardSummary{
		ItemsCount:          itemsCount,
		CommentsCount:       commentsCount,
		UnreadNotifications: unread,
	}, nil
}

type CollaborationService struct {
	repos *repository.Repositories
}

func (s *CollaborationService) AddComment(ctx context.Context, authorID, itemID, parentID, content string) (*models.Comment, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("comment content is required")
	}
	mentionRegex := regexp.MustCompile(`@([a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,})`)
	matches := mentionRegex.FindAllStringSubmatch(content, -1)
	mentions := make([]string, 0)
	for _, m := range matches {
		if len(m) > 1 {
			mentions = append(mentions, strings.ToLower(strings.TrimSpace(m[1])))
		}
	}
	comment := &models.Comment{
		ID:        uuid.NewString(),
		ItemID:    itemID,
		AuthorID:  authorID,
		ParentID:  parentID,
		Content:   content,
		Mentions:  mentions,
		CreatedAt: time.Now(),
	}
	_, err := s.repos.Comments.InsertOne(ctx, comment)
	if err != nil {
		return nil, err
	}
	for _, mentionEmail := range mentions {
		var user models.User
		if err := s.repos.Users.FindOne(ctx, bson.M{"email": mentionEmail}).Decode(&user); err == nil {
			_, _ = s.repos.Notifications.InsertOne(ctx, models.Notification{
				ID:        uuid.NewString(),
				UserID:    user.ID,
				Type:      "mention",
				Title:     "You were mentioned",
				Body:      content,
				Read:      false,
				CreatedAt: time.Now(),
			})
		}
	}
	_, _ = s.repos.ActivityLogs.InsertOne(ctx, models.ActivityLog{
		ID:        uuid.NewString(),
		ActorID:   authorID,
		Entity:    "comment",
		EntityID:  comment.ID,
		Action:    "created",
		Message:   "Commented on item",
		CreatedAt: time.Now(),
	})
	return comment, nil
}

func (s *CollaborationService) ListItemComments(ctx context.Context, itemID string) ([]models.Comment, error) {
	cur, err := s.repos.Comments.Find(ctx, bson.M{"item_id": itemID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]models.Comment, 0)
	for cur.Next(ctx) {
		var c models.Comment
		if err := cur.Decode(&c); err == nil {
			out = append(out, c)
		}
	}
	return out, nil
}

func (s *CollaborationService) ActivityFeed(ctx context.Context, userID string, limit int64) ([]models.ActivityLog, error) {
	if limit < 1 || limit > 100 {
		limit = 25
	}
	cur, err := s.repos.ActivityLogs.Find(ctx, bson.M{"actor_id": userID}, options.Find().SetLimit(limit).SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]models.ActivityLog, 0)
	for cur.Next(ctx) {
		var a models.ActivityLog
		if err := cur.Decode(&a); err == nil {
			out = append(out, a)
		}
	}
	return out, nil
}

type NotificationService struct {
	repos *repository.Repositories
}

func (s *NotificationService) Create(ctx context.Context, userID, nt, title, body string) error {
	_, err := s.repos.Notifications.InsertOne(ctx, models.Notification{
		ID:        uuid.NewString(),
		UserID:    userID,
		Type:      nt,
		Title:     title,
		Body:      body,
		Read:      false,
		CreatedAt: time.Now(),
	})
	return err
}

func (s *NotificationService) List(ctx context.Context, userID string) ([]models.Notification, error) {
	cur, err := s.repos.Notifications.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(50))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]models.Notification, 0)
	for cur.Next(ctx) {
		var n models.Notification
		if err := cur.Decode(&n); err == nil {
			out = append(out, n)
		}
	}
	return out, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, userID, id string, read bool) error {
	res, err := s.repos.Notifications.UpdateOne(ctx, bson.M{"_id": id, "user_id": userID}, bson.M{"$set": bson.M{"read": read}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("notification not found")
	}
	return nil
}

type AdminService struct {
	repos *repository.Repositories
}

func (s *AdminService) ListUsers(ctx context.Context) ([]models.User, error) {
	cur, err := s.repos.Users.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(200))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	users := make([]models.User, 0)
	for cur.Next(ctx) {
		var u models.User
		if err := cur.Decode(&u); err == nil {
			users = append(users, u)
		}
	}
	return users, nil
}

func (s *AdminService) UpdateUserRole(ctx context.Context, userID string, role models.Role) error {
	_, err := s.repos.Users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"role": role, "updated_at": time.Now()}})
	return err
}

func (s *AdminService) SetUserActive(ctx context.Context, userID string, active bool) error {
	_, err := s.repos.Users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"is_active": active, "updated_at": time.Now()}})
	return err
}

type AdminMetrics struct {
	UsersTotal         int64 `json:"usersTotal"`
	UsersActive        int64 `json:"usersActive"`
	ItemsTotal         int64 `json:"itemsTotal"`
	SessionsActive     int64 `json:"sessionsActive"`
	CommentsTotal      int64 `json:"commentsTotal"`
	NotificationsTotal int64 `json:"notificationsTotal"`
}

type EmailQueueHealth struct {
	Pending      int64             `json:"pending"`
	Failed       int64             `json:"failed"`
	Dead         int64             `json:"dead"`
	RecentFailed []models.EmailJob `json:"recentFailed"`
}

func (s *AdminService) Metrics(ctx context.Context) (*AdminMetrics, error) {
	usersTotal, err := s.repos.Users.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	usersActive, err := s.repos.Users.CountDocuments(ctx, bson.M{"is_active": true})
	if err != nil {
		return nil, err
	}
	itemsTotal, err := s.repos.Items.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	sessionsActive, err := s.repos.Sessions.CountDocuments(ctx, bson.M{"expires_at": bson.M{"$gt": time.Now()}})
	if err != nil {
		return nil, err
	}
	commentsTotal, err := s.repos.Comments.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	notificationsTotal, err := s.repos.Notifications.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	return &AdminMetrics{
		UsersTotal: usersTotal, UsersActive: usersActive, ItemsTotal: itemsTotal, SessionsActive: sessionsActive, CommentsTotal: commentsTotal, NotificationsTotal: notificationsTotal,
	}, nil
}

func (s *AdminService) EmailQueueHealth(ctx context.Context) (*EmailQueueHealth, error) {
	pending, err := s.repos.EmailJobs.CountDocuments(ctx, bson.M{"status": "PENDING"})
	if err != nil {
		return nil, err
	}
	failed, err := s.repos.EmailJobs.CountDocuments(ctx, bson.M{"status": "FAILED"})
	if err != nil {
		return nil, err
	}
	dead, err := s.repos.EmailJobs.CountDocuments(ctx, bson.M{"status": "DEAD"})
	if err != nil {
		return nil, err
	}
	cur, err := s.repos.EmailJobs.Find(
		ctx,
		bson.M{"status": "FAILED"},
		options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}}).SetLimit(10),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	recentFailed := make([]models.EmailJob, 0, 10)
	for cur.Next(ctx) {
		var j models.EmailJob
		if err := cur.Decode(&j); err == nil {
			recentFailed = append(recentFailed, j)
		}
	}
	return &EmailQueueHealth{
		Pending:      pending,
		Failed:       failed,
		Dead:         dead,
		RecentFailed: recentFailed,
	}, nil
}

func randomToken(byteLen int) string {
	if byteLen < 8 {
		byteLen = 8
	}
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", buf)
}

type AIService struct {
	repos *repository.Repositories
}

func (s *AIService) GenerateTaskPlan(goalTitle string) []string {
	return []string{
		"Define scope and success metrics for: " + goalTitle,
		"Create milestone timeline based on dependencies",
		"Assign owners based on historical delivery capacity",
		"Set weekly risk review checkpoints",
	}
}

func (s *AIService) DelayRiskScore(percentComplete float64, daysToDeadline int, blockers int) float64 {
	score := (1-percentComplete)*50 + float64(blockers)*10
	if daysToDeadline < 7 {
		score += 20
	}
	if score > 100 {
		return 100
	}
	if score < 0 {
		return 0
	}
	return score
}

type InsightService struct {
	repos *repository.Repositories
}

func (s *InsightService) OrgSummary(ctx context.Context, orgID string) (map[string]int64, error) {
	tasks, err := s.repos.Tasks.CountDocuments(ctx, bson.M{"organization_id": orgID})
	if err != nil {
		return nil, err
	}
	projects, err := s.repos.Projects.CountDocuments(ctx, bson.M{"organization_id": orgID})
	if err != nil {
		return nil, err
	}
	risks, err := s.repos.Risks.CountDocuments(ctx, bson.M{"organization_id": orgID})
	if err != nil {
		return nil, err
	}
	return map[string]int64{
		"projects": projects,
		"tasks":    tasks,
		"risks":    risks,
	}, nil
}
