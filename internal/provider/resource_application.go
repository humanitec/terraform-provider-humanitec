package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceApplication{}
var _ resource.ResourceWithImportState = &ResourceApplication{}

var defaultApplicationReadTimeout = 2 * time.Minute
var defaultApplicationDeleteTimeout = 2 * time.Minute

func NewResourceApplication() resource.Resource {
	return &ResourceApplication{}
}

// ResourceApplication defines the resource implementation.
type ResourceApplication struct {
	client *humanitec.Client
	orgId  string
}

// ApplicationEnvironmentModel describes the app env data model.
type ApplicationEnvironmentModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

// ApplicationModel describes the app data model.
type ApplicationModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`

	Env *ApplicationEnvironmentModel `tfsdk:"env"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *ResourceApplication) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ResourceApplication) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An Application is a collection of Workloads that work together. When deployed, all Workloads in an Application are deployed to the same namespace.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID which refers to a specific application.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The Human-friendly name for the Application.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"env": schema.SingleNestedAttribute{
				MarkdownDescription: "Initial environment to create. Will be `development` by default.",
				Optional:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: "The ID the Environment is referenced as.",
						Required:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "The Human-friendly name for the Environment.",
						Required:            true,
					},
					"type": schema.StringAttribute{
						MarkdownDescription: "The Environment Type. This is used for organizing and managing Environments.",
						Required:            true,
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Read:   true,
				Delete: true,
			}),
		},
	}
}

func (r *ResourceApplication) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func parseApplicationResponse(res *client.ApplicationResponse, data *ApplicationModel) {
	data.ID = types.StringValue(res.Id)
	data.Name = types.StringValue(res.Name)
}

func (r *ResourceApplication) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ApplicationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	name := data.Name.ValueString()

	var env *client.EnvironmentBaseRequest

	if data.Env != nil {
		env = &client.EnvironmentBaseRequest{
			Id:   data.Env.ID.ValueString(),
			Name: data.Env.Name.ValueString(),
			Type: data.Env.Type.ValueString(),
		}
	}

	httpResp, err := r.client.PostOrgsOrgIdAppsWithResponse(ctx, r.orgId, client.PostOrgsOrgIdAppsJSONRequestBody{
		Id:   id,
		Name: name,
		Env:  env,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create app, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 201 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create app, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseApplicationResponse(httpResp.JSON201, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceApplication) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ApplicationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := data.Timeouts.Create(ctx, defaultApplicationReadTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var httpResp *client.GetOrgsOrgIdAppsAppIdResponse

	err := retry.RetryContext(ctx, readTimeout, func() *retry.RetryError {
		var err error

		httpResp, err = r.client.GetOrgsOrgIdAppsAppIdWithResponse(ctx, r.orgId, data.ID.ValueString())
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if httpResp.StatusCode() == 404 {
			return nil
		}

		if httpResp.StatusCode() != 200 {
			return retry.RetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read application, got error: %s", err))
		return
	}

	if httpResp.StatusCode() == 404 {
		resp.Diagnostics.AddWarning("Application not found", fmt.Sprintf("The app (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	parseApplicationResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceApplication) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("UNSUPPORTED_OPERATION", "Updating an application is currently not supported")
}

func (r *ResourceApplication) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ApplicationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove all active resource before removing the app
	appID := data.ID.ValueString()
	if err := deleteActiveAppResources(ctx, r.client, r.orgId, appID); err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, err.Error())
		return
	}

	deleteTimeout, diags := data.Timeouts.Create(ctx, defaultApplicationDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		httpResp, err := r.client.DeleteOrgsOrgIdAppsAppIdWithResponse(ctx, r.orgId, data.ID.ValueString())
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if httpResp.StatusCode() != 204 {
			return retry.RetryableError(fmt.Errorf("unable to delete application, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read application, got error: %s", err))
		return
	}
}

func (r *ResourceApplication) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func deleteActiveAppResources(ctx context.Context, client *humanitec.Client, orgID, appID string) error {
	envResp, err := client.GetOrgsOrgIdAppsAppIdEnvsWithResponse(ctx, orgID, appID)
	if err != nil {
		return fmt.Errorf("unable to read app envs, got error: %s", err)
	}
	if envResp.StatusCode() != 200 {
		return fmt.Errorf("unable to read app envs, unexpected status code: %d, body: %s", envResp.StatusCode(), envResp.Body)
	}

	for _, env := range *envResp.JSON200 {
		resResp, err := client.GetOrgsOrgIdAppsAppIdEnvsEnvIdResourcesWithResponse(ctx, orgID, appID, env.Id)
		if err != nil {
			return fmt.Errorf("unable to read app env res, got error: %s", err)
		}
		if resResp.StatusCode() != 200 {
			return fmt.Errorf("unable to read app env res, unexpected status code: %d, body: %s", resResp.StatusCode(), resResp.Body)
		}

		for _, res := range *resResp.JSON200 {
			delResResp, err := client.DeleteOrgsOrgIdAppsAppIdEnvsEnvIdResourcesTypeResIdWithResponse(ctx, orgID, appID, env.Id, res.Type, res.ResId)
			if err != nil {
				return fmt.Errorf("unable to delete app env res, got error: %s", err)
			}
			if delResResp.StatusCode() != 204 {
				return fmt.Errorf("unable to delete app env res, unexpected status code: %d, body: %s", delResResp.StatusCode(), delResResp.Body)
			}
		}
	}

	return nil
}
