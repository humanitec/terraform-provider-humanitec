package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceAccountResource{}
var _ resource.ResourceWithImportState = &ResourceAccountResource{}

func NewResourceAccountResource() resource.Resource {
	return &ResourceAccountResource{}
}

// ResourceDefinitionResource defines the resource implementation.
type ResourceAccountResource struct {
	client *client.ClientWithResponses
	orgId  string
}

// DefinitionResourceModel describes the resource data model.
type ResourceAccountModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Credentials types.String `tfsdk:"credentials"`
}

func (r *ResourceAccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_account"
}

func (r *ResourceAccountResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Resource Accounts hold credentials that are required to provision and manage resources.",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Unique identifier for the account (in scope of the organization it belongs to).",
				Required:            true,
				Type:                types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"name": {
				MarkdownDescription: "Display name.",
				Required:            true,
				Type:                types.StringType,
			},
			"type": {
				MarkdownDescription: "The type of the account",
				Required:            true,
				Type:                types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"credentials": {
				MarkdownDescription: "Credentials associated with the account.",
				Required:            true,
				Type:                types.StringType,
			},
		},
	}, nil
}

func (r *ResourceAccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func parseResourceAccountResponse(res *client.ResourceAccountResponse, data *ResourceAccountModel) {
	data.ID = types.StringValue(res.Id)
	data.Name = types.StringValue(res.Name)
	data.Type = types.StringValue(res.Type)
}

func (r *ResourceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ResourceAccountModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	name := data.Name.ValueString()
	accountType := data.Type.ValueString()
	credentialsJSON := data.Credentials.ValueString()

	var credentials map[string]interface{}
	if err := json.Unmarshal([]byte(credentialsJSON), &credentials); err != nil {
		resp.Diagnostics.AddError(HUM_INPUT_ERR, fmt.Sprintf("Unable unmarshal credentials json: %s", err))
		return
	}

	httpResp, err := r.client.PostOrgsOrgIdResourcesAccountsWithResponse(ctx, r.orgId, client.PostOrgsOrgIdResourcesAccountsJSONRequestBody{
		Id:          &id,
		Name:        &name,
		Type:        &accountType,
		Credentials: &credentials,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create resource account, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create resource account, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseResourceAccountResponse(httpResp.JSON200, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ResourceAccountModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.GetOrgsOrgIdResourcesAccountsAccIdWithResponse(ctx, r.orgId, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read resource account, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read resource account, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseResourceAccountResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ResourceAccountModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	credentialsJSON := data.Credentials.ValueString()

	var credentials map[string]interface{}
	if err := json.Unmarshal([]byte(credentialsJSON), &credentials); err != nil {
		resp.Diagnostics.AddError(HUM_INPUT_ERR, fmt.Sprintf("Unable unmarshal credentials json: %s", err))
		return
	}

	httpResp, err := r.client.PatchOrgsOrgIdResourcesAccountsAccIdWithResponse(ctx, r.orgId, data.ID.ValueString(), client.PatchOrgsOrgIdResourcesAccountsAccIdJSONRequestBody{
		Name:        &name,
		Credentials: &credentials,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update resource account, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update resource account, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseResourceAccountResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Unsupported delete", "Deleting a Resource Accounts will only remove it from Terraform state but will not deleted it from Humanitec")
}

func (r *ResourceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
