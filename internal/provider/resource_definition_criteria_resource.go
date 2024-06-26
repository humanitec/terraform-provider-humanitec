package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceDefinitionCriteriaResource{}
var _ resource.ResourceWithImportState = &ResourceDefinitionCriteriaResource{}

var defaultResourceDefinitionCriteriaDeleteTimeout = 10 * time.Minute

func NewResourceDefinitionCriteriaResource() resource.Resource {
	return &ResourceDefinitionCriteriaResource{}
}

// ResourceDefinitionCriteriaResource defines the resource implementation.
type ResourceDefinitionCriteriaResource struct {
	data *HumanitecData
}

func (r *ResourceDefinitionCriteriaResource) client() *humanitec.Client {
	return r.data.Client
}

func (r *ResourceDefinitionCriteriaResource) orgId() string {
	return r.data.OrgID
}

// ResourceDefinitionCriteriaResourceModel describes the resource data model.
type ResourceDefinitionCriteriaResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ResourceDefinitionID types.String `tfsdk:"resource_definition_id"`
	AppID                types.String `tfsdk:"app_id"`
	EnvID                types.String `tfsdk:"env_id"`
	EnvType              types.String `tfsdk:"env_type"`
	ResID                types.String `tfsdk:"res_id"`
	Class                types.String `tfsdk:"class"`

	ForceDelete types.Bool     `tfsdk:"force_delete"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

func (r *ResourceDefinitionCriteriaResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_definition_criteria"
}

func (r *ResourceDefinitionCriteriaResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Visit the [docs](https://docs.humanitec.com/reference/concepts/resources/definitions) to learn more about resource definitions.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Matching Criteria ID",
				Computed:            true,
			},
			"resource_definition_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The Resource Definition ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Application that the Resources should belong to.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"class": schema.StringAttribute{
				MarkdownDescription: "The class of the Resource in the Deployment Set. Can not be empty, if is not defined, set to `default`.",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Default: stringdefault.StaticString("default"),
			},
			"env_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Environment that the Resources should belong to. If `env_type` is also set, it must match the Type of the Environment for the Criteria to match.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"env_type": schema.StringAttribute{
				MarkdownDescription: "The Type of the Environment that the Resources should belong to. If `env_id` is also set, it must have an Environment Type that matches this parameter for the Criteria to match.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"res_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Resource in the Deployment Set. The ID is normally a `.` separated path to the definition in the set, e.g. `modules.my-module.externals.my-database`.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"force_delete": schema.BoolAttribute{
				MarkdownDescription: "If set to `true`, the Matching Criteria is deleted immediately, even if this action affects existing Active Resources.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Delete: true,
			}),
		},
	}
}

func (r *ResourceDefinitionCriteriaResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*HumanitecData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.data = data
}

func parseResourceDefinitionCriteriaResponse(res *client.MatchingCriteriaResponse, data *ResourceDefinitionCriteriaResourceModel) {
	data.ID = types.StringValue(res.Id)
	data.AppID = parseOptionalString(res.AppId)
	data.EnvID = parseOptionalString(res.EnvId)
	data.EnvType = parseOptionalString(res.EnvType)
	data.ResID = parseOptionalString(res.ResId)
	data.Class = types.StringValue(res.Class)
}

func (r *ResourceDefinitionCriteriaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ResourceDefinitionCriteriaResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client().CreateResourceDefinitionCriteriaWithResponse(ctx, r.orgId(), data.ResourceDefinitionID.ValueString(), client.CreateResourceDefinitionCriteriaJSONRequestBody{
		AppId:   data.AppID.ValueStringPointer(),
		EnvId:   data.EnvID.ValueStringPointer(),
		EnvType: data.EnvType.ValueStringPointer(),
		ResId:   data.ResID.ValueStringPointer(),
		Class:   data.Class.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create resource definition criteria, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create resource definition criteria, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseResourceDefinitionCriteriaResponse(httpResp.JSON200, data)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceDefinitionCriteriaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ResourceDefinitionCriteriaResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client().GetResourceDefinitionWithResponse(ctx, r.orgId(), data.ResourceDefinitionID.ValueString(), &client.GetResourceDefinitionParams{Deleted: toPtr(false)})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read resource definition, got error: %s", err))
		return
	}

	if httpResp.StatusCode() == 404 {
		resp.Diagnostics.AddWarning("Resource definition not found", fmt.Sprintf("The resource definition (%s) was deleted outside Terraform", data.ResourceDefinitionID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read resource definition, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddWarning("Resource definition criteria not found", fmt.Sprintf("The resource definition criteria (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	res := *httpResp.JSON200

	if res.Criteria == nil {
		resp.Diagnostics.AddWarning("Resource definition criteria not found", fmt.Sprintf("The resource definition criteria (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	criteria := *res.Criteria

	found := false
	for _, c := range criteria {
		if c.Id == data.ID.ValueString() {
			found = true
			parseResourceDefinitionCriteriaResponse(&c, data)
		}
	}

	if !found {
		resp.Diagnostics.AddWarning("Resource definition criteria not found", fmt.Sprintf("The resource definition criteria (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceDefinitionCriteriaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *ResourceDefinitionCriteriaResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update client-only attributes
	state.ForceDelete = data.ForceDelete
	state.Timeouts = data.Timeouts

	// All other attributes require a replacement, so no API calls here

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceDefinitionCriteriaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ResourceDefinitionCriteriaResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, defaultResourceDefinitionCriteriaDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	force := data.ForceDelete.ValueBool()

	err := retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		httpResp, err := r.client().DeleteResourceDefinitionCriteriaWithResponse(ctx, r.orgId(), data.ResourceDefinitionID.ValueString(), data.ID.ValueString(), &client.DeleteResourceDefinitionCriteriaParams{
			Force: &force,
		})
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if httpResp.StatusCode() == 409 {
			return retry.RetryableError(fmt.Errorf("resource definition criteria has still active resources, status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		if httpResp.StatusCode() != 204 {
			return retry.NonRetryableError(fmt.Errorf("unable to delete definition criteria, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete definition criteria, got error: %s", err))
		return
	}
}

func (r *ResourceDefinitionCriteriaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: resource_definition_id/id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_definition_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
