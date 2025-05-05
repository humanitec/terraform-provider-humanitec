package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceUserGroup{}
var _ resource.ResourceWithImportState = &ResourceUserGroup{}

func NewResourceUserGroup() resource.Resource {
	return &ResourceUserGroup{}
}

// ResourceValue defines the resource implementation.
type ResourceUserGroup struct {
	client *humanitec.Client
	orgId  string
}

// ValueModel describes the app data model.
type UserGroupModel struct {
	ID        types.String `tfsdk:"id"`
	GroupId   types.String `tfsdk:"group_id"`
	Role      types.String `tfsdk:"role"`
	IdPId     types.String `tfsdk:"idp_id"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *ResourceUserGroup) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group"
}

func (r *ResourceUserGroup) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A collection of Organization users.",

		Attributes: map[string]schema.Attribute{
			"role": schema.StringAttribute{
				MarkdownDescription: "The role that the group should have on the organization it is created in. Could be `member`, `artefactContributor`, `manager`, `orgViewer` or `administrator`.",
				Required:            true,
			},
			"group_id": schema.StringAttribute{
				MarkdownDescription: "The name of the group in the IdP",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier assigned from Humanitec to this group",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"idp_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the IdP the group belongs to, as it is registered with Humanitec",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The time the group was first registered with Humanitec",
				Computed:            true,
			},
		},
	}
}

func (r *ResourceUserGroup) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ResourceUserGroup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *UserGroupModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	groupId := data.GroupId.ValueString()
	role := client.OrganizationRole(data.Role.ValueString())
	idpId := data.IdPId.ValueString()

	httpResp, err := r.client.CreateGroupWithResponse(ctx, r.orgId, client.CreateGroupJSONRequestBody{
		GroupId: groupId,
		IdpId:   idpId,
		Role:    role,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create a user group, got error: %s", err))
		return
	}
	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create a user group, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseUserGroupResponse(httpResp.JSON200, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceUserGroup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *UserGroupModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()

	httpResp, err := r.client.GetUserRoleInOrgWithResponse(ctx, r.orgId, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read group, got error: %s", err))
		return
	}
	if httpResp.StatusCode() == 404 {
		resp.Diagnostics.AddWarning("Group not found", fmt.Sprintf("The group (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to fetch group, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	okResp := httpResp.JSON200
	if okResp.Type != "group" {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Resource with id %s should be a group but it actually is %s", id, okResp.Type))
		return
	}

	parseUserGroupAsUserResponse(okResp, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceUserGroup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *UserGroupModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	role := data.Role.ValueString()

	httpResp, err := r.client.UpdateUserRoleInOrgWithResponse(ctx, r.orgId, id, client.RoleRequest{
		Role: role,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update group organization role, got error: %s", err))
		return
	}
	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update group organization role, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseUserGroupAsUserResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceUserGroup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *UserGroupModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	httpResp, err := r.client.DeleteUserRoleInOrgWithResponse(ctx, r.orgId, id)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete user group, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 204 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete user group, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}
}

func (r *ResourceUserGroup) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func parseUserGroupResponse(res *client.GroupWithRole, data *UserGroupModel) {
	data.ID = types.StringValue(res.Id)
	data.IdPId = types.StringValue(res.IdpId)
	data.CreatedAt = types.StringValue(res.CreatedAt.String())
	data.Role = types.StringValue(string(res.Role))
	data.GroupId = types.StringValue(res.GroupId)
}

func parseUserGroupAsUserResponse(res *client.UserRoleResponse, data *UserGroupModel) {
	data.ID = types.StringValue(res.Id)
	data.IdPId = types.StringValue(*res.IdpId)
	data.CreatedAt = types.StringValue(res.CreatedAt)
	data.Role = types.StringValue(string(res.Role))
	data.GroupId = types.StringValue(res.Name)
}
