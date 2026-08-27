package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

var (
	_ resource.Resource                = (*solutionCategoryResource)(nil)
	_ resource.ResourceWithConfigure   = (*solutionCategoryResource)(nil)
	_ resource.ResourceWithImportState = (*solutionCategoryResource)(nil)
	_ resource.Resource                = (*solutionFolderResource)(nil)
	_ resource.Resource                = (*solutionArticleResource)(nil)
)

// --- categories ----------------------------------------------------------

type solutionCategoryResource = crud[solutionCategoryModel, freshdesk.SolutionCategory, *solutionCategoryModel]

type solutionCategoryModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	VisibleInPortals types.Set    `tfsdk:"visible_in_portals"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *solutionCategoryModel) GetID() types.String { return m.ID }

// Apply copies an API solution category into the model.
func (m *solutionCategoryModel) Apply(c *freshdesk.SolutionCategory) {
	m.ID = idString(c.ID)
	m.Name = types.StringValue(c.Name)
	m.Description = optString(c.Description)
	m.CreatedAt = timeString(c.CreatedAt)
	m.UpdatedAt = timeString(c.UpdatedAt)

	m.VisibleInPortals = applyInt64Set(m.VisibleInPortals, c.VisibleInPortals)
}

// NewSolutionCategoryResource returns the freshdesk_solution_category resource.
func NewSolutionCategoryResource() resource.Resource {
	build := func(
		ctx context.Context,
		plan, prior *solutionCategoryModel,
		d *diagnostics,
	) freshdesk.SolutionCategoryRequest {
		priorDescription := types.StringNull()
		if prior != nil {
			priorDescription = prior.Description
		}

		return freshdesk.SolutionCategoryRequest{
			Name:             strPtr(plan.Name),
			Description:      clearableString(plan.Description, priorDescription),
			VisibleInPortals: toInt64Slice(ctx, plan.VisibleInPortals, d),
		}
	}

	return &solutionCategoryResource{
		name:  "solution_category",
		label: "solution category",
		schema: schema.Schema{
			MarkdownDescription: "A knowledge-base category, the top level of the solutions tree.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("category"),
				"name": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Name of the category.",
				},
				"description": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Description of the category.",
				},
				"visible_in_portals": schema.SetAttribute{
					Optional:    true,
					ElementType: types.Int64Type,
					MarkdownDescription: "IDs of the portals the category appears in. Only " +
						"meaningful on accounts with more than one portal.",
				},
			}, timestampAttributes("category")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *solutionCategoryModel, d *diagnostics,
		) (*freshdesk.SolutionCategory, error) {
			return c.CreateSolutionCategory(ctx, build(ctx, plan, nil, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *solutionCategoryModel,
		) (*freshdesk.SolutionCategory, error) {
			return c.GetSolutionCategory(ctx, id, "")
		},
		updateFn: func(
			ctx context.Context,
			c *freshdesk.Client,
			id int64,
			plan, prior *solutionCategoryModel,
			d *diagnostics,
		) (*freshdesk.SolutionCategory, error) {
			return c.UpdateSolutionCategory(ctx, id, "", build(ctx, plan, prior, d))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *solutionCategoryModel) error {
			return c.DeleteSolutionCategory(ctx, id)
		},
	}
}

// --- folders -------------------------------------------------------------

type solutionFolderResource = crud[solutionFolderModel, freshdesk.SolutionFolder, *solutionFolderModel]

type solutionFolderModel struct {
	ID             types.String `tfsdk:"id"`
	CategoryID     types.Int64  `tfsdk:"category_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Visibility     types.Int64  `tfsdk:"visibility"`
	CompanyIDs     types.Set    `tfsdk:"company_ids"`
	ParentFolderID types.Int64  `tfsdk:"parent_folder_id"`
	ArticlesCount  types.Int64  `tfsdk:"articles_count"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *solutionFolderModel) GetID() types.String { return m.ID }

// Apply copies an API solution folder into the model.
func (m *solutionFolderModel) Apply(f *freshdesk.SolutionFolder) {
	m.ID = idString(f.ID)
	m.Name = types.StringValue(f.Name)
	m.Description = optString(f.Description)
	m.Visibility = types.Int64Value(int64(f.Visibility))
	m.ParentFolderID = optInt64(f.ParentFolderID)
	m.ArticlesCount = types.Int64Value(int64(f.ArticlesCount))
	m.CreatedAt = timeString(f.CreatedAt)
	m.UpdatedAt = timeString(f.UpdatedAt)

	// The API omits category_id on some reads; keep the configured parent.
	if f.CategoryID != 0 {
		m.CategoryID = types.Int64Value(f.CategoryID)
	}

	m.CompanyIDs = applyInt64Set(m.CompanyIDs, f.CompanyIDs)
}

// NewSolutionFolderResource returns the freshdesk_solution_folder resource.
func NewSolutionFolderResource() resource.Resource {
	build := func(
		ctx context.Context,
		plan, prior *solutionFolderModel,
		d *diagnostics,
	) freshdesk.SolutionFolderRequest {
		priorDescription := types.StringNull()
		if prior != nil {
			priorDescription = prior.Description
		}

		return freshdesk.SolutionFolderRequest{
			Name:           strPtr(plan.Name),
			Description:    clearableString(plan.Description, priorDescription),
			Visibility:     intPtr(plan.Visibility),
			CompanyIDs:     toInt64Slice(ctx, plan.CompanyIDs, d),
			ParentFolderID: int64Ptr(plan.ParentFolderID),
		}
	}

	return &solutionFolderResource{
		name:  "solution_folder",
		label: "solution folder",
		schema: schema.Schema{
			MarkdownDescription: "A knowledge-base folder, which holds articles inside a category.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("folder"),
				"category_id": schema.Int64Attribute{
					Required: true,
					MarkdownDescription: "ID of the parent category. Changing it forces a new " +
						"folder, because Freshdesk cannot move a folder between categories.",
					PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				},
				"name": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Name of the folder.",
				},
				"description": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Description of the folder.",
				},
				"visibility": schema.Int64Attribute{
					Optional:   true,
					Computed:   true,
					Validators: []validatorInt64{folderVisibilityValidator()},
					MarkdownDescription: "Who may read the folder: `1` all users, `2` logged-in " +
						"users, `3` agents, `4` selected companies, `5` bots.",
				},
				"company_ids": schema.SetAttribute{
					Optional:    true,
					ElementType: types.Int64Type,
					MarkdownDescription: "IDs of the companies that may read the folder. Only used " +
						"when `visibility` is `4`.",
				},
				"parent_folder_id": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "ID of the parent folder, making this one a subfolder.",
				},
				"articles_count": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "Number of articles in the folder.",
				},
			}, timestampAttributes("folder")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *solutionFolderModel, d *diagnostics,
		) (*freshdesk.SolutionFolder, error) {
			return c.CreateSolutionFolder(ctx, plan.CategoryID.ValueInt64(), build(ctx, plan, nil, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *solutionFolderModel,
		) (*freshdesk.SolutionFolder, error) {
			return c.GetSolutionFolder(ctx, id, "")
		},
		updateFn: func(
			ctx context.Context,
			c *freshdesk.Client,
			id int64,
			plan, prior *solutionFolderModel,
			d *diagnostics,
		) (*freshdesk.SolutionFolder, error) {
			return c.UpdateSolutionFolder(ctx, id, "", build(ctx, plan, prior, d))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *solutionFolderModel) error {
			return c.DeleteSolutionFolder(ctx, id)
		},
	}
}

// --- articles ------------------------------------------------------------

type solutionArticleResource = crud[solutionArticleModel, freshdesk.SolutionArticle, *solutionArticleModel]

type solutionArticleModel struct {
	ID              types.String `tfsdk:"id"`
	FolderID        types.Int64  `tfsdk:"folder_id"`
	Title           types.String `tfsdk:"title"`
	Description     types.String `tfsdk:"description"`
	Status          types.Int64  `tfsdk:"status"`
	Type            types.Int64  `tfsdk:"type"`
	AgentID         types.Int64  `tfsdk:"agent_id"`
	Tags            types.Set    `tfsdk:"tags"`
	MetaTitle       types.String `tfsdk:"meta_title"`
	MetaDescription types.String `tfsdk:"meta_description"`
	MetaKeywords    types.Set    `tfsdk:"meta_keywords"`

	CategoryID      types.Int64  `tfsdk:"category_id"`
	DescriptionText types.String `tfsdk:"description_text"`
	Hits            types.Int64  `tfsdk:"hits"`
	ThumbsUp        types.Int64  `tfsdk:"thumbs_up"`
	ThumbsDown      types.Int64  `tfsdk:"thumbs_down"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *solutionArticleModel) GetID() types.String { return m.ID }

// Apply copies an API solution article into the model.
func (m *solutionArticleModel) Apply(a *freshdesk.SolutionArticle) {
	m.ID = idString(a.ID)
	m.Title = types.StringValue(a.Title)
	m.Description = optString(a.Description)
	m.Status = types.Int64Value(int64(a.Status))
	m.AgentID = optInt64(a.AgentID)
	m.CategoryID = optInt64(a.CategoryID)
	m.DescriptionText = optString(a.DescriptionText)
	m.Hits = types.Int64Value(int64(a.Hits))
	m.ThumbsUp = types.Int64Value(int64(a.ThumbsUp))
	m.ThumbsDown = types.Int64Value(int64(a.ThumbsDown))
	m.CreatedAt = timeString(a.CreatedAt)
	m.UpdatedAt = timeString(a.UpdatedAt)

	if a.Type != 0 {
		m.Type = types.Int64Value(int64(a.Type))
	}

	if a.FolderID != 0 {
		m.FolderID = types.Int64Value(a.FolderID)
	}

	m.Tags = applyStringSet(m.Tags, a.Tags)

	if a.SEOData != nil {
		m.MetaTitle = optString(a.SEOData.MetaTitle)
		m.MetaDescription = optString(a.SEOData.MetaDescription)

		if a.SEOData.MetaKeywords != nil {
			m.MetaKeywords = stringSet(a.SEOData.MetaKeywords)
		}
	}
}

// NewSolutionArticleResource returns the freshdesk_solution_article resource.
func NewSolutionArticleResource() resource.Resource {
	build := func(
		ctx context.Context,
		plan, prior *solutionArticleModel,
		d *diagnostics,
	) freshdesk.SolutionArticleRequest {
		req := freshdesk.SolutionArticleRequest{
			Title:       strPtr(plan.Title),
			Description: strPtr(plan.Description),
			Status:      intPtr(plan.Status),
			Type:        intPtr(plan.Type),
			AgentID:     int64Ptr(plan.AgentID),
			Tags:        toStringSlice(ctx, plan.Tags, d),
		}

		if prior != nil {
			req.ClearTags = plan.Tags.IsNull() && !prior.Tags.IsNull()
		}

		keywords := toStringSlice(ctx, plan.MetaKeywords, d)
		if !plan.MetaTitle.IsNull() || !plan.MetaDescription.IsNull() || keywords != nil {
			req.SEOData = &freshdesk.SEOData{
				MetaTitle:       plan.MetaTitle.ValueString(),
				MetaDescription: plan.MetaDescription.ValueString(),
				MetaKeywords:    keywords,
			}
		}

		return req
	}

	return &solutionArticleResource{
		name:  "solution_article",
		label: "solution article",
		schema: schema.Schema{
			MarkdownDescription: "A knowledge-base article.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("article"),
				"folder_id": schema.Int64Attribute{
					Required: true,
					MarkdownDescription: "ID of the folder holding the article. Changing it forces " +
						"a new article, because Freshdesk cannot move an article between folders.",
					PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				},
				"title": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Title of the article.",
				},
				"description": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Body of the article, in HTML.",
				},
				"status": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					Validators:          []validatorInt64{articleStatusValidator()},
					MarkdownDescription: "Publication state: `1` draft, `2` published.",
				},
				"type": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "Article type: `1` permanent, `2` workaround. Freshdesk " +
						"defaults new articles to `1`.",
				},
				"agent_id": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "ID of the agent credited as the author.",
				},
				"tags": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Tags applied to the article.",
				},
				"meta_title": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "SEO title of the article.",
				},
				"meta_description": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "SEO description of the article.",
				},
				"meta_keywords": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "SEO keywords for the article.",
				},
				"category_id": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "ID of the category the article's folder belongs to.",
				},
				"description_text": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Plain-text rendering of `description`.",
				},
				"hits": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "How many times the article has been viewed.",
				},
				"thumbs_up": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "How many readers marked the article helpful.",
				},
				"thumbs_down": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "How many readers marked the article unhelpful.",
				},
			}, timestampAttributes("article")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *solutionArticleModel, d *diagnostics,
		) (*freshdesk.SolutionArticle, error) {
			return c.CreateSolutionArticle(ctx, plan.FolderID.ValueInt64(), build(ctx, plan, nil, d))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *solutionArticleModel,
		) (*freshdesk.SolutionArticle, error) {
			return c.GetSolutionArticle(ctx, id, "")
		},
		updateFn: func(
			ctx context.Context,
			c *freshdesk.Client,
			id int64,
			plan, prior *solutionArticleModel,
			d *diagnostics,
		) (*freshdesk.SolutionArticle, error) {
			return c.UpdateSolutionArticle(ctx, id, "", build(ctx, plan, prior, d))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *solutionArticleModel) error {
			return c.DeleteSolutionArticle(ctx, id)
		},
	}
}
