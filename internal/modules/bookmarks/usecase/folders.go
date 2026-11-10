package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/bookmarks/domain"
)

func folderView(f domain.Folder) FolderView {
	return FolderView{ID: f.ID(), Name: f.Name(), Position: f.Position()}
}

type createFolder struct{ d Deps }

// NewCreateFolder returns the CreateFolder use case.
func NewCreateFolder(d Deps) CreateFolder { return &createFolder{d: d} }

// Execute implements CreateFolder.
func (uc *createFolder) Execute(ctx context.Context, name string) (FolderView, error) {
	existing, err := uc.d.Repo.ListFolders(ctx)
	if err != nil {
		return FolderView{}, fmt.Errorf("list folders: %w", err)
	}
	f, err := domain.NewFolder(uc.d.IDs.NewID(), name, len(existing))
	if err != nil {
		return FolderView{}, err
	}
	if err := uc.d.Repo.SaveFolder(ctx, f); err != nil {
		return FolderView{}, fmt.Errorf("save folder: %w", err)
	}
	return folderView(f), nil
}

type renameFolder struct{ d Deps }

// NewRenameFolder returns the RenameFolder use case.
func NewRenameFolder(d Deps) RenameFolder { return &renameFolder{d: d} }

// Execute implements RenameFolder.
func (uc *renameFolder) Execute(ctx context.Context, id, name string) (FolderView, error) {
	f, err := uc.d.Repo.GetFolder(ctx, id)
	if err != nil {
		return FolderView{}, err
	}
	if err := f.Rename(name); err != nil {
		return FolderView{}, err
	}
	if err := uc.d.Repo.SaveFolder(ctx, f); err != nil {
		return FolderView{}, fmt.Errorf("save folder: %w", err)
	}
	return folderView(f), nil
}

type deleteFolder struct{ d Deps }

// NewDeleteFolder returns the DeleteFolder use case.
func NewDeleteFolder(d Deps) DeleteFolder { return &deleteFolder{d: d} }

// Execute implements DeleteFolder.
func (uc *deleteFolder) Execute(ctx context.Context, id string) error {
	if _, err := uc.d.Repo.GetFolder(ctx, id); err != nil {
		return err
	}
	if err := uc.d.Repo.DeleteFolder(ctx, id); err != nil {
		return fmt.Errorf("delete folder: %w", err)
	}
	return nil
}
