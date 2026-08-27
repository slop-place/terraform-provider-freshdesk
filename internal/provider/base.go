package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// diagnostics is the framework's diagnostic collection, aliased so resource
// helpers read a little more plainly.
type diagnostics = diag.Diagnostics

// validatorString is the framework's string validator interface, aliased for
// brevity in the many schemas that use one.
type validatorString = validator.String

// validatorInt64 is the framework's int64 validator interface.
type validatorInt64 = validator.Int64

// base carries the configured API client into a resource. Embedding it
// supplies the Configure method every resource needs.
type base struct {
	client *freshdesk.Client
}

// Configure stores the provider's client on the resource.
func (b *base) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		// Configure runs before the provider is configured during validation.
		return
	}

	client, ok := req.ProviderData.(*freshdesk.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data",
			fmt.Sprintf("Expected *freshdesk.Client, got %T. This is a bug in the provider.",
				req.ProviderData))

		return
	}
	b.client = client
}

// dataSourceBase carries the configured API client into a data source.
type dataSourceBase struct {
	client *freshdesk.Client
}

// Configure stores the provider's client on the data source.
func (b *dataSourceBase) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*freshdesk.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data",
			fmt.Sprintf("Expected *freshdesk.Client, got %T. This is a bug in the provider.",
				req.ProviderData))

		return
	}
	b.client = client
}

// planModifierString is the framework's string plan modifier interface.
type planModifierString = planmodifier.String

// useStateForUnknown keeps a computed attribute stable across plans when the
// API will not change it.
func useStateForUnknown() planModifierString {
	return stringplanmodifier.UseStateForUnknown()
}

// requiresReplace marks an attribute that cannot be changed in place.
func requiresReplace() planModifierString {
	return stringplanmodifier.RequiresReplace()
}

// requiresReplaceInt64 marks an int64 attribute that cannot be changed in place.
func requiresReplaceInt64() planmodifier.Int64 {
	return int64planmodifier.RequiresReplace()
}
