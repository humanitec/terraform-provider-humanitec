package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
	"github.com/humanitec/terraform-provider-humanitec/internal/hashcode"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &UsersDataSource{}

func NewUserGroupDataSource() datasource.DataSource {
	return &UserGroupsDataSource{}
}

// UsersDataSource defines the data source implementation.
type UserGroupsDataSource struct {
	client *humanitec.Client
	orgId  string
}

// UsersDataSourceModel describes the data source data model.
type UserGroupsDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Filter types.Object `tfsdk:"filter"`
	Groups types.List   `tfsdk:"groups"`
}

type UserGroupsFilterDataSourceModel struct {
	Id      types.String `tfsdk:"id"`
	GroupId types.String `tfsdk:"group_id"`
	IdpId   types.String `tfsdk:"idp_id"`
}

var userGroupAttrTypes = map[string]attr.Type{
	"id":         types.StringType,
	"group_id":   types.StringType,
	"role":       types.StringType,
	"idp_id":     types.StringType,
	"created_at": types.StringType,
}

func (d *UserGroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_groups"
}

func (d *UserGroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Humanitec user groups",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The identifier assigned to the User Group by Humanitec",
					},
					"group_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The name of the group in the IdP. It can be specified only if idp_id is specified as well",
					},
					"idp_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The identifier of the IdP the group belongs to, as it is registered with Humanitec",
					},
				},
				Optional: true,
			},
			"groups": schema.ListAttribute{
				ElementType: types.ObjectType{
					AttrTypes: userGroupAttrTypes,
				},
				Computed: true,
			},
		},
	}
}

func (d *UserGroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = resdata.Client
	d.orgId = resdata.OrgID
}

func (d *UserGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserGroupsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	typeGroup := client.SubjectTypeEnumGroup
	httpResp, err := d.client.ListUserRolesInOrgWithResponse(ctx, d.orgId, &client.ListUserRolesInOrgParams{Type: &typeGroup})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to list user groups, got error: %s", err))
		return
	}
	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to list user groups, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	userGroupIds := []string{}
	userGroups := []basetypes.ObjectValue{}
	for _, userGroupRole := range *httpResp.JSON200 {
		isUserRoleMatchesFilters, diags := matchesUserGroupFilters(ctx, data.Filter, userGroupRole)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if isUserRoleMatchesFilters {
			userGroup, diags := types.ObjectValueFrom(ctx, userGroupAttrTypes, &UserGroupModel{
				ID:        types.StringValue(userGroupRole.Id),
				GroupId:   types.StringValue(userGroupRole.Name),
				Role:      types.StringValue(userGroupRole.Role),
				IdPId:     types.StringValue(*userGroupRole.IdpId),
				CreatedAt: types.StringValue(userGroupRole.CreatedAt),
			})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			userGroupIds = append(userGroupIds, userGroupRole.Id)
			userGroups = append(userGroups, userGroup)
		}
	}

	userGroupsList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: userGroupAttrTypes}, userGroups)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Groups = userGroupsList
	data.ID = types.StringValue(hashcode.Strings(userGroupIds))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func matchesUserGroupFilters(ctx context.Context, filter basetypes.ObjectValue, userGroupRole client.UserRoleResponse) (bool, diag.Diagnostics) {
	var id, idpId, groupId *string
	if !filter.IsNull() {
		var parsedFilter UserGroupsFilterDataSourceModel
		diags := filter.As(ctx, &parsedFilter, basetypes.ObjectAsOptions{})
		if len(diags) != 0 {
			return false, diags
		}

		id = parsedFilter.Id.ValueStringPointer()
		idpId = parsedFilter.IdpId.ValueStringPointer()
		groupId = parsedFilter.GroupId.ValueStringPointer()
		if groupId != nil && idpId == nil {
			diags := diag.Diagnostics{}
			diags.AddError("UNSUPPORTED_OPERATION", "It is required to specify the idp_id along with the group_id")
			return false, diags
		}
	}

	matchesIdIfSet := id == nil || userGroupRole.Id == *id
	matchesIdpIdSet := idpId == nil || (userGroupRole.IdpId != nil && *userGroupRole.IdpId == *idpId)
	matchesIdpAndGroupIdIfSet := idpId == nil || *userGroupRole.IdpId == *idpId && (groupId == nil || userGroupRole.Name == *groupId)

	return matchesIdIfSet && matchesIdpIdSet && matchesIdpAndGroupIdIfSet, diag.Diagnostics{}
}
