package freshdesk

import "context"

// Canned response visibility values.
const (
	CannedResponseVisibilityAllAgents = 0
	CannedResponseVisibilityPersonal  = 1
	CannedResponseVisibilityGroups    = 2
)

// CannedResponseFolder groups canned responses.
type CannedResponseFolder struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Personal marks the agent's private folder.
	Personal bool `json:"personal"`
	// ResponsesCount is the number of responses in the folder.
	ResponsesCount int  `json:"responses_count"`
	CreatedAt      Time `json:"created_at"`
	UpdatedAt      Time `json:"updated_at"`
}

// CannedResponseFolderRequest is the create/update payload for a folder.
type CannedResponseFolderRequest struct {
	Name *string `json:"name,omitempty"`
}

// CannedResponse is a reusable reply template.
type CannedResponse struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	// ContentHTML is the rich-text form of Content.
	ContentHTML string `json:"content_html"`
	FolderID    int64  `json:"folder_id"`
	// Visibility is one of the CannedResponseVisibility* constants.
	Visibility int `json:"visibility"`
	// GroupIDs scopes a response with CannedResponseVisibilityGroups.
	GroupIDs    []int64      `json:"group_ids"`
	Attachments []Attachment `json:"attachments"`
	CreatedAt   Time         `json:"created_at"`
	UpdatedAt   Time         `json:"updated_at"`
}

// CannedResponseRequest is the create/update payload for a canned response.
type CannedResponseRequest struct {
	Title       *string `json:"title,omitempty"`
	ContentHTML *string `json:"content_html,omitempty"`
	FolderID    *int64  `json:"folder_id,omitempty"`
	Visibility  *int    `json:"visibility,omitempty"`
	GroupIDs    []int64 `json:"group_ids,omitempty"`
	// Attachments are local file paths; setting them forces a multipart request.
	Attachments []string `json:"-"`
}

const (
	// cannedResponseFoldersPath serves every folder verb. The published
	// documentation names a singular "canned_response_folder" path for the
	// writes, but that answers 404; the plural path is the live one.
	cannedResponseFoldersPath = "canned_response_folders"
	cannedResponsesPath       = "canned_responses"
)

// ListCannedResponseFolders returns every canned-response folder.
func (c *Client) ListCannedResponseFolders(ctx context.Context, opts ListOptions) ([]CannedResponseFolder, error) {
	return listAll[CannedResponseFolder](ctx, c, cannedResponseFoldersPath, opts)
}

// GetCannedResponseFolder fetches a folder along with its responses.
func (c *Client) GetCannedResponseFolder(ctx context.Context, id int64) (*CannedResponseFolder, error) {
	return getResource[CannedResponseFolder](ctx, c, cannedResponseFoldersPath, id, nil)
}

// ListCannedResponsesInFolder returns the responses in a folder, with bodies.
func (c *Client) ListCannedResponsesInFolder(
	ctx context.Context,
	folderID int64,
	opts ListOptions,
) ([]CannedResponse, error) {
	return listAll[CannedResponse](ctx, c, pathFor(cannedResponseFoldersPath, folderID)+"/responses", opts)
}

// CreateCannedResponseFolder creates a folder.
func (c *Client) CreateCannedResponseFolder(
	ctx context.Context,
	req CannedResponseFolderRequest,
) (*CannedResponseFolder, error) {
	// The plural path is the one that accepts a create; the singular path the
	// documentation names answers 404.
	return createResource[CannedResponseFolder](ctx, c, cannedResponseFoldersPath, req)
}

// UpdateCannedResponseFolder renames a folder.
func (c *Client) UpdateCannedResponseFolder(
	ctx context.Context,
	id int64,
	req CannedResponseFolderRequest,
) (*CannedResponseFolder, error) {
	return updateResource[CannedResponseFolder](ctx, c, cannedResponseFoldersPath, id, req)
}

// DeleteCannedResponseFolder deletes a folder and its responses.
//
// Freshdesk answers 405 here: the folder endpoints accept only GET, POST and
// PUT. The method is kept because the behaviour is account-dependent and the
// documentation describes a delete, but callers should expect it to fail.
func (c *Client) DeleteCannedResponseFolder(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, cannedResponseFoldersPath, id)
}

// GetCannedResponse fetches a canned response.
func (c *Client) GetCannedResponse(ctx context.Context, id int64) (*CannedResponse, error) {
	return getResource[CannedResponse](ctx, c, cannedResponsesPath, id, nil)
}

// CreateCannedResponse creates a canned response.
func (c *Client) CreateCannedResponse(ctx context.Context, req CannedResponseRequest) (*CannedResponse, error) {
	if len(req.Attachments) > 0 {
		fields, err := multipartFields(req)
		if err != nil {
			return nil, err
		}
		var out CannedResponse
		if err := c.PostMultipart(ctx, cannedResponsesPath, fields,
			map[string][]string{AttachmentsField: req.Attachments}, &out); err != nil {
			return nil, err
		}
		return &out, nil
	}
	return createResource[CannedResponse](ctx, c, cannedResponsesPath, req)
}

// UpdateCannedResponse applies a partial update to a canned response.
func (c *Client) UpdateCannedResponse(
	ctx context.Context,
	id int64,
	req CannedResponseRequest,
) (*CannedResponse, error) {
	return updateResource[CannedResponse](ctx, c, cannedResponsesPath, id, req)
}

// DeleteCannedResponse deletes a canned response.
//
// As with folders, Freshdesk answers 405: the canned response endpoints accept
// only GET, PATCH and PUT.
func (c *Client) DeleteCannedResponse(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, cannedResponsesPath, id)
}

// CreateCannedResponses creates many canned responses in one call.
func (c *Client) CreateCannedResponses(
	ctx context.Context,
	folderID int64,
	reqs []CannedResponseRequest,
) ([]CannedResponse, error) {
	body := map[string]any{"folder_id": folderID, "canned_responses": reqs}
	var out []CannedResponse
	if err := c.Post(ctx, cannedResponsesPath+"/bulk", body, &out); err != nil {
		return nil, err
	}
	return out, nil
}
