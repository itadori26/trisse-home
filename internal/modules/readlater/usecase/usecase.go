package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"trisse-home/internal/modules/readlater/domain"
	"trisse-home/internal/shared/kernel"
)

// Page is a fetched and sanitised web page (see the gateway).
type Page struct {
	Title, Site, Excerpt, Content string
	Images                        []string
	Words                         int
}

// Image is a stored picture of an article.
type Image struct {
	Bytes       []byte
	ContentType string
}

// PageReader fetches and extracts a page (output port).
type PageReader interface {
	Read(ctx context.Context, url string) (Page, error)
}

// ImageFetcher downloads one image of at most domain.MaxImageBytes (output port).
type ImageFetcher interface {
	Fetch(ctx context.Context, url string) (Image, error)
}

// Repository stores articles and their images (output port).
type Repository interface {
	List(ctx context.Context) ([]domain.Article, error)
	Get(ctx context.Context, id string) (domain.Article, error)
	Save(ctx context.Context, a domain.Article, images map[int]Image) error
	SetStatus(ctx context.Context, id string, s domain.Status) error
	Delete(ctx context.Context, id string) error
	Image(ctx context.Context, id string, n int) (Image, error)
	TotalBytes(ctx context.Context) (int64, error)
}

// ReadLater is the input port of the module.
type ReadLater interface {
	// Save stores a page with its reader copy; saving the same address
	// again returns the existing article.
	Save(ctx context.Context, url string) (domain.Article, error)
	// List returns articles of one status ("" = unread and read), newest first.
	List(ctx context.Context, status domain.Status) ([]domain.Article, error)
	Get(ctx context.Context, id string) (domain.Article, error)
	SetStatus(ctx context.Context, id string, s domain.Status) error
	Delete(ctx context.Context, id string) error
	Image(ctx context.Context, id string, n int) (Image, error)
}

type readLater struct {
	repo   Repository
	pages  PageReader
	images ImageFetcher
	clock  kernel.Clock
	ids    kernel.IDGenerator
}

// New returns the use cases.
func New(repo Repository, pages PageReader, images ImageFetcher, clock kernel.Clock, ids kernel.IDGenerator) ReadLater {
	return &readLater{repo: repo, pages: pages, images: images, clock: clock, ids: ids}
}

// ImagePath is where a stored image is served; it is written into the copy.
func ImagePath(id string, n int) string {
	return "/api/readlater/" + id + "/images/" + strconv.Itoa(n)
}

func (uc *readLater) Save(ctx context.Context, rawURL string) (domain.Article, error) {
	a, err := domain.NewArticle(uc.ids.NewID(), rawURL, uc.clock.Now())
	if err != nil {
		return domain.Article{}, err
	}
	all, err := uc.repo.List(ctx)
	if err != nil {
		return domain.Article{}, err
	}
	if i := slices.IndexFunc(all, func(x domain.Article) bool { return x.URL == a.URL }); i >= 0 {
		return all[i], nil
	}
	used, err := uc.repo.TotalBytes(ctx)
	if err != nil {
		return domain.Article{}, err
	}
	if used >= domain.MaxStorageBytes {
		return domain.Article{}, domain.ErrStorageFull
	}
	images := map[int]Image{}
	page, err := uc.pages.Read(ctx, a.URL)
	if err != nil {
		// The link is kept; the reader copy is missing.
		a.Error = "The page could not be loaded: " + publicReason(err)
	} else {
		stored := map[int]string{}
		var imageBytes int64
		for n, src := range page.Images {
			if len(images) >= domain.MaxImages || imageBytes >= domain.MaxArticleImageBytes {
				break
			}
			img, err := uc.images.Fetch(ctx, src)
			if err != nil || int64(len(img.Bytes))+imageBytes > domain.MaxArticleImageBytes {
				continue
			}
			images[n], stored[n] = img, ImagePath(a.ID, n)
			imageBytes += int64(len(img.Bytes))
		}
		content := domain.ResolveImages(page.Content, stored)
		if len(content) > domain.MaxContentBytes {
			content, images = "", map[int]Image{}
			a.Error = "The page is too long for a reader copy."
		}
		a.ApplyPage(page.Title, page.Site, page.Excerpt, content, page.Words)
		if content == "" && a.Error == "" {
			a.Error = "No article text was found on the page."
		}
		a.Bytes = int64(len(content)) + imageBytes
	}
	if err := uc.repo.Save(ctx, a, images); err != nil {
		return domain.Article{}, fmt.Errorf("save article: %w", err)
	}
	return a, nil
}

func (uc *readLater) List(ctx context.Context, status domain.Status) ([]domain.Article, error) {
	all, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := all[:0:0]
	for _, a := range all {
		if a.Status == status || (status == "" && a.Status != domain.StatusArchived) {
			a.Content = "" // the list does not carry the copies
			out = append(out, a)
		}
	}
	slices.SortStableFunc(out, func(x, y domain.Article) int { return y.SavedAt.Compare(x.SavedAt) })
	return out, nil
}

func (uc *readLater) Get(ctx context.Context, id string) (domain.Article, error) {
	return uc.repo.Get(ctx, id)
}

func (uc *readLater) SetStatus(ctx context.Context, id string, s domain.Status) error {
	if _, ok := domain.ParseStatus(string(s)); !ok {
		return kernel.NewValidationError("status", "The status must be unread, read or archived.")
	}
	if _, err := uc.repo.Get(ctx, id); err != nil {
		return err
	}
	return uc.repo.SetStatus(ctx, id, s)
}

func (uc *readLater) Delete(ctx context.Context, id string) error {
	if _, err := uc.repo.Get(ctx, id); err != nil {
		return err
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *readLater) Image(ctx context.Context, id string, n int) (Image, error) {
	return uc.repo.Image(ctx, id, n)
}

// publicReason keeps public messages and hides technical details.
func publicReason(err error) string {
	var pe *kernel.PublicError
	if errors.As(err, &pe) {
		return pe.Message
	}
	return "the server did not answer as expected"
}
