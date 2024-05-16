package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceDefinitionResource{}
var _ resource.ResourceWithImportState = &ResourceDefinitionResource{}

var defaultResourceDefinitionDeleteTimeout = 10 * time.Minute

func NewResourceDefinitionResource() resource.Resource {
	return &ResourceDefinitionResource{}
}

// ResourceDefinitionResource defines the resource implementation.
type ResourceDefinitionResource struct {
	data *HumanitecData
}

func (r *ResourceDefinitionResource) client() *humanitec.Client {
	return r.data.Client
}

func (r *ResourceDefinitionResource) orgId() string {
	return r.data.OrgID
}

// DefinitionResourceDriverInputsModel describes the resource data model.
type DefinitionResourceDriverInputsModel struct {
	ValuesString  types.String `tfsdk:"values_string"`
	SecretsString types.String `tfsdk:"secrets_string"`
	SecretRefs    types.String `tfsdk:"secret_refs"`
}

// DefinitionResourceCriteriaModel describes the resource data model.
type DefinitionResourceCriteriaModel struct {
	ID      types.String `tfsdk:"id"`
	AppID   types.String `tfsdk:"app_id"`
	EnvID   types.String `tfsdk:"env_id"`
	EnvType types.String `tfsdk:"env_type"`
	ResID   types.String `tfsdk:"res_id"`
}

// DefinitionResourceProvisionModel describes the resource definition provision model.
type DefinitionResourceProvisionModel struct {
	IsDependant     types.Bool `tfsdk:"is_dependent"`
	MatchDependents types.Bool `tfsdk:"match_dependents"`
}

// DefinitionResourceModel describes the resource data model.
type DefinitionResourceModel struct {
	ID            types.String                                 `tfsdk:"id"`
	Name          types.String                                 `tfsdk:"name"`
	Type          types.String                                 `tfsdk:"type"`
	DriverType    types.String                                 `tfsdk:"driver_type"`
	DriverAccount types.String                                 `tfsdk:"driver_account"`
	DriverInputs  *DefinitionResourceDriverInputsModel         `tfsdk:"driver_inputs"`
	Provision     *map[string]DefinitionResourceProvisionModel `tfsdk:"provision"`

	ForceDelete types.Bool     `tfsdk:"force_delete"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

func (r *ResourceDefinitionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_definition"
}

func (r *ResourceDefinitionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Visit the [docs](https://docs.humanitec.com/reference/concepts/resources/definitions) to learn more about resource definitions.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The Resource Definition ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The Resource Type.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"driver_type": schema.StringAttribute{
				MarkdownDescription: "The driver to be used to create the resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"driver_account": schema.StringAttribute{
				MarkdownDescription: "Security account required by the driver.",
				Optional:            true,
			},
			"driver_inputs": schema.SingleNestedAttribute{
				MarkdownDescription: "Data that should be passed around split by sensitivity.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"values_string": schema.StringAttribute{
						MarkdownDescription: "JSON encoded input data set. Passed around as-is.",
						Optional:            true,
					},
					"secrets_string": schema.StringAttribute{
						MarkdownDescription: "JSON encoded secret data set. Passed around as-is. Can't be used together with secret_refs.",
						Optional:            true,
					},
					"secret_refs": schema.StringAttribute{
						MarkdownDescription: "JSON encoded secrets section of the data set. They can hold sensitive information that will be stored in the primary organization secret store and replaced with the secret store paths when sent outside, or secret references stored in a defined secret store. Can't be used together with secrets.",
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.Expressions{
								path.MatchRelative().AtParent().AtName("secrets_string"),
							}...),
						},
					},
				},
			},
			"provision": schema.MapNestedAttribute{
				MarkdownDescription: "ProvisionDependencies defines resources which are needed to be co-provisioned with the current resource.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"is_dependent": schema.BoolAttribute{
							MarkdownDescription: "If the co-provisioned resource is dependent on the current one.",
							Optional:            true,
						},
						"match_dependents": schema.BoolAttribute{
							MarkdownDescription: "If the resources dependant on the main resource, are also dependant on the co-provisioned one.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
					},
				},
			},
			"force_delete": schema.BoolAttribute{
				MarkdownDescription: "If set to `true`, will mark the Resource Definition for deletion, even if it affects existing Active Resources.",
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

func (r *ResourceDefinitionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func parseOptionalString(input *string) types.String {
	if input == nil {
		return types.StringNull()
	}

	return types.StringValue(*input)
}

func parseProvisionInput(provision *map[string]client.ProvisionDependenciesResponse) *map[string]DefinitionResourceProvisionModel {
	if provision == nil {
		return nil
	}

	data := make(map[string]DefinitionResourceProvisionModel, len(*provision))
	for k, v := range *provision {
		data[k] = DefinitionResourceProvisionModel{
			IsDependant:     types.BoolValue(v.IsDependent),
			MatchDependents: defaultFalseBoolValuePointer(v.MatchDependents),
		}
	}

	return &data
}

// defaultFalseBoolValuePointer returns a types.Bool value of false if the pointer is nil, otherwise it returns the value of the pointer.
func defaultFalseBoolValuePointer(b *bool) types.Bool {
	if b == nil {
		return types.BoolValue(false)
	}

	return types.BoolValue(*b)
}

func parseResourceDefinitionResponse(ctx context.Context, driverInputSchema map[string]interface{}, res *client.ResourceDefinitionResponse, data *DefinitionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(res.Id)
	data.Name = types.StringValue(res.Name)
	data.Type = types.StringValue(res.Type)
	data.DriverType = types.StringValue(res.DriverType)
	data.DriverAccount = parseOptionalString(res.DriverAccount)
	data.Provision = parseProvisionInput(res.Provision)

	driverInputs := res.DriverInputs

	if driverInputs != nil && driverInputs.Values != nil {
		if data.DriverInputs == nil {
			data.DriverInputs = &DefinitionResourceDriverInputsModel{
				SecretsString: types.StringNull(),
				SecretRefs:    types.StringNull(),
			}
		}

		b, err := json.Marshal(driverInputs.Values)
		if err != nil {
			diags.AddError(HUM_API_ERR, fmt.Sprintf("Failed to marshal values: %s", err.Error()))
		}
		data.DriverInputs.ValuesString = types.StringValue(string(b))
	}

	if data.DriverInputs != nil {
		if driverInputs.SecretRefs == nil {
			data.DriverInputs.SecretRefs = types.StringNull()
		} else {
			if !strings.Contains(data.DriverInputs.SecretRefs.ValueString(), `{"value":"`) {
				b, err := json.Marshal(driverInputs.SecretRefs)
				if err != nil {
					diags.AddError(HUM_API_ERR, fmt.Sprintf("Failed to marshal secret_refs: %s", err.Error()))
				}
				data.DriverInputs.SecretRefs = types.StringValue(string(b))
			}
		}
	}
	return diags
}

func provisionFromModel(data *map[string]DefinitionResourceProvisionModel) *map[string]client.ProvisionDependenciesRequest {
	if data == nil {
		return nil
	}

	provision := make(map[string]client.ProvisionDependenciesRequest, len(*data))

	for k, v := range *data {
		provision[k] = client.ProvisionDependenciesRequest{
			IsDependent:     v.IsDependant.ValueBoolPointer(),
			MatchDependents: v.MatchDependents.ValueBoolPointer(),
		}
	}

	return &provision
}

func driverInputsFromModel(ctx context.Context, inputSchema map[string]interface{}, data *DefinitionResourceModel) (*client.ValuesSecretsRefsRequest, diag.Diagnostics) {
	if data.DriverInputs == nil {
		return nil, nil
	}

	var diags diag.Diagnostics

	driverInputs := &client.ValuesSecretsRefsRequest{}

	var secrets map[string]interface{}
	var secretRefs map[string]interface{}
	var secretsDiag diag.Diagnostics

	if !data.DriverInputs.SecretsString.IsNull() {
		if err := json.Unmarshal([]byte(data.DriverInputs.SecretsString.ValueString()), &secrets); err != nil {
			secretsDiag.AddError(HUM_INPUT_ERR, fmt.Sprintf("Failed to unmarshal secrets_string: %s", err.Error()))
		}
	} else if !data.DriverInputs.SecretRefs.IsUnknown() {
		if err := json.Unmarshal([]byte(data.DriverInputs.SecretRefs.ValueString()), &secretRefs); err != nil {
			secretsDiag.AddError(HUM_INPUT_ERR, fmt.Sprintf("Failed to unmarshal secret_refs: %s", err.Error()))
		}
	}
	diags.Append(secretsDiag...)
	if secrets != nil {
		driverInputs.Secrets = &secrets
	}
	if secretRefs != nil {
		driverInputs.SecretRefs = &secretRefs
	}

	var values map[string]interface{}
	var valuesDiag diag.Diagnostics

	if !data.DriverInputs.ValuesString.IsNull() {
		if err := json.Unmarshal([]byte(data.DriverInputs.ValuesString.ValueString()), &values); err != nil {
			valuesDiag.AddError(HUM_INPUT_ERR, fmt.Sprintf("Failed to unmarshal values_string: %s", err.Error()))
		}
	}
	diags.Append(valuesDiag...)
	if values != nil {
		driverInputs.Values = &values
	}

	return driverInputs, diags
}

func (r *ResourceDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DefinitionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	provision := provisionFromModel(data.Provision)
	driverType := data.DriverType.ValueString()
	driverInputSchema, diag := r.data.DriverInputSchemaByDriverTypeOrType(ctx, driverType, data.Type.ValueString())
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	driverInputs, diag := driverInputsFromModel(ctx, driverInputSchema, data)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client().CreateResourceDefinitionWithResponse(ctx, r.orgId(), client.CreateResourceDefinitionRequestRequest{
		Provision:     provision,
		DriverAccount: data.DriverAccount.ValueStringPointer(),
		DriverInputs:  driverInputs,
		DriverType:    data.DriverType.ValueString(),
		Id:            data.ID.ValueString(),
		Name:          data.Name.ValueString(),
		Type:          data.Type.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create resource definition, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create resource definition, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	resp.Diagnostics.Append(parseResourceDefinitionResponse(ctx, driverInputSchema, httpResp.JSON200, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DefinitionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client().GetResourceDefinitionWithResponse(ctx, r.orgId(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read resource definition, got error: %s", err))
		return
	}

	if httpResp.StatusCode() == 404 {
		resp.Diagnostics.AddWarning("Resource definition not found", fmt.Sprintf("The resource definition (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read resource definition, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	driverInputSchema, diag := r.data.DriverInputSchemaByDriverTypeOrType(ctx, httpResp.JSON200.DriverType, httpResp.JSON200.Type)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(parseResourceDefinitionResponse(ctx, driverInputSchema, httpResp.JSON200, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *DefinitionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	driverType := data.DriverType.ValueString()
	driverInputSchema, diag := r.data.DriverInputSchemaByDriverTypeOrType(ctx, driverType, data.Type.ValueString())
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	driverInputs, diag := driverInputsFromModel(ctx, driverInputSchema, data)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	defID := data.ID.ValueString()

	provision := provisionFromModel(data.Provision)

	httpResp, err := r.client().UpdateResourceDefinitionWithResponse(ctx, r.orgId(), defID, client.UpdateResourceDefinitionRequestRequest{
		DriverAccount: data.DriverAccount.ValueStringPointer(),
		DriverInputs:  driverInputs,
		Name:          name,
		Provision:     provision,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read definition, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read definition, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	resp.Diagnostics.Append(parseResourceDefinitionResponse(ctx, driverInputSchema, httpResp.JSON200, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DefinitionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, defaultResourceDefinitionDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	force := data.ForceDelete.ValueBool()

	err := retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		httpResp, err := r.client().DeleteResourceDefinitionWithResponse(ctx, r.orgId(), data.ID.ValueString(), &client.DeleteResourceDefinitionParams{
			Force: &force,
		})
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if httpResp.StatusCode() == 409 {
			return retry.RetryableError(fmt.Errorf("resource definition has still active resources, status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		if httpResp.StatusCode() != 204 {
			return retry.NonRetryableError(fmt.Errorf("unable to delete resource definition, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete resource definition, got error: %s", err))
		return
	}
}

func (r *ResourceDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
