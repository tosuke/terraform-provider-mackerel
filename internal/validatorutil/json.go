package validatorutil

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type jsonValidator struct{}

var _ validator.String = (*jsonValidator)(nil)

func JSONString() validator.String {
	return &jsonValidator{}
}

func (v *jsonValidator) Description(context.Context) string {
	return "Value must be a json string"
}

func (v *jsonValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *jsonValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	s := req.ConfigValue.ValueString()
	if s == "" {
		return
	}

	var j any
	if err := json.Unmarshal([]byte(s), &j); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid JSON",
			err.Error(),
		)
		return
	}
}
