package models

import "time"

type Role string

const (
	RoleSuperAdmin Role = "SUPER_ADMIN"
	RoleAdmin      Role = "ADMIN"
	RoleModerator  Role = "MODERATOR"
	RoleUser       Role = "USER"
	RoleGuest      Role = "GUEST"
)

type User struct {
	ID                  string                  `bson:"_id" json:"id"`
	Name                string                  `bson:"name" json:"name"`
	Email               string                  `bson:"email" json:"email"`
	PasswordHash        string                  `bson:"password_hash" json:"-"`
	Role                Role                    `bson:"role" json:"role"`
	IsActive            bool                    `bson:"is_active" json:"isActive"`
	IsEmailVerified     bool                    `bson:"is_email_verified" json:"isEmailVerified"`
	FailedLoginAttempts int                     `bson:"failed_login_attempts" json:"-"`
	LockUntil           time.Time               `bson:"lock_until" json:"-"`
	TwoFAEnabled        bool                    `bson:"two_fa_enabled" json:"twoFAEnabled"`
	TwoFASecret         string                  `bson:"two_fa_secret" json:"-"`
	AvatarURL           string                  `bson:"avatar_url" json:"avatarUrl"`
	CreatedAt           time.Time               `bson:"created_at" json:"createdAt"`
	UpdatedAt           time.Time               `bson:"updated_at" json:"updatedAt"`
	NotificationPrefs   NotificationPreferences `bson:"notification_prefs" json:"notificationPrefs"`
	PrivacySettings     PrivacySettings         `bson:"privacy_settings" json:"privacySettings"`
}

type NotificationPreferences struct {
	EmailMentions bool `bson:"email_mentions" json:"emailMentions"`
	InAppMentions bool `bson:"in_app_mentions" json:"inAppMentions"`
	ProductNews   bool `bson:"product_news" json:"productNews"`
}

type PrivacySettings struct {
	ShowEmail   bool `bson:"show_email" json:"showEmail"`
	ShowProfile bool `bson:"show_profile" json:"showProfile"`
}

type Organization struct {
	ID        string    `bson:"_id" json:"id"`
	Name      string    `bson:"name" json:"name"`
	Slug      string    `bson:"slug" json:"slug"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
}

type Membership struct {
	ID             string    `bson:"_id" json:"id"`
	UserID         string    `bson:"user_id" json:"userId"`
	OrganizationID string    `bson:"organization_id" json:"organizationId"`
	Role           Role      `bson:"role" json:"role"`
	CreatedAt      time.Time `bson:"created_at" json:"createdAt"`
}

type Goal struct {
	ID             string    `bson:"_id" json:"id"`
	OrganizationID string    `bson:"organization_id" json:"organizationId"`
	Title          string    `bson:"title" json:"title"`
	Description    string    `bson:"description" json:"description"`
	KPI            string    `bson:"kpi" json:"kpi"`
	TargetValue    float64   `bson:"target_value" json:"targetValue"`
	CurrentValue   float64   `bson:"current_value" json:"currentValue"`
	CreatedAt      time.Time `bson:"created_at" json:"createdAt"`
}

type Project struct {
	ID             string    `bson:"_id" json:"id"`
	OrganizationID string    `bson:"organization_id" json:"organizationId"`
	GoalID         string    `bson:"goal_id" json:"goalId"`
	Name           string    `bson:"name" json:"name"`
	Description    string    `bson:"description" json:"description"`
	Status         string    `bson:"status" json:"status"`
	OwnerUserID    string    `bson:"owner_user_id" json:"ownerUserId"`
	StartDate      time.Time `bson:"start_date" json:"startDate"`
	EndDate        time.Time `bson:"end_date" json:"endDate"`
	CreatedAt      time.Time `bson:"created_at" json:"createdAt"`
}

type Task struct {
	ID             string    `bson:"_id" json:"id"`
	OrganizationID string    `bson:"organization_id" json:"organizationId"`
	ProjectID      string    `bson:"project_id" json:"projectId"`
	Title          string    `bson:"title" json:"title"`
	Description    string    `bson:"description" json:"description"`
	Status         string    `bson:"status" json:"status"`
	Priority       string    `bson:"priority" json:"priority"`
	AssigneeUserID string    `bson:"assignee_user_id" json:"assigneeUserId"`
	DueDate        time.Time `bson:"due_date" json:"dueDate"`
	CreatedAt      time.Time `bson:"created_at" json:"createdAt"`
}

type Risk struct {
	ID             string    `bson:"_id" json:"id"`
	OrganizationID string    `bson:"organization_id" json:"organizationId"`
	ProjectID      string    `bson:"project_id" json:"projectId"`
	Title          string    `bson:"title" json:"title"`
	Severity       int       `bson:"severity" json:"severity"`
	Probability    int       `bson:"probability" json:"probability"`
	MitigationPlan string    `bson:"mitigation_plan" json:"mitigationPlan"`
	CreatedAt      time.Time `bson:"created_at" json:"createdAt"`
}

type Decision struct {
	ID              string    `bson:"_id" json:"id"`
	OrganizationID  string    `bson:"organization_id" json:"organizationId"`
	ProjectID       string    `bson:"project_id" json:"projectId"`
	Description     string    `bson:"description" json:"description"`
	Alternatives    []string  `bson:"alternatives" json:"alternatives"`
	ResponsibleUser string    `bson:"responsible_user" json:"responsibleUser"`
	ImpactAnalysis  string    `bson:"impact_analysis" json:"impactAnalysis"`
	CreatedAt       time.Time `bson:"created_at" json:"createdAt"`
}

type Session struct {
	ID           string    `bson:"_id" json:"id"`
	UserID       string    `bson:"user_id" json:"userId"`
	RefreshToken string    `bson:"refresh_token" json:"-"`
	ExpiresAt    time.Time `bson:"expires_at" json:"expiresAt"`
	DeviceName   string    `bson:"device_name" json:"deviceName"`
	IPAddress    string    `bson:"ip_address" json:"ipAddress"`
	LastSeenAt   time.Time `bson:"last_seen_at" json:"lastSeenAt"`
	CreatedAt    time.Time `bson:"created_at" json:"createdAt"`
}

type Item struct {
	ID          string    `bson:"_id" json:"id"`
	Title       string    `bson:"title" json:"title"`
	Description string    `bson:"description" json:"description"`
	OwnerID     string    `bson:"owner_id" json:"ownerId"`
	CreatedAt   time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updatedAt"`
}

type PasswordResetToken struct {
	ID        string    `bson:"_id" json:"id"`
	UserID    string    `bson:"user_id" json:"userId"`
	Email     string    `bson:"email" json:"email"`
	Token     string    `bson:"token" json:"token"`
	ExpiresAt time.Time `bson:"expires_at" json:"expiresAt"`
	UsedAt    time.Time `bson:"used_at" json:"usedAt"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
}

type EmailVerificationToken struct {
	ID        string    `bson:"_id" json:"id"`
	UserID    string    `bson:"user_id" json:"userId"`
	Email     string    `bson:"email" json:"email"`
	Token     string    `bson:"token" json:"token"`
	ExpiresAt time.Time `bson:"expires_at" json:"expiresAt"`
	UsedAt    time.Time `bson:"used_at" json:"usedAt"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
}

type Comment struct {
	ID        string    `bson:"_id" json:"id"`
	ItemID    string    `bson:"item_id" json:"itemId"`
	AuthorID  string    `bson:"author_id" json:"authorId"`
	ParentID  string    `bson:"parent_id" json:"parentId"`
	Content   string    `bson:"content" json:"content"`
	Mentions  []string  `bson:"mentions" json:"mentions"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
}

type Notification struct {
	ID        string    `bson:"_id" json:"id"`
	UserID    string    `bson:"user_id" json:"userId"`
	Type      string    `bson:"type" json:"type"`
	Title     string    `bson:"title" json:"title"`
	Body      string    `bson:"body" json:"body"`
	Read      bool      `bson:"read" json:"read"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
}

type ActivityLog struct {
	ID        string    `bson:"_id" json:"id"`
	ActorID   string    `bson:"actor_id" json:"actorId"`
	Entity    string    `bson:"entity" json:"entity"`
	EntityID  string    `bson:"entity_id" json:"entityId"`
	Action    string    `bson:"action" json:"action"`
	Message   string    `bson:"message" json:"message"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
}

type OAuthAccount struct {
	ID             string    `bson:"_id" json:"id"`
	UserID         string    `bson:"user_id" json:"userId"`
	Provider       string    `bson:"provider" json:"provider"`
	ProviderUserID string    `bson:"provider_user_id" json:"providerUserId"`
	Email          string    `bson:"email" json:"email"`
	CreatedAt      time.Time `bson:"created_at" json:"createdAt"`
}

type AuthChallenge struct {
	ID        string    `bson:"_id" json:"id"`
	UserID    string    `bson:"user_id" json:"userId"`
	Token     string    `bson:"token" json:"token"`
	ExpiresAt time.Time `bson:"expires_at" json:"expiresAt"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
}

type FileAsset struct {
	ID          string    `bson:"_id" json:"id"`
	OwnerID     string    `bson:"owner_id" json:"ownerId"`
	ItemID      string    `bson:"item_id" json:"itemId"`
	FileName    string    `bson:"file_name" json:"fileName"`
	ContentType string    `bson:"content_type" json:"contentType"`
	SizeBytes   int64     `bson:"size_bytes" json:"sizeBytes"`
	URL         string    `bson:"url" json:"url"`
	StorageKind string    `bson:"storage_kind" json:"storageKind"`
	CreatedAt   time.Time `bson:"created_at" json:"createdAt"`
}

type ModerationReport struct {
	ID         string    `bson:"_id" json:"id"`
	CommentID  string    `bson:"comment_id" json:"commentId"`
	ReporterID string    `bson:"reporter_id" json:"reporterId"`
	Reason     string    `bson:"reason" json:"reason"`
	Status     string    `bson:"status" json:"status"`
	ReviewedBy string    `bson:"reviewed_by" json:"reviewedBy"`
	ReviewedAt time.Time `bson:"reviewed_at" json:"reviewedAt"`
	CreatedAt  time.Time `bson:"created_at" json:"createdAt"`
}

type OAuthState struct {
	ID        string    `bson:"_id" json:"id"`
	Provider  string    `bson:"provider" json:"provider"`
	State     string    `bson:"state" json:"state"`
	ExpiresAt time.Time `bson:"expires_at" json:"expiresAt"`
	UsedAt    time.Time `bson:"used_at" json:"usedAt"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
}

type EmailJob struct {
	ID            string    `bson:"_id" json:"id"`
	To            string    `bson:"to" json:"to"`
	Subject       string    `bson:"subject" json:"subject"`
	Body          string    `bson:"body" json:"body"`
	Status        string    `bson:"status" json:"status"`
	Attempts      int       `bson:"attempts" json:"attempts"`
	LastError     string    `bson:"last_error" json:"lastError"`
	NextAttemptAt time.Time `bson:"next_attempt_at" json:"nextAttemptAt"`
	CreatedAt     time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt     time.Time `bson:"updated_at" json:"updatedAt"`
}

type FileAccessToken struct {
	ID        string    `bson:"_id" json:"id"`
	FileID    string    `bson:"file_id" json:"fileId"`
	UserID    string    `bson:"user_id" json:"userId"`
	Token     string    `bson:"token" json:"token"`
	ExpiresAt time.Time `bson:"expires_at" json:"expiresAt"`
	UsedAt    time.Time `bson:"used_at" json:"usedAt"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
}
