package provider

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/mackerelio-labs/terraform-provider-mackerel/internal/mackerel"
	"github.com/mackerelio-labs/terraform-provider-mackerel/internal/validatorutil"
)

type mackerelProvider struct {
	config MackerelProviderConfig
}
type MackerelProviderConfig struct {
	// if nil, all components are enabled
	enabledResourceTypes    []string
	disabledResourceTypes   []string
	enabledDataSourceTypes  []string
	disabledDataSourceTypes []string
}
type MackerelProviderOption func(*MackerelProviderConfig)

func WithResourceEnabled(types ...string) MackerelProviderOption {
	return func(mpc *MackerelProviderConfig) {
		mpc.enabledResourceTypes = append(mpc.enabledResourceTypes, types...)
	}
}
func WithResourceDisabled(types ...string) MackerelProviderOption {
	return func(mpc *MackerelProviderConfig) {
		mpc.disabledResourceTypes = append(mpc.disabledResourceTypes, types...)
	}
}
func WithDataSourceEnabled(types ...string) MackerelProviderOption {
	return func(mpc *MackerelProviderConfig) {
		mpc.enabledDataSourceTypes = append(mpc.enabledDataSourceTypes, types...)
	}
}
func WithDataSourceDisabled(types ...string) MackerelProviderOption {
	return func(mpc *MackerelProviderConfig) {
		mpc.disabledDataSourceTypes = append(mpc.disabledDataSourceTypes, types...)
	}
}

const providerTypeName = "mackerel"

var _ provider.Provider = (*mackerelProvider)(nil)

func New(opts ...MackerelProviderOption) provider.Provider {
	var config MackerelProviderConfig
	for _, opt := range opts {
		opt(&config)
	}
	return &mackerelProvider{
		config: config,
	}
}

func (m *mackerelProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = providerTypeName
}

func (m *mackerelProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Mackerel API Key",
				Optional:    true,
				Sensitive:   true,
			},
			"api_base": schema.StringAttribute{
				Description: "Mackerel API BASE URL",
				Optional:    true,
				Sensitive:   true,
				Validators:  []validator.String{validatorutil.IsURLWithHTTPorHTTPS()},
			},
		},
	}
}

func (m *mackerelProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var schemaConfig mackerel.ClientConfigModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &schemaConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := mackerel.NewClientConfigFromEnv()
	// merge config
	if config.APIKey.IsUnknown() {
		config.APIKey = schemaConfig.APIKey
	}
	if config.APIBase.IsUnknown() {
		config.APIBase = schemaConfig.APIBase
	}

	client, err := config.NewClient()
	if err != nil {
		if errors.Is(err, mackerel.ErrNoAPIKey) {
			resp.Diagnostics.AddError(
				"No API Key",
				err.Error(),
			)
		} else {
			resp.Diagnostics.AddError(
				"Unable to create Mackerel Client",
				err.Error(),
			)
		}
		return
	}

	resp.ResourceData = client
	resp.DataSourceData = client
}

func (m *mackerelProvider) Resources(ctx context.Context) []func() resource.Resource {
	factories := []func() resource.Resource{
		NewMackerelServiceResource,
	}
	if m.config.enabledResourceTypes == nil && m.config.disabledResourceTypes == nil {
		return factories
	}

	filteredFactories := make([]func() resource.Resource, 0, len(factories))
	for _, f := range factories {
		req := resource.MetadataRequest{ProviderTypeName: providerTypeName}
		resp := resource.MetadataResponse{}
		f().Metadata(ctx, req, &resp)

		enabled := true
		if m.config.enabledResourceTypes != nil {
			enabled = enabled && slices.Contains(m.config.enabledResourceTypes, resp.TypeName)
		}
		if m.config.disabledResourceTypes != nil {
			enabled = enabled && !slices.Contains(m.config.disabledResourceTypes, resp.TypeName)
		}
		if enabled {
			filteredFactories = append(filteredFactories, f)
		}
	}
	return filteredFactories
}

func (m *mackerelProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	factories := []func() datasource.DataSource{
		NewMackerelServiceDataSource,
	}
	if m.config.enabledDataSourceTypes == nil && m.config.disabledDataSourceTypes == nil {
		return factories
	}

	filteredFactories := make([]func() datasource.DataSource, 0, len(factories))
	for _, f := range factories {
		req := datasource.MetadataRequest{ProviderTypeName: providerTypeName}
		resp := datasource.MetadataResponse{}
		f().Metadata(ctx, req, &resp)

		enabled := true
		if m.config.enabledDataSourceTypes != nil {
			enabled = enabled && slices.Contains(m.config.enabledDataSourceTypes, resp.TypeName)
		}
		if m.config.disabledDataSourceTypes != nil {
			enabled = enabled && !slices.Contains(m.config.disabledDataSourceTypes, resp.TypeName)
		}
		if enabled {
			filteredFactories = append(filteredFactories, f)
		}
	}
	return filteredFactories
}

func retrieveClient(_ context.Context, providerData any) (client *mackerel.Client, diags diag.Diagnostics) {
	if /* ConfigureProvider RPC is not called */ providerData == nil {
		return
	}

	client, ok := providerData.(*mackerel.Client)
	if !ok {
		diags.AddError(

			"Unconfigured Mackerel Client",
			fmt.Sprintf("Expected configured Mackerel client, but got: %T. Please report this issue.", providerData),
		)
		return
	}
	return
}
