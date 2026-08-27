package freshdesk

import (
	"context"
	"net/url"
)

// Article statuses.
const (
	ArticleStatusDraft     = 1
	ArticleStatusPublished = 2
)

// Folder visibility values.
const (
	FolderVisibilityAllUsers        = 1
	FolderVisibilityLoggedInUsers   = 2
	FolderVisibilityAgents          = 3
	FolderVisibilitySelectCompanies = 4
	FolderVisibilityBots            = 5
)

// SolutionCategory groups solution folders.
type SolutionCategory struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// VisibleInPortals lists the portal IDs the category appears in.
	VisibleInPortals []int64 `json:"visible_in_portals"`
	// LanguageID and Language identify a translated copy of the category.
	LanguageID int64  `json:"language_id,omitempty"`
	Language   string `json:"language,omitempty"`
	CreatedAt  Time   `json:"created_at"`
	UpdatedAt  Time   `json:"updated_at"`
}

// SolutionCategoryRequest is the create/update payload for a category.
type SolutionCategoryRequest struct {
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	VisibleInPortals []int64 `json:"visible_in_portals,omitempty"`
}

// SolutionFolder groups solution articles.
type SolutionFolder struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Visibility controls who may read the folder; see FolderVisibility*.
	Visibility int `json:"visibility"`
	// CompanyIDs restricts a folder with FolderVisibilitySelectCompanies.
	CompanyIDs []int64 `json:"company_ids"`
	// CategoryID is the parent category.
	CategoryID int64 `json:"category_id"`
	// ParentFolderID is set for a subfolder.
	ParentFolderID  int64 `json:"parent_folder_id,omitempty"`
	ArticlesCount   int   `json:"articles_count,omitempty"`
	SubFoldersCount int   `json:"sub_folders_count,omitempty"`

	LanguageID int64  `json:"language_id,omitempty"`
	Language   string `json:"language,omitempty"`
	CreatedAt  Time   `json:"created_at"`
	UpdatedAt  Time   `json:"updated_at"`
}

// SolutionFolderRequest is the create/update payload for a folder.
type SolutionFolderRequest struct {
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	Visibility     *int    `json:"visibility,omitempty"`
	CompanyIDs     []int64 `json:"company_ids,omitempty"`
	ParentFolderID *int64  `json:"parent_folder_id,omitempty"`
}

// SolutionArticle is a knowledge-base article.
type SolutionArticle struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// DescriptionText is the plain-text rendering of Description.
	DescriptionText string `json:"description_text"`
	// Status is ArticleStatusDraft or ArticleStatusPublished.
	Status int `json:"status"`
	// Type distinguishes permanent (1) from workaround (2) articles.
	Type       int      `json:"type,omitempty"`
	AgentID    int64    `json:"agent_id"`
	CategoryID int64    `json:"category_id"`
	FolderID   int64    `json:"folder_id"`
	Tags       []string `json:"tags"`
	SEOData    *SEOData `json:"seo_data,omitempty"`

	// Hits, ThumbsUp and ThumbsDown are read-only engagement counters.
	Hits       int `json:"hits"`
	ThumbsUp   int `json:"thumbs_up"`
	ThumbsDown int `json:"thumbs_down"`

	Attachments []Attachment `json:"attachments"`

	LanguageID int64  `json:"language_id,omitempty"`
	Language   string `json:"language,omitempty"`
	CreatedAt  Time   `json:"created_at"`
	UpdatedAt  Time   `json:"updated_at"`
}

// SEOData holds an article's search-engine metadata.
type SEOData struct {
	MetaTitle       string   `json:"meta_title,omitempty"`
	MetaDescription string   `json:"meta_description,omitempty"`
	MetaKeywords    []string `json:"meta_keywords,omitempty"`
}

// SolutionArticleRequest is the create/update payload for an article.
type SolutionArticleRequest struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Status      *int     `json:"status,omitempty"`
	Type        *int     `json:"type,omitempty"`
	AgentID     *int64   `json:"agent_id,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	SEOData     *SEOData `json:"seo_data,omitempty"`

	// ClearTags sends an empty tags array, removing every tag.
	ClearTags bool `json:"-"`
}

// MarshalJSON applies the empty-array clearing semantics.
func (r SolutionArticleRequest) MarshalJSON() ([]byte, error) {
	type alias SolutionArticleRequest
	m, err := structToMap(alias(r))
	if err != nil {
		return nil, err
	}

	if r.ClearTags {
		m["tags"] = []string{}
	}

	return marshalMap(m)
}

// langSuffix appends a language code path segment when one is given.
func langSuffix(path, lang string) string {
	if lang == "" {
		return path
	}
	return path + "/" + url.PathEscape(lang)
}

const solutionCategoriesPath = "solutions/categories"

// GetSolutionCategory fetches a category. Pass lang to read a translation.
func (c *Client) GetSolutionCategory(ctx context.Context, id int64, lang string) (*SolutionCategory, error) {
	var out SolutionCategory
	if err := c.Get(ctx, langSuffix(pathFor(solutionCategoriesPath, id), lang), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSolutionCategories returns every category, optionally in one language.
func (c *Client) ListSolutionCategories(
	ctx context.Context,
	lang string,
	opts ListOptions,
) ([]SolutionCategory, error) {
	return listAll[SolutionCategory](ctx, c, langSuffix(solutionCategoriesPath, lang), opts)
}

// CreateSolutionCategory creates a category.
func (c *Client) CreateSolutionCategory(ctx context.Context, req SolutionCategoryRequest) (*SolutionCategory, error) {
	return createResource[SolutionCategory](ctx, c, solutionCategoriesPath, req)
}

// CreateSolutionCategoryTranslation creates a translated copy of a category.
func (c *Client) CreateSolutionCategoryTranslation(
	ctx context.Context,
	id int64,
	lang string,
	req SolutionCategoryRequest,
) (*SolutionCategory, error) {
	var out SolutionCategory
	if err := c.Post(ctx, langSuffix(pathFor(solutionCategoriesPath, id), lang), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateSolutionCategory updates a category, or its translation when lang is set.
func (c *Client) UpdateSolutionCategory(
	ctx context.Context,
	id int64,
	lang string,
	req SolutionCategoryRequest,
) (*SolutionCategory, error) {
	var out SolutionCategory
	if err := c.Put(ctx, langSuffix(pathFor(solutionCategoriesPath, id), lang), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSolutionCategory deletes a category and everything under it.
func (c *Client) DeleteSolutionCategory(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, solutionCategoriesPath, id)
}

const solutionFoldersPath = "solutions/folders"

// GetSolutionFolder fetches a folder. Pass lang to read a translation.
func (c *Client) GetSolutionFolder(ctx context.Context, id int64, lang string) (*SolutionFolder, error) {
	var out SolutionFolder
	if err := c.Get(ctx, langSuffix(pathFor(solutionFoldersPath, id), lang), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSolutionFolders returns the folders in a category.
func (c *Client) ListSolutionFolders(
	ctx context.Context,
	categoryID int64,
	lang string,
	opts ListOptions,
) ([]SolutionFolder, error) {
	path := langSuffix(pathFor(solutionCategoriesPath, categoryID)+"/folders", lang)
	return listAll[SolutionFolder](ctx, c, path, opts)
}

// ListSolutionSubFolders returns the subfolders of a folder.
func (c *Client) ListSolutionSubFolders(
	ctx context.Context,
	folderID int64,
	lang string,
	opts ListOptions,
) ([]SolutionFolder, error) {
	path := langSuffix(pathFor(solutionFoldersPath, folderID)+"/subfolders", lang)
	return listAll[SolutionFolder](ctx, c, path, opts)
}

// CreateSolutionFolder creates a folder in a category.
func (c *Client) CreateSolutionFolder(
	ctx context.Context,
	categoryID int64,
	req SolutionFolderRequest,
) (*SolutionFolder, error) {
	return createResource[SolutionFolder](ctx, c, pathFor(solutionCategoriesPath, categoryID)+"/folders", req)
}

// CreateSolutionFolderTranslation creates a translated copy of a folder.
func (c *Client) CreateSolutionFolderTranslation(
	ctx context.Context,
	id int64,
	lang string,
	req SolutionFolderRequest,
) (*SolutionFolder, error) {
	var out SolutionFolder
	if err := c.Post(ctx, langSuffix(pathFor(solutionFoldersPath, id), lang), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateSolutionFolder updates a folder, or its translation when lang is set.
func (c *Client) UpdateSolutionFolder(
	ctx context.Context,
	id int64,
	lang string,
	req SolutionFolderRequest,
) (*SolutionFolder, error) {
	var out SolutionFolder
	if err := c.Put(ctx, langSuffix(pathFor(solutionFoldersPath, id), lang), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSolutionFolder deletes a folder and its articles.
func (c *Client) DeleteSolutionFolder(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, solutionFoldersPath, id)
}

const solutionArticlesPath = "solutions/articles"

// GetSolutionArticle fetches an article. Pass lang to read a translation.
func (c *Client) GetSolutionArticle(ctx context.Context, id int64, lang string) (*SolutionArticle, error) {
	var out SolutionArticle
	if err := c.Get(ctx, langSuffix(pathFor(solutionArticlesPath, id), lang), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSolutionArticles returns the articles in a folder.
func (c *Client) ListSolutionArticles(
	ctx context.Context,
	folderID int64,
	lang string,
	opts ListOptions,
) ([]SolutionArticle, error) {
	path := langSuffix(pathFor(solutionFoldersPath, folderID)+"/articles", lang)
	return listAll[SolutionArticle](ctx, c, path, opts)
}

// CreateSolutionArticle creates an article in a folder.
func (c *Client) CreateSolutionArticle(
	ctx context.Context,
	folderID int64,
	req SolutionArticleRequest,
) (*SolutionArticle, error) {
	return createResource[SolutionArticle](ctx, c, pathFor(solutionFoldersPath, folderID)+"/articles", req)
}

// CreateSolutionArticleTranslation creates a translated copy of an article.
func (c *Client) CreateSolutionArticleTranslation(
	ctx context.Context,
	id int64,
	lang string,
	req SolutionArticleRequest,
) (*SolutionArticle, error) {
	var out SolutionArticle
	if err := c.Post(ctx, langSuffix(pathFor(solutionArticlesPath, id), lang), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateSolutionArticle updates an article, or its translation when lang is set.
func (c *Client) UpdateSolutionArticle(
	ctx context.Context,
	id int64,
	lang string,
	req SolutionArticleRequest,
) (*SolutionArticle, error) {
	var out SolutionArticle
	if err := c.Put(ctx, langSuffix(pathFor(solutionArticlesPath, id), lang), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSolutionArticle deletes an article.
func (c *Client) DeleteSolutionArticle(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, solutionArticlesPath, id)
}

// SearchSolutions runs a full-text search over published articles.
func (c *Client) SearchSolutions(ctx context.Context, term string, opts ListOptions) ([]SolutionArticle, error) {
	v := opts.Values()
	v.Set("term", term)
	opts.Extra = v
	return listAll[SolutionArticle](ctx, c, "search/solutions", opts)
}
