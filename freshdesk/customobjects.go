package freshdesk

import (
	"context"
	"net/url"
)

// CustomObjectSchema describes a custom object and its fields.
//
// Custom objects require a plan that enables the feature; accounts without it
// answer 403 on these endpoints.
type CustomObjectSchema struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Prefix      string              `json:"prefix"`
	Fields      []CustomObjectField `json:"fields"`
	CreatedAt   Time                `json:"created_at"`
	UpdatedAt   Time                `json:"updated_at"`
}

// CustomObjectField is one field on a custom object schema.
type CustomObjectField struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Label string `json:"label"`
	Type  string `json:"type"`
	// Required marks a field that must be supplied on every record.
	Required bool `json:"required"`
	// PrimaryField marks the field used as the record's display value.
	PrimaryField bool     `json:"primary_field"`
	Choices      []string `json:"choices,omitempty"`
}

// CustomObjectRecord is one row of a custom object.
type CustomObjectRecord struct {
	// DisplayID is the human-facing record identifier, e.g. "BKG-1".
	DisplayID string `json:"display_id"`
	// Data holds the record's field values keyed by field name.
	Data      map[string]any `json:"data"`
	CreatedAt Time           `json:"created_at"`
	UpdatedAt Time           `json:"updated_at"`
}

const customObjectSchemasPath = "custom_objects/schemas"

// ListCustomObjectSchemas returns every custom object schema.
func (c *Client) ListCustomObjectSchemas(ctx context.Context) ([]CustomObjectSchema, error) {
	// This endpoint wraps its collection in a "schemas" envelope.
	var out struct {
		Schemas []CustomObjectSchema `json:"schemas"`
	}
	if err := c.Get(ctx, customObjectSchemasPath, nil, &out); err != nil {
		return nil, err
	}
	return out.Schemas, nil
}

// GetCustomObjectSchema fetches one custom object schema.
func (c *Client) GetCustomObjectSchema(ctx context.Context, schemaID string) (*CustomObjectSchema, error) {
	return getResource[CustomObjectSchema](ctx, c, customObjectSchemasPath, schemaID, nil)
}

// recordsPath builds the records collection path for a schema.
func recordsPath(schemaID string) string {
	return customObjectSchemasPath + "/" + url.PathEscape(schemaID) + "/records"
}

// ListCustomObjectRecords returns records of a custom object. Pass filters as
// field-name/value pairs to narrow the result.
func (c *Client) ListCustomObjectRecords(
	ctx context.Context,
	schemaID string,
	filters url.Values,
) ([]CustomObjectRecord, error) {
	// This endpoint wraps its collection in a "records" envelope.
	var out struct {
		Records []CustomObjectRecord `json:"records"`
	}
	if err := c.Get(ctx, recordsPath(schemaID), filters, &out); err != nil {
		return nil, err
	}
	return out.Records, nil
}

// CountCustomObjectRecords returns how many records a custom object holds.
func (c *Client) CountCustomObjectRecords(ctx context.Context, schemaID string) (int, error) {
	var out struct {
		Count int `json:"count"`
	}
	if err := c.Get(ctx, recordsPath(schemaID)+"/count", nil, &out); err != nil {
		return 0, err
	}
	return out.Count, nil
}

// GetCustomObjectRecord fetches one record by display ID.
func (c *Client) GetCustomObjectRecord(ctx context.Context, schemaID, recordID string) (*CustomObjectRecord, error) {
	return getResource[CustomObjectRecord](ctx, c, recordsPath(schemaID), recordID, nil)
}

// CreateCustomObjectRecord creates a record. Data is keyed by field name.
func (c *Client) CreateCustomObjectRecord(
	ctx context.Context,
	schemaID string,
	data map[string]any,
) (*CustomObjectRecord, error) {
	return createResource[CustomObjectRecord](ctx, c, recordsPath(schemaID),
		map[string]any{"data": data})
}

// UpdateCustomObjectRecord updates a record.
func (c *Client) UpdateCustomObjectRecord(
	ctx context.Context,
	schemaID,
	recordID string,
	data map[string]any,
) (*CustomObjectRecord, error) {
	return updateResource[CustomObjectRecord](ctx, c, recordsPath(schemaID), recordID,
		map[string]any{"data": data})
}

// DeleteCustomObjectRecord deletes a record.
func (c *Client) DeleteCustomObjectRecord(ctx context.Context, schemaID, recordID string) error {
	return deleteResource(ctx, c, recordsPath(schemaID), recordID)
}
