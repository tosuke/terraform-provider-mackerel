package validatorutil_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mackerelio-labs/terraform-provider-mackerel/internal/validatorutil"
)

func Test_Validator_JSON(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		val       types.String
		wantError bool
	}{
		"valid": {
			val: types.StringValue(`{ "str": "s", "bool": true, "num": 42, "null": null, "array": [[]] }`),
		},
		"raw value": {
			val: types.StringValue("3.14"),
		},
		"empty string": {
			val: types.StringValue(""),
		},
		"invalid key": {
			val: types.StringValue("{ key: 0 }"),
			wantError: true,
		},
		"invalid comma": {
			val: types.StringValue(`{"key": 0,}`),
			wantError: true,
		},
	}

	ctx := context.Background()
	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    tt.val,
			}
			resp := &validator.StringResponse{}
			validatorutil.JSONString().ValidateString(ctx, req, resp)
			for _, d := range resp.Diagnostics {
				assertDiagMatchPathExpr(t, d, path.MatchRoot("test"))
			}

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				if tt.wantError {
					t.Error("expected to have errors, but got no error")
				} else {
					t.Errorf("unexpected error: %+v", resp.Diagnostics.Errors())
				}
			}
		})
	}
}
