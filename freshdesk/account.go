package freshdesk

import "context"

// Account describes the helpdesk account the API key belongs to.
type Account struct {
	AccountID        int64  `json:"account_id"`
	AccountName      string `json:"account_name"`
	AccountDomain    string `json:"account_domain"`
	OrganisationID   int64  `json:"organisation_id"`
	OrganisationName string `json:"organisation_name"`
	// TierType names the plan, e.g. "Enterprise Omni".
	TierType string `json:"tier_type"`
	// Type distinguishes classic from unified-omni accounts.
	Type      string `json:"type"`
	BundleID  any    `json:"bundle_id"`
	CloudType any    `json:"cloud_type"`
	// DataCenter is the region the account is hosted in.
	DataCenter     string             `json:"data_center"`
	Timezone       string             `json:"timezone"`
	HIPAACompliant bool               `json:"hipaa_compliant"`
	TotalAgents    AccountAgentCounts `json:"total_agents"`
	Address        AccountAddress     `json:"address"`
	ContactPerson  AccountContact     `json:"contact_person"`
}

// AccountAgentCounts breaks agent seats down by kind.
type AccountAgentCounts struct {
	FullTime      int `json:"full_time"`
	Occasional    int `json:"occasional"`
	FieldService  int `json:"field_service"`
	Collaborators int `json:"collaborators"`
}

// AccountAddress is the account's registered address.
type AccountAddress struct {
	Country       string `json:"country"`
	State         string `json:"state"`
	City          string `json:"city"`
	Street        string `json:"street"`
	PrimaryStreet string `json:"primary_street"`
	PostalCode    string `json:"postalcode"`
}

// AccountContact is the account's primary contact.
type AccountContact struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
}

// GetAccount reads the account the API key belongs to.
func (c *Client) GetAccount(ctx context.Context) (*Account, error) {
	var out Account
	if err := c.Get(ctx, "account", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ExportAccount starts an account data export and returns the job.
func (c *Client) ExportAccount(ctx context.Context, body map[string]any) (*Job, error) {
	var out Job
	if err := c.Post(ctx, "account/export", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// HelpdeskSettings holds the account's language configuration.
type HelpdeskSettings struct {
	PrimaryLanguage     string   `json:"primary_language"`
	SupportedLanguages  []string `json:"supported_languages"`
	PortalLanguages     []string `json:"portal_languages"`
	HelpWidgetLanguages []string `json:"help_widget_languages"`
}

// GetHelpdeskSettings reads the account's language settings.
func (c *Client) GetHelpdeskSettings(ctx context.Context) (*HelpdeskSettings, error) {
	var out HelpdeskSettings
	if err := c.Get(ctx, "settings/helpdesk", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Job is an asynchronous operation such as a bulk create or an export.
type Job struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	// DownloadURL is set once an export job completes.
	DownloadURL string `json:"download_url,omitempty"`
	// Records reports per-record outcomes for bulk jobs.
	Records []map[string]any `json:"records,omitempty"`
	// Errors reports per-record failures for bulk jobs.
	Errors []map[string]any `json:"errors,omitempty"`
}

// GetJob polls an asynchronous job.
func (c *Client) GetJob(ctx context.Context, id string) (*Job, error) {
	return getResource[Job](ctx, c, "jobs", id, nil)
}
