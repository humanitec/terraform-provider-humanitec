package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceWorkloadProfile{}
var _ resource.ResourceWithImportState = &ResourceWorkloadProfile{}

func NewResourceWorkloadProfile() resource.Resource {
	return &ResourceWorkloadProfile{}
}

// ResourceRule defines the resource implementation.
type ResourceWorkloadProfile struct {
	client *humanitec.Client
	orgID  string
}

func (r *ResourceWorkloadProfile) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workload_profile"
}

func (r *ResourceWorkloadProfile) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Workload Profile",

		Attributes: map[string]schema.Attribute{
			"deprecation_message": schema.StringAttribute{
				MarkdownDescription: "A not-empty string indicates that the workload profile is deprecated.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Describes the workload profile",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Workload Profile ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"spec_definition": schema.StringAttribute{
				MarkdownDescription: "Workload spec definition",
				Required:            true,
			},
			"version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Version identifier. The version must be unique, but the API doesn't not enforce any ordering. Currently workloads will always use the latest update.",
			},
			"workload_profile_chart": schema.SingleNestedAttribute{
				MarkdownDescription: "References a workload profile chart.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Workload Profile Chart ID",
					},
					"version": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Version",
					},
				},
			},
		},
	}
}

func (r *ResourceWorkloadProfile) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type WorkloadProfileChartReferenceModel struct {
	ID      types.String `tfsdk:"id"`
	Version types.String `tfsdk:"version"`
}

type WorkloadProfileModel struct {
	ID                   types.String                        `tfsdk:"id"`
	Description          types.String                        `tfsdk:"description"`
	DeprecationMessage   types.String                        `tfsdk:"deprecation_message"`
	SpecDefinition       types.String                        `tfsdk:"spec_definition"`
	Version              types.String                        `tfsdk:"version"`
	WorkloadProfileChart *WorkloadProfileChartReferenceModel `tfsdk:"workload_profile_chart"`
}

func (r *ResourceWorkloadProfile) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *WorkloadProfileModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	specDefinition, diags := toWorkloadProfileSpecDefinition(data.SpecDefinition)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	createRes, err := r.client.CreateWorkloadProfileWithResponse(ctx, r.orgID, client.CreateWorkloadProfileJSONRequestBody{
		DeprecationMessage: data.DeprecationMessage.ValueStringPointer(),
		Description:        data.Description.ValueStringPointer(),
		Id:                 data.ID.ValueString(),
		SpecDefinition:     specDefinition,
		Version:            data.Version.ValueStringPointer(),
		WorkloadProfileChart: client.WorkloadProfileChartReference{
			Id:      data.WorkloadProfileChart.ID.ValueString(),
			Version: data.WorkloadProfileChart.Version.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create workload profile, got error: %s", err))
		return
	}
	if createRes.StatusCode() != 201 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create workload profile, unexpected status code: %d, body: %s", createRes.StatusCode(), createRes.Body))
		return
	}

	parseWorkloadProfileResponse(createRes.JSON201, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceWorkloadProfile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *WorkloadProfileModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()

	getRes, err := r.client.GetWorkloadProfileWithResponse(ctx, r.orgID, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to get workload profile, got error: %s", err))
		return
	}
	if getRes.StatusCode() == 404 {
		resp.Diagnostics.AddWarning("Workload Profile not found", fmt.Sprintf("The Workload Profile (%s) was deleted outside Terraform", id))
		resp.State.RemoveResource(ctx)
		return
	}
	if getRes.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to get workload profile, unexpected status code: %d, body: %s", getRes.StatusCode(), getRes.Body))
		return
	}

	parseWorkloadProfileResponse(getRes.JSON200, data)

	tflog.Error(ctx, "WorkloadProfileModel: %v", map[string]interface{}{
		"data": data,
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceWorkloadProfile) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *WorkloadProfileModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()

	specDefinition, diags := toWorkloadProfileSpecDefinition(data.SpecDefinition)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateRes, err := r.client.UpdateWorkloadProfileWithResponse(ctx, r.orgID, id, client.UpdateWorkloadProfileJSONRequestBody{
		DeprecationMessage: data.DeprecationMessage.ValueStringPointer(),
		Description:        data.Description.ValueStringPointer(),
		SpecDefinition:     specDefinition,
		Version:            data.Version.ValueStringPointer(),
		WorkloadProfileChart: client.WorkloadProfileChartReference{
			Id:      data.WorkloadProfileChart.ID.ValueString(),
			Version: data.WorkloadProfileChart.Version.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update workload profile, got error: %s", err))
		return
	}

	if updateRes.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update workload profile, unexpected status code: %d, body: %s", updateRes.StatusCode(), updateRes.Body))
		return
	}

	parseWorkloadProfileResponse(updateRes.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceWorkloadProfile) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *WorkloadProfileModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()

	deleteRes, err := r.client.DeleteWorkloadProfileWithResponse(ctx, r.orgID, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete webhook, got error: %s", err))
		return
	}

	if deleteRes.StatusCode() != 204 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete webhook, unexpected status code: %d, body: %s", deleteRes.StatusCode(), deleteRes.Body))
		return
	}
}

func (r *ResourceWorkloadProfile) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func toWorkloadProfileSpecDefinition(modelSpecDefinition types.String) (client.WorkloadProfileSpecDefinition, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	specDefinition := client.WorkloadProfileSpecDefinition{}
	if err := json.Unmarshal([]byte(modelSpecDefinition.ValueString()), &specDefinition); err != nil {
		diags.AddError(HUM_INPUT_ERR, fmt.Sprintf("Unable to unmarshal spec definition, got error: %s", err))
	}

	return specDefinition, diags
}

func parseWorkloadProfileResponse(cv *client.WorkloadProfileResponse, data *WorkloadProfileModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	data.DeprecationMessage = types.StringPointerValue(cv.DeprecationMessage)
	data.Description = types.StringValue(cv.Description)
	data.ID = types.StringValue(cv.Id)

	specDefinition, err := json.Marshal(cv.SpecDefinition)
	if err != nil {
		diags.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to marshal spec definition, got error: %s", err))
	}

	data.SpecDefinition = types.StringValue(string(specDefinition))
	data.Version = types.StringValue(cv.Version)
	data.WorkloadProfileChart = &WorkloadProfileChartReferenceModel{
		ID:      types.StringValue(cv.WorkloadProfileChart.Id),
		Version: types.StringValue(cv.WorkloadProfileChart.Version),
	}

	return diags
}
