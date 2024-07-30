package provider

import (
	"context"
	"crypto/sha256"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Agent{}
var _ resource.ResourceWithImportState = &Agent{}

func NewResourceAgent() resource.Resource {
	return &Agent{}
}

// Agent defines the resource implementation.
type Agent struct {
	client *humanitec.Client
	orgId  string
}

// KeyModel describes the app data model.
type KeyModel struct {
	Key types.String `tfsdk:"key"`
}

// AgentModel describes the app data model.
type AgentModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	PublicKeys  []KeyModel   `tfsdk:"public_keys"`
}

func (*Agent) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (*Agent) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An Agent represents an instance of the Humanitec Agent that will be used by the Platform Orchestrator to deploy into a private cluster.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Agent.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description to show future users. It can be empty.",
				Optional:            true,
				Default:             stringdefault.StaticString(""),
				Computed:            true,
			},
			"public_keys": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:    true,
							Description: "A pcks8 RSA public keys PEM encoded (as the ones produced by openssl), whose module length is greater or equal than 4096 bits.",
						},
					},
				},
				MarkdownDescription: "A non-empty list of pcks8 RSA public keys PEM encoded (as the ones produced by openssl), whose module length is greater or equal than 4096 bits.",
				Required:            true,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			},
		},
	}
}

func (a *Agent) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	a.client = resdata.Client
	a.orgId = resdata.OrgID
}

func (a *AgentModel) updateFromContent(res *client.Agent, keys *[]client.Key) {
	a.ID = types.StringValue(res.Id)
	if res.Description == nil {
		a.Description = types.StringValue("")
	} else {
		a.Description = types.StringValue(*res.Description)
	}

	a.PublicKeys = []KeyModel{}
	for _, key := range *keys {
		a.PublicKeys = append(a.PublicKeys, KeyModel{Key: types.StringValue(key.PublicKey)})
	}
}

func (a *AgentModel) getKeysMap() map[string]string {
	var modelKeysMap = make(map[string]string)
	for _, modelKey := range a.PublicKeys {
		key := modelKey.Key.ValueString()
		modelKeysMap[getFingerprintByKey(key)] = key
	}
	return modelKeysMap
}

func fromKeyListToMap(keys []client.Key) map[string]string {
	var keyMap = make(map[string]string)
	for _, key := range keys {
		keyMap[key.Fingerprint] = key.PublicKey
	}
	return keyMap
}

func (a *Agent) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data *AgentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	description := data.Description.ValueString()
	var keys []client.Key
	var agent *client.Agent
	for _, key := range data.PublicKeys {
		keyString := key.Key.ValueString()
		if agent == nil {
			// we have to create the agent
			clientResp, err := a.client.CreateAgentWithResponse(ctx, a.orgId, client.AgentCreateBody{
				Id:          id,
				Description: &description,
				PublicKey:   keyString,
			})
			if err != nil {
				resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create an agent, got error: %s", err))
				return
			}
			switch clientResp.StatusCode() {
			case http.StatusOK:
				agent = clientResp.JSON200
				keys = append(keys, client.Key{PublicKey: keyString, Fingerprint: getFingerprintByKey(keyString)})
			case http.StatusBadRequest:
				resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create an agent, Humanitec returned bad request: %s", clientResp.Body))
				return
			case http.StatusConflict:
				resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create an agent due to a conflicts: %s", clientResp.Body))
				return
			default:
				resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Received unexpected status code when creating an agent: %d, body: %s", clientResp.StatusCode(), clientResp.Body))
				return
			}
		} else {
			registeredKey, diags := a.addKeyToAgent(ctx, id, keyString)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			keys = append(keys, *registeredKey)
		}
	}
	data.updateFromContent(agent, &keys)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

// Read implements resource.Resource.
func (a *Agent) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var data *AgentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := data.ID.ValueString()

	// read agent metadata
	clientResp, err := a.client.ListAgentsWithResponse(ctx, a.orgId, nil)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to list agents, got error: %s", err))
		return
	}
	var agent *client.Agent
	switch clientResp.StatusCode() {
	case http.StatusOK:
		for _, registeredAgent := range *clientResp.JSON200 {
			if registeredAgent.Id == id {
				agent = &registeredAgent
				break
			}
		}
		if agent == nil {
			resp.Diagnostics.AddWarning("Agent not found", fmt.Sprintf("The agent (%s) was deleted outside Terraform", data.ID.ValueString()))
			resp.State.RemoveResource(ctx)
			return
		}
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Received unexpected status code when reading agent list: %d, body: %s", clientResp.StatusCode(), clientResp.Body))
		return
	}

	registeredKeys, diags := a.getKeysForAnAgent(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.updateFromContent(agent, registeredKeys)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (a *Agent) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *AgentModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	// update agent description
	clientResp, err := a.client.PatchAgentWithResponse(ctx, a.orgId, id, client.AgentPatchBody{Description: data.Description.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update agent %s description, got error: %s", id, err))
		return
	}
	var agent *client.Agent
	switch clientResp.StatusCode() {
	case http.StatusOK:
		agent = clientResp.JSON200
	case http.StatusBadRequest:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update the agent %s, Humanitec returned bad request: %s", id, clientResp.Body))
		return
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update the agent %s, Humanitec returned the agent does not exist: %s", id, clientResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Received unexpected status code when updating the agent %s: %d, body: %s", id, clientResp.StatusCode(), clientResp.Body))
		return
	}

	registeredKeys, diags := a.getKeysForAnAgent(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var registeredKeysMap = fromKeyListToMap(*registeredKeys)
	var modelKeysMap = data.getKeysMap()

	var keysToAdd []string
	var keysToRemove []string
	var keys []client.Key
	for fingerprint := range registeredKeysMap {
		if _, ok := modelKeysMap[fingerprint]; !ok {
			keysToRemove = append(keysToRemove, fingerprint)
		}
	}

	for fingerprint, key := range modelKeysMap {
		if _, ok := registeredKeysMap[fingerprint]; ok {
			keys = append(keys, client.Key{Fingerprint: fingerprint, PublicKey: key})
		} else {
			keysToAdd = append(keysToAdd, key)
		}
	}

	for _, fingerprint := range keysToRemove {
		diags := a.removeKeyFromAnAgent(ctx, id, fingerprint)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	for _, key := range keysToAdd {
		registeredKey, diags := a.addKeyToAgent(ctx, id, key)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		keys = append(keys, *registeredKey)
	}
	data.updateFromContent(agent, &keys)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (a *Agent) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *AgentModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()

	clientResp, err := a.client.DeleteAgentWithResponse(ctx, a.orgId, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete agent %s, got error: %s", id, err))
		return
	}

	switch clientResp.StatusCode() {
	case http.StatusNoContent:
		return
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete missing agent %s: %s", id, clientResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Received unexpected status code when deleting the agent %s: %d, body: %s", id, clientResp.StatusCode(), clientResp.Body))
		return
	}
}

func (a *Agent) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (a *Agent) addKeyToAgent(ctx context.Context, agentId, key string) (*client.Key, diag.Diagnostics) {
	totalDiags := diag.Diagnostics{}
	clientResp, err := a.client.CreateKeyWithResponse(ctx, a.orgId, agentId, client.KeyCreateBody{PublicKey: key})
	if err != nil {
		totalDiags.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to register a key under the agent %s, got error: %s", agentId, err))
		return nil, totalDiags
	}
	switch clientResp.StatusCode() {
	case http.StatusOK:
		return clientResp.JSON200, totalDiags
	case http.StatusBadRequest:
		totalDiags.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to register a key under the agent %s, Humanitec returned bad request: %s", agentId, clientResp.Body))
		return nil, totalDiags
	case http.StatusNotFound:
		totalDiags.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to register a key under the agent %s, Humanitec returned the agent does not exist: %s", agentId, clientResp.Body))
		return nil, totalDiags
	case http.StatusConflict:
		totalDiags.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to register a key under the agent %s due to a conflicts: %s", agentId, clientResp.Body))
		return nil, totalDiags
	default:
		totalDiags.AddError(HUM_API_ERR, fmt.Sprintf("Received unexpected status code when registering a key under the agent %s: %d, body: %s", agentId, clientResp.StatusCode(), clientResp.Body))
		return nil, totalDiags
	}
}

func (a *Agent) removeKeyFromAnAgent(ctx context.Context, agentId, fingerprint string) diag.Diagnostics {
	totalDiags := diag.Diagnostics{}
	clientResp, err := a.client.DeleteKeyInAgentWithResponse(ctx, a.orgId, agentId, fingerprint)
	if err != nil {
		totalDiags.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to register a key under the agent %s, got error: %s", agentId, err))
		return totalDiags
	}
	switch clientResp.StatusCode() {
	case http.StatusNoContent:
		return totalDiags
	case http.StatusNotFound:
		totalDiags.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete a key under the agent %s, Humanitec returned resource does not exist: %s", agentId, clientResp.Body))
		return totalDiags
	default:
		totalDiags.AddError(HUM_API_ERR, fmt.Sprintf("Received unexpected status code when deleting a key under the agent %s: %d, body: %s", agentId, clientResp.StatusCode(), clientResp.Body))
		return totalDiags
	}
}

func (a *Agent) getKeysForAnAgent(ctx context.Context, agentId string) (*[]client.Key, diag.Diagnostics) {
	totalDiags := diag.Diagnostics{}
	clientResp, err := a.client.ListKeysInAgentWithResponse(ctx, a.orgId, agentId)
	if err != nil {
		totalDiags.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to list keys in agent %s, got error: %s", agentId, err))
		return nil, totalDiags
	}
	switch clientResp.StatusCode() {
	case http.StatusOK:
		return clientResp.JSON200, nil
	case http.StatusNotFound:
		totalDiags.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Humanitec returned the agent %s does not exist: %s", agentId, clientResp.Body))
		return nil, totalDiags
	default:
		totalDiags.AddError(HUM_API_ERR, fmt.Sprintf("Received unexpected status code when listing keys under the agent %s: %d, body: %s", agentId, clientResp.StatusCode(), clientResp.Body))
		return nil, totalDiags
	}
}

func getFingerprintByKey(key string) string {
	pem, _ := pem.Decode([]byte(key))
	sha256sum := sha256.Sum256(pem.Bytes)
	return fmt.Sprintf("%x", sha256sum)
}
