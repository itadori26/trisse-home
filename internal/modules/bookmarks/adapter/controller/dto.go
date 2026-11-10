package controller

// BookmarkInputDTO is the body of POST /api/bookmarks and PUT /api/bookmarks/{id}.
type BookmarkInputDTO struct {
	FolderID    string   `json:"folderId"`
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// FolderInputDTO is the body of POST /api/bookmarks/folders and PATCH /api/bookmarks/folders/{id}.
type FolderInputDTO struct {
	Name string `json:"name"`
}
