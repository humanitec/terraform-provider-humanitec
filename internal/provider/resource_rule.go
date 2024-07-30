package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceRule{}
var _ resource.ResourceWithImportState = &ResourceRule{}

func NewResourceRule() resource.Resource {
	return &ResourceRule{}
}

// ResourceRule defines the resource implementation.
type ResourceRule struct {
	client *humanitec.Client
	orgId  string
}

// RuleModel describes the app data model.
type RuleModel struct {
	ID    types.String `tfsdk:"id"`
	AppID types.String `tfsdk:"app_id"`
	EnvID types.String `tfsdk:"env_id"`

	Active                 types.Bool     `tfsdk:"active"`
	ArtefactsFilter        []types.String `tfsdk:"artefacts_filter"`
	ExcludeArtefactsFilter types.Bool     `tfsdk:"exclude_artefacts_filter"`
	MatchRef               types.String   `tfsdk:"match_ref"`
	Type                   types.String   `tfsdk:"type"`
}

func (r *ResourceRule) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rule"
}

func (r *ResourceRule) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An Automation Rule defining how and when artefacts in an environment should be updated.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Rule.",
				Computed:            true,
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Application that the Rule should belong to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"env_id": schema.StringAttribute{
				MarkdownDescription: "The Environment ID.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the rule will be processed or not.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"artefacts_filter": schema.ListAttribute{
				MarkdownDescription: "A list of artefact names to be processed by the rule. If the array is empty, it implies include all. If `exclude_artefacts_filter` is true, this list describes the artefacts to exclude.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"exclude_artefacts_filter": schema.BoolAttribute{
				MarkdownDescription: "Whether the artefacts specified in `artefacts_filter` should be excluded (`true`) or included (`false`) in the automation rule.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"match_ref": schema.StringAttribute{
				MarkdownDescription: "A regular expression applied to the ref of a new artefact version. Defaults to match all if omitted or empty.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Specifies the type of event. Currently, only updates to either branches or tags are supported. Must be `update`.",
				Required:            true,
			},
		},
	}
}

func (r *ResourceRule) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func parseAutomationRuleResponse(res *client.AutomationRuleResponse, data *RuleModel) {
	data.ID = types.StringValue(res.Id)
	data.Active = types.BoolValue(res.Active)

	data.ArtefactsFilter = []types.String{}

	for _, v := range res.ArtefactsFilter {
		data.ArtefactsFilter = append(data.ArtefactsFilter, types.StringValue(v))
	}

	data.ExcludeArtefactsFilter = types.BoolValue(res.ExcludeArtefactsFilter)
	data.MatchRef = types.StringValue(res.MatchRef)
	data.Type = types.StringValue(res.Type)
}

func toAutomationRuleRequest(data *RuleModel) (*client.AutomationRuleRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	artefactsFilter := []string{}
	for _, f := range data.ArtefactsFilter {
		artefactsFilter = append(artefactsFilter, f.ValueString())
	}

	return &client.AutomationRuleRequest{
		Active:                 data.Active.ValueBoolPointer(),
		ArtefactsFilter:        &artefactsFilter,
		ExcludeArtefactsFilter: data.ExcludeArtefactsFilter.ValueBoolPointer(),
		MatchRef:               data.MatchRef.ValueStringPointer(),
		Type:                   data.Type.ValueString(),
	}, diags
}

func (r *ResourceRule) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *RuleModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()
	envID := data.EnvID.ValueString()

	httpBody, diags := toAutomationRuleRequest(data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.CreateAutomationRuleWithResponse(ctx, r.orgId, appID, envID, *httpBody)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create rule, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 201 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create rule, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseAutomationRuleResponse(httpResp.JSON201, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceRule) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *RuleModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()
	envID := data.EnvID.ValueString()
	id := data.ID.ValueString()

	httpResp, err := r.client.GetAutomationRuleWithResponse(ctx, r.orgId, appID, envID, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read rule, got error: %s", err))
		return
	}
	if httpResp.StatusCode() == 404 {
		resp.Diagnostics.AddWarning("Rule not found", fmt.Sprintf("The rule (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read rule, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseAutomationRuleResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceRule) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *RuleModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.AppID.ValueString()
	envID := state.EnvID.ValueString()
	id := state.ID.ValueString()

	httpBody, diags := toAutomationRuleRequest(data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.UpdateAutomationRuleWithResponse(ctx, r.orgId, appID, envID, id, *httpBody)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update rule, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update rule, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseAutomationRuleResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceRule) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *RuleModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()
	envID := data.EnvID.ValueString()
	id := data.ID.ValueString()

	httpResp, err := r.client.DeleteAutomationRuleWithResponse(ctx, r.orgId, appID, envID, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete rule, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 204 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete rule, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}
}

func (r *ResourceRule) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	// ensure idParts elements are not empty
	for _, idPart := range idParts {
		if idPart == "" {
			resp.Diagnostics.AddError(
				"Unexpected Import Identifier",
				fmt.Sprintf("Expected import identifier with format: app_id/env_id/rule_id. Got: %q", req.ID),
			)
			return
		}
	}

	if len(idParts) == 3 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("env_id"), idParts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[2])...)
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: app_id/env_id/rule_id. Got: %q", req.ID),
		)
		return
	}
}
