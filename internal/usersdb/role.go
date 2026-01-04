package usersdb

import (
	"context"

	"baas-api/internal/dto"
	"baas-api/internal/models"

	"github.com/google/uuid"
)

func (s *service) GetUsers(ctx context.Context, jwt string, in *dto.GetRolesInput) ([]models.User, error) {
	db, err := s.GetDB(ctx, jwt, in.Ref, "superuser")
	if err != nil {
		return nil, err
	}

	var users []models.User

	if _, err := uuid.Parse(in.Query); err == nil {
		// 如果輸入的是 UUID 格式，則直接以 ID 查詢
		err = db.WithContext(ctx).
			Table("auth.users").
			Where("id = ?", in.Query).
			Limit(1).
			Find(&users).Error
		if err != nil {
			return nil, err
		}
		return users, nil
	}

	err = db.WithContext(ctx).
		Table("auth.users").
		Where("name ILIKE ? OR email ILIKE ?", "%"+in.Query+"%", "%"+in.Query+"%").
		Limit(in.Limit).
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *service) GetGroups(ctx context.Context, jwt string, in *dto.GetRolesInput) ([]models.Group, error) {
	db, err := s.GetDB(ctx, jwt, in.Ref, "superuser")
	if err != nil {
		return nil, err
	}

	var groups []models.Group

	if _, err := uuid.Parse(in.Query); err == nil {
		// 如果輸入的是 UUID 格式，則直接以 ID 查詢
		err = db.WithContext(ctx).
			Table("auth.groups").
			Where("id = ?", in.Query).
			Limit(1).
			Find(&groups).Error
		if err != nil {
			return nil, err
		}
		return groups, nil
	}

	err = db.WithContext(ctx).
		Table("auth.groups").
		Where("name ILIKE ?", "%"+in.Query+"%").
		Limit(in.Limit).
		Find(&groups).Error
	if err != nil {
		return nil, err
	}

	return groups, nil
}
