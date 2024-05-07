package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceRegistry{}
var _ resource.ResourceWithImportState = &ResourceRegistry{}

func NewResourceRegistry() resource.Resource {
	return &ResourceRegistry{}
}

// ResourceRule defines the resource implementation.
type ResourceRegistry struct {
	client *humanitec.Client
	orgID  string
}

func (r *ResourceRegistry) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry"
}

func (r *ResourceRegistry) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Container Registries store and manage container images ready for use when they are needed in a deployment.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Registry ID, unique within the Organization.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[a-z0-9][a-z0-9-]+[a-z0-9]$"), "must follow standard Humanitec id pattern"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"registry": schema.StringAttribute{
				MarkdownDescription: "Registry name, usually in a \"{domain}\" or \"{domain}/{project}\" format.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Registry type, describes the registry authentication method, and defines the schema for the credentials.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("basic", "google_gcr", "amazon_ecr", "secret_ref"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"enable_ci": schema.BoolAttribute{
				MarkdownDescription: "Indicates if registry secrets and credentials should be exposed to CI agents.",
				Optional:            true,
			},
			"creds": schema.ObjectAttribute{
				MarkdownDescription: "AccountCreds represents an account credentials (either, username- or token-based).",
				Optional:            true,
				AttributeTypes: map[string]attr.Type{
					"password": types.StringType,
					"username": types.StringType,
				},
				Sensitive: true,
			},
			"secrets": schema.MapNestedAttribute{
				MarkdownDescription: "ClusterSecretsMap stores a list of Kuberenetes secret references for the target deployment clusters.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"namespace": schema.StringAttribute{
							Required: true,
						},
						"secret": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (r *ResourceRegistry) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type RegistryCredsModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type RegistryModel struct {
	ID       types.String             `tfsdk:"id"`
	Registry types.String             `tfsdk:"registry"`
	Type     types.String             `tfsdk:"type"`
	EnableCI types.Bool               `tfsdk:"enable_ci"`
	Creds    *RegistryCredsModel      `tfsdk:"creds"`
	Secrets  *map[string]SecretsModel `tfsdk:"secrets"`
}

type SecretsModel struct {
	Namespace types.String `tfsdk:"namespace"`
	Secret    types.String `tfsdk:"secret"`
}

func (r *ResourceRegistry) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *RegistryModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, diags := parseRegistryModel(data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var registry *client.RegistryResponse
	createRegistryResp, err := r.client.PostOrgsOrgIdRegistriesWithResponse(ctx, r.orgID, *request)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create registry, got error: %s", err))
		return
	}
	switch createRegistryResp.StatusCode() {
	case http.StatusCreated:
		registry = createRegistryResp.JSON201
	case http.StatusBadRequest:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create registry, Humanitec returned bad request: %s", createRegistryResp.Body))
		return
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create registry, organization not found: %s", createRegistryResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create registry unexpected status code: %d, body: %s", createRegistryResp.StatusCode(), createRegistryResp.Body))
		return
	}

	diags = parseRegistryResponse(registry, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceRegistry) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *RegistryModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()

	var registry *client.RegistryResponse
	getRegistryResp, err := r.client.GetOrgsOrgIdRegistriesRegIdWithResponse(ctx, r.orgID, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to get registry, got error: %s", err))
		return
	}

	switch getRegistryResp.StatusCode() {
	case http.StatusOK:
		registry = getRegistryResp.JSON200
	case http.StatusNotFound:
		resp.Diagnostics.AddWarning("registry not found", fmt.Sprintf("The registry (%s) was deleted outside Terraform", id))
		resp.State.RemoveResource(ctx)
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to get registry, unexpected status code: %d, body: %s", getRegistryResp.StatusCode(), getRegistryResp.Body))
		return
	}

	diags := parseRegistryResponse(registry, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceRegistry) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *RegistryModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	request, diags := parseRegistryModel(data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var registry *client.RegistryResponse
	updateRegistryResp, err := r.client.PatchOrgsOrgIdRegistriesRegIdWithResponse(ctx, r.orgID, id, *request)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update registry, got error: %s", err))
		return
	}
	switch updateRegistryResp.StatusCode() {
	case http.StatusOK:
		registry = updateRegistryResp.JSON200
	case http.StatusBadRequest:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update registry, Humanitec returned bad request: %s", updateRegistryResp.Body))
		return
	case http.StatusForbidden:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update humanitec build-in registry: %s", updateRegistryResp.Body))
		return
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update registry, organization or registry not found: %s", updateRegistryResp.Body))
		return
	case http.StatusConflict:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update registry, registry already registered: %s", updateRegistryResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update registry, unexpected status code: %d, body: %s", updateRegistryResp.StatusCode(), updateRegistryResp.Body))
		return
	}

	diags = parseRegistryResponse(registry, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceRegistry) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *RegistryModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()

	deleteRegistryResp, err := r.client.DeleteOrgsOrgIdRegistriesRegIdWithResponse(ctx, r.orgID, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete registry, got error: %s", err))
		return
	}
	switch deleteRegistryResp.StatusCode() {
	case http.StatusNoContent:
		// Do nothing
	case http.StatusForbidden:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete humanitec build-in registry: %s", deleteRegistryResp.Body))
		return
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete registry, registry not found: %s", deleteRegistryResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete registry, unexpected status code: %d, body: %s", deleteRegistryResp.StatusCode(), deleteRegistryResp.Body))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceRegistry) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			"Expected import identifier with registry id. Got an empty string",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func parseRegistryModel(data *RegistryModel) (*client.RegistryRequest, diag.Diagnostics) {
	totalDiags := diag.Diagnostics{}

	var creds *client.AccountCredsRequest
	if data.Creds != nil {
		creds = &client.AccountCredsRequest{
			Password: data.Creds.Password.ValueString(),
			Username: data.Creds.Username.ValueString(),
		}
	}

	var secrets *client.ClusterSecretsMapRequest
	if data.Secrets != nil {
		secretsMap := client.ClusterSecretsMapRequest{}
		for key, value := range *data.Secrets {
			secretsMap[key] = client.ClusterSecretRequest{
				Namespace: value.Namespace.ValueString(),
				Secret:    value.Secret.ValueString(),
			}
		}
		secrets = &secretsMap
	}

	return &client.RegistryRequest{
		Id:       data.ID.ValueString(),
		Registry: data.Registry.ValueString(),
		Type:     data.Type.ValueString(),
		EnableCi: data.EnableCI.ValueBoolPointer(),
		Creds:    creds,
		Secrets:  secrets,
	}, totalDiags
}

func parseRegistryResponse(res *client.RegistryResponse, data *RegistryModel) diag.Diagnostics {
	totalDiags := diag.Diagnostics{}

	data.ID = types.StringValue(res.Id)
	data.Registry = types.StringValue(res.Registry)
	data.Type = types.StringValue(res.Type)
	data.EnableCI = types.BoolValue(res.EnableCi)

	if res.Secrets != nil {
		secrets := make(map[string]SecretsModel)
		for key, value := range *res.Secrets {
			secrets[key] = SecretsModel{
				Namespace: types.StringValue(value.Namespace),
				Secret:    types.StringValue(value.Secret),
			}
		}
		data.Secrets = &secrets
	}

	return totalDiags
}
