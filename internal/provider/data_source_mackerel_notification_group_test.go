package provider_test

import (
	"context"
	"testing"

	fwdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/mackerelio-labs/terraform-provider-mackerel/internal/provider"
)

func Test_MackerelNotificationGroupDataSource_schema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	req := fwdatasource.SchemaRequest{}
	resp := &fwdatasource.SchemaResponse{}
	provider.NewMackerelNotificationGroupDataSource().Schema(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema method diagnostics: %+v", resp.Diagnostics)
	}

	if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation diagnostics: %+v", diags)
	}
}
