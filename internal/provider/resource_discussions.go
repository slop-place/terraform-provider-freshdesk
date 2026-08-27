package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

var (
	_ resource.Resource                = (*forumCategoryResource)(nil)
	_ resource.ResourceWithConfigure   = (*forumCategoryResource)(nil)
	_ resource.ResourceWithImportState = (*forumCategoryResource)(nil)
	_ resource.Resource                = (*forumResource)(nil)
	_ resource.Resource                = (*topicResource)(nil)
	_ resource.Resource                = (*commentResource)(nil)
)

// --- forum categories ----------------------------------------------------

type forumCategoryResource = crud[forumCategoryModel, freshdesk.ForumCategory, *forumCategoryModel]

type forumCategoryModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *forumCategoryModel) GetID() types.String { return m.ID }

// Apply copies an API forum category into the model.
func (m *forumCategoryModel) Apply(c *freshdesk.ForumCategory) {
	m.ID = idString(c.ID)
	m.Name = types.StringValue(c.Name)
	m.Description = optString(c.Description)
	m.CreatedAt = timeString(c.CreatedAt)
	m.UpdatedAt = timeString(c.UpdatedAt)
}

// NewForumCategoryResource returns the freshdesk_forum_category resource.
func NewForumCategoryResource() resource.Resource {
	build := func(plan *forumCategoryModel) freshdesk.ForumCategoryRequest {
		return freshdesk.ForumCategoryRequest{
			Name:        strPtr(plan.Name),
			Description: strPtr(plan.Description),
		}
	}

	return &forumCategoryResource{
		name:  "forum_category",
		label: "forum category",
		schema: schema.Schema{
			MarkdownDescription: "A community forum category, the top level of the discussions tree.",
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
			}, timestampAttributes("category")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *forumCategoryModel, _ *diagnostics,
		) (*freshdesk.ForumCategory, error) {
			return c.CreateForumCategory(ctx, build(plan))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *forumCategoryModel,
		) (*freshdesk.ForumCategory, error) {
			return c.GetForumCategory(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *forumCategoryModel, _ *diagnostics,
		) (*freshdesk.ForumCategory, error) {
			return c.UpdateForumCategory(ctx, id, build(plan))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *forumCategoryModel) error {
			return c.DeleteForumCategory(ctx, id)
		},
	}
}

// --- forums --------------------------------------------------------------

type forumResource = crud[forumModel, freshdesk.Forum, *forumModel]

type forumModel struct {
	ID              types.String `tfsdk:"id"`
	ForumCategoryID types.Int64  `tfsdk:"forum_category_id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	ForumType       types.Int64  `tfsdk:"forum_type"`
	ForumVisibility types.Int64  `tfsdk:"forum_visibility"`
	CompanyIDs      types.Set    `tfsdk:"company_ids"`
	TopicsCount     types.Int64  `tfsdk:"topics_count"`
	PostsCount      types.Int64  `tfsdk:"posts_count"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *forumModel) GetID() types.String { return m.ID }

// Apply copies an API forum into the model.
func (m *forumModel) Apply(f *freshdesk.Forum) {
	m.ID = idString(f.ID)
	m.Name = types.StringValue(f.Name)
	m.Description = optString(f.Description)
	m.ForumType = types.Int64Value(int64(f.ForumType))
	m.ForumVisibility = types.Int64Value(int64(f.ForumVisibility))
	m.TopicsCount = types.Int64Value(int64(f.TopicsCount))
	m.PostsCount = types.Int64Value(int64(f.PostsCount))
	m.CreatedAt = timeString(f.CreatedAt)
	m.UpdatedAt = timeString(f.UpdatedAt)

	if f.ForumCategoryID != 0 {
		m.ForumCategoryID = types.Int64Value(f.ForumCategoryID)
	}

	m.CompanyIDs = applyInt64Set(m.CompanyIDs, f.CompanyIDs)
}

// NewForumResource returns the freshdesk_forum resource.
func NewForumResource() resource.Resource {
	build := func(
		ctx context.Context,
		plan *forumModel,
		d *diagnostics,
		creating bool,
	) freshdesk.ForumRequest {
		req := freshdesk.ForumRequest{
			Name:            strPtr(plan.Name),
			Description:     strPtr(plan.Description),
			ForumVisibility: intPtr(plan.ForumVisibility),
			CompanyIDs:      toInt64Slice(ctx, plan.CompanyIDs, d),
		}

		// Freshdesk refuses a forum_type on a forum that already has topics,
		// even when the value is unchanged, so it is only sent on create.
		if creating {
			req.ForumType = intPtr(plan.ForumType)
		}

		return req
	}

	return &forumResource{
		name:  "forum",
		label: "forum",
		schema: schema.Schema{
			MarkdownDescription: "A community forum inside a forum category.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("forum"),
				"forum_category_id": schema.Int64Attribute{
					Required: true,
					MarkdownDescription: "ID of the parent forum category. Changing it forces a " +
						"new forum.",
					PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				},
				"name": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Name of the forum.",
				},
				"description": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Description of the forum.",
				},
				"forum_type": schema.Int64Attribute{
					Optional:   true,
					Computed:   true,
					Validators: []validatorInt64{forumTypeValidator()},
					MarkdownDescription: "Forum type: `1` how-to, `2` idea, `3` problem, " +
						"`4` announcement.\n\n~> Freshdesk refuses to change the type of a forum " +
						"that already has topics, so changing it forces a new forum.",
					PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				},
				"forum_visibility": schema.Int64Attribute{
					Optional:   true,
					Computed:   true,
					Validators: []validatorInt64{forumVisibilityValidator()},
					MarkdownDescription: "Who may read the forum: `1` everyone, `2` logged-in " +
						"users, `3` agents, `4` selected companies.",
				},
				"company_ids": schema.SetAttribute{
					Optional:    true,
					ElementType: types.Int64Type,
					MarkdownDescription: "IDs of the companies that may read the forum. Only used " +
						"when `forum_visibility` is `4`.",
				},
				"topics_count": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "Number of topics in the forum.",
				},
				"posts_count": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "Number of posts across the forum's topics.",
				},
			}, timestampAttributes("forum")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *forumModel, d *diagnostics,
		) (*freshdesk.Forum, error) {
			return c.CreateForum(ctx, plan.ForumCategoryID.ValueInt64(), build(ctx, plan, d, true))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *forumModel,
		) (*freshdesk.Forum, error) {
			return c.GetForum(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *forumModel, d *diagnostics,
		) (*freshdesk.Forum, error) {
			return c.UpdateForum(ctx, id, build(ctx, plan, d, false))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *forumModel) error {
			return c.DeleteForum(ctx, id)
		},
	}
}

// --- topics --------------------------------------------------------------

type topicResource = crud[topicModel, freshdesk.Topic, *topicModel]

type topicModel struct {
	ID            types.String `tfsdk:"id"`
	ForumID       types.Int64  `tfsdk:"forum_id"`
	Title         types.String `tfsdk:"title"`
	Message       types.String `tfsdk:"message"`
	Sticky        types.Bool   `tfsdk:"sticky"`
	Locked        types.Bool   `tfsdk:"locked"`
	StampType     types.Int64  `tfsdk:"stamp_type"`
	UserID        types.Int64  `tfsdk:"user_id"`
	CommentsCount types.Int64  `tfsdk:"comments_count"`
	Hits          types.Int64  `tfsdk:"hits"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *topicModel) GetID() types.String { return m.ID }

// Apply copies an API topic into the model.
func (m *topicModel) Apply(t *freshdesk.Topic) {
	m.ID = idString(t.ID)
	m.Title = types.StringValue(t.Title)
	m.Sticky = types.BoolValue(t.Sticky)
	m.Locked = types.BoolValue(t.Locked)
	m.UserID = optInt64(t.UserID)
	m.CommentsCount = types.Int64Value(int64(t.CommentsCount))
	m.Hits = types.Int64Value(int64(t.HitsCount))
	m.CreatedAt = timeString(t.CreatedAt)
	m.UpdatedAt = timeString(t.UpdatedAt)

	if t.ForumID != 0 {
		m.ForumID = types.Int64Value(t.ForumID)
	}

	if t.StampType != 0 {
		m.StampType = types.Int64Value(int64(t.StampType))
	}

	// Freshdesk never echoes the opening message back, so the configured value
	// is kept rather than blanked. A message edited in the portal therefore
	// cannot be detected as drift.
	if t.Message != "" {
		m.Message = types.StringValue(t.Message)
	}
}

// NewTopicResource returns the freshdesk_topic resource.
func NewTopicResource() resource.Resource {
	build := func(plan *topicModel) freshdesk.TopicRequest {
		return freshdesk.TopicRequest{
			Title:     strPtr(plan.Title),
			Message:   strPtr(plan.Message),
			Sticky:    boolPtr(plan.Sticky),
			Locked:    boolPtr(plan.Locked),
			StampType: intPtr(plan.StampType),
			UserID:    int64Ptr(plan.UserID),
		}
	}

	return &topicResource{
		name:  "topic",
		label: "topic",
		schema: schema.Schema{
			MarkdownDescription: "A discussion topic inside a community forum.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("topic"),
				"forum_id": schema.Int64Attribute{
					Required:            true,
					MarkdownDescription: "ID of the parent forum. Changing it forces a new topic.",
					PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				},
				"title": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Title of the topic.",
				},
				"message": schema.StringAttribute{
					Required: true,
					MarkdownDescription: "Opening message of the topic, in HTML.\n\n" +
						"~> Freshdesk does not return this field, so a message edited in the " +
						"portal cannot be detected as drift.",
				},
				"sticky": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the topic is pinned to the top of the forum.",
				},
				"locked": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the topic is closed to further comments.",
				},
				"stamp_type": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					MarkdownDescription: "State stamp for idea and problem topics, for example " +
						"planned or solved. The accepted values depend on the forum type.",
				},
				"user_id": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "ID of the user credited as the author.",
				},
				"comments_count": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "Number of comments on the topic.",
				},
				"hits": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "How many times the topic has been viewed.",
				},
			}, timestampAttributes("topic")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *topicModel, _ *diagnostics,
		) (*freshdesk.Topic, error) {
			return c.CreateTopic(ctx, plan.ForumID.ValueInt64(), build(plan))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, _ *topicModel,
		) (*freshdesk.Topic, error) {
			return c.GetTopic(ctx, id)
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *topicModel, _ *diagnostics,
		) (*freshdesk.Topic, error) {
			return c.UpdateTopic(ctx, id, build(plan))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *topicModel) error {
			return c.DeleteTopic(ctx, id)
		},
	}
}

// --- comments ------------------------------------------------------------

type commentResource = crud[commentModel, freshdesk.Comment, *commentModel]

type commentModel struct {
	ID        types.String `tfsdk:"id"`
	TopicID   types.Int64  `tfsdk:"topic_id"`
	Body      types.String `tfsdk:"body"`
	UserID    types.Int64  `tfsdk:"user_id"`
	Answer    types.Bool   `tfsdk:"answer"`
	BodyText  types.String `tfsdk:"body_text"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// GetID reports the identifier held in state.
func (m *commentModel) GetID() types.String { return m.ID }

// Apply copies an API comment into the model.
func (m *commentModel) Apply(c *freshdesk.Comment) {
	m.ID = idString(c.ID)
	m.Body = types.StringValue(c.Body)
	m.BodyText = optString(c.BodyText)
	m.UserID = optInt64(c.UserID)
	m.Answer = types.BoolValue(c.Answer)
	m.CreatedAt = timeString(c.CreatedAt)
	m.UpdatedAt = timeString(c.UpdatedAt)

	if c.TopicID != 0 {
		m.TopicID = types.Int64Value(c.TopicID)
	}
}

// NewCommentResource returns the freshdesk_comment resource.
func NewCommentResource() resource.Resource {
	build := func(plan *commentModel) freshdesk.CommentRequest {
		return freshdesk.CommentRequest{
			Body:   strPtr(plan.Body),
			UserID: int64Ptr(plan.UserID),
			Answer: boolPtr(plan.Answer),
		}
	}

	return &commentResource{
		name:  "comment",
		label: "comment",
		schema: schema.Schema{
			MarkdownDescription: "A comment on a community forum topic.",
			Attributes: withAttributes(map[string]schema.Attribute{
				"id": idAttribute("comment"),
				"topic_id": schema.Int64Attribute{
					Required:            true,
					MarkdownDescription: "ID of the topic. Changing it forces a new comment.",
					PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				},
				"body": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Body of the comment, in HTML.",
				},
				"user_id": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "ID of the user credited as the author.",
				},
				"answer": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether the comment is marked as the topic's answer.",
				},
				"body_text": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Plain-text rendering of `body`.",
				},
			}, timestampAttributes("comment")),
		},
		createFn: func(
			ctx context.Context, c *freshdesk.Client, plan *commentModel, _ *diagnostics,
		) (*freshdesk.Comment, error) {
			return c.CreateComment(ctx, plan.TopicID.ValueInt64(), build(plan))
		},
		readFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, state *commentModel,
		) (*freshdesk.Comment, error) {
			// Freshdesk exposes no single-comment read, so the comment is
			// located by walking the comments of the topic held in state.
			comments, err := c.ListComments(ctx, state.TopicID.ValueInt64(), freshdesk.ListOptions{})
			if err != nil {
				return nil, fmt.Errorf("listing the topic's comments: %w", err)
			}

			for i := range comments {
				if comments[i].ID == id {
					return &comments[i], nil
				}
			}

			return nil, &freshdesk.Error{
				StatusCode: http.StatusNotFound,
				Code:       "not_found",
				Message:    "comment " + strconv.FormatInt(id, 10) + " is no longer on its topic",
			}
		},
		updateFn: func(
			ctx context.Context, c *freshdesk.Client, id int64, plan, _ *commentModel, _ *diagnostics,
		) (*freshdesk.Comment, error) {
			return c.UpdateComment(ctx, id, build(plan))
		},
		deleteFn: func(ctx context.Context, c *freshdesk.Client, id int64, _ *commentModel) error {
			return c.DeleteComment(ctx, id)
		},
	}
}
