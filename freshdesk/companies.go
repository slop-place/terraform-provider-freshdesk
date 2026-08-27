package freshdesk

import (
	"context"
	"net/url"
	"strconv"
)

// Company groups contacts that share an organisation.
type Company struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Note        string   `json:"note"`
	Domains     []string `json:"domains"`
	// HealthScore, AccountTier and Industry are dropdown fields whose allowed
	// values are configured per account.
	HealthScore string `json:"health_score"`
	AccountTier string `json:"account_tier"`
	Industry    string `json:"industry"`
	RenewalDate Time   `json:"renewal_date"`

	CustomFields CustomFields `json:"custom_fields"`

	// OrgCompanyID identifies the company across the Freshworks organisation.
	// Freshdesk quotes it on list responses but not on create, hence FlexString.
	OrgCompanyID FlexString `json:"org_company_id"`

	CreatedAt Time `json:"created_at"`
	UpdatedAt Time `json:"updated_at"`
}

// CompanyRequest is the create/update payload for a company.
type CompanyRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Note        *string  `json:"note,omitempty"`
	Domains     []string `json:"domains,omitempty"`
	HealthScore *string  `json:"health_score,omitempty"`
	AccountTier *string  `json:"account_tier,omitempty"`
	Industry    *string  `json:"industry,omitempty"`
	// RenewalDate accepts a YYYY-MM-DD date.
	RenewalDate *string `json:"renewal_date,omitempty"`

	CustomFields CustomFields `json:"custom_fields,omitempty"`
	// LookupParameter is "display_id" or "primary_field_value".
	LookupParameter *string `json:"lookup_parameter,omitempty"`

	// ClearDomains sends an empty domains array.
	ClearDomains bool `json:"-"`
	// ClearRenewalDate sends "renewal_date": null, unsetting the date.
	ClearRenewalDate bool `json:"-"`
}

// MarshalJSON applies the empty-array clearing semantics.
func (r CompanyRequest) MarshalJSON() ([]byte, error) {
	type alias CompanyRequest
	m, err := structToMap(alias(r))
	if err != nil {
		return nil, err
	}
	if r.ClearDomains {
		m["domains"] = []string{}
	}

	if r.ClearRenewalDate {
		m["renewal_date"] = nil
	}

	return marshalMap(m)
}

const companiesPath = "companies"

// GetCompany fetches a company by ID.
func (c *Client) GetCompany(ctx context.Context, id int64) (*Company, error) {
	return getResource[Company](ctx, c, companiesPath, id, nil)
}

// ListCompanies returns every company.
func (c *Client) ListCompanies(ctx context.Context, opts ListOptions) ([]Company, error) {
	return listAll[Company](ctx, c, companiesPath, opts)
}

// CreateCompany creates a company.
func (c *Client) CreateCompany(ctx context.Context, req CompanyRequest) (*Company, error) {
	return createResource[Company](ctx, c, companiesPath, req)
}

// UpdateCompany applies a partial update to a company.
func (c *Client) UpdateCompany(ctx context.Context, id int64, req CompanyRequest) (*Company, error) {
	return updateResource[Company](ctx, c, companiesPath, id, req)
}

// DeleteCompany deletes a company. Its contacts are retained.
func (c *Client) DeleteCompany(ctx context.Context, id int64) error {
	return deleteResource(ctx, c, companiesPath, id)
}

// SearchCompanies matches companies by name prefix.
func (c *Client) SearchCompanies(ctx context.Context, name string) ([]Company, error) {
	var out struct {
		Companies []Company `json:"companies"`
	}
	if err := c.Get(ctx, companiesPath+"/autocomplete", url.Values{"name": {name}}, &out); err != nil {
		return nil, err
	}
	return out.Companies, nil
}

// CompanySearchResult is the envelope returned by the company filter endpoint.
type CompanySearchResult struct {
	Total   int       `json:"total"`
	Results []Company `json:"results"`
}

// FilterCompanies runs a Freshdesk query-language filter over companies.
func (c *Client) FilterCompanies(ctx context.Context, query string, page int) (*CompanySearchResult, error) {
	q := url.Values{"query": {quoteQuery(query)}}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	var out CompanySearchResult
	if err := c.Get(ctx, "search/companies", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ExportCompanies starts a company export and returns the job to poll.
func (c *Client) ExportCompanies(ctx context.Context, defaultFields, customFields []string) (*ExportJob, error) {
	body := map[string]any{
		"fields": map[string]any{
			"default_fields": defaultFields,
			"custom_fields":  customFields,
		},
	}
	var out ExportJob
	if err := c.Post(ctx, companiesPath+"/export", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCompanyExport polls a company export job.
func (c *Client) GetCompanyExport(ctx context.Context, jobID string) (*ExportJob, error) {
	return getResource[ExportJob](ctx, c, companiesPath+"/export", jobID, nil)
}

// ImportCompanies starts a company import from an uploaded CSV.
func (c *Client) ImportCompanies(ctx context.Context, csvPath string, fields map[string]string) (*ImportJob, error) {
	form := map[string][]string{}
	for k, v := range fields {
		form["fields["+k+"]"] = []string{v}
	}
	var out ImportJob
	if err := c.PostMultipart(ctx, companiesPath+"/imports", form,
		map[string][]string{"file": {csvPath}}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCompanyImport polls a company import job.
func (c *Client) GetCompanyImport(ctx context.Context, id int64) (*ImportJob, error) {
	return getResource[ImportJob](ctx, c, companiesPath+"/imports", id, nil)
}

// CancelCompanyImport cancels a running company import.
func (c *Client) CancelCompanyImport(ctx context.Context, id int64) error {
	return c.Post(ctx, pathFor(companiesPath+"/imports", id)+"/cancel", nil, nil)
}
