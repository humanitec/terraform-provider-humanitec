package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceApplicationUser{}
var _ resource.ResourceWithImportState = &ResourceApplicationUser{}

var (
	defaultApplicationUserCreateTimeout = 30 * time.Second
	defaultApplicationUserReadTimeout   = 30 * time.Second
)

func NewResourceApplicationUser() resource.Resource {
	return &ResourceApplicationUser{}
}

// ResourceApplicationUser defines the application user implementation.
type ResourceApplicationUser struct {
	client *humanitec.Client
	orgId  string
}

// ResourceApplicationUserModel describes the application user data model.
type ResourceApplicationUserModel struct {
	ID     types.String `tfsdk:"id"`
	AppID  types.String `tfsdk:"app_id"`
	UserID types.String `tfsdk:"user_id"`

	Role types.String `tfsdk:"role"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *ResourceApplicationUser) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_user"
}

func (r *ResourceApplicationUser) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource Application User holds the mapping of role to user for an application.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The Application ID.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "The user ID that hold the role",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role that this user holds. Could be `viewer`, `developer` or `owner`.",
				Required:            true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
			}),
		},
	}
}

func (r *ResourceApplicationUser) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ResourceApplicationUser) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ResourceApplicationUserModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, defaultApplicationUserCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.UserID.ValueString()
	appID := data.AppID.ValueString()
	role := data.Role.ValueString()

	var httpResp *client.CreateUserRoleInAppResponse
	err := retry.RetryContext(ctx, createTimeout, func() *retry.RetryError {
		var err error
		httpResp, err = r.client.CreateUserRoleInAppWithResponse(ctx, r.orgId, appID, client.CreateUserRoleInAppJSONRequestBody{
			Id:   &userID,
			Role: &role,
		})
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if httpResp.StatusCode() == 404 {
			return retry.RetryableError(fmt.Errorf("waiting for application to be ready, status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		if httpResp.StatusCode() != 200 {
			return retry.NonRetryableError(fmt.Errorf("unable to create resource application user, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create resource application user, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", appID, userID))
	data.Role = types.StringValue(httpResp.JSON200.Role)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceApplicationUser) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ResourceApplicationUserModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var httpResp *client.GetUserRoleInAppResponse
	err := retry.RetryContext(ctx, defaultApplicationUserReadTimeout, func() *retry.RetryError {
		var err error
		httpResp, err = r.client.GetUserRoleInAppWithResponse(ctx, r.orgId, data.AppID.ValueString(), data.UserID.ValueString())
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if httpResp.StatusCode() == 404 {
			return retry.RetryableError(fmt.Errorf("waiting for application to be ready, status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		if httpResp.StatusCode() != 200 {
			return retry.NonRetryableError(fmt.Errorf("unable to read resource application user, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read resource application user, got error: %s", err))
		return
	}

	if httpResp.StatusCode() == 404 {
		resp.Diagnostics.AddWarning("Application user not found", fmt.Sprintf("The application user (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	data.Role = types.StringValue(httpResp.JSON200.Role)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceApplicationUser) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ResourceApplicationUserModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.UserID.ValueString()
	appID := data.AppID.ValueString()
	role := data.Role.ValueString()

	httpResp, err := r.client.UpdateUserRoleInAppWithResponse(ctx, r.orgId, appID, userID, client.UpdateUserRoleInAppJSONRequestBody{
		Role: &role,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update resource application user, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update resource application user, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", appID, userID))
	data.Role = types.StringValue(httpResp.JSON200.Role)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceApplicationUser) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ResourceApplicationUserModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.UserID.ValueString()
	applicationID := data.AppID.ValueString()

	httpResp, err := r.client.DeleteUserRoleInAppWithResponse(ctx, r.orgId, applicationID, userID)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete resource application user, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 204 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete resource application user, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}
}

func (r *ResourceApplicationUser) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: app_id/user_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
