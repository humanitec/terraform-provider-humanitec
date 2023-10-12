package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &SecretStore{}
var _ resource.ResourceWithImportState = &SecretStore{}

func NewResourceSecretStore() resource.Resource {
	return &SecretStore{}
}

// SecretStore defines the resource implementation.
type SecretStore struct {
	client *humanitec.Client
	orgId  string
}

// SecretStoreModel describes the app data model.
type SecretStoreModel struct {
	ID      types.String  `tfsdk:"id"`
	Primary types.Bool    `tfsdk:"primary"`
	AwsSM   *AwsSMModel   `tfsdk:"awssm"`
	AzureKV *AzureKVModel `tfsdk:"azurekv"`
	GcpSM   *GcpSMModel   `tfsdk:"gcpsm"`
	Vault   *VaultModel   `tfsdk:"vault"`
}

type AwsSMModel struct {
	Auth   *AwsAuthModel `tfsdk:"auth"`
	Region types.String  `tfsdk:"region"`
}

type AwsAuthModel struct {
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
}

type AzureKVModel struct {
	Auth     *AzureKVAuthModel `tfsdk:"auth"`
	TenantID types.String      `tfsdk:"tenant_id"`
	Url      types.String      `tfsdk:"url"`
}

type AzureKVAuthModel struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

type GcpSMModel struct {
	Auth      *GcpAuthModel `tfsdk:"auth"`
	ProjectID types.String  `tfsdk:"project_id"`
}

type GcpAuthModel struct {
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
}

type VaultModel struct {
	Auth    *VaultAuthModel `tfsdk:"auth"`
	AgentID types.String    `tfsdk:"agent_id"`
	Path    types.String    `tfsdk:"path"`
	Url     types.String    `tfsdk:"url"`
}

type VaultAuthModel struct {
	Role  types.String `tfsdk:"role"`
	Token types.String `tfsdk:"token"`
}

func (*SecretStore) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secretstore"
}

func (*SecretStore) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An external secret management system used by an organization to store secrets referenced in Humanitec.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Secret Store.",
				Required:            true,
			},
			"primary": schema.BoolAttribute{
				MarkdownDescription: "Whether the Secret Store is the Primary one for the organization.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"awssm": schema.SingleNestedAttribute{
				MarkdownDescription: "AWS Secret Manager specification.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"region": schema.StringAttribute{
						MarkdownDescription: "The region of AWS Secret Manager.",
						Required:            true,
					},
					"auth": schema.SingleNestedAttribute{
						MarkdownDescription: "Credentials to authenticate to AWS Secret Manager.",
						Sensitive:           true,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"access_key_id": schema.StringAttribute{
								MarkdownDescription: "The Access Key ID.",
								Required:            true,
							},
							"secret_access_key": schema.StringAttribute{
								MarkdownDescription: "The Secret Access Key.",
								Required:            true,
							},
						},
					},
				},
			},
			"azurekv": schema.SingleNestedAttribute{
				MarkdownDescription: "Azure KV Secret Manager specification.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"tenant_id": schema.StringAttribute{
						MarkdownDescription: "The AzureKV Tenant ID.",
						Required:            true,
					},
					"url": schema.StringAttribute{
						MarkdownDescription: "The AzureKV URL.",
						Required:            true,
					},
					"auth": schema.SingleNestedAttribute{
						MarkdownDescription: "Credentials to authenticate to Azure Key Vault.",
						Sensitive:           true,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"client_id": schema.StringAttribute{
								MarkdownDescription: "The AzureKV Client ID.",
								Required:            true,
							},
							"client_secret": schema.StringAttribute{
								MarkdownDescription: "The AzureKV Client Secret.",
								Required:            true,
							},
						},
					},
				},
			},
			"gcpsm": schema.SingleNestedAttribute{
				MarkdownDescription: "GCP Secret Manager specification.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"project_id": schema.StringAttribute{
						MarkdownDescription: "The project ID of the GCPSM.",
						Required:            true,
					},
					"auth": schema.SingleNestedAttribute{
						MarkdownDescription: "Credentials to authenticate the GCPSM.",
						Sensitive:           true,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"secret_access_key": schema.StringAttribute{
								MarkdownDescription: "The Secret Access Key.",
								Required:            true,
							},
						},
					},
				},
			},
			"vault": schema.SingleNestedAttribute{
				MarkdownDescription: "Vault specification.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The Vault URL.",
					},
					"path": schema.StringAttribute{
						MarkdownDescription: "The path used to read / write secrets.",
						Optional:            true,
					},
					"agent_id": schema.StringAttribute{
						MarkdownDescription: "Reference to the agent to use to hit Vault.",
						Optional:            true,
					},
					"auth": schema.SingleNestedAttribute{
						MarkdownDescription: "Credentials to authenticate the Vault.",
						Sensitive:           true,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"token": schema.StringAttribute{
								MarkdownDescription: "Token to access Vault.",
								Optional:            true,
							},
							"role": schema.StringAttribute{
								MarkdownDescription: "Role to assume to access Vault.",
								Optional:            true,
							},
						},
					},
				},
			},
		},
	}
}

func (s *SecretStore) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	s.client = resdata.Client
	s.orgId = resdata.OrgID
}

func parseSecretStoreResponse(res *client.SecretStoreResponse, data *SecretStoreModel) {
	data.ID = types.StringValue(res.Id)
	data.Primary = types.BoolValue(res.Primary)
	if res.Awssm != nil {
		if data.AwsSM == nil {
			data.AwsSM = &AwsSMModel{}
		}
		if res.Awssm.Region != nil {
			data.AwsSM.Region = types.StringValue(*res.Awssm.Region)
		}
	} else if res.Azurekv != nil {
		if data.AzureKV == nil {
			data.AzureKV = &AzureKVModel{}
		}
		if res.Azurekv.TenantId != nil {
			data.AzureKV.TenantID = types.StringValue(*res.Azurekv.TenantId)
		}
		if res.Azurekv.Url != nil {
			data.AzureKV.Url = types.StringValue(*res.Azurekv.Url)
		}
	} else if res.Gcpsm != nil {
		if data.GcpSM == nil {
			data.GcpSM = &GcpSMModel{}
		}
		if res.Gcpsm.ProjectId != nil {
			data.GcpSM.ProjectID = types.StringValue(*res.Gcpsm.ProjectId)
		}
	} else if res.Vault != nil {
		if data.Vault == nil {
			data.Vault = &VaultModel{}
		}
		if res.Vault.AgentId != nil {
			data.Vault.AgentID = types.StringValue(*res.Vault.AgentId)
		}
		if res.Vault.Path != nil {
			data.Vault.Path = types.StringValue(*res.Vault.Path)
		}
		if res.Vault.Url != nil {
			data.Vault.Url = types.StringValue(*res.Vault.Url)
		}
	}
}

func toSecretStoreRequest(data *SecretStoreModel) (*client.CreateSecretStorePayloadRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	secretStorePayload := &client.CreateSecretStorePayloadRequest{
		Id:      data.ID.ValueString(),
		Primary: data.Primary.ValueBool(),
	}
	if data.AwsSM != nil {
		secretStorePayload.Awssm = &client.AWSSMRequest{
			Region: data.AwsSM.Region.ValueStringPointer(),
		}
		if data.AwsSM.Auth != nil {
			secretStorePayload.Awssm.Auth = &client.AWSAuthRequest{
				AccessKeyId:     data.AwsSM.Auth.AccessKeyID.ValueStringPointer(),
				SecretAccessKey: data.AwsSM.Auth.SecretAccessKey.ValueStringPointer(),
			}
		}
	} else if data.AzureKV != nil {
		secretStorePayload.Azurekv = &client.AzureKVRequest{
			TenantId: data.AzureKV.TenantID.ValueStringPointer(),
			Url:      data.AzureKV.Url.ValueStringPointer(),
		}
		if data.AzureKV.Auth != nil {
			secretStorePayload.Azurekv.Auth = &client.AzureAuthRequest{
				ClientId:     data.AzureKV.Auth.ClientID.ValueStringPointer(),
				ClientSecret: data.AzureKV.Auth.ClientSecret.ValueStringPointer(),
			}
		}
	} else if data.GcpSM != nil {
		secretStorePayload.Gcpsm = &client.GCPSMRequest{
			ProjectId: data.GcpSM.ProjectID.ValueStringPointer(),
		}
		if data.GcpSM.Auth != nil {
			secretStorePayload.Gcpsm.Auth = &client.GCPAuthRequest{
				SecretAccessKey: data.GcpSM.Auth.SecretAccessKey.ValueStringPointer(),
			}
		}
	} else if data.Vault != nil {
		secretStorePayload.Vault = &client.VaultRequest{
			AgentId: data.Vault.AgentID.ValueStringPointer(),
			Url:     data.Vault.Url.ValueStringPointer(),
			Path:    data.Vault.Path.ValueStringPointer(),
		}
		if data.Vault.Auth != nil {
			secretStorePayload.Vault.Auth = &client.VaultAuthRequest{
				Token: data.Vault.Auth.Token.ValueStringPointer(),
				Role:  data.Vault.Auth.Role.ValueStringPointer(),
			}
		}
	}

	return secretStorePayload, diags
}

func (s *SecretStore) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SecretStoreModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpBody, diags := toSecretStoreRequest(data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := s.client.PostOrgsOrgIdSecretstoresWithResponse(ctx, s.orgId, *httpBody)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create secret role, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 201 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create secret store, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource.
func (s *SecretStore) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *SecretStoreModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()

	httpResp, err := s.client.GetOrgsOrgIdSecretstoresStoreIdWithResponse(ctx, s.orgId, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read secret store, got error: %s", err))
		return
	}
	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read secret store, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read secret store, missing body, body: %s", httpResp.Body))
		return
	}

	parseSecretStoreResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (s *SecretStore) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *SecretStoreModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	createBody, diags := toSecretStoreRequest(data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var updateBody client.UpdateSecretStorePayloadRequest
	updateBody.Primary = &createBody.Primary
	if createBody.Awssm != nil {
		updateBody.Awssm = createBody.Awssm
	} else if createBody.Azurekv != nil {
		updateBody.Azurekv = createBody.Azurekv
	} else if createBody.Gcpsm != nil {
		updateBody.Gcpsm = createBody.Gcpsm
	} else if createBody.Vault != nil {
		updateBody.Vault = createBody.Vault
	}

	httpResp, err := s.client.PatchOrgsOrgIdSecretstoresStoreIdWithResponse(ctx, s.orgId, id, updateBody)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update secret store, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update secret store, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseSecretStoreResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (s *SecretStore) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SecretStoreModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()

	httpResp, err := s.client.DeleteOrgsOrgIdSecretstoresStoreIdWithResponse(ctx, s.orgId, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete secret store, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 204 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete secret store, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}
}

func (s *SecretStore) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
