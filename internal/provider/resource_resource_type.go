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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceResourceType{}
var _ resource.ResourceWithImportState = &ResourceResourceType{}

func NewResourceResourceType() resource.Resource {
	return &ResourceResourceType{}
}

// ResourceResourceType defines the resource implementation.
type ResourceResourceType struct {
	client *humanitec.Client
	orgId  string
}

// ResourceTypeModel describes the app data model.
type ResourceTypeModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Category      types.String `tfsdk:"category"`
	Use           types.String `tfsdk:"use"`
	InputsSchema  types.String `tfsdk:"inputs_schema"`
	OutputsSchema types.String `tfsdk:"outputs_schema"`
}

func (r *ResourceResourceType) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_type"
}

func (r *ResourceResourceType) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource Types are templates for resources, which are used to drive Humanitec's resource management engine.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the resource type. It should start with the Humanitec Organization ID followed by '/'.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name.",
				Optional:            true,
			},
			"category": schema.StringAttribute{
				MarkdownDescription: "Category name (used to group similar resources on the UI).",
				Optional:            true,
			},
			"use": schema.StringAttribute{
				MarkdownDescription: "Kind of dependency between resource of this type and a workload. It should be one of: `direct`, `indirect`.",
				Required:            true,
			},
			"inputs_schema": schema.StringAttribute{
				MarkdownDescription: "A JSON Schema specifying the type-specific parameters for the driver (input).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("{}"),
			},
			"outputs_schema": schema.StringAttribute{
				MarkdownDescription: "A JSON Schema specifying the type-specific data passed to the deployment (output).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("{}"),
			},
		},
	}
}

func (r *ResourceResourceType) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ResourceResourceType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ResourceTypeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var inputsSchema map[string]interface{}
	if err := json.Unmarshal([]byte(data.InputsSchema.ValueString()), &inputsSchema); err != nil {
		resp.Diagnostics.AddError(HUM_INPUT_ERR, fmt.Sprintf("Failed to unmarshal inputs_schema: %s", err.Error()))
		return
	}

	var outputsSchema map[string]interface{}
	if err := json.Unmarshal([]byte(data.OutputsSchema.ValueString()), &outputsSchema); err != nil {
		resp.Diagnostics.AddError(HUM_INPUT_ERR, fmt.Sprintf("Failed to unmarshal outputs_schema: %s", err.Error()))
		return
	}

	httpResp, err := r.client.CreateResourceTypeWithResponse(ctx, r.orgId, client.ResourceTypeRequest{
		Type:          data.ID.ValueString(),
		Name:          data.Name.ValueStringPointer(),
		Category:      data.Category.ValueStringPointer(),
		Use:           data.Use.ValueString(),
		InputsSchema:  &inputsSchema,
		OutputsSchema: &outputsSchema,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create resource type, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create resource type, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	diags := parseResourceTypeResponse(httpResp.JSON200, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceResourceType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ResourceTypeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.ListResourceTypesWithResponse(ctx, r.orgId)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read resource type, got error: %s", err))
		return
	}

	if httpResp.StatusCode() == 404 {
		resp.Diagnostics.AddWarning("Resource type not found", fmt.Sprintf("The resource type (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read resource type, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	found := false
	if httpResp.JSON200 != nil {
		for _, res := range *httpResp.JSON200 {
			if res.Type == data.ID.ValueString() {
				diags := parseResourceTypeResponse(&res, data)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				found = true
				break
			}
		}
	}

	if !found {
		resp.Diagnostics.AddWarning("Resource type not found", fmt.Sprintf("The resource type (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceResourceType) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *ResourceTypeModel

	// Read Terraform plan and state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var inputsSchema map[string]interface{}
	if err := json.Unmarshal([]byte(plan.InputsSchema.ValueString()), &inputsSchema); err != nil {
		resp.Diagnostics.AddError(HUM_INPUT_ERR, fmt.Sprintf("Failed to unmarshal inputs_schema: %s", err.Error()))
		return
	}

	var outputsSchema map[string]interface{}
	if err := json.Unmarshal([]byte(plan.OutputsSchema.ValueString()), &outputsSchema); err != nil {
		resp.Diagnostics.AddError(HUM_INPUT_ERR, fmt.Sprintf("Failed to unmarshal outputs_schema: %s", err.Error()))
		return
	}

	httpResp, err := r.client.UpdateResourceTypeWithResponse(ctx, r.orgId, plan.ID.ValueString(), client.UpdateResourceTypeRequestRequest{
		Name:          plan.Name.ValueStringPointer(),
		Category:      plan.Category.ValueStringPointer(),
		Use:           plan.Use.ValueString(),
		InputsSchema:  &inputsSchema,
		OutputsSchema: &outputsSchema,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update resource type, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update resource type, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	diags := parseResourceTypeResponse(httpResp.JSON200, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ResourceResourceType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ResourceTypeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.DeleteResourceTypeWithResponse(ctx, r.orgId, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete resource type, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 204 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete resource type, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}
}

func (r *ResourceResourceType) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func parseResourceTypeResponse(res *client.ResourceTypeResponse, data *ResourceTypeModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(res.Type)
	data.Name = types.StringValue(res.Name)
	data.Category = types.StringValue(res.Category)
	data.Use = types.StringValue(res.Use)

	if res.InputsSchema != nil {
		b, err := json.Marshal(res.InputsSchema)
		if err != nil {
			diags.AddError(HUM_API_ERR, fmt.Sprintf("Failed to marshal inputs_schema: %s", err.Error()))
		}
		data.InputsSchema = types.StringValue(string(b))
	}

	if res.OutputsSchema != nil {
		b, err := json.Marshal(res.OutputsSchema)
		if err != nil {
			diags.AddError(HUM_API_ERR, fmt.Sprintf("Failed to marshal outputs_schema: %s", err.Error()))
		}
		data.OutputsSchema = types.StringValue(string(b))
	}

	return diags
}
