package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceDefinitionResource{}
var _ resource.ResourceWithImportState = &ResourceDefinitionResource{}

var defaultResourceDefinitionDeleteTimeout = 3 * time.Minute

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
	Values  types.Map `tfsdk:"values"`
	Secrets types.Map `tfsdk:"secrets"`
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
	Criteria      *[]DefinitionResourceCriteriaModel           `tfsdk:"criteria"`
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
					"values": schema.MapAttribute{
						MarkdownDescription: "Values section of the data set. Passed around as-is.",
						ElementType:         types.StringType,
						Optional:            true,
					},
					"secrets": schema.MapAttribute{
						MarkdownDescription: "Secrets section of the data set.",
						ElementType:         types.StringType,
						Optional:            true,
						Sensitive:           true,
					},
				},
			},
			"criteria": schema.SetNestedAttribute{
				MarkdownDescription: "The criteria to use when looking for a Resource Definition during the deployment.",
				Optional:            true,
				DeprecationMessage:  "Inline criteria management is deprecated and should be done using the dedicated humanitec_resource_definition_criteria resource instead",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Matching Criteria ID",
							Computed:            true,
						},
						"app_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the Application that the Resources should belong to.",
							Optional:            true,
						},
						"env_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the Environment that the Resources should belong to. If env_type is also set, it must match the Type of the Environment for the Criteria to match.",
							Optional:            true,
						},
						"env_type": schema.StringAttribute{
							MarkdownDescription: "The Type of the Environment that the Resources should belong to. If env_id is also set, it must have an Environment Type that matches this parameter for the Criteria to match.",
							Optional:            true,
						},
						"res_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the Resource in the Deployment Set. The ID is normally a . separated path to the definition in the set, e.g. modules.my-module.externals.my-database.",
							Optional:            true,
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

func parseMapInput(input map[string]interface{}, inputSchema map[string]interface{}, field string) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	inputSchemaJSON, err := json.MarshalIndent(inputSchema, "", "\t")
	if err != nil {
		diags.AddError(HUM_API_ERR, fmt.Sprintf("Failed to marshal driver schema: %s", err.Error()))
	}

	lenDriverInput := len(input)
	inputProperties, ok := valueAtPath[map[string]interface{}](inputSchema, []string{"properties", field, "properties"})
	if lenDriverInput > 0 && !ok {
		diags.AddError(HUM_INPUT_ERR, fmt.Sprintf("No value inputs expected in driver input schema:\n%s", string(inputSchemaJSON)))
		return nil, diags
	}

	inputMap := make(map[string]string, lenDriverInput)
	for k, v := range input {
		propertyType, ok := valueAtPath[string](inputProperties, []string{k, "type"})
		if !ok {
			diags.AddError(HUM_INPUT_ERR, fmt.Sprintf("Property \"%s\" not found in driver input schema:\n%s", k, string(inputSchemaJSON)))
			continue
		}

		switch propertyType {
		case "string":
			inputMap[k] = v.(string)
		case "integer":
			if isReference(v) {
				inputMap[k] = v.(string)
				continue
			}
			inputMap[k] = strconv.FormatInt(int64(v.(float64)), 10)
		case "boolean":
			if isReference(v) {
				inputMap[k] = v.(string)
				continue
			}
			inputMap[k] = strconv.FormatBool(v.(bool))
		case "object":
			if isReference(v) {
				inputMap[k] = v.(string)
				continue
			}
			obj, err := json.Marshal(v)
			if err != nil {
				diags.AddError(HUM_INPUT_ERR, fmt.Sprintf("Failed to marshal property \"%s\": %s", k, err.Error()))
				continue
			}
			inputMap[k] = string(obj)
		default:
			diags.AddError(HUM_PROVIDER_ERR, fmt.Sprintf("Unexpected property type \"%s\" for property \"%s\" in driver input schema:\n%s", propertyType, k, string(inputSchemaJSON)))
			continue
		}
	}
	return inputMap, diags
}

func parseCriteriaInput(criteria *[]client.MatchingCriteriaResponse) *[]DefinitionResourceCriteriaModel {
	if criteria == nil {
		return nil
	}

	data := []DefinitionResourceCriteriaModel{}

	for _, c := range *criteria {
		data = append(data, DefinitionResourceCriteriaModel{
			AppID:   parseOptionalString(c.AppId),
			EnvID:   parseOptionalString(c.EnvId),
			EnvType: parseOptionalString(c.EnvType),
			ResID:   parseOptionalString(c.ResId),
		})
	}

	return &data
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
	data.DriverType = parseOptionalString(res.DriverType)
	data.DriverAccount = parseOptionalString(res.DriverAccount)
	data.Criteria = parseCriteriaInput(res.Criteria)
	data.Provision = parseProvisionInput(res.Provision)

	driverInputs := res.DriverInputs

	if driverInputs != nil {
		if data.DriverInputs == nil {
			data.DriverInputs = &DefinitionResourceDriverInputsModel{
				Secrets: types.MapNull(types.StringType),
			}
		}

		if driverInputs.Values == nil {
			data.DriverInputs.Values = types.MapNull(types.StringType)
		} else {
			valuesMap, diag := parseMapInput(*driverInputs.Values, driverInputSchema, "values")
			diags.Append(diag...)

			m, diag := types.MapValueFrom(ctx, types.StringType, valuesMap)
			diags.Append(diag...)
			data.DriverInputs.Values = m
		}
	}

	if res.Criteria != nil {
		criteria := []DefinitionResourceCriteriaModel{}
		for _, c := range *res.Criteria {
			criteria = append(criteria, DefinitionResourceCriteriaModel{
				ID:      types.StringValue(c.Id),
				AppID:   parseOptionalString(c.AppId),
				EnvID:   parseOptionalString(c.EnvId),
				EnvType: parseOptionalString(c.EnvType),
				ResID:   parseOptionalString(c.ResId),
			})
		}
		data.Criteria = &criteria
	} else {
		data.Criteria = nil
	}

	return diags
}

func criteriaFromModel(data *DefinitionResourceModel) *[]client.MatchingCriteriaRequest {
	if data.Criteria == nil {
		return nil
	}

	criteria := []client.MatchingCriteriaRequest{}
	for _, c := range *data.Criteria {
		criteria = append(criteria, client.MatchingCriteriaRequest{
			Id:      c.ID.ValueStringPointer(),
			AppId:   c.AppID.ValueStringPointer(),
			EnvId:   c.EnvID.ValueStringPointer(),
			EnvType: c.EnvType.ValueStringPointer(),
			ResId:   c.ResID.ValueStringPointer(),
		})
	}

	return &criteria
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

// isReference returns whether a value is a resource reference https://docs.humanitec.com/reference/concepts/resources/references
func isReference(v interface{}) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	return strings.HasPrefix(s, "${resources")
}

func driverInputToMap(ctx context.Context, data types.Map, inputSchema map[string]interface{}, field string) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	inputSchemaJSON, err := json.MarshalIndent(inputSchema, "", "\t")
	if err != nil {
		diags.AddError(HUM_API_ERR, fmt.Sprintf("Failed to marshal driver schema: %s", err.Error()))
	}

	var driverInput map[string]string
	diags.Append(data.ElementsAs(ctx, &driverInput, false)...)

	if driverInput == nil {
		return nil, diags
	}

	lenDriverInput := len(driverInput)
	inputProperties, ok := valueAtPath[map[string]interface{}](inputSchema, []string{"properties", field, "properties"})
	if lenDriverInput > 0 && !ok {
		diags.AddError(HUM_INPUT_ERR, fmt.Sprintf("No %s inputs expected in driver input schema:\n%s", field, string(inputSchemaJSON)))
		return nil, diags
	}

	inputMap := make(map[string]interface{}, lenDriverInput)
	for k, v := range driverInput {
		propertyType, ok := valueAtPath[string](inputProperties, []string{k, "type"})
		if !ok {
			diags.AddError(HUM_INPUT_ERR, fmt.Sprintf("Property \"%s\" not found in driver input schema %s:\n%s", k, field, string(inputSchemaJSON)))
			continue
		}

		switch propertyType {
		case "string":
			inputMap[k] = v
		case "integer":
			if isReference(v) {
				inputMap[k] = v
				continue
			}
			intVar, err := strconv.Atoi(v)
			if err != nil {
				diags.AddError(HUM_INPUT_ERR, fmt.Sprintf("Failed to convert property \"%s\" with value \"%s\" to integer: %s", k, v, err.Error()))
				continue
			}
			inputMap[k] = intVar
		case "boolean":
			if isReference(v) {
				inputMap[k] = v
				continue
			}
			booleanVar, err := strconv.ParseBool(v)
			if err != nil {
				diags.AddError(HUM_INPUT_ERR, fmt.Sprintf("Failed to convert property \"%s\" with value \"%s\" to boolean: %s", k, v, err.Error()))
				continue
			}
			inputMap[k] = booleanVar
		case "object":
			if isReference(v) {
				inputMap[k] = v
				continue
			}
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(v), &obj); err != nil {
				diags.AddError(HUM_INPUT_ERR, fmt.Sprintf("Failed to unmarshal property \"%s\": %s", k, err.Error()))
				continue
			}
			inputMap[k] = obj
		default:
			diags.AddError(HUM_PROVIDER_ERR, fmt.Sprintf("Unexpected property type \"%s\" for property \"%s\" driver input schema %s:\n%s", propertyType, k, field, string(inputSchemaJSON)))
			continue
		}

	}

	return inputMap, diags
}

func driverInputsFromModel(ctx context.Context, inputSchema map[string]interface{}, data *DefinitionResourceModel) (*client.ValuesSecretsRequest, diag.Diagnostics) {
	var diag diag.Diagnostics

	driverInputs := &client.ValuesSecretsRequest{}

	secrets, secretsDiag := driverInputToMap(ctx, data.DriverInputs.Secrets, inputSchema, "secrets")
	diag.Append(secretsDiag...)
	if secrets != nil {
		driverInputs.Secrets = &secrets
	}

	values, valueDiag := driverInputToMap(ctx, data.DriverInputs.Values, inputSchema, "values")
	diag.Append(valueDiag...)
	if values != nil {
		driverInputs.Values = &values
	}

	return driverInputs, diag
}

func (r *ResourceDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DefinitionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	criteria := criteriaFromModel(data)
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

	httpResp, err := r.client().PostOrgsOrgIdResourcesDefsWithResponse(ctx, r.orgId(), client.PostOrgsOrgIdResourcesDefsJSONRequestBody{
		Criteria:      criteria,
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

	httpResp, err := r.client().GetOrgsOrgIdResourcesDefsDefIdWithResponse(ctx, r.orgId(), data.ID.ValueString())
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

	driverInputSchema, diag := r.data.DriverInputSchemaByDriverTypeOrType(ctx, *httpResp.JSON200.DriverType, httpResp.JSON200.Type)
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

func diffCriteria(previousCriteria *[]DefinitionResourceCriteriaModel, currentCriteria *[]DefinitionResourceCriteriaModel) ([]DefinitionResourceCriteriaModel, []DefinitionResourceCriteriaModel) {
	addedCriteria := []DefinitionResourceCriteriaModel{}
	removedCriteria := []DefinitionResourceCriteriaModel{}

	if previousCriteria == nil {
		if currentCriteria != nil {
			// All criteria are new
			addedCriteria = append(addedCriteria, *currentCriteria...)
		}
	} else {
		if currentCriteria == nil {
			// All criteria are deleted
			removedCriteria = append(removedCriteria, *previousCriteria...)
		} else {
			toKey := func(c *DefinitionResourceCriteriaModel) string {
				return fmt.Sprintf("%s/%s/%s/%s", c.AppID, c.EnvID, c.EnvType, c.ResID)
			}

			// Diff old vs. new
			previousCriteriaMap := map[string]*DefinitionResourceCriteriaModel{}
			for _, c := range *previousCriteria {
				c := c
				previousCriteriaMap[toKey(&c)] = &c
			}
			currentCriteriaMap := map[string]*DefinitionResourceCriteriaModel{}
			for _, c := range *currentCriteria {
				c := c
				key := toKey(&c)
				currentCriteriaMap[key] = &c

				if _, ok := previousCriteriaMap[key]; !ok {
					addedCriteria = append(addedCriteria, c)
				}
			}

			for k, v := range previousCriteriaMap {
				if _, ok := currentCriteriaMap[k]; !ok {
					removedCriteria = append(removedCriteria, *v)
				}
			}
		}
	}

	return addedCriteria, removedCriteria
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
	addedCriteria, removedCriteria := diffCriteria(state.Criteria, data.Criteria)

	// Add criteria
	for _, c := range addedCriteria {
		httpResp, err := r.client().PostOrgsOrgIdResourcesDefsDefIdCriteriaWithResponse(ctx, r.orgId(), defID, client.PostOrgsOrgIdResourcesDefsDefIdCriteriaJSONRequestBody{
			AppId:   c.AppID.ValueStringPointer(),
			EnvId:   c.EnvID.ValueStringPointer(),
			EnvType: c.EnvType.ValueStringPointer(),
			ResId:   c.ResID.ValueStringPointer(),
		})
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create resource definition criteria, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 200 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create resource definition criteria, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}
	}

	// Remove criteria
	force := true
	for _, c := range removedCriteria {
		criteriaID := c.ID.ValueString()
		if criteriaID == "" {
			// This shouldn't be possible, the patch and state override below will unset never saved values
			continue
		}

		httpResp, err := r.client().DeleteOrgsOrgIdResourcesDefsDefIdCriteriaCriteriaIdWithResponse(ctx, r.orgId(), defID, criteriaID, &client.DeleteOrgsOrgIdResourcesDefsDefIdCriteriaCriteriaIdParams{
			Force: &force,
		})
		if err != nil {
			resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete resource definition criteria, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 204 {
			resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete resource definition criteria, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
			return
		}
	}

	provision := provisionFromModel(data.Provision)

	httpResp, err := r.client().PutOrgsOrgIdResourcesDefsDefIdWithResponse(ctx, r.orgId(), defID, client.PutOrgsOrgIdResourcesDefsDefIdJSONRequestBody{
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
		httpResp, err := r.client().DeleteOrgsOrgIdResourcesDefsDefIdWithResponse(ctx, r.orgId(), data.ID.ValueString(), &client.DeleteOrgsOrgIdResourcesDefsDefIdParams{
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
