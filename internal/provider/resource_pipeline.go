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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourcePipeline{}
var _ resource.ResourceWithImportState = &ResourcePipeline{}

func NewResourcePipeline() resource.Resource {
	return &ResourcePipeline{}
}

// ResourceRule defines the resource implementation.
type ResourcePipeline struct {
	client *humanitec.Client
	orgID  string
}

func (r *ResourcePipeline) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline"
}

func (r *ResourcePipeline) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Pipeline defining a configurable automated process that will run one or more jobs.",

		Attributes: map[string]schema.Attribute{
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The id of the Application containing this Pipeline.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"definition": schema.StringAttribute{
				MarkdownDescription: "The YAML definition of the pipeline.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The id of the Pipeline.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Pipeline.",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The unique id of the current Pipeline Version.",
				Computed:            true,
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: "The map of key value pipeline additional information.",
				ElementType: types.StringType,
				Computed:            true,
			},
			"trigger_types": schema.SetAttribute{
				MarkdownDescription: "The list of trigger types in the current schema.",
				ElementType: types.StringType,
				Computed:            true,
			},
		},
	}
}

func (r *ResourcePipeline) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type PipelineModel struct {
	AppID        types.String `tfsdk:"app_id"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Version      types.String `tfsdk:"version"`
	Metadata     types.Map    `tfsdk:"metadata"`
	TriggerTypes types.Set    `tfsdk:"trigger_types"`
	Definition   types.String `tfsdk:"definition"`
}

func (r *ResourcePipeline) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *PipelineModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()
	definition := data.Definition.ValueString()

	var pipeline *client.PipelineResponse
	createPipelineResp, err := r.client.CreatePipelineWithBodyWithResponse(ctx, r.orgID, appID, "application/x-yaml", strings.NewReader(definition))
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create pipeline, got error: %s", err))
		return
	}
	switch createPipelineResp.StatusCode() {
	case http.StatusCreated:
		pipeline = createPipelineResp.JSON201
	case http.StatusBadRequest:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create pipeline, Humanitec returned bad request: %s", createPipelineResp.Body))
		return
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create pipeline, organization or application not found: %s", createPipelineResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create pipeline unexpected status code: %d, body: %s", createPipelineResp.StatusCode(), createPipelineResp.Body))
		return
	}

	diags := parsePipelineResponse(ctx, pipeline, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourcePipeline) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *PipelineModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()
	id := data.ID.ValueString()

	var pipeline *client.PipelineResponse
	getPipelineResp, err := r.client.GetPipelineWithResponse(ctx, r.orgID, appID, id, &client.GetPipelineParams{})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to get pipeline, got error: %s", err))
		return
	}
	switch getPipelineResp.StatusCode() {
	case http.StatusOK:
		pipeline = getPipelineResp.JSON200
	case http.StatusNotFound:
		resp.Diagnostics.AddWarning("Pipeline not found", fmt.Sprintf("The Pipeline (%s) was deleted outside Terraform", id))
		resp.State.RemoveResource(ctx)
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to get pipeline, unexpected status code: %d, body: %s", getPipelineResp.StatusCode(), getPipelineResp.Body))
		return
	}

	diags := parsePipelineResponse(ctx, pipeline, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	contentType := "application/x.humanitec-pipelines-v1.0+yaml"
	getPipelineDefinitionResp, err := r.client.GetPipelineSchemaWithResponse(ctx, r.orgID, appID, id, &client.GetPipelineSchemaParams{
		Accept: &contentType,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to get pipeline definition, got error: %s", err))
		return
	}
	switch getPipelineResp.StatusCode() {
	case http.StatusOK:
		definition := string(getPipelineDefinitionResp.Body)
		data.Definition = types.StringValue(definition)
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to get pipeline definition, unexpected status code: %d, body: %s", getPipelineDefinitionResp.StatusCode(), getPipelineDefinitionResp.Body))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourcePipeline) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *PipelineModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.AppID.ValueString()
	id := state.ID.ValueString()
	definition := data.Definition.ValueString()

	var pipeline *client.PipelineResponse
	updatePipelineResp, err := r.client.UpdatePipelineWithBodyWithResponse(ctx, r.orgID, appID, id, &client.UpdatePipelineParams{}, "application/x-yaml", strings.NewReader(definition))
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update pipeline, got error: %s", err))
		return
	}
	switch updatePipelineResp.StatusCode() {
	case http.StatusOK:
		pipeline = updatePipelineResp.JSON200
	case http.StatusBadRequest:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update pipeline, Humanitec returned bad request: %s", updatePipelineResp.Body))
		return
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update pipeline, organization or application not found: %s", updatePipelineResp.Body))
		return
	case http.StatusPreconditionFailed:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update pipeline, the state of Terraform resource do not match resource in Humanitec: %s", updatePipelineResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update pipeline, unexpected status code: %d, body: %s", updatePipelineResp.StatusCode(), updatePipelineResp.Body))
		return
	}

	diags := parsePipelineResponse(ctx, pipeline, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourcePipeline) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *PipelineModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()
	id := data.ID.ValueString()

	deletePipelineResp, err := r.client.DeletePipelineWithResponse(ctx, r.orgID, appID, id, &client.DeletePipelineParams{})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete pipeline, got error: %s", err))
		return
	}
	switch deletePipelineResp.StatusCode() {
	case http.StatusNoContent, http.StatusAccepted:
		// Do nothing
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete pipeline, pipeline not found: %s", deletePipelineResp.Body))
		return
	case http.StatusPreconditionFailed:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete pipeline, the state of Terraform resource do not match resource in Humanitec: %s", deletePipelineResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update pipeline, unexpected status code: %d, body: %s", deletePipelineResp.StatusCode(), deletePipelineResp.Body))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourcePipeline) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	// ensure idParts elements are not empty
	for _, idPart := range idParts {
		if idPart == "" {
			resp.Diagnostics.AddError(
				"Unexpected Import Identifier",
				fmt.Sprintf("Expected import identifier with format: app_id/pipeline_id. Got: %q", req.ID),
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
			fmt.Sprintf("Expected import identifier with format: app_id/pipeline_id. Got: %q", req.ID),
		)
		return
	}
}

func parsePipelineResponse(ctx context.Context, res *client.PipelineResponse, data *PipelineModel) diag.Diagnostics {
	totalDiags := diag.Diagnostics{}

	data.AppID = types.StringValue(res.AppId)
	data.ID = types.StringValue(res.Id)
	data.Name = types.StringValue(res.Name)
	data.Version = types.StringValue(res.Version)

	triggers, diags := types.SetValueFrom(ctx, types.StringType, res.TriggerTypes)
	totalDiags.Append(diags...)
	data.TriggerTypes = triggers

	if res.Metadata == nil {
		res.Metadata = &map[string]string{}
	}

	metadata, diags := types.MapValueFrom(ctx, types.StringType, *res.Metadata)
	totalDiags.Append(diags...)
	data.Metadata = metadata

	return totalDiags
}
