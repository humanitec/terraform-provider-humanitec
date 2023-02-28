package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceEnvironmentType{}
var _ resource.ResourceWithImportState = &ResourceEnvironmentType{}

func NewResourceEnvironmentType() resource.Resource {
	return &ResourceEnvironmentType{}
}

// ResourceEnvironmentType defines the resource implementation.
type ResourceEnvironmentType struct {
	client *humanitec.Client
	orgId  string
}

// EnvironmentTypeModel describes the app data model.
type EnvironmentTypeModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
}

func (r *ResourceEnvironmentType) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_type"
}

func (r *ResourceEnvironmentType) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Environment Types are a way of grouping and managing Environments.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Environment Type.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A Human-readable description of the Environment Type",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ResourceEnvironmentType) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	resdata, ok := req.ProviderData.(*HumanitecData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = resdata.Client
	r.orgId = resdata.OrgID
}

func parseEnvironmentTypeResponse(res *client.EnvironmentTypeResponse, data *EnvironmentTypeModel) {
	data.ID = types.StringValue(res.Id)
	data.Description = types.StringValue(res.Description)
}

func (r *ResourceEnvironmentType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *EnvironmentTypeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()

	var description *string
	if !data.Description.IsNull() {
		val := data.Description.ValueString()
		description = &val
	}

	httpResp, err := r.client.PostOrgsOrgIdEnvTypesWithResponse(ctx, r.orgId, client.PostOrgsOrgIdEnvTypesJSONRequestBody{
		Id:          id,
		Description: description,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create environment type, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 201 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create environment type, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseEnvironmentTypeResponse(httpResp.JSON201, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceEnvironmentType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *EnvironmentTypeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.GetOrgsOrgIdEnvTypesEnvTypeIdWithResponse(ctx, r.orgId, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read environment type, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read environment type, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseEnvironmentTypeResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceEnvironmentType) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("UNSUPPORTED_OPERATION", "Updating an environment type is currently not supported")
}

func (r *ResourceEnvironmentType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *EnvironmentTypeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.DeleteOrgsOrgIdEnvTypesEnvTypeIdWithResponse(ctx, r.orgId, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete environment type, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 204 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete environment type, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}
}

func (r *ResourceEnvironmentType) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
