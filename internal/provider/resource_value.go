package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceValue{}
var _ resource.ResourceWithImportState = &ResourceValue{}

func NewResourceValue() resource.Resource {
	return &ResourceValue{}
}

// ResourceValue defines the resource implementation.
type ResourceValue struct {
	client *humanitec.Client
	orgId  string
}

// ValueModel describes the app data model.
type ValueModel struct {
	ID    types.String `tfsdk:"id"`
	AppID types.String `tfsdk:"app_id"`
	EnvID types.String `tfsdk:"env_id"`

	Key         types.String `tfsdk:"key"`
	Description types.String `tfsdk:"description"`
	IsSecret    types.Bool   `tfsdk:"is_secret"`
	Value       types.String `tfsdk:"value"`
}

func (r *ResourceValue) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_value"
}

func (r *ResourceValue) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Shared Values can be used to manage variables and configuration that might vary between environments. They are also the way that secrets can be stored securely.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Application that the Shared Value should belong to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"env_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Environment that the Shared Value should belong to.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "The unique key by which the Shared Value can be referenced.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A Human friendly description of what the Shared Value is.",
				Required:            true,
			},
			"is_secret": schema.BoolAttribute{
				MarkdownDescription: "Specified that the Shared Value contains a secret.",
				Required:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "The value that will be stored.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *ResourceValue) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func parseValueResponse(res *client.ValueResponse, data *ValueModel) {
	data.ID = types.StringValue(res.Key)
	data.Key = types.StringValue(res.Key)
	data.Description = types.StringValue(res.Description)
	data.IsSecret = types.BoolValue(res.IsSecret)
	if !res.IsSecret {
		data.Value = types.StringValue(res.Value)
	}
}

func (r *ResourceValue) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ValueModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()
	key := data.Key.ValueString()

	var res *client.ValueResponse
	if data.EnvID.IsNull() {
		httpResp, err := r.client.PostOrgsOrgIdAppsAppIdValuesWithResponse(ctx, r.orgId, appID, client.PostOrgsOrgIdAppsAppIdValuesJSONRequestBody{
			Key:         key,
			Description: data.Description.ValueStringPointer(),
			IsSecret:    data.IsSecret.ValueBoolPointer(),
			Value:       data.Value.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 201 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}
		res = httpResp.JSON201
	} else {
		envID := data.EnvID.ValueString()
		httpResp, err := r.client.PostOrgsOrgIdAppsAppIdEnvsEnvIdValuesWithResponse(ctx, r.orgId, appID, envID, client.PostOrgsOrgIdAppsAppIdEnvsEnvIdValuesJSONRequestBody{
			Key:         key,
			Description: data.Description.ValueStringPointer(),
			IsSecret:    data.IsSecret.ValueBoolPointer(),
			Value:       data.Value.ValueString(),
		})

		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 201 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}

		res = httpResp.JSON201
	}

	parseValueResponse(res, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceValue) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ValueModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var res *[]client.ValueResponse
	if data.EnvID.IsNull() {
		httpResp, err := r.client.GetOrgsOrgIdAppsAppIdValuesWithResponse(ctx, r.orgId, data.AppID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 200 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}

		res = httpResp.JSON200
	} else {
		httpResp, err := r.client.GetOrgsOrgIdAppsAppIdEnvsEnvIdValuesWithResponse(ctx, r.orgId, data.AppID.ValueString(), data.EnvID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 200 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}

		res = httpResp.JSON200
	}

	// TODO Ideally the API should allow to fetch a value by KEY
	key := data.Key.ValueString()
	value, found := findInSlicePtr(res, func(a client.ValueResponse) bool {
		return a.Key == key
	})

	if !found {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read value, key (%s) not found in response, %+v", key, res))
		return
	}

	parseValueResponse(&value, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceValue) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ValueModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var res *client.ValueResponse
	if data.EnvID.IsNull() {
		httpResp, err := r.client.PutOrgsOrgIdAppsAppIdValuesKeyWithResponse(ctx, r.orgId, data.AppID.ValueString(), data.Key.ValueString(), client.ValueEditPayloadRequest{
			Description: data.Description.ValueStringPointer(),
			IsSecret:    data.IsSecret.ValueBoolPointer(),
			Value:       data.Value.ValueStringPointer(),
		})
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 200 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}

		res = httpResp.JSON200
	} else {
		httpResp, err := r.client.PutOrgsOrgIdAppsAppIdEnvsEnvIdValuesKeyWithResponse(ctx, r.orgId, data.AppID.ValueString(), data.EnvID.ValueString(), data.Key.ValueString(), client.ValueEditPayloadRequest{
			Description: data.Description.ValueStringPointer(),
			IsSecret:    data.IsSecret.ValueBoolPointer(),
			Value:       data.Value.ValueStringPointer(),
		})
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 200 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}

		res = httpResp.JSON200
	}

	parseValueResponse(res, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceValue) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ValueModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.EnvID.IsNull() {
		httpResp, err := r.client.DeleteOrgsOrgIdAppsAppIdValuesKeyWithResponse(ctx, r.orgId, data.AppID.ValueString(), data.Key.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 204 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}
	} else {
		httpResp, err := r.client.DeleteOrgsOrgIdAppsAppIdEnvsEnvIdValuesKeyWithResponse(ctx, r.orgId, data.AppID.ValueString(), data.EnvID.ValueString(), data.Key.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 204 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}
	}

}

func (r *ResourceValue) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	// ensure idParts elements are not empty
	for _, idPart := range idParts {
		if idPart == "" {
			resp.Diagnostics.AddError(
				"Unexpected Import Identifier",
				fmt.Sprintf("Expected import identifier with format: app_id/key or app_id/env_id/key. Got: %q", req.ID),
			)
			return
		}
	}

	if len(idParts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), idParts[1])...)
	} else if len(idParts) == 3 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("env_id"), idParts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[2])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), idParts[2])...)
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: app_id/key or app_id/env_id/key. Got: %q", req.ID),
		)
		return
	}
}
