package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/mackerelio-labs/terraform-provider-mackerel/internal/mackerel"
)

var (
	_ resource.Resource                   = (*mackerelServiceMetadataResource)(nil)
	_ resource.ResourceWithValidateConfig = (*mackerelServiceMetadataResource)(nil)
	_ resource.ResourceWithConfigure      = (*mackerelServiceMetadataResource)(nil)
	_ resource.ResourceWithImportState    = (*mackerelServiceMetadataResource)(nil)
)

func NewMackerelServiceMetadataResource() resource.Resource {
	return &mackerelServiceMetadataResource{}
}

type mackerelServiceMetadataResource struct {
	Client *mackerel.Client
}

func (r *mackerelServiceMetadataResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_metadata"
}

func (r *mackerelServiceMetadataResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource allows creating and management of Service Metadata.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service": schema.StringAttribute{
				Description: "The name of the service.",

				Required: true,
				Validators: []validator.String{
					mackerel.ServiceNameValidator(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"namespace": schema.StringAttribute{
				Description: "Identifier for the metadata.",

				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"metadata_json": schema.StringAttribute{
				Description: "Arbitrary JSON data for the service.",

				Required:   true,
				CustomType: jsontypes.NormalizedType{},
			},
		},
	}
}

func (r *mackerelServiceMetadataResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data mackerel.ServiceMetadataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		serviceName, namespace, err := mackerel.ParseServiceMetadataID(data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Invalid Service Metadata ID",
				err.Error(),
			)
		} else {
			// TODO: validate service name part
			// match
			if !data.ServiceName.IsNull() && !data.ServiceName.IsUnknown() && data.ServiceName.ValueString() != serviceName {
				resp.Diagnostics.AddAttributeWarning(
					path.Root("service"),
					"Service name is unmatch with ID",
					fmt.Sprintf("expected to %s, but got: %s", serviceName, data.ServiceName.ValueString()),
				)
			}
			if !data.Namespace.IsNull() && !data.Namespace.IsUnknown() && data.Namespace.ValueString() != namespace {
				resp.Diagnostics.AddAttributeWarning(
					path.Root("namespace"),
					"Namespace is unmatch with ID",
					fmt.Sprintf("expected to %s, but got: %s", namespace, data.Namespace.ValueString()),
				)
			}
		}
	}
}

func (r *mackerelServiceMetadataResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := retrieveClient(ctx, req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.Client = client
}

func (r *mackerelServiceMetadataResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data mackerel.ServiceMetadataModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.createOrUpdate(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *mackerelServiceMetadataResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data mackerel.ServiceMetadataModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.read(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *mackerelServiceMetadataResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data mackerel.ServiceMetadataModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.createOrUpdate(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *mackerelServiceMetadataResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data mackerel.ServiceMetadataModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.Client.DeleteServiceMetaData(
		data.ServiceName.ValueString(),
		data.Namespace.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Service Metadata",
			err.Error(),
		)
		return
	}
}

func (r *mackerelServiceMetadataResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *mackerelServiceMetadataResource) read(ctx context.Context, data *mackerel.ServiceMetadataModel) (diags diag.Diagnostics) {
	if data.ID.IsNull() || data.ID.IsUnknown() {
		diags.AddAttributeError(
			path.Root("id"),
			"No ID",
			"Please report this issue.",
		)
		return
	}
	serviceName, namespace, err := mackerel.ParseServiceMetadataID(data.ID.ValueString())
	if err != nil {
		diags.AddAttributeError(
			path.Root("id"),
			"Invalid ID",
			err.Error(),
		)
		return
	}

	remoteData, err := mackerel.ReadServiceMetadata(ctx, r.Client, serviceName, namespace)
	if err != nil {
		diags.AddError(
			"Unable to read Service Metadata",
			err.Error(),
		)
		return
	}

	data.Set(*remoteData)

	return
}

func (r *mackerelServiceMetadataResource) createOrUpdate(ctx context.Context, data *mackerel.ServiceMetadataModel) (diags diag.Diagnostics) {
	if err := data.CreateOrUpdateMetadata(ctx, r.Client); err != nil {
		diags.AddError(
			"Unable to put Service Metadata",
			err.Error(),
		)
		return
	}

	diags.Append(r.read(ctx, data)...)
	if diags.HasError() {
		return
	}

	return
}
