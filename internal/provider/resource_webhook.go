package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceWebhook{}
var _ resource.ResourceWithImportState = &ResourceWebhook{}

func NewResourceWebhook() resource.Resource {
	return &ResourceWebhook{}
}

// ResourceWebhook defines the resource implementation.
type ResourceWebhook struct {
	client *humanitec.Client
	orgId  string
}

// WebhookModel describes the app data model.
type WebhookTriggerModel struct {
	Scope types.String `tfsdk:"scope"`
	Type  types.String `tfsdk:"type"`
}

// WebhookModel describes the app data model.
type WebhookModel struct {
	ID    types.String `tfsdk:"id"`
	AppID types.String `tfsdk:"app_id"`

	Disabled types.Bool            `tfsdk:"disabled"`
	Headers  types.Map             `tfsdk:"headers"`
	Payload  types.Map             `tfsdk:"payload"`
	Triggers []WebhookTriggerModel `tfsdk:"triggers"`
	URL      types.String          `tfsdk:"url"`
}

func (r *ResourceWebhook) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *ResourceWebhook) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Webhook is a special type of a Job, it performs a HTTPS request to a specified URL with specified headers.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Webhook.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Application that the Webhook should belong to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Defines whether this job is currently disabled.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"headers": schema.MapAttribute{
				MarkdownDescription: "Custom webhook headers.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             mapdefault.StaticValue(types.MapValueMust(types.StringType, map[string]attr.Value{})),
			},
			"payload": schema.MapAttribute{
				MarkdownDescription: "Customize payload.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"triggers": schema.SetNestedAttribute{
				MarkdownDescription: `
A list of Events by which the Job is triggered, supported triggers are:

  | scope | type |
	|-------|------|
	| environment  | created |
	| environment  | deleted |
	| deployment  | started |
	| deployment  | finished |
				`,
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"scope": schema.StringAttribute{
							MarkdownDescription: "Scope of the trigger",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the trigger",
							Required:            true,
						},
					},
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Thw webhook's URL (without protocol, only HTTPS is supported)",
				Required:            true,
			},
		},
	}
}

func (r *ResourceWebhook) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func parseWebhookResponse(ctx context.Context, res *client.WebhookResponse, data *WebhookModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	data.ID = types.StringValue(res.Id)
	data.Disabled = types.BoolPointerValue(res.Disabled)

	headers, diag := types.MapValueFrom(ctx, types.StringType, res.Headers)
	diags.Append(diag...)
	data.Headers = headers

	payload, diag := types.MapValueFrom(ctx, types.StringType, res.Payload)
	diags.Append(diag...)
	data.Payload = payload

	triggers := []WebhookTriggerModel{}
	for _, trigger := range res.Triggers {
		triggers = append(triggers, WebhookTriggerModel{
			Scope: types.StringValue(trigger.Scope),
			Type:  types.StringValue(trigger.Type),
		})
	}
	data.Triggers = triggers

	data.URL = types.StringPointerValue(res.Url)

	return diags
}

func parseWebhookUpdateResponse(ctx context.Context, res *client.WebhookUpdateResponse, data *WebhookModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	data.Disabled = types.BoolPointerValue(res.Disabled)

	headers, diag := types.MapValueFrom(ctx, types.StringType, res.Headers)
	diags.Append(diag...)
	data.Headers = headers

	payload, diag := types.MapValueFrom(ctx, types.StringType, res.Payload)
	diags.Append(diag...)
	data.Payload = payload

	triggers := []WebhookTriggerModel{}

	if res.Triggers != nil {
		for _, trigger := range *res.Triggers {
			triggers = append(triggers, WebhookTriggerModel{
				Scope: types.StringValue(trigger.Scope),
				Type:  types.StringValue(trigger.Type),
			})
		}
	}
	data.Triggers = triggers

	data.URL = types.StringPointerValue(res.Url)

	return diags
}

// mapToJSONFieldRequest converts a tf string map to a client.JSONFieldRequest.
func mapToJSONFieldRequest(ctx context.Context, tfmap basetypes.MapValue) (client.JSONFieldRequest, diag.Diagnostics) {
	if tfmap.IsNull() {
		return nil, nil
	}

	// ElementsAs doesn't support map[string]interface{}, so make a map[string]string first.
	var m map[string]string
	diags := tfmap.ElementsAs(ctx, &m, false)

	j := make(client.JSONFieldRequest, len(m))
	for k, v := range m {
		j[k] = v
	}
	return j, diags
}

func toWebhookRequest(ctx context.Context, data *WebhookModel) (*client.WebhookRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	headers, fieldDiags := mapToJSONFieldRequest(ctx, data.Headers)
	diags.Append(fieldDiags...)

	payload, fieldDiags := mapToJSONFieldRequest(ctx, data.Payload)
	diags.Append(fieldDiags...)

	triggers := []client.EventBaseRequest{}
	for _, trigger := range data.Triggers {
		triggers = append(triggers, client.EventBaseRequest{
			Scope: trigger.Scope.ValueStringPointer(),
			Type:  trigger.Type.ValueStringPointer(),
		})
	}

	return &client.WebhookRequest{
		Disabled: data.Disabled.ValueBoolPointer(),
		Headers:  &headers,
		Id:       data.ID.ValueStringPointer(),
		Payload:  &payload,
		Triggers: &triggers,
		Url:      data.URL.ValueStringPointer(),
	}, diags
}

func (r *ResourceWebhook) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *WebhookModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()

	httpBody, diags := toWebhookRequest(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.PostOrgsOrgIdAppsAppIdWebhooksWithResponse(ctx, r.orgId, appID, *httpBody)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create webhook, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 201 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create webhook, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	resp.Diagnostics.Append(parseWebhookResponse(ctx, httpResp.JSON201, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceWebhook) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *WebhookModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()
	id := data.ID.ValueString()

	httpResp, err := r.client.GetOrgsOrgIdAppsAppIdWebhooksJobIdWithResponse(ctx, r.orgId, appID, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read webhook, got error: %s", err))
		return
	}

	if httpResp.StatusCode() == 404 {
		resp.Diagnostics.AddWarning("Webook not found", fmt.Sprintf("The webhook (%s) was deleted outside Terraform", id))
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read webhook, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseWebhookResponse(ctx, httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceWebhook) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *WebhookModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()
	id := data.ID.ValueString()

	httpBody, diags := toWebhookRequest(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.PatchOrgsOrgIdAppsAppIdWebhooksJobIdWithResponse(ctx, r.orgId, appID, id, *httpBody)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update value, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update value, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseWebhookUpdateResponse(ctx, httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceWebhook) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *WebhookModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.AppID.ValueString()
	id := data.ID.ValueString()

	httpResp, err := r.client.DeleteOrgsOrgIdAppsAppIdWebhooksJobIdWithResponse(ctx, r.orgId, appID, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete webhook, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 204 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete webhook, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}
}

func (r *ResourceWebhook) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	// ensure idParts elements are not empty
	for _, idPart := range idParts {
		if idPart == "" {
			resp.Diagnostics.AddError(
				"Unexpected Import Identifier",
				fmt.Sprintf("Expected import identifier with format: app_id/webhook_id. Got: %q", req.ID),
			)
			return
		}
	}

	if len(idParts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: app_id/webhook_id. Got: %q", req.ID),
		)
		return
	}
}
