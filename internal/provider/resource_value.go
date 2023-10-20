package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

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
	SecretRef   types.Object `tfsdk:"secret_ref"`
}

// SecretRef describes a secret reference that might contain a secret value or a reference to an already stored secret.
type SecretRef struct {
	Ref     types.String `tfsdk:"ref"`
	Store   types.String `tfsdk:"store"`
	Version types.String `tfsdk:"version"`
	Value   types.String `tfsdk:"value"`
}

func SecretRefAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ref":     types.StringType,
		"store":   types.StringType,
		"version": types.StringType,
		"value":   types.StringType,
	}
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
				MarkdownDescription: "The value that will be stored. It can't be defined if secret_ref is defined.",
				Optional:            true,
				Sensitive:           true,
			},
			"secret_ref": schema.SingleNestedAttribute{
				MarkdownDescription: "The sensitive value that will be stored in the primary organization store or a reference to a sensitive value already stored in one of the registered stores. It can't be defined if is_secret is false or value is defined.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"ref": schema.StringAttribute{
						MarkdownDescription: "Secret reference in the format of the target store. It can't be defined if value is defined.",
						Optional:            true,
						Computed:            true,
					},
					"store": schema.StringAttribute{
						MarkdownDescription: "Secret Store id. This can't be humanitec (our internal Secret Store). It's mandatory if ref is defined and can't be used if value is defined.",
						Optional:            true,
						Computed:            true,
					},
					"version": schema.StringAttribute{
						MarkdownDescription: "Only valid if ref is defined. It's the version of the secret as defined in the target store.",
						Optional:            true,
						Computed:            true,
					},
					"value": schema.StringAttribute{
						MarkdownDescription: "Value to store in the secret store. It can't be defined if ref is defined.",
						Optional:            true,
						Sensitive:           true,
					},
				},
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

func envValueIdPrefix(appID, envID string) string {
	return strings.Join([]string{appID, envID}, "/")
}

func parseValueResponse(ctx context.Context, res *client.ValueResponse, data *ValueModel, idPrefix string) {
	data.ID = types.StringValue(strings.Join([]string{idPrefix, res.Key}, "/"))
	data.Key = types.StringValue(res.Key)
	data.Description = types.StringValue(res.Description)
	data.IsSecret = types.BoolValue(res.IsSecret)
	if !res.IsSecret {
		data.Value = types.StringValue(res.Value)
		data.SecretRef = basetypes.NewObjectNull(SecretRefAttributeTypes())
	} else {
		var secretRef SecretRef
		if data.SecretRef.IsUnknown() {
			secretRef = SecretRef{}
		} else {
			diags := data.SecretRef.As(ctx, &secretRef, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				tflog.Debug(ctx, "can't populate secretRef from model", map[string]interface{}{"err": diags.Errors()})
				return
			}
		}

		secretRef.Ref = types.StringValue(*res.SecretKey)
		secretRef.Store = types.StringValue(*res.SecretStoreId)
		if res.SecretVersion != nil {
			secretRef.Version = types.StringValue(*res.SecretVersion)
		}

		objectValue, diags := types.ObjectValueFrom(ctx, SecretRefAttributeTypes(), secretRef)
		if diags.HasError() {
			tflog.Debug(ctx, "can't decode object from secret ref", map[string]interface{}{"err": diags})
			return
		}
		data.SecretRef = objectValue
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
	var idPrefix string
	var createPayload = client.PostOrgsOrgIdAppsAppIdValuesJSONRequestBody{
		Key:         key,
		Description: data.Description.ValueStringPointer(),
		IsSecret:    data.IsSecret.ValueBoolPointer(),
	}
	if !data.Value.IsNull() {
		createPayload.Value = data.Value.ValueStringPointer()
	} else {
		var secretRef SecretRef
		diags := data.SecretRef.As(ctx, &secretRef, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			tflog.Debug(ctx, "can't populate secretRef from model", map[string]interface{}{"err": diags.Errors()})
			return
		}
		if !secretRef.Value.IsNull() {
			createPayload.SecretRef = &client.SecretReference{
				Value: secretRef.Value.ValueStringPointer(),
			}
		} else {
			createPayload.SecretRef = &client.SecretReference{
				Ref:     secretRef.Ref.ValueStringPointer(),
				Store:   secretRef.Store.ValueStringPointer(),
				Version: secretRef.Version.ValueStringPointer(),
			}
		}
	}

	if data.EnvID.IsNull() {
		httpResp, err := r.client.PostOrgsOrgIdAppsAppIdValuesWithResponse(ctx, r.orgId, appID, createPayload)
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 201 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}
		res = httpResp.JSON201
		idPrefix = appID
	} else {
		envID := data.EnvID.ValueString()
		httpResp, err := r.client.PostOrgsOrgIdAppsAppIdEnvsEnvIdValuesWithResponse(ctx, r.orgId, appID, envID, createPayload)

		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 201 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}

		res = httpResp.JSON201
		idPrefix = envValueIdPrefix(appID, envID)
	}

	parseValueResponse(ctx, res, data, idPrefix)

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

	appID := data.AppID.ValueString()

	var res *[]client.ValueResponse
	var idPrefix string
	if data.EnvID.IsNull() {
		httpResp, err := r.client.GetOrgsOrgIdAppsAppIdValuesWithResponse(ctx, r.orgId, appID)
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 200 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}

		res = httpResp.JSON200
		idPrefix = appID
	} else {
		envID := data.EnvID.ValueString()
		httpResp, err := r.client.GetOrgsOrgIdAppsAppIdEnvsEnvIdValuesWithResponse(ctx, r.orgId, appID, envID)
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 200 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}

		res = httpResp.JSON200
		idPrefix = envValueIdPrefix(appID, envID)
	}

	// TODO Ideally the API should allow to fetch a value by KEY
	key := data.Key.ValueString()
	value, found := findInSlicePtr(res, func(a client.ValueResponse) bool {
		return a.Key == key
	})

	if !found {
		resp.Diagnostics.AddWarning("Value not found", fmt.Sprintf("The value (%s) was deleted outside Terraform", key))
		resp.State.RemoveResource(ctx)
		return
	}

	parseValueResponse(ctx, &value, data, idPrefix)

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
	var idPrefix string
	appID := data.AppID.ValueString()
	var editPayload = client.ValueEditPayloadRequest{
		Description: data.Description.ValueStringPointer(),
		IsSecret:    data.IsSecret.ValueBoolPointer(),
	}
	if !data.Value.IsNull() {
		editPayload.Value = data.Value.ValueStringPointer()
	} else {
		var secretRef SecretRef
		diags := data.SecretRef.As(ctx, &secretRef, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			tflog.Debug(ctx, "can't populate secretRef from model", map[string]interface{}{"err": diags.Errors()})
			return
		}
		if !secretRef.Value.IsNull() {
			editPayload.SecretRef = &client.SecretReference{
				Value: secretRef.Value.ValueStringPointer(),
			}
		} else {
			editPayload.SecretRef = &client.SecretReference{
				Ref:     secretRef.Ref.ValueStringPointer(),
				Store:   secretRef.Store.ValueStringPointer(),
				Version: secretRef.Version.ValueStringPointer(),
			}
		}
	}
	if data.EnvID.IsNull() {
		httpResp, err := r.client.PutOrgsOrgIdAppsAppIdValuesKeyWithResponse(ctx, r.orgId, appID, data.Key.ValueString(), editPayload)
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 200 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}

		res = httpResp.JSON200
		idPrefix = appID
	} else {
		envID := data.EnvID.ValueString()
		httpResp, err := r.client.PutOrgsOrgIdAppsAppIdEnvsEnvIdValuesKeyWithResponse(ctx, r.orgId, appID, envID, data.Key.ValueString(), editPayload)
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update value, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 200 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}

		res = httpResp.JSON200
		idPrefix = envValueIdPrefix(appID, envID)
	}

	parseValueResponse(ctx, res, data, idPrefix)

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
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), idParts[1])...)
	} else if len(idParts) == 3 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("env_id"), idParts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), idParts[2])...)
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: app_id/key or app_id/env_id/key. Got: %q", req.ID),
		)
		return
	}
}
