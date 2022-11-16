package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/humanitec/terraform-provider-humanitec/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceDefinitionResource{}
var _ resource.ResourceWithImportState = &ResourceDefinitionResource{}

func NewResourceDefinitionResource() resource.Resource {
	return &ResourceDefinitionResource{}
}

// ResourceDefinitionResource defines the resource implementation.
type ResourceDefinitionResource struct {
	client *client.ClientWithResponses
	orgId  string
}

// DefinitionResourceDriverInputsModel describes the resource data model.
type DefinitionResourceDriverInputsModel struct {
	Values  types.Map `tfsdk:"values"`
	Secrets types.Map `tfsdk:"secrets"`
}

// DefinitionResourceDriverInputsModel describes the resource data model.
// type DefinitionResourceCriteriaModel struct {
// 	ID      types.String `tfsdk:"id"`
// 	AppID   types.String `tfsdk:"app_id"`
// 	EnvID   types.String `tfsdk:"env_id"`
// 	EnvType types.String `tfsdk:"env_type"`
// 	ResID   types.String `tfsdk:"res_id"`
// }

// DefinitionResourceModel describes the resource data model.
type DefinitionResourceModel struct {
	ID            types.String                         `tfsdk:"id"`
	Name          types.String                         `tfsdk:"name"`
	Type          types.String                         `tfsdk:"type"`
	DriverType    types.String                         `tfsdk:"driver_type"`
	DriverAccount types.String                         `tfsdk:"driver_account"`
	DriverInputs  *DefinitionResourceDriverInputsModel `tfsdk:"driver_inputs"`
	// Criteria      *[]DefinitionResourceCriteriaModel   `tfsdk:"criteria"`
}

func (r *ResourceDefinitionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_definition"
}

func (r *ResourceDefinitionResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Required:            true,
				MarkdownDescription: "",
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "",
				Required:            true,
				Type:                types.StringType,
			},
			"type": {
				MarkdownDescription: "",
				Required:            true,
				Type:                types.StringType,
			},
			"driver_type": {
				MarkdownDescription: "",
				Required:            true,
				Type:                types.StringType,
			},
			"driver_account": {
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.StringType,
			},
			"driver_inputs": {
				MarkdownDescription: "",
				Optional:            true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"values": {
						MarkdownDescription: "",
						Optional:            true,
						Type: types.MapType{
							ElemType: types.StringType,
						},
					},
					"secrets": {
						MarkdownDescription: "",
						Optional:            true,
						Type: types.MapType{
							ElemType: types.StringType,
						},
						Sensitive: true,
					},
				}),
			},
			// "criteria": {
			// 	MarkdownDescription: "",
			// 	Optional:            true,
			// 	Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			// 		"id": {
			// 			MarkdownDescription: "",
			// 			Required:            true,
			// 			Type:                types.StringType,
			// 		},
			// 		"app_id": {
			// 			MarkdownDescription: "",
			// 			Optional:            true,
			// 			Type:                types.StringType,
			// 		},
			// 		"env_id": {
			// 			MarkdownDescription: "",
			// 			Optional:            true,
			// 			Type:                types.StringType,
			// 		},
			// 		"env_type": {
			// 			MarkdownDescription: "",
			// 			Optional:            true,
			// 			Type:                types.StringType,
			// 		},
			// 		"res_id": {
			// 			MarkdownDescription: "",
			// 			Optional:            true,
			// 			Type:                types.StringType,
			// 		},
			// 	}),
			// },
		},
	}, nil
}

func (r *ResourceDefinitionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	resdata, ok := req.ProviderData.(*HumanitecResourceData)

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

func parseOptionalString(input *string) types.String {
	if input == nil {
		return types.StringNull()
	}

	return types.StringValue(*input)
}

func parseMapInput(driver *client.DriverDefinitionResponse, input map[string]interface{}) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	inputSchema := driver.InputsSchema.AdditionalProperties

	lenDriverInput := len(input)
	inputProperties, ok := valueAtPath[map[string]interface{}](inputSchema, []string{"properties", "values", "properties"})
	if lenDriverInput > 0 && !ok {
		diags.AddError("Client Error", fmt.Sprintf("No value inputs expected in driver input schema: %v", inputSchema))
		return nil, diags
	}

	inputMap := make(map[string]string, lenDriverInput)
	for k, v := range input {
		propertyType, ok := valueAtPath[string](inputProperties, []string{k, "type"})
		if !ok {
			diags.AddError("Client Error", fmt.Sprintf("Property \"%s\" not found in schema: %v", k, inputSchema))
			continue
		}

		switch propertyType {
		case "string":
			inputMap[k] = v.(string)
		case "integer":
			inputMap[k] = strconv.FormatInt(int64(v.(float64)), 10)
		case "object":
			obj, err := json.Marshal(v)
			if err != nil {
				diags.AddError("Client Error", fmt.Sprintf("Failed to marshal property \"%s\": %s", k, err.Error()))
				continue
			}
			inputMap[k] = string(obj)
		default:
			diags.AddError("Client Error", fmt.Sprintf("Unexpected property type \"%s\" for property \"%s\": %v", propertyType, k, inputSchema))
			continue
		}
	}
	return inputMap, diags
}

func parseResourceDefinitionResponse(ctx context.Context, driver *client.DriverDefinitionResponse, res *client.ResourceDefinitionResponse, data *DefinitionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(res.Id)
	data.Name = types.StringValue(res.Name)
	data.Type = types.StringValue(res.Type)
	data.DriverType = parseOptionalString(res.DriverType)
	data.DriverAccount = parseOptionalString(res.DriverAccount)

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
			valuesMap, diag := parseMapInput(driver, driverInputs.Values.AdditionalProperties)
			diags.Append(diag...)

			m, diag := types.MapValueFrom(ctx, types.StringType, valuesMap)
			diags.Append(diag...)
			data.DriverInputs.Values = m
		}
	}

	// if res.Criteria != nil {
	// 	criteria := []DefinitionResourceCriteriaModel{}
	// 	for _, critera := range *res.Criteria {
	// 		criteria = append(criteria, DefinitionResourceCriteriaModel{
	// 			ID:      types.StringValue(critera.Id),
	// 			AppID:   parseOptionalString(critera.AppId),
	// 			EnvID:   parseOptionalString(critera.EnvId),
	// 			EnvType: parseOptionalString(critera.EnvType),
	// 			ResID:   parseOptionalString(critera.ResId),
	// 		})
	// 	}
	// 	data.Criteria = &criteria
	// } else {
	// 	data.Criteria = nil
	// }

	return diags
}

func optionalStringFromModel(input types.String) *string {
	if input.IsNull() {
		return nil
	}

	v := input.ValueString()

	return &v
}

// func criteriaFromModel(data *DefinitionResourceModel) *[]client.MatchingCriteriaRequest {
// 	if data.Criteria == nil {
// 		return nil
// 	}

// 	criteria := []client.MatchingCriteriaRequest{}

// 	for _, c := range *data.Criteria {
// 		id := c.ID.ValueString()
// 		criteria = append(criteria, client.MatchingCriteriaRequest{
// 			Id:      &id,
// 			AppId:   optionalStringFromModel(c.AppID),
// 			EnvId:   optionalStringFromModel(c.EnvID),
// 			EnvType: optionalStringFromModel(c.EnvType),
// 			ResId:   optionalStringFromModel(c.ResID),
// 		})
// 	}

// 	return &criteria
// }

func driverInputToMap(ctx context.Context, data types.Map, inputSchema map[string]interface{}, field string) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	var driverInput map[string]string
	diags.Append(data.ElementsAs(ctx, &driverInput, false)...)

	if driverInput == nil {
		return nil, diags
	}

	lenDriverInput := len(driverInput)
	inputProperties, ok := valueAtPath[map[string]interface{}](inputSchema, []string{"properties", field, "properties"})
	if lenDriverInput > 0 && !ok {
		diags.AddError("Client Error", fmt.Sprintf("No %s inputs expected in driver input schema: %v", field, inputSchema))
		return nil, diags
	}

	inputMap := make(map[string]interface{}, lenDriverInput)
	for k, v := range driverInput {
		propertyType, ok := valueAtPath[string](inputProperties, []string{k, "type"})
		if !ok {
			diags.AddError("Client Error", fmt.Sprintf("Property \"%s\" not found in schema: %v", k, inputSchema))
			continue
		}

		switch propertyType {
		case "string":
			inputMap[k] = v
		case "integer":
			intVar, err := strconv.Atoi(v)
			if err != nil {
				diags.AddError("Client Error", fmt.Sprintf("Failed to convert property \"%s\" with value \"%s\" to int: %s", k, v, err.Error()))
				continue
			}
			inputMap[k] = intVar
		case "object":
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(v), &obj); err != nil {
				diags.AddError("Client Error", fmt.Sprintf("Failed to unmarshal property \"%s\": %s", k, err.Error()))
				continue
			}
			inputMap[k] = obj
		default:
			diags.AddError("Client Error", fmt.Sprintf("Unexpected property type \"%s\" for property \"%s\": %v", propertyType, k, inputSchema))
			continue
		}

	}

	return inputMap, diags
}

func driverInputsFromModel(ctx context.Context, driver *client.DriverDefinitionResponse, data *DefinitionResourceModel) (*client.ValuesSecretsRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	driverInputs := &client.ValuesSecretsRequest{}

	secrets, diags := driverInputToMap(ctx, data.DriverInputs.Secrets, driver.InputsSchema.AdditionalProperties, "secrets")
	diags.Append(diags...)
	if secrets != nil {
		driverInputs.Secrets = &client.ValuesSecretsRequest_Secrets{
			AdditionalProperties: secrets,
		}
	}

	values, diags := driverInputToMap(ctx, data.DriverInputs.Values, driver.InputsSchema.AdditionalProperties, "values")
	diags.Append(diags...)
	if values != nil {
		driverInputs.Values = &client.ValuesSecretsRequest_Values{
			AdditionalProperties: values,
		}
	}

	return driverInputs, diags
}

func (r *ResourceDefinitionResource) driverByDriverType(ctx context.Context, driverType string) (*client.DriverDefinitionResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	httpResp, err := r.client.GetOrgsOrgIdResourcesDriversWithResponse(ctx, r.orgId)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to get drivers, got error: %s", err))
		return nil, diags
	}

	if httpResp.StatusCode() != 200 {
		diags.AddError("Client Error", fmt.Sprintf("Unable to get drivers, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return nil, diags
	}

	if httpResp.JSON200 == nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to get drivers, missing body, body: %s", httpResp.Body))
		return nil, diags
	}

	driversByType := map[string]*client.DriverDefinitionResponse{}
	for _, d := range *httpResp.JSON200 {
		d := d
		driversByType[fmt.Sprintf("%s/%s", d.OrgId, d.Id)] = &d
	}

	driver, ok := driversByType[driverType]
	if !ok {
		diags.AddError("Client Error", fmt.Sprintf("Not driver found for type: %s", driverType))
		return nil, diags
	}

	return driver, diags
}

func (r *ResourceDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DefinitionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// criteria := criteriaFromModel(data)

	driverType := data.DriverType.ValueString()
	driver, diag := r.driverByDriverType(ctx, driverType)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	driverInputs, diag := driverInputsFromModel(ctx, driver, data)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.PostOrgsOrgIdResourcesDefsWithResponse(ctx, r.orgId, client.PostOrgsOrgIdResourcesDefsJSONRequestBody{
		// Criteria:      criteria,
		DriverAccount: optionalStringFromModel(data.DriverAccount),
		DriverInputs:  driverInputs,
		DriverType:    data.DriverType.ValueString(),
		Id:            data.ID.ValueString(),
		Name:          data.Name.ValueString(),
		Type:          data.Type.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create definition, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create definition, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	tflog.Info(ctx, "response", map[string]interface{}{"string": string(httpResp.Body)})

	resp.Diagnostics.Append(parseResourceDefinitionResponse(ctx, driver, httpResp.JSON200, data)...)
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

	httpResp, err := r.client.GetOrgsOrgIdResourcesDefsDefIdWithResponse(ctx, r.orgId, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read definition, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read definition, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	driver, diag := r.driverByDriverType(ctx, *httpResp.JSON200.DriverType)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(parseResourceDefinitionResponse(ctx, driver, httpResp.JSON200, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DefinitionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	driverType := data.DriverType.ValueString()
	driver, diag := r.driverByDriverType(ctx, driverType)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	driverInputs, diag := driverInputsFromModel(ctx, driver, data)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.PatchOrgsOrgIdResourcesDefsDefIdWithResponse(ctx, r.orgId, data.ID.ValueString(), client.PatchOrgsOrgIdResourcesDefsDefIdJSONRequestBody{
		DriverAccount: optionalStringFromModel(data.DriverAccount),
		DriverInputs:  driverInputs,
		Name:          &name,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read definition, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read definition, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	tflog.Info(ctx, "response", map[string]interface{}{"string": string(httpResp.Body)})

	resp.Diagnostics.Append(parseResourceDefinitionResponse(ctx, driver, httpResp.JSON200, data)...)
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
	httpResp, err := r.client.DeleteOrgsOrgIdResourcesDefsDefIdWithResponse(ctx, r.orgId, data.ID.ValueString(), &client.DeleteOrgsOrgIdResourcesDefsDefIdParams{
		Force: &force,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete definition, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 204 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete definition, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}
}

func (r *ResourceDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	httpResp, err := r.client.GetOrgsOrgIdResourcesDefsDefIdWithResponse(ctx, r.orgId, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read definition, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read definition, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	driver, diag := r.driverByDriverType(ctx, *httpResp.JSON200.DriverType)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	data := &DefinitionResourceModel{}
	resp.Diagnostics.Append(parseResourceDefinitionResponse(ctx, driver, httpResp.JSON200, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
