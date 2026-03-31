package service

import (
	"context"
	"errors"
	errorsPostgres "evo-ai-core-service/internal/infra/postgres"
	folderShareModel "evo-ai-core-service/pkg/folder_share/model"

	"evo-ai-core-service/pkg/folder/model"
	"evo-ai-core-service/pkg/folder/repository"

	"github.com/google/uuid"
)

type FolderService interface {
	Create(ctx context.Context, request model.Folder) (*model.Folder, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Folder, error)
	List(ctx context.Context, page int, pageSize int) (*model.FolderListResponse, error)
	Update(ctx context.Context, request *model.Folder, id uuid.UUID) (*model.Folder, error)
	Delete(ctx context.Context, id uuid.UUID) (bool, error)
	ListOwnedFolders(ctx context.Context, page int, pageSize int) ([]*folderShareModel.FolderWithSharingResponse, error)
}

type folderService struct {
	folderRepository repository.FolderRepository
}

func NewFolderService(folderRepository repository.FolderRepository) FolderService {
	return &folderService{
		folderRepository: folderRepository,
	}
}

func (s *folderService) Create(ctx context.Context, request model.Folder) (*model.Folder, error) {
	folder, err := s.folderRepository.Create(ctx, request)

	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderErrors)
	}

	return folder, nil
}

func (s *folderService) GetByID(ctx context.Context, id uuid.UUID) (*model.Folder, error) {
	folder, err := s.folderRepository.GetByID(ctx, id)

	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderErrors)
	}

	return folder, nil
}

func (s *folderService) List(ctx context.Context, page int, pageSize int) (*model.FolderListResponse, error) {
	// Get paginated items
	folders, err := s.folderRepository.List(ctx, page, pageSize)
	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderErrors)
	}

	// Get total count
	totalItems, err := s.folderRepository.Count(ctx)
	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderErrors)
	}

	// Convert to response items
	items := make([]model.FolderResponse, len(folders))
	for i, folder := range folders {
		items[i] = *folder.ToResponse()
	}

	// Calculate pagination metadata
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	skip := (page - 1) * pageSize
	limit := pageSize

	return &model.FolderListResponse{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Skip:       skip,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
}

func (s *folderService) Update(ctx context.Context, request *model.Folder, id uuid.UUID) (*model.Folder, error) {
	_, err := s.GetByID(ctx, id)

	if err != nil {
		return nil, errors.New("Folder not found")
	}

	folder, err := s.folderRepository.Update(ctx, request, id)

	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderErrors)
	}

	return folder, nil
}

func (s *folderService) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := s.GetByID(ctx, id)

	if err != nil {
		return false, errors.New("Folder not found")
	}

	deleted, err := s.folderRepository.Delete(ctx, id)

	if err != nil {
		return false, errorsPostgres.MapDBError(err, model.FolderErrors)
	}

	return deleted, nil
}
func (s *folderService) ListOwnedFolders(ctx context.Context, page int, pageSize int) ([]*folderShareModel.FolderWithSharingResponse, error) {
	// Use repository directly to get the folder entities
	ownedFolders, err := s.List(ctx, page, pageSize)
	if err != nil {
		return nil, errors.New("Failed to list folders")
	}

	result := make([]*folderShareModel.FolderWithSharingResponse, 0)

	for _, folder := range ownedFolders.Items {
		result = append(result, &folderShareModel.FolderWithSharingResponse{
			ID:              folder.ID,
			Name:            folder.Name,
			Description:     folder.Description,
			CreatedAt:       folder.CreatedAt,
			UpdatedAt:       folder.UpdatedAt,
			IsShared:        false,
			PermissionLevel: "write",
			SharedBy:        nil,
			ShareID:         nil,
		})
	}

	return result, nil
}
