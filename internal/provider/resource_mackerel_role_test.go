package provider_test

import (
	"context"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/mackerelio-labs/terraform-provider-mackerel/internal/provider"
)

func Test_MackerelRoleResource_schema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	req := fwresource.SchemaRequest{}
	resp := fwresource.SchemaResponse{}
	provider.NewMackerelRoleResource().Schema(ctx, req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema: %+v", resp.Diagnostics)
	}

	if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation: %+v", diags)
	}
}
