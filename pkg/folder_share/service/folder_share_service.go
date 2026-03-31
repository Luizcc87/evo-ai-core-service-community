package service

import (
	"context"
	"errors"
	errorsPostgres "evo-ai-core-service/internal/infra/postgres"
	"evo-ai-core-service/internal/utils/contextutils"
	folderService "evo-ai-core-service/pkg/folder/service"
	"evo-ai-core-service/pkg/folder_share/model"
	"evo-ai-core-service/pkg/folder_share/repository"

	"github.com/google/uuid"
)

type FolderShareService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.FolderShare, error)
	Create(ctx context.Context, request model.FolderShare) (*model.FolderShare, error)
	Update(ctx context.Context, id uuid.UUID, request *model.FolderShare) (*model.FolderShare, error)
	Delete(ctx context.Context, id uuid.UUID) (bool, error)
	GetSharedFolder(ctx context.Context, folderId uuid.UUID, page int, pageSize int) (*model.FolderShareListResponse, error)
	GetSharedFoldersWithEmail(ctx context.Context, page int, pageSize int) (*model.FolderShareListResponse, error)
	ListSharedFolders(ctx context.Context, page int, pageSize int) (*model.FolderWithSharingListResponse, error)
	ListSharedFoldersByEmail(ctx context.Context, page int, pageSize int) ([]*model.FolderWithSharingResponse, error)
	CheckFolderAccess(ctx context.Context, folderID uuid.UUID, email string, requiredPermission string) (bool, error)
}

type folderShareService struct {
	folderShareRepository repository.FolderShareRepository
	folderService         folderService.FolderService
}

func NewFolderShareService(folderShareRepository repository.FolderShareRepository, folderService folderService.FolderService) FolderShareService {
	return &folderShareService{
		folderShareRepository: folderShareRepository,
		folderService:         folderService,
	}
}

func (s *folderShareService) GetByID(ctx context.Context, id uuid.UUID) (*model.FolderShare, error) {
	folderShare, err := s.folderShareRepository.GetByID(ctx, id)

	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderShareErrors)
	}

	return folderShare, nil
}

func (s *folderShareService) Create(ctx context.Context, request model.FolderShare) (*model.FolderShare, error) {
	_, err := s.folderService.GetByID(ctx, request.FolderID)

	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderShareErrors)
	}

	email, err := contextutils.GetUserEmail(ctx)
	if err != nil {
		return nil, err
	}

	if request.SharedWithEmail == email {
		return nil, errors.New("You cannot share a folder with yourself")
	}

	userID, err := contextutils.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	existingShare, _ := s.folderShareRepository.GetByFolderIDAndSharedWithEmail(ctx, request.FolderID, request.SharedWithEmail)

	if existingShare != nil {
		return nil, errors.New("Folder is already shared with this user")
	}

	request.SharedByUserID = userID
	request.SharedWithUserID = userID

	folderShare, err := s.folderShareRepository.Create(ctx, request)

	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderShareErrors)
	}

	return folderShare, nil
}

func (s *folderShareService) Update(ctx context.Context, id uuid.UUID, request *model.FolderShare) (*model.FolderShare, error) {
	folderShare, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_, err = s.folderService.GetByID(ctx, folderShare.FolderID)
	if err != nil {
		return nil, errors.New("Share references a folder that no longer exists")
	}

	folderShare, err = s.folderShareRepository.Update(ctx, id, request)

	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderShareErrors)
	}

	return folderShare, nil
}

func (s *folderShareService) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := s.GetByID(ctx, id)

	if err != nil {
		return false, err
	}

	deleted, err := s.folderShareRepository.Delete(ctx, id)

	if err != nil {
		return false, errorsPostgres.MapDBError(err, model.FolderShareErrors)
	}

	return deleted, nil
}

func (s *folderShareService) GetSharedFolder(ctx context.Context, folderId uuid.UUID, page int, pageSize int) (*model.FolderShareListResponse, error) {
	_, err := s.folderService.GetByID(ctx, folderId)
	if err != nil {
		return nil, errors.New("Folder not found")
	}

	sharedFolders, err := s.folderShareRepository.GetSharedFolder(ctx, folderId, page, pageSize)
	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderShareErrors)
	}

	totalItems, err := s.folderShareRepository.CountSharedFolder(ctx, folderId)
	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderShareErrors)
	}

	items := make([]model.FolderShareResponse, len(sharedFolders))
	for i, folderShare := range sharedFolders {
		items[i] = *folderShare.ToResponse()
	}

	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	skip := (page - 1) * pageSize
	limit := pageSize

	return &model.FolderShareListResponse{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Skip:       skip,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
}

func (s *folderShareService) GetSharedFoldersWithEmail(ctx context.Context, page int, pageSize int) (*model.FolderShareListResponse, error) {
	email, err := contextutils.GetUserEmail(ctx)
	if err != nil {
		return nil, err
	}

	folderShares, err := s.folderShareRepository.GetSharedFoldersWithEmail(ctx, email, page, pageSize)
	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderShareErrors)
	}

	totalItems, err := s.folderShareRepository.CountSharedFoldersWithEmail(ctx, email)
	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderShareErrors)
	}

	items := make([]model.FolderShareResponse, len(folderShares))
	for i, folderShare := range folderShares {
		items[i] = *folderShare.ToResponse()
	}

	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	skip := (page - 1) * pageSize
	limit := pageSize

	return &model.FolderShareListResponse{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Skip:       skip,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
}

func (s *folderShareService) ListSharedFolders(ctx context.Context, page int, pageSize int) (*model.FolderWithSharingListResponse, error) {
	email, err := contextutils.GetUserEmail(ctx)
	if err != nil {
		return nil, err
	}

	// Get owned folders
	ownedFolders, err := s.folderService.ListOwnedFolders(ctx, page, pageSize)
	if err != nil {
		return nil, errors.New("Failed to list owned folders")
	}

	// Get shared folders
	sharedFolders, err := s.folderShareRepository.ListSharedFoldersByEmail(ctx, email, page, pageSize)
	if err != nil {
		return nil, err
	}

	sharedFoldersResponse := make([]*model.FolderWithSharingResponse, len(sharedFolders))
	for i, share := range sharedFolders {
		sharedFoldersResponse[i] = &model.FolderWithSharingResponse{
			ID:              share.FolderID,
			Name:            share.Folder.Name,
			Description:     share.Folder.Description,
			CreatedAt:       share.Folder.CreatedAt,
			UpdatedAt:       share.Folder.UpdatedAt,
			IsShared:        true,
			PermissionLevel: share.PermissionLevel,
			SharedBy:        nil,
			ShareID:         &share.ID,
		}
	}

	// Combine results
	allFolders := append(ownedFolders, sharedFoldersResponse...)

	// Convert to response items
	items := make([]model.FolderWithSharingResponse, len(allFolders))
	for i, folder := range allFolders {
		items[i] = *folder
	}

	// For simplicity, we'll return the combined results without precise total count
	// In a real scenario, you might want to implement more sophisticated pagination
	totalItems := int64(len(allFolders))
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	skip := (page - 1) * pageSize
	limit := pageSize

	return &model.FolderWithSharingListResponse{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Skip:       skip,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
}

func (s *folderShareService) ListSharedFoldersByEmail(ctx context.Context, page int, pageSize int) ([]*model.FolderWithSharingResponse, error) {
	email, err := contextutils.GetUserEmail(ctx)
	if err != nil {
		return nil, err
	}

	sharedFolders, err := s.folderShareRepository.ListSharedFoldersByEmail(ctx, email, page, pageSize)
	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.FolderShareErrors)
	}

	result := make([]*model.FolderWithSharingResponse, 0)

	for _, share := range sharedFolders {
		result = append(result, &model.FolderWithSharingResponse{
			ID:              share.FolderID,
			Name:            share.Folder.Name,
			Description:     share.Folder.Description,
			CreatedAt:       share.Folder.CreatedAt,
			UpdatedAt:       share.Folder.UpdatedAt,
			IsShared:        true,
			PermissionLevel: share.PermissionLevel,
			SharedBy:        nil,
			ShareID:         &share.ID,
		})
	}

	return result, nil
}

func (s *folderShareService) CheckFolderAccess(
	ctx context.Context,
	folderID uuid.UUID,
	email string,
	requiredPermission string,
) (bool, error) {
	_, err := s.folderService.GetByID(ctx, folderID)
	if err != nil {
		return false, nil
	}

	share, err := s.folderShareRepository.GetByFolderIDAndSharedWithEmail(ctx, folderID, email)
	if err != nil {
		return false, errorsPostgres.MapDBError(err, model.FolderShareErrors)
	}

	if share == nil {
		return false, nil
	}

	if share.PermissionLevel != requiredPermission {
		return false, nil
	}

	return true, nil
}
