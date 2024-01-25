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

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceResourceDriver{}
var _ resource.ResourceWithImportState = &ResourceResourceDriver{}

func NewResourceResourceDriver() resource.Resource {
	return &ResourceResourceDriver{}
}

// ResourceResourceDriver defines the resource implementation.
type ResourceResourceDriver struct {
	client *humanitec.Client
	orgId  string
}

// ResourceDriverModel describes the app data model.
type ResourceDriverModel struct {
	ID           types.String   `tfsdk:"id"`
	AccountTypes []types.String `tfsdk:"account_types"`
	InputsSchema types.String   `tfsdk:"inputs_schema"`
	Target       types.String   `tfsdk:"target"`
	Template     types.String   `tfsdk:"template"`
	Type         types.String   `tfsdk:"type"`
}

func (r *ResourceResourceDriver) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_driver"
}

func (r *ResourceResourceDriver) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `DriverDefinition describes the resource driver.

Resource Drivers are code that fulfils the Humanitec Resource Driver Interface.
This interface allows for certain actions to be performed on resources such as creation and destruction.
Humanitec provides numerous Resource Drivers “out of the box”.
It is also possible to use 3rd party Resource Drivers or write your own.`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID for this driver. Is used as `driver_type`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account_types": schema.ListAttribute{
				MarkdownDescription: "List of resources accounts types supported by the driver",
				Required:            true,
				ElementType:         types.StringType,
			},
			"inputs_schema": schema.StringAttribute{
				MarkdownDescription: "A JSON Schema specifying the driver-specific input parameters.",
				Required:            true,
			},
			"target": schema.StringAttribute{
				MarkdownDescription: "The prefix where the driver resides or, if the driver is a virtual driver, the reference to an existing driver using the `driver://` schema of the format `driver://{orgId}/{driverId}`. Only members of the organization the driver belongs to can see `target`.",
				Required:            true,
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "If the driver is a virtual driver, template defines a Go template that converts the driver inputs supplied in the resource definition into the driver inputs for the target driver.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of resource produced by this driver",
				Required:            true,
			},
		},
	}
}

func (r *ResourceResourceDriver) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func parseResourceDriverResponse(res *client.DriverDefinitionResponse, data *ResourceDriverModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(res.Id)
	data.AccountTypes = []types.String{}
	for _, v := range res.AccountTypes {
		data.AccountTypes = append(data.AccountTypes, types.StringValue(v))
	}

	bi, err := json.Marshal(res.InputsSchema)
	if err != nil {
		diags.AddError(HUM_API_ERR, fmt.Sprintf("Failed to marshal driver input_schema: %s", err.Error()))
	}
	data.InputsSchema = types.StringValue(string(bi))

	if res.Template != nil {
		bt, err := json.Marshal(res.Template)
		if err != nil {
			diags.AddError(HUM_API_ERR, fmt.Sprintf("Failed to marshal driver template: %s", err.Error()))
		}
		data.Template = types.StringValue(string(bt))
	} else {
		data.Template = types.StringNull()
	}

	data.Target = types.StringPointerValue(res.Target)

	data.Type = types.StringValue(res.Type)

	return diags
}

func (r *ResourceResourceDriver) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ResourceDriverModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()

	var inputsSchema map[string]interface{}
	if err := json.Unmarshal([]byte(data.InputsSchema.ValueString()), &inputsSchema); err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(HUM_API_ERR, fmt.Sprintf("Failed to unmarshal driver input_schema: %s", err.Error())))
		return
	}

	var template *interface{}

	if data.Template.ValueStringPointer() != nil {
		if err := json.Unmarshal([]byte(data.Template.ValueString()), &template); err != nil {
			resp.Diagnostics.Append(diag.NewErrorDiagnostic(HUM_API_ERR, fmt.Sprintf("Failed to unmarshal driver template: %s", err.Error())))
			return
		}
	}

	accountTypes := []string{}
	for _, v := range data.AccountTypes {
		accountTypes = append(accountTypes, v.ValueString())
	}

	httpResp, err := r.client.CreateResourceDriverWithResponse(ctx, r.orgId, client.CreateDriverRequestRequest{
		Id:           id,
		AccountTypes: accountTypes,
		InputsSchema: inputsSchema,
		Target:       data.Target.ValueString(),
		Template:     template,
		Type:         data.Type.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create resource driver, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create resource driver, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseResourceDriverResponse(httpResp.JSON200, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceResourceDriver) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ResourceDriverModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.GetResourceDriverWithResponse(ctx, r.orgId, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read resource driver, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read resource driver, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseResourceDriverResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceResourceDriver) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ResourceDriverModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	var inputsSchema map[string]interface{}
	if err := json.Unmarshal([]byte(data.InputsSchema.ValueString()), &inputsSchema); err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(HUM_API_ERR, fmt.Sprintf("Failed to unmarshal driver input_schema: %s", err.Error())))
		return
	}

	accountTypes := []string{}
	for _, v := range data.AccountTypes {
		accountTypes = append(accountTypes, v.ValueString())
	}

	var template *interface{}

	if data.Template.ValueStringPointer() != nil {
		if err := json.Unmarshal([]byte(data.Template.ValueString()), &template); err != nil {
			resp.Diagnostics.Append(diag.NewErrorDiagnostic(HUM_API_ERR, fmt.Sprintf("Failed to unmarshal driver template: %s", err.Error())))
			return
		}
	}

	httpResp, err := r.client.UpdateResourceDriverWithResponse(ctx, r.orgId, id, client.UpdateDriverRequestRequest{
		AccountTypes: accountTypes,
		InputsSchema: inputsSchema,
		Target:       data.Target.ValueString(),
		Template:     template,
		Type:         data.Type.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update value, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseResourceDriverResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceResourceDriver) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ResourceDriverModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.DeleteResourceDriverWithResponse(ctx, r.orgId, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete resource driver, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 204 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete resource driver, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}
}

func (r *ResourceResourceDriver) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
