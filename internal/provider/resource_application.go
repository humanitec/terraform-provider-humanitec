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
var _ resource.Resource = &ResourceApplication{}
var _ resource.ResourceWithImportState = &ResourceApplication{}

func NewResourceApplication() resource.Resource {
	return &ResourceApplication{}
}

// ResourceApplication defines the resource implementation.
type ResourceApplication struct {
	client *humanitec.Client
	orgId  string
}

// ApplicationModel describes the app data model.
type ApplicationModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (r *ResourceApplication) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ResourceApplication) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An Application is a collection of Modules that work together. When deployed, all Modules in an Application are deployed to the same namespace.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID which refers to a specific application.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The Human-friendly name for the Application.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ResourceApplication) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func parseApplicationResponse(res *client.ApplicationResponse, data *ApplicationModel) {
	data.ID = types.StringValue(res.Id)
	data.Name = types.StringValue(res.Name)
}

func (r *ResourceApplication) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ApplicationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	name := data.Name.ValueString()

	httpResp, err := r.client.PostOrgsOrgIdAppsWithResponse(ctx, r.orgId, client.PostOrgsOrgIdAppsJSONRequestBody{
		Id:   id,
		Name: name,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create app, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 201 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create app, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseApplicationResponse(httpResp.JSON201, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceApplication) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ApplicationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.GetOrgsOrgIdAppsAppIdWithResponse(ctx, r.orgId, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read application, got error: %s", err))
		return
	}

	if httpResp.StatusCode() == 404 {
		resp.Diagnostics.AddWarning("Application not found", fmt.Sprintf("The app (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read application, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseApplicationResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceApplication) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("UNSUPPORTED_OPERATION", "Updating an application is currently not supported")
}

func (r *ResourceApplication) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ApplicationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.DeleteOrgsOrgIdAppsAppIdWithResponse(ctx, r.orgId, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete application, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 204 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete application, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}
}

func (r *ResourceApplication) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
