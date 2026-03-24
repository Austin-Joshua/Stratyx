package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repositories struct {
	Users              *mongo.Collection
	Organizations      *mongo.Collection
	Memberships        *mongo.Collection
	Goals              *mongo.Collection
	Projects           *mongo.Collection
	Tasks              *mongo.Collection
	Risks              *mongo.Collection
	Decisions          *mongo.Collection
	Sessions           *mongo.Collection
	Items              *mongo.Collection
	PasswordResets     *mongo.Collection
	EmailVerifications *mongo.Collection
	Comments           *mongo.Collection
	Notifications      *mongo.Collection
	ActivityLogs       *mongo.Collection
	OAuthAccounts      *mongo.Collection
	AuthChallenges     *mongo.Collection
	Files              *mongo.Collection
	ModerationReports  *mongo.Collection
	OAuthStates        *mongo.Collection
	EmailJobs          *mongo.Collection
	FileAccessTokens   *mongo.Collection
}

func NewMongo(uri, dbName string) (*mongo.Client, *mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, err
	}
	return client, client.Database(dbName), nil
}

func NewRepositories(db *mongo.Database) *Repositories {
	return &Repositories{
		Users:              db.Collection("users"),
		Organizations:      db.Collection("organizations"),
		Memberships:        db.Collection("memberships"),
		Goals:              db.Collection("goals"),
		Projects:           db.Collection("projects"),
		Tasks:              db.Collection("tasks"),
		Risks:              db.Collection("risks"),
		Decisions:          db.Collection("decisions"),
		Sessions:           db.Collection("sessions"),
		Items:              db.Collection("items"),
		PasswordResets:     db.Collection("password_resets"),
		EmailVerifications: db.Collection("email_verifications"),
		Comments:           db.Collection("comments"),
		Notifications:      db.Collection("notifications"),
		ActivityLogs:       db.Collection("activity_logs"),
		OAuthAccounts:      db.Collection("oauth_accounts"),
		AuthChallenges:     db.Collection("auth_challenges"),
		Files:              db.Collection("files"),
		ModerationReports:  db.Collection("moderation_reports"),
		OAuthStates:        db.Collection("oauth_states"),
		EmailJobs:          db.Collection("email_jobs"),
		FileAccessTokens:   db.Collection("file_access_tokens"),
	}
}

func (r *Repositories) EnsureIndexes(ctx context.Context) error {
	_, err := r.Users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	_, err = r.Items.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "owner_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "updated_at", Value: -1}},
		},
	})
	if err != nil {
		return err
	}

	_, err = r.Sessions.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "refresh_token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		return err
	}

	_, err = r.Notifications.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "read", Value: 1}},
		},
	})
	if err != nil {
		return err
	}
	_, err = r.OAuthAccounts.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "provider", Value: 1}, {Key: "provider_user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		return err
	}
	_, err = r.Files.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "owner_id", Value: 1}}})
	if err != nil {
		return err
	}
	_, err = r.ModerationReports.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "comment_id", Value: 1}}},
	})
	if err != nil {
		return err
	}
	_, err = r.OAuthStates.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "state", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	_, err = r.OAuthStates.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	if err != nil {
		return err
	}
	_, err = r.OAuthStates.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "used_at", Value: 1}},
		Options: options.Index().SetPartialFilterExpression(bson.M{
			"used_at": bson.M{"$exists": true, "$ne": time.Time{}},
		}).SetExpireAfterSeconds(86400),
	})
	if err != nil {
		return err
	}
	_, err = r.EmailJobs.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "next_attempt_at", Value: 1}}},
		{
			Keys: bson.D{{Key: "updated_at", Value: 1}},
			Options: options.Index().
				SetPartialFilterExpression(bson.M{"status": "SENT"}).
				SetExpireAfterSeconds(2592000),
		},
		{
			Keys: bson.D{{Key: "updated_at", Value: 1}},
			Options: options.Index().
				SetPartialFilterExpression(bson.M{"status": "DEAD"}).
				SetExpireAfterSeconds(604800),
		},
	})
	if err != nil {
		return err
	}
	_, err = r.FileAccessTokens.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "token", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	_, err = r.FileAccessTokens.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	if err != nil {
		return err
	}
	_, err = r.FileAccessTokens.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "used_at", Value: 1}},
		Options: options.Index().SetPartialFilterExpression(bson.M{
			"used_at": bson.M{"$exists": true, "$ne": time.Time{}},
		}).SetExpireAfterSeconds(86400),
	})
	if err != nil {
		return err
	}
	_, err = r.PasswordResets.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "token", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
		{
			Keys: bson.D{{Key: "used_at", Value: 1}},
			Options: options.Index().SetPartialFilterExpression(bson.M{
				"used_at": bson.M{"$exists": true, "$ne": time.Time{}},
			}).SetExpireAfterSeconds(86400),
		},
	})
	if err != nil {
		return err
	}
	_, err = r.EmailVerifications.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "token", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
		{
			Keys: bson.D{{Key: "used_at", Value: 1}},
			Options: options.Index().SetPartialFilterExpression(bson.M{
				"used_at": bson.M{"$exists": true, "$ne": time.Time{}},
			}).SetExpireAfterSeconds(86400),
		},
	})
	return err
}
