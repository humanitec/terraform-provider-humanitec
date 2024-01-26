package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourcePipelineCriteria{}
var _ resource.ResourceWithImportState = &ResourcePipelineCriteria{}

func NewResourcePipelineCriteria() resource.Resource {
	return &ResourcePipelineCriteria{}
}

// ResourcePipelineCriteria defines the resource implementation.
type ResourcePipelineCriteria struct {
	client *humanitec.Client
	orgID  string
}

func (r *ResourcePipelineCriteria) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_criteria"
}

func (r *ResourcePipelineCriteria) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Pipeline criteria link Pipelines to applicable triggers in the application. The only
supported trigger type today is "deployment_request" which specifies that the Pipeline should be used for deployments
in any environment which matches the criteria.
`,
		Attributes: map[string]schema.Attribute{
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The id of the Application containing the Pipeline.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pipeline_id": schema.StringAttribute{
				MarkdownDescription: "The id of the Pipeline.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pipeline_name": schema.StringAttribute{
				MarkdownDescription: "The name of the Pipeline.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The id of the Pipeline Criteria.",
				Computed:            true,
			},
			"deployment_request": schema.SingleNestedAttribute{
				MarkdownDescription: "The criteria required to match a deployment request.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"app_id": schema.StringAttribute{
						MarkdownDescription: "The Application id for this criteria to match.",
						Computed:            true,
					},
					"env_type": schema.StringAttribute{
						MarkdownDescription: "The environment type for this criteria to match.",
						Optional:            true,
						Computed:            true,
					},
					"env_id": schema.StringAttribute{
						MarkdownDescription: "The environment id for this criteria to match.",
						Optional:            true,
						Computed:            true,
					},
					"deployment_type": schema.StringAttribute{
						MarkdownDescription: "The deployment type for this criteria to match ('deploy' or 're-deploy').",
						Optional:            true,
						Computed:            true,
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ResourcePipelineCriteria) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type pipelineCriteriaDeploymentRequestModel struct {
	EnvType        types.String `tfsdk:"env_type"`
	AppID          types.String `tfsdk:"app_id"`
	EnvId          types.String `tfsdk:"env_id"`
	DeploymentType types.String `tfsdk:"deployment_type"`
}

// pipelineCriteriaModel is used to deserialize the plan or state in order to access its attributes
type pipelineCriteriaModel struct {
	AppID             types.String                            `tfsdk:"app_id"`
	PipelineId        types.String                            `tfsdk:"pipeline_id"`
	PipelineName      types.String                            `tfsdk:"pipeline_name"`
	Id                types.String                            `tfsdk:"id"`
	DeploymentRequest *pipelineCriteriaDeploymentRequestModel `tfsdk:"deployment_request"`
}

func (pcm *pipelineCriteriaModel) updateFromContent(res *client.PipelineCriteria) diag.Diagnostics {
	totalDiags := diag.Diagnostics{}
	drc, err := res.AsPipelineDeploymentRequestCriteria()
	if err != nil {
		totalDiags.AddError(HUM_PROVIDER_ERR, "provider does not support trigger type "+res.Trigger)
		return totalDiags
	}
	pcm.Id = types.StringValue(drc.Id)
	pcm.AppID = types.StringPointerValue(drc.AppId)
	pcm.PipelineId = types.StringValue(drc.PipelineId)
	pcm.PipelineName = types.StringValue(drc.PipelineName)
	pcm.DeploymentRequest = &pipelineCriteriaDeploymentRequestModel{
		EnvType:        types.StringPointerValue(drc.EnvType),
		AppID:          types.StringPointerValue(drc.AppId),
		EnvId:          types.StringPointerValue(drc.EnvId),
		DeploymentType: types.StringPointerValue(drc.DeploymentType),
	}
	return totalDiags
}

func (r *ResourcePipelineCriteria) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *pipelineCriteriaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestBody := client.CreatePipelineCriteriaJSONRequestBody{}
	request := client.PipelineDeploymentRequestCriteriaCreateBody{
		AppId: data.AppID.ValueStringPointer(),
	}
	if v := data.DeploymentRequest.EnvType.ValueString(); v != "" {
		request.EnvType = &v
	}
	if v := data.DeploymentRequest.EnvId.ValueString(); v != "" {
		request.EnvId = &v
	}
	if v := data.DeploymentRequest.DeploymentType.ValueString(); v != "" {
		request.DeploymentType = &v
	}
	_ = requestBody.FromPipelineDeploymentRequestCriteriaCreateBody(request)
	clientResp, err := r.client.CreatePipelineCriteriaWithResponse(ctx, r.orgID, data.AppID.ValueString(), data.PipelineId.ValueString(), requestBody)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create pipeline criteria, got error: %s", err))
		return
	}
	switch clientResp.StatusCode() {
	case http.StatusCreated:
		diags := data.updateFromContent(clientResp.JSON201)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	case http.StatusBadRequest:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create pipeline criteria, Humanitec returned bad request: %s", clientResp.Body))
		return
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create pipeline criteria, organization or application not found: %s", clientResp.Body))
		return
	case http.StatusConflict:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create pipeline criteria due to a conflicts: %s", clientResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Received unexpected status code when creating pipeline criteria: %d, body: %s", clientResp.StatusCode(), clientResp.Body))
		return
	}
}

func (r *ResourcePipelineCriteria) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var data *pipelineCriteriaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientResp, err := r.client.GetPipelineCriteriaWithResponse(ctx, r.orgID, data.AppID.ValueString(), data.PipelineId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to get pipeline criteria, got error: %s", err))
		return
	}
	switch clientResp.StatusCode() {
	case http.StatusOK:
		diags := data.updateFromContent(clientResp.JSON200)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to get pipeline criteria, organization or application not found: %s", clientResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Received unexpected status code when reading pipeline criteria: %d, body: %s", clientResp.StatusCode(), clientResp.Body))
		return
	}
}

func (r *ResourcePipelineCriteria) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// you can't update criteria in place, all updates should be done with a replacement
	resp.Diagnostics.AddError(HUM_CLIENT_ERR, "Unable to update pipeline criteria")
	return
}

func (r *ResourcePipelineCriteria) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *pipelineCriteriaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	clientResp, err := r.client.DeletePipelineCriteriaWithResponse(ctx, r.orgID, data.AppID.ValueString(), data.PipelineId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete pipeline criteria, got error: %s", err))
		return
	}
	switch clientResp.StatusCode() {
	case http.StatusNoContent:
		return
	case http.StatusBadRequest:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete pipeline criteria, Humanitec returned bad request: %s", clientResp.Body))
		return
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete missing pipeline criteria: %s", clientResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Received unexpected status code when deleting pipeline criteria: %d, body: %s", clientResp.StatusCode(), clientResp.Body))
		return
	}
}

func (r *ResourcePipelineCriteria) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 3 {
		resp.Diagnostics.AddError("Unexpected Import Identifier", "expected a 3 part import id like <app_id>/<pipeline_id>/<criteria_id>")
		return
	}
	appId, pipelineId, criteriaId := idParts[0], idParts[1], idParts[2]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), appId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pipeline_id"), pipelineId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), criteriaId)...)
	return
}
