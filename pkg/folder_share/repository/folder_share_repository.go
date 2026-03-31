package repository

import (
	"context"
	"evo-ai-core-service/pkg/folder_share/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FolderShareRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.FolderShare, error)
	GetByFolderIDAndSharedWithEmail(ctx context.Context, folderID uuid.UUID, sharedWithEmail string) (*model.FolderShare, error)
	Create(ctx context.Context, folderShare model.FolderShare) (*model.FolderShare, error)
	Update(ctx context.Context, id uuid.UUID, folderShare *model.FolderShare) (*model.FolderShare, error)
	Delete(ctx context.Context, id uuid.UUID) (bool, error)
	GetSharedFolder(ctx context.Context, folderId uuid.UUID, page int, pageSize int) ([]model.FolderShare, error)
	CountSharedFolder(ctx context.Context, folderId uuid.UUID) (int64, error)
	CountSharedFoldersWithEmail(ctx context.Context, email string) (int64, error)
	GetSharedFoldersWithEmail(ctx context.Context, sharedWithEmail string, page int, pageSize int) ([]model.FolderShare, error)
	ListSharedFoldersByEmail(ctx context.Context, userEmail string, page int, pageSize int) ([]*model.FolderShare, error)
}

type folderShareRepository struct {
	db *gorm.DB
}

func NewFolderShareRepository(db *gorm.DB) FolderShareRepository {
	return &folderShareRepository{db: db}
}

func (r *folderShareRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.FolderShare, error) {
	var folderShare model.FolderShare
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&folderShare).Error; err != nil {
		return nil, err
	}

	return &folderShare, nil
}

func (r *folderShareRepository) GetByFolderIDAndSharedWithEmail(ctx context.Context, folderID uuid.UUID, sharedWithEmail string) (*model.FolderShare, error) {
	var folderShare model.FolderShare
	if err := r.db.WithContext(ctx).Where("folder_id = ? AND shared_with_email = ? AND is_active = ?", folderID, sharedWithEmail, true).First(&folderShare).Error; err != nil {
		return nil, err
	}

	return &folderShare, nil
}

func (r *folderShareRepository) Create(ctx context.Context, folderShare model.FolderShare) (*model.FolderShare, error) {
	if err := r.db.WithContext(ctx).Create(&folderShare).Error; err != nil {
		return nil, err
	}

	return &folderShare, nil
}

func (r *folderShareRepository) Update(ctx context.Context, id uuid.UUID, folderShare *model.FolderShare) (*model.FolderShare, error) {
	if err := r.db.WithContext(ctx).Model(&model.FolderShare{}).Where("id = ?", id).Updates(folderShare).First(&folderShare).Error; err != nil {
		return nil, err
	}

	return folderShare, nil
}

func (r *folderShareRepository) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	if err := r.db.WithContext(ctx).Model(&model.FolderShare{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return false, err
	}

	return true, nil
}

func (r *folderShareRepository) GetSharedFolder(ctx context.Context, folderId uuid.UUID, page int, pageSize int) ([]model.FolderShare, error) {
	var folderShares []model.FolderShare
	query := r.db.WithContext(ctx).Where("folder_id = ? AND is_active = ?", folderId, true)

	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&folderShares).Error; err != nil {
		return nil, err
	}

	return folderShares, nil
}

func (r *folderShareRepository) GetSharedFoldersWithEmail(ctx context.Context, sharedWithEmail string, page int, pageSize int) ([]model.FolderShare, error) {
	var folderShares []model.FolderShare
	query := r.db.WithContext(ctx).Where("shared_with_email = ? AND is_active = ?", sharedWithEmail, true)

	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&folderShares).Error; err != nil {
		return nil, err
	}

	return folderShares, nil
}

func (r *folderShareRepository) ListSharedFoldersByEmail(ctx context.Context, userEmail string, page int, pageSize int) ([]*model.FolderShare, error) {
	var folderShares []*model.FolderShare

	if err := r.db.WithContext(ctx).
		Preload("Folder").
		Where("folder_shares.shared_with_email = ? AND folder_shares.is_active = ?", userEmail, true).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&folderShares).Error; err != nil {
		return nil, err
	}

	return folderShares, nil
}

func (r *folderShareRepository) CountSharedFolder(ctx context.Context, folderId uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.FolderShare{}).Where("folder_id = ?", folderId).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *folderShareRepository) CountSharedFoldersWithEmail(ctx context.Context, email string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.FolderShare{}).Where("shared_with_email = ?", email).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
