package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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
var _ resource.Resource = &ResourceEnvironment{}
var _ resource.ResourceWithImportState = &ResourceEnvironment{}

func NewResourceEnvironment() resource.Resource {
	return &ResourceEnvironment{}
}

// ResourceEnvironment defines the resource implementation.
type ResourceEnvironment struct {
	client *humanitec.Client
	orgID  string
}

type EnvironmentModel struct {
	AppID        types.String `tfsdk:"app_id"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	FromDeployID types.String `tfsdk:"from_deploy_id"`
}

func (r *ResourceEnvironment) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *ResourceEnvironment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An Environment is a space where an instance of an Application can be deployed. Environments consist of a Kubernetes namespace and any shared Resources (as configured by relevant Matching Rules).",

		Attributes: map[string]schema.Attribute{
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The Application ID.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID the Environment is referenced as.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The Human-friendly name for the Environment.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The Environment Type. This is used for organizing and managing Environments.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"from_deploy_id": schema.StringAttribute{
				MarkdownDescription: "Defines the existing Deployment the new Environment will be based on.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ResourceEnvironment) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.orgID = resdata.OrgID
}

func (r *ResourceEnvironment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *EnvironmentModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()

	var environment *client.EnvironmentResponse
	createEnvironmentResp, err := r.client.CreateEnvironmentWithResponse(ctx, r.orgID, appID, client.EnvironmentDefinitionRequest{
		Id:           data.ID.ValueString(),
		Name:         data.Name.ValueString(),
		Type:         data.Type.ValueStringPointer(),
		FromDeployId: data.FromDeployID.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create environment, got error: %s", err))
		return
	}
	switch createEnvironmentResp.StatusCode() {
	case http.StatusCreated:
		environment = createEnvironmentResp.JSON201
	case http.StatusBadRequest:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create environment, Humanitec returned bad request: %s", createEnvironmentResp.Body))
		return
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create environment, environment not found: %s", createEnvironmentResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create environment unexpected status code: %d, body: %s", createEnvironmentResp.StatusCode(), createEnvironmentResp.Body))
		return
	}

	parseEnvironmentResponse(appID, environment, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceEnvironment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *EnvironmentModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()
	id := data.ID.ValueString()

	var environment *client.EnvironmentResponse
	getEnvironmentResp, err := r.client.GetEnvironmentWithResponse(ctx, r.orgID, appID, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to get environment, got error: %s", err))
		return
	}
	switch getEnvironmentResp.StatusCode() {
	case http.StatusOK:
		environment = getEnvironmentResp.JSON200
	case http.StatusNotFound:
		resp.Diagnostics.AddWarning("Environment not found", fmt.Sprintf("The environment (%s) was deleted outside Terraform", id))
		resp.State.RemoveResource(ctx)
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to get environment, unexpected status code: %d, body: %s", getEnvironmentResp.StatusCode(), getEnvironmentResp.Body))
		return
	}

	parseEnvironmentResponse(appID, environment, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceEnvironment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *EnvironmentModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.AppID.ValueString()
	id := state.ID.ValueString()

	var environment *client.EnvironmentResponse
	updateEnvironmentResp, err := r.client.UpdateEnvironmentWithResponse(ctx, r.orgID, appID, id, client.UpdateEnvironmentJSONRequestBody{
		Name: data.Name.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update environment, got error: %s", err))
		return
	}
	switch updateEnvironmentResp.StatusCode() {
	case http.StatusOK:
		environment = updateEnvironmentResp.JSON200
	case http.StatusBadRequest:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update environment, Humanitec returned bad request: %s", updateEnvironmentResp.Body))
		return
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update environment, environment not found: %s", updateEnvironmentResp.Body))
		return
	case http.StatusPreconditionFailed:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update environment, the state of Terraform resource do not match resource in Humanitec: %s", updateEnvironmentResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update environment, unexpected status code: %d, body: %s", updateEnvironmentResp.StatusCode(), updateEnvironmentResp.Body))
		return
	}

	parseEnvironmentResponse(appID, environment, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceEnvironment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *EnvironmentModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()
	id := data.ID.ValueString()

	deleteEnvironmentResp, err := r.client.DeleteEnvironmentWithResponse(ctx, r.orgID, appID, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete environment, got error: %s", err))
		return
	}
	switch deleteEnvironmentResp.StatusCode() {
	case http.StatusNoContent, http.StatusAccepted:
		// Do nothing
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete environment, environment not found: %s", deleteEnvironmentResp.Body))
		return
	case http.StatusPreconditionFailed:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete environment, the state of Terraform resource do not match resource in Humanitec: %s", deleteEnvironmentResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete environment, unexpected status code: %d, body: %s", deleteEnvironmentResp.StatusCode(), deleteEnvironmentResp.Body))
		return
	}
}

func (r *ResourceEnvironment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	// ensure idParts elements are not empty
	for _, idPart := range idParts {
		if idPart == "" {
			resp.Diagnostics.AddError(
				"Unexpected Import Identifier",
				fmt.Sprintf("Expected import identifier with format: app_id/env_id. Got: %q", req.ID),
			)
			return
		}
	}

	if len(idParts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: app_id/env_id. Got: %q", req.ID),
		)
		return
	}
}

func parseEnvironmentResponse(appID string, res *client.EnvironmentResponse, data *EnvironmentModel) {
	var fromDeployId *string
	if res.FromDeploy != nil {
		fromDeployId = &res.FromDeploy.Id
	}

	data.FromDeployID = types.StringPointerValue(fromDeployId)
	data.AppID = types.StringValue(appID)
	data.ID = types.StringValue(res.Id)
	data.Name = types.StringValue(res.Name)
	data.Type = types.StringValue(res.Type)
}
