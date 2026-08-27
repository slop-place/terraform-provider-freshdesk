package freshdesk

import "context"

// Forum types.
const (
	ForumTypeHowTo        = 1
	ForumTypeIdea         = 2
	ForumTypeProblem      = 3
	ForumTypeAnnouncement = 4
)

// Forum visibility values.
const (
	ForumVisibilityAll             = 1
	ForumVisibilityLoggedInUsers   = 2
	ForumVisibilityAgents          = 3
	ForumVisibilitySelectCompanies = 4
)

// ForumCategory groups forums in the community portal.
type ForumCategory struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   Time   `json:"created_at"`
	UpdatedAt   Time   `json:"updated_at"`
}

// ForumCategoryRequest is the create/update payload for a forum category.
type ForumCategoryRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Forum is a discussion board within a category.
type Forum struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// ForumType is one of the ForumType* constants.
	ForumType int `json:"forum_type"`
	// ForumVisibility is one of the ForumVisibility* constants.
	ForumVisibility int   `json:"forum_visibility"`
	ForumCategoryID int64 `json:"forum_category_id"`
	// CompanyIDs restricts a forum with ForumVisibilitySelectCompanies.
	CompanyIDs  []int64 `json:"company_ids"`
	TopicsCount int     `json:"topics_count"`
	PostsCount  int     `json:"posts_count"`
	CreatedAt   Time    `json:"created_at"`
	UpdatedAt   Time    `json:"updated_at"`
}

// ForumRequest is the create/update payload for a forum.
type ForumRequest struct {
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	ForumType       *int    `json:"forum_type,omitempty"`
	ForumVisibility *int    `json:"forum_visibility,omitempty"`
	CompanyIDs      []int64 `json:"company_ids,omitempty"`
}

// Topic is a thread within a forum.
type Topic struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Message string `json:"message"`
	ForumID int64  `json:"forum_id"`
	UserID  int64  `json:"user_id"`
	// Sticky pins the topic to the top of the forum.
	Sticky bool `json:"sticky"`
	// Locked prevents further comments.
	Locked bool `json:"locked"`
	// StampType marks an idea or problem topic's state.
	StampType int  `json:"stamp_type,omitempty"`
	RepliedAt Time `json:"replied_at"`
	// RepliedBy is the user who posted the most recent comment.
	RepliedBy int64 `json:"replied_by"`
	// Published reports whether the topic is visible in the portal.
	Published bool `json:"published"`
	// UserVotes counts the votes an idea topic has received.
	UserVotes int `json:"user_votes"`
	// MergedTopicID is set when the topic was merged into another.
	MergedTopicID int64 `json:"merged_topic_id"`
	HitsCount     int   `json:"hits"`
	CommentsCount int   `json:"comments_count"`
	CreatedAt     Time  `json:"created_at"`
	UpdatedAt     Time  `json:"updated_at"`
}

// TopicRequest is the create/update payload for a topic.
type TopicRequest struct {
	Title   *string `json:"title,omitempty"`
	Message *string `json:"message,omitempty"`
	Sticky  *bool   `json:"sticky,omitempty"`
	Locked  *bool   `json:"locked,omitempty"`
	// StampType marks an idea or problem topic's state.
	StampType *int   `json:"stamp_type,omitempty"`
	UserID    *int64 `json:"user_id,omitempty"`
}

// Comment is a reply on a topic.
type Comment struct {
	ID       int64  `json:"id"`
	Body     string `json:"body"`
	BodyText string `json:"body_text"`
	TopicID  int64  `json:"topic_id"`
	UserID   int64  `json:"user_id"`
	// Answer marks the comment that resolved the topic.
	Answer    bool `json:"answer"`
	CreatedAt Time `json:"created_at"`
	UpdatedAt Time `json:"updated_at"`
}

// CommentRequest is the create/update payload for a comment.
type CommentRequest struct {
	Body   *string `json:"body,omitempty"`
	UserID *int64  `json:"user_id,omitempty"`
	Answer *bool   `json:"answer,omitempty"`
}

const forumCategoriesPath = "discussions/categories"

// GetForumCategory fetches a forum category.
func (c *Client) GetForumCategory(ctx context.Context, id int64) (*ForumCategory, error) {
	return getResource[ForumCategory](ctx, c, forumCategoriesPath, id, nil)
}

// ListForumCategories returns every forum category.
func (c *Client) ListForumCategories(ctx context.Context, opts ListOptions) ([]ForumCategory, error) {
	return listAll[ForumCategory](ctx, c, forumCategoriesPath, opts)
}

// CreateForumCategory creates a forum category.
func (c *Client) CreateForumCategory(ctx context.Context, req ForumCategoryRequest) (*ForumCategory, error) {
	return createResource[ForumCategory](ctx, c, forumCategoriesPath, req)
}

// UpdateForumCategory updates a forum category.
func (c *Client) UpdateForumCategory(ctx context.Context, id int64, req ForumCategoryRequest) (*ForumCategory, error) {
	return updateResource[ForumCategory](ctx, c, forumCategoriesPath, id, req)
}

// DeleteForumCategory deletes a forum category and its forums.
func (c *Client) DeleteForumCategory(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, forumCategoriesPath, id)
}

const forumsPath = "discussions/forums"

// GetForum fetches a forum.
func (c *Client) GetForum(ctx context.Context, id int64) (*Forum, error) {
	return getResource[Forum](ctx, c, forumsPath, id, nil)
}

// ListForums returns the forums in a category.
func (c *Client) ListForums(ctx context.Context, categoryID int64, opts ListOptions) ([]Forum, error) {
	return listAll[Forum](ctx, c, pathFor(forumCategoriesPath, categoryID)+"/forums", opts)
}

// CreateForum creates a forum in a category.
func (c *Client) CreateForum(ctx context.Context, categoryID int64, req ForumRequest) (*Forum, error) {
	return createResource[Forum](ctx, c, pathFor(forumCategoriesPath, categoryID)+"/forums", req)
}

// UpdateForum updates a forum.
func (c *Client) UpdateForum(ctx context.Context, id int64, req ForumRequest) (*Forum, error) {
	return updateResource[Forum](ctx, c, forumsPath, id, req)
}

// DeleteForum deletes a forum and its topics.
func (c *Client) DeleteForum(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, forumsPath, id)
}

// FollowForum subscribes a user to a forum.
func (c *Client) FollowForum(ctx context.Context, forumID, userID int64) error {
	return c.Post(ctx, pathFor(forumsPath, forumID)+"/follow",
		map[string]any{"user_id": userID}, nil)
}

// UnfollowForum unsubscribes a user from a forum.
func (c *Client) UnfollowForum(ctx context.Context, forumID, userID int64) error {
	return c.Delete(ctx, pathFor(forumsPath, forumID)+"/follow?user_id="+idString(userID))
}

const topicsPath = "discussions/topics"

// GetTopic fetches a topic.
func (c *Client) GetTopic(ctx context.Context, id int64) (*Topic, error) {
	return getResource[Topic](ctx, c, topicsPath, id, nil)
}

// ListTopics returns the topics in a forum.
func (c *Client) ListTopics(ctx context.Context, forumID int64, opts ListOptions) ([]Topic, error) {
	return listAll[Topic](ctx, c, pathFor(forumsPath, forumID)+"/topics", opts)
}

// CreateTopic creates a topic in a forum.
func (c *Client) CreateTopic(ctx context.Context, forumID int64, req TopicRequest) (*Topic, error) {
	return createResource[Topic](ctx, c, pathFor(forumsPath, forumID)+"/topics", req)
}

// UpdateTopic updates a topic.
func (c *Client) UpdateTopic(ctx context.Context, id int64, req TopicRequest) (*Topic, error) {
	return updateResource[Topic](ctx, c, topicsPath, id, req)
}

// DeleteTopic deletes a topic and its comments.
func (c *Client) DeleteTopic(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, topicsPath, id)
}

// FollowTopic subscribes a user to a topic.
func (c *Client) FollowTopic(ctx context.Context, topicID, userID int64) error {
	return c.Post(ctx, pathFor(topicsPath, topicID)+"/follow",
		map[string]any{"user_id": userID}, nil)
}

// UnfollowTopic unsubscribes a user from a topic.
func (c *Client) UnfollowTopic(ctx context.Context, topicID, userID int64) error {
	return c.Delete(ctx, pathFor(topicsPath, topicID)+"/follow?user_id="+idString(userID))
}

// ListTopicsFollowedBy returns the topics a user follows.
func (c *Client) ListTopicsFollowedBy(ctx context.Context, userID int64, opts ListOptions) ([]Topic, error) {
	v := opts.Values()
	v.Set("user_id", idString(userID))
	opts.Extra = v
	return listAll[Topic](ctx, c, topicsPath+"/followed_by", opts)
}

// ListTopicsParticipatedBy returns the topics a user has commented on.
func (c *Client) ListTopicsParticipatedBy(ctx context.Context, userID int64, opts ListOptions) ([]Topic, error) {
	v := opts.Values()
	v.Set("user_id", idString(userID))
	opts.Extra = v
	return listAll[Topic](ctx, c, topicsPath+"/participated_by", opts)
}

const commentsPath = "discussions/comments"

// ListComments returns the comments on a topic.
func (c *Client) ListComments(ctx context.Context, topicID int64, opts ListOptions) ([]Comment, error) {
	return listAll[Comment](ctx, c, pathFor(topicsPath, topicID)+"/comments", opts)
}

// CreateComment adds a comment to a topic.
func (c *Client) CreateComment(ctx context.Context, topicID int64, req CommentRequest) (*Comment, error) {
	return createResource[Comment](ctx, c, pathFor(topicsPath, topicID)+"/comments", req)
}

// UpdateComment updates a comment.
func (c *Client) UpdateComment(ctx context.Context, id int64, req CommentRequest) (*Comment, error) {
	return updateResource[Comment](ctx, c, commentsPath, id, req)
}

// DeleteComment deletes a comment.
func (c *Client) DeleteComment(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, commentsPath, id)
}
