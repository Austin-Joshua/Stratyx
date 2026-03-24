package service

import (
	"context"
	"errors"
	"time"

	"stratyx/backend/internal/domain/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *CollaborationService) ReportComment(ctx context.Context, reporterID, commentID, reason string) (*models.ModerationReport, error) {
	if reason == "" {
		return nil, errors.New("reason is required")
	}
	report := &models.ModerationReport{
		ID:         uuid.NewString(),
		CommentID:  commentID,
		ReporterID: reporterID,
		Reason:     reason,
		Status:     "OPEN",
		CreatedAt:  time.Now(),
	}
	_, err := s.repos.ModerationReports.InsertOne(ctx, report)
	return report, err
}

func (s *AdminService) ModerationQueue(ctx context.Context, status string) ([]models.ModerationReport, error) {
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}
	cur, err := s.repos.ModerationReports.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(200))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]models.ModerationReport, 0)
	for cur.Next(ctx) {
		var r models.ModerationReport
		if err := cur.Decode(&r); err == nil {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *AdminService) ReviewModeration(ctx context.Context, reviewerID, reportID, status string) error {
	if status != "APPROVED" && status != "REJECTED" {
		return errors.New("status must be APPROVED or REJECTED")
	}
	_, err := s.repos.ModerationReports.UpdateOne(ctx, bson.M{"_id": reportID}, bson.M{"$set": bson.M{
		"status":      status,
		"reviewed_by": reviewerID,
		"reviewed_at": time.Now(),
	}})
	if err != nil {
		return err
	}
	_, _ = s.repos.ActivityLogs.InsertOne(ctx, models.ActivityLog{
		ID:        uuid.NewString(),
		ActorID:   reviewerID,
		Entity:    "moderation_report",
		EntityID:  reportID,
		Action:    "reviewed",
		Message:   "Reviewed moderation report as " + status,
		CreatedAt: time.Now(),
	})
	return nil
}

func (s *AdminService) AuditLogs(ctx context.Context, limit int64) ([]models.ActivityLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	cur, err := s.repos.ActivityLogs.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit))
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
