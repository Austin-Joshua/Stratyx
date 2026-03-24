package service

import (
	"context"
	"log"
	"time"

	"stratyx/backend/internal/domain/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *AuthService) QueueEmail(ctx context.Context, to, subject, body string) error {
	now := time.Now()
	_, err := s.repos.EmailJobs.InsertOne(ctx, models.EmailJob{
		ID:            uuid.NewString(),
		To:            to,
		Subject:       subject,
		Body:          body,
		Status:        "PENDING",
		Attempts:      0,
		NextAttemptAt: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	return err
}

func (s *Services) StartEmailWorker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processEmailBatch(ctx)
		}
	}
}

func (s *Services) processEmailBatch(ctx context.Context) {
	now := time.Now()
	cur, err := s.Auth.repos.EmailJobs.Find(ctx, bson.M{
		"status":          bson.M{"$in": []string{"PENDING", "FAILED"}},
		"next_attempt_at": bson.M{"$lte": now},
	}, options.Find().SetLimit(10).SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var job models.EmailJob
		if err := cur.Decode(&job); err != nil {
			continue
		}
		if job.Attempts >= 5 {
			_, _ = s.Auth.repos.EmailJobs.UpdateOne(ctx, bson.M{"_id": job.ID}, bson.M{
				"$set": bson.M{"status": "DEAD", "updated_at": time.Now()},
			})
			continue
		}
		err := s.Auth.SendEmail(job.To, job.Subject, job.Body)
		if err != nil {
			log.Printf("email send failed for job %s: %v", job.ID, err)
			backoff := time.Duration(job.Attempts+1) * time.Minute
			_, _ = s.Auth.repos.EmailJobs.UpdateOne(ctx, bson.M{"_id": job.ID}, bson.M{
				"$set": bson.M{
					"status":          "FAILED",
					"last_error":      err.Error(),
					"next_attempt_at": time.Now().Add(backoff),
					"updated_at":      time.Now(),
				},
				"$inc": bson.M{"attempts": 1},
			})
			continue
		}
		_, _ = s.Auth.repos.EmailJobs.UpdateOne(ctx, bson.M{"_id": job.ID}, bson.M{
			"$set": bson.M{
				"status":     "SENT",
				"updated_at": time.Now(),
				"last_error": "",
			},
		})
	}
}
