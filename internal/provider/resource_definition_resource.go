package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceDefinitionResource{}
var _ resource.ResourceWithImportState = &ResourceDefinitionResource{}

func NewResourceDefinitionResource() resource.Resource {
	return &ResourceDefinitionResource{}
}

// ResourceDefinitionResource defines the resource implementation.
type ResourceDefinitionResource struct {
	data *HumanitecData
}

func (r *ResourceDefinitionResource) client() *client.ClientWithResponses {
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

// DefinitionResourceModel describes the resource data model.
type DefinitionResourceModel struct {
	ID            types.String                         `tfsdk:"id"`
	Name          types.String                         `tfsdk:"name"`
	Type          types.String                         `tfsdk:"type"`
	DriverType    types.String                         `tfsdk:"driver_type"`
	DriverAccount types.String                         `tfsdk:"driver_account"`
	DriverInputs  *DefinitionResourceDriverInputsModel `tfsdk:"driver_inputs"`
	Criteria      *[]DefinitionResourceCriteriaModel   `tfsdk:"criteria"`
}

func (r *ResourceDefinitionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_definition"
}

func (r *ResourceDefinitionResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Visit the [docs](https://docs.humanitec.com/reference/concepts/resources/definitions) to learn more about resource definitions.",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Required:            true,
				MarkdownDescription: "The Resource Definition ID.",
				Type:                types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"name": {
				MarkdownDescription: "The display name.",
				Required:            true,
				Type:                types.StringType,
			},
			"type": {
				MarkdownDescription: "The Resource Type.",
				Required:            true,
				Type:                types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"driver_type": {
				MarkdownDescription: "The driver to be used to create the resource.",
				Required:            true,
				Type:                types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"driver_account": {
				MarkdownDescription: "Security account required by the driver.",
				Optional:            true,
				Type:                types.StringType,
			},
			"driver_inputs": {
				MarkdownDescription: "Data that should be passed around split by sensitivity.",
				Optional:            true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"values": {
						MarkdownDescription: "Values section of the data set. Passed around as-is.",
						Optional:            true,
						Type: types.MapType{
							ElemType: types.StringType,
						},
					},
					"secrets": {
						MarkdownDescription: "Secrets section of the data set.",
						Optional:            true,
						Type: types.MapType{
							ElemType: types.StringType,
						},
						Sensitive: true,
					},
				}),
			},
			"criteria": {
				MarkdownDescription: "The criteria to use when looking for a Resource Definition during the deployment.",
				Optional:            true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						MarkdownDescription: "Matching Criteria ID",
						Computed:            true,
						Type:                types.StringType,
					},
					"app_id": {
						MarkdownDescription: "The ID of the Application that the Resources should belong to.",
						Optional:            true,
						Type:                types.StringType,
					},
					"env_id": {
						MarkdownDescription: "The ID of the Environment that the Resources should belong to. If env_type is also set, it must match the Type of the Environment for the Criteria to match.",
						Optional:            true,
						Type:                types.StringType,
					},
					"env_type": {
						MarkdownDescription: "The Type of the Environment that the Resources should belong to. If env_id is also set, it must have an Environment Type that matches this parameter for the Criteria to match.",
						Optional:            true,
						Type:                types.StringType,
					},
					"res_id": {
						MarkdownDescription: "The ID of the Resource in the Deployment Set. The ID is normally a . separated path to the definition in the set, e.g. modules.my-module.externals.my-database.",
						Optional:            true,
						Type:                types.StringType,
					},
				}),
			},
		},
	}, nil
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

func parseResourceDefinitionResponse(ctx context.Context, driverInputSchema map[string]interface{}, res *client.ResourceDefinitionResponse, data *DefinitionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(res.Id)
	data.Name = types.StringValue(res.Name)
	data.Type = types.StringValue(res.Type)
	data.DriverType = parseOptionalString(res.DriverType)
	data.DriverAccount = parseOptionalString(res.DriverAccount)
	data.Criteria = parseCriteriaInput(res.Criteria)

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

func optionalStringFromModel(input types.String) *string {
	if input.IsNull() {
		return nil
	}

	v := input.ValueString()

	return &v
}

func criteriaFromModel(data *DefinitionResourceModel) *[]client.MatchingCriteriaRequest {
	if data.Criteria == nil {
		return nil
	}

	criteria := []client.MatchingCriteriaRequest{}
	for _, c := range *data.Criteria {
		criteria = append(criteria, client.MatchingCriteriaRequest{
			Id:      optionalStringFromModel(c.ID),
			AppId:   optionalStringFromModel(c.AppID),
			EnvId:   optionalStringFromModel(c.EnvID),
			EnvType: optionalStringFromModel(c.EnvType),
			ResId:   optionalStringFromModel(c.ResID),
		})
	}

	return &criteria
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
				diags.AddError(HUM_INPUT_ERR, fmt.Sprintf("Failed to convert property \"%s\" with value \"%s\" to int: %s", k, v, err.Error()))
				continue
			}
			inputMap[k] = intVar
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
		DriverAccount: optionalStringFromModel(data.DriverAccount),
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

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read resource definition, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	driverInputSchema, diag := r.data.DriverInputSchemaByDriverTypeOrType(ctx, *httpResp.JSON200.DriverType, *&httpResp.JSON200.Type)
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
			AppId:   optionalStringFromModel(c.AppID),
			EnvId:   optionalStringFromModel(c.EnvID),
			EnvType: optionalStringFromModel(c.EnvType),
			ResId:   optionalStringFromModel(c.ResID),
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

	httpResp, err := r.client().PatchOrgsOrgIdResourcesDefsDefIdWithResponse(ctx, r.orgId(), defID, client.PatchOrgsOrgIdResourcesDefsDefIdJSONRequestBody{
		DriverAccount: optionalStringFromModel(data.DriverAccount),
		DriverInputs:  driverInputs,
		Name:          &name,
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

	force := false
	httpResp, err := r.client().DeleteOrgsOrgIdResourcesDefsDefIdWithResponse(ctx, r.orgId(), data.ID.ValueString(), &client.DeleteOrgsOrgIdResourcesDefsDefIdParams{
		Force: &force,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete definition, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 204 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete definition, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}
}

func (r *ResourceDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
