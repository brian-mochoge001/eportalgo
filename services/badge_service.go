package services

import (
	"context"
	"database/sql"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type BadgeService struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewBadgeService(q *db.Queries, d *sql.DB) *BadgeService {
	return &BadgeService{Queries: q, DB: d}
}

func (s *BadgeService) CreateBadge(ctx context.Context, params db.CreateBadgeParams) (db.Badge, error) {
	return s.Queries.CreateBadge(ctx, params)
}

func (s *BadgeService) GetBadgesBySchool(ctx context.Context, schoolID uuid.UUID) ([]db.Badge, error) {
	return s.Queries.GetBadgesBySchool(ctx, schoolID)
}

func (s *BadgeService) GetBadgeByID(ctx context.Context, badgeID uuid.UUID, schoolID uuid.UUID) (db.Badge, error) {
	return s.Queries.GetBadgeByID(ctx, db.GetBadgeByIDParams{
		BadgeID:  badgeID,
		SchoolID: schoolID,
	})
}

func (s *BadgeService) UpdateBadge(ctx context.Context, params db.UpdateBadgeParams) (db.Badge, error) {
	return s.Queries.UpdateBadge(ctx, params)
}

func (s *BadgeService) DeleteBadge(ctx context.Context, badgeID uuid.UUID, schoolID uuid.UUID) error {
	return s.Queries.DeleteBadge(ctx, db.DeleteBadgeParams{
		BadgeID:  badgeID,
		SchoolID: schoolID,
	})
}

func (s *BadgeService) AwardBadge(ctx context.Context, params db.AwardBadgeParams) (db.StudentBadge, error) {
	return s.Queries.AwardBadge(ctx, params)
}

func (s *BadgeService) RevokeBadge(ctx context.Context, params db.RevokeBadgeParams) error {
	return s.Queries.RevokeBadge(ctx, params)
}

func (s *BadgeService) GetStudentBadges(ctx context.Context, studentID uuid.UUID, schoolID uuid.UUID) ([]db.GetStudentBadgesRow, error) {
	return s.Queries.GetStudentBadges(ctx, db.GetStudentBadgesParams{
		StudentID: studentID,
		SchoolID:  schoolID,
	})
}
