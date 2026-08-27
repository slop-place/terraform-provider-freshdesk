package freshdesk

import (
	"context"
	"strings"
)

// Survey is a customer-satisfaction survey on the newer CSAT endpoint.
//
// Surveys are identified by a UUID rather than a numeric ID.
type Survey struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// HeaderMessage introduces the survey to the customer.
	HeaderMessage string `json:"header_message"`
	// State is "ACTIVE" or "INACTIVE".
	State     string           `json:"state"`
	Language  string           `json:"language"`
	Questions []SurveyQuestion `json:"questions"`
	CreatedAt Time             `json:"created_at"`
	UpdatedAt Time             `json:"updated_at"`
}

// Active reports whether the survey is being sent.
func (s Survey) Active() bool { return strings.EqualFold(s.State, "ACTIVE") }

// SurveyQuestion is one question on a survey.
type SurveyQuestion struct {
	ID string `json:"id"`
	// Type is "range", "paragraph", "thank_you", or a choice type.
	Type string `json:"type"`
	// QuestionIdentifier names the question's role, e.g. "CSAT".
	QuestionIdentifier string `json:"question_identifier"`
	// Text is the wording shown to the customer.
	Text     string `json:"text"`
	Required bool   `json:"required"`
	// Scale describes a range question's presentation.
	Scale any `json:"scale,omitempty"`
	// Ratings lists the permitted values for a range question.
	Ratings any `json:"ratings,omitempty"`
	// SkipLogic routes the customer to a later question.
	SkipLogic any `json:"skip_logic,omitempty"`
}

// ListSurveys returns every satisfaction survey via the newer endpoint.
//
// This endpoint wraps its collection in a "data" envelope rather than
// returning a bare array, so it does not use the shared pagination walker.
func (c *Client) ListSurveys(ctx context.Context, opts ListOptions) ([]Survey, error) {
	var out struct {
		Data []Survey `json:"data"`
	}
	if err := c.Get(ctx, "customer-satisfaction/surveys", opts.Values(), &out); err != nil {
		return nil, err
	}

	return out.Data, nil
}

// LegacySurvey is a survey as returned by the older /surveys endpoint.
type LegacySurvey struct {
	ID        int64            `json:"id"`
	Title     string           `json:"title"`
	Active    bool             `json:"active"`
	Questions []SurveyQuestion `json:"questions"`
	CreatedAt Time             `json:"created_at"`
	UpdatedAt Time             `json:"updated_at"`
}

// ListLegacySurveys returns surveys through the older /surveys endpoint.
func (c *Client) ListLegacySurveys(ctx context.Context, opts ListOptions) ([]LegacySurvey, error) {
	return listAll[LegacySurvey](ctx, c, "surveys", opts)
}

// SatisfactionRating is one customer response to a survey.
type SatisfactionRating struct {
	ID       int64 `json:"id"`
	TicketID int64 `json:"ticket_id"`
	// SurveyID identifies the survey answered.
	SurveyID int64 `json:"survey_id"`
	AgentID  int64 `json:"agent_id"`
	GroupID  int64 `json:"group_id"`
	UserID   int64 `json:"user_id"`
	// Ratings maps a question key to the rating given.
	Ratings map[string]int `json:"ratings"`
	// Feedback is the free-text comment, when the customer left one.
	Feedback  string `json:"feedback"`
	CreatedAt Time   `json:"created_at"`
	UpdatedAt Time   `json:"updated_at"`
}

// SatisfactionRatingRequest records a rating against a ticket.
type SatisfactionRatingRequest struct {
	// Ratings maps a question key to the rating given; "default_question" is
	// the primary rating.
	Ratings  map[string]int `json:"ratings"`
	Feedback *string        `json:"feedback,omitempty"`
	// CreatedAt backdates the rating.
	CreatedAt *string `json:"created_at,omitempty"`
}

// SatisfactionRatingListOptions filters the ratings endpoint.
type SatisfactionRatingListOptions struct {
	ListOptions

	// CreatedSince restricts results to ratings at or after a timestamp.
	CreatedSince string
	UserID       int64
}

// ListSatisfactionRatings returns every satisfaction rating.
func (c *Client) ListSatisfactionRatings(
	ctx context.Context,
	opts SatisfactionRatingListOptions,
) ([]SatisfactionRating, error) {
	base := opts.ListOptions
	v := base.Values()
	if opts.CreatedSince != "" {
		v.Set("created_since", opts.CreatedSince)
	}
	if opts.UserID > 0 {
		v.Set("user_id", idString(opts.UserID))
	}
	base.Extra = v
	return listAll[SatisfactionRating](ctx, c, "surveys/satisfaction_ratings", base)
}

// ListTicketSatisfactionRatings returns the ratings left on a ticket.
func (c *Client) ListTicketSatisfactionRatings(
	ctx context.Context,
	ticketID int64,
	opts ListOptions,
) ([]SatisfactionRating, error) {
	return listAll[SatisfactionRating](ctx, c, pathFor(ticketsPath, ticketID)+"/satisfaction_ratings", opts)
}

// CreateSatisfactionRating records a rating against a ticket.
func (c *Client) CreateSatisfactionRating(
	ctx context.Context,
	ticketID int64,
	req SatisfactionRatingRequest,
) (*SatisfactionRating, error) {
	return createResource[SatisfactionRating](ctx, c,
		pathFor(ticketsPath, ticketID)+"/satisfaction_ratings", req)
}

// SurveyResponse is a response to a survey on the newer CSAT endpoint.
type SurveyResponse struct {
	ID        string         `json:"id"`
	SurveyID  string         `json:"survey_id"`
	TicketID  int64          `json:"ticket_id"`
	Responses map[string]any `json:"responses"`
	CreatedAt Time           `json:"created_at"`
	UpdatedAt Time           `json:"updated_at"`
}

// surveyResponsesPath builds the responses path for a survey UUID.
func surveyResponsesPath(surveyID string) string {
	return "customer-satisfaction/surveys/" + idString(surveyID) + "/responses"
}

// ListSurveyResponses returns the responses to a survey.
func (c *Client) ListSurveyResponses(ctx context.Context, surveyID string, opts ListOptions) ([]SurveyResponse, error) {
	return listAll[SurveyResponse](ctx, c, surveyResponsesPath(surveyID), opts)
}

// CreateSurveyResponse records a response to a survey.
func (c *Client) CreateSurveyResponse(
	ctx context.Context,
	surveyID string,
	body map[string]any,
) (*SurveyResponse, error) {
	var out SurveyResponse
	if err := c.Post(ctx, surveyResponsesPath(surveyID), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
