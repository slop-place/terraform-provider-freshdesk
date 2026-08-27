package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// testProvider builds the provider under test.
func testProvider() *freshdeskProvider {
	return &freshdeskProvider{version: "test"}
}

// TestProviderSchemaIsValid asks the framework to validate the provider schema,
// which catches malformed attributes that would otherwise only fail at plan.
func TestProviderSchemaIsValid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := testProvider()

	resp := &provider.SchemaResponse{}
	p.Schema(ctx, provider.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("provider schema: %v", resp.Diagnostics)
	}

	if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Errorf("provider schema is invalid: %v", diags)
	}
}

// TestResourceSchemasAreValid validates every resource's schema and checks that
// each declares a distinct type name.
func TestResourceSchemasAreValid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	seen := map[string]bool{}

	for _, newResource := range testProvider().Resources(ctx) {
		r := newResource()

		metaResp := &resource.MetadataResponse{}
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "freshdesk"}, metaResp)

		if metaResp.TypeName == "" {
			t.Errorf("%T: empty type name", r)

			continue
		}

		if seen[metaResp.TypeName] {
			t.Errorf("duplicate resource type name %q", metaResp.TypeName)
		}
		seen[metaResp.TypeName] = true

		schemaResp := &resource.SchemaResponse{}
		r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

		if schemaResp.Diagnostics.HasError() {
			t.Errorf("%s: schema: %v", metaResp.TypeName, schemaResp.Diagnostics)

			continue
		}

		if diags := schemaResp.Schema.ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s: invalid schema: %v", metaResp.TypeName, diags)
		}

		if _, ok := schemaResp.Schema.Attributes["id"]; !ok {
			t.Errorf("%s: no id attribute", metaResp.TypeName)
		}
	}

	if len(seen) == 0 {
		t.Fatal("the provider registered no resources")
	}
	t.Logf("validated %d resources", len(seen))
}

// TestDataSourceSchemasAreValid validates every data source's schema.
func TestDataSourceSchemasAreValid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	seen := map[string]bool{}

	for _, newDataSource := range testProvider().DataSources(ctx) {
		d := newDataSource()

		metaResp := &datasource.MetadataResponse{}
		d.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "freshdesk"}, metaResp)

		if metaResp.TypeName == "" {
			t.Errorf("%T: empty type name", d)

			continue
		}

		if seen[metaResp.TypeName] {
			t.Errorf("duplicate data source type name %q", metaResp.TypeName)
		}
		seen[metaResp.TypeName] = true

		schemaResp := &datasource.SchemaResponse{}
		d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)

		if schemaResp.Diagnostics.HasError() {
			t.Errorf("%s: schema: %v", metaResp.TypeName, schemaResp.Diagnostics)

			continue
		}

		if diags := schemaResp.Schema.ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s: invalid schema: %v", metaResp.TypeName, diags)
		}
	}

	if len(seen) == 0 {
		t.Fatal("the provider registered no data sources")
	}
	t.Logf("validated %d data sources", len(seen))
}
