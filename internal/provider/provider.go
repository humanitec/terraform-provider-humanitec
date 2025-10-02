package provider

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/justinrixx/retryhttp"

	"github.com/humanitec/humanitec-go-autogen"
)

// Ensure HumanitecProvider satisfies various provider interfaces.
var _ provider.Provider = &HumanitecProvider{}

// HumanitecProvider defines the provider implementation.
type HumanitecProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// HumanitecProviderModel describes the provider data model.
type HumanitecProviderModel struct {
	APIPrefix types.String `tfsdk:"api_prefix"`
	Host      types.String `tfsdk:"host"`
	OrgID     types.String `tfsdk:"org_id"`
	Token     types.String `tfsdk:"token"`
	Config    types.String `tfsdk:"config"`

	DisableSSLCertificateVerification types.Bool `tfsdk:"disable_ssl_certificate_verification"`
}

const (
	HUM_CLIENT_ERR   = "Humanitec client error"
	HUM_API_ERR      = "Humanitec API error"
	HUM_PROVIDER_ERR = "Provider error"
	HUM_INPUT_ERR    = "Input error"
)

func (p *HumanitecProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "humanitec"
	resp.Version = p.version
}

func (p *HumanitecProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform Provider for [Humanitec](https://humanitec.com/).",

		Attributes: map[string]schema.Attribute{
			"api_prefix": schema.StringAttribute{
				MarkdownDescription: "Humanitec API prefix (or using the `HUMANITEC_API_PREFIX` environment variable)",
				Optional:            true,
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "Humanitec API host (or using the `HUMANITEC_HOST` environment variable)",
				Optional:            true,
				DeprecationMessage:  "This attribute is deprecated in favor of api_prefix (`HUMANITEC_API_PREFIX` environment variable).",
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: "Humanitec Organization ID (or using the `HUMANITEC_ORG` environment variable)",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Humanitec Token (or using the `HUMANITEC_TOKEN` environment variable)",
				Optional:            true,
				Sensitive:           true,
			},
			"disable_ssl_certificate_verification": schema.BoolAttribute{
				MarkdownDescription: "Disables SSL certificate verification",
				Optional:            true,
			},
			"config": schema.StringAttribute{
				MarkdownDescription: "Location of Humanitec configuration",
				Optional:            true,
			},
		},
	}
}

func (p *HumanitecProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data HumanitecProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Reading config or .humctl file in the home directory of the system
	config, diags := readConfig(data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPrefix := config.ApiPrefix
	orgID := config.Org
	token := config.Token

	// Environment variables have precedence over config file, if found
	if hostOld := os.Getenv("HUMANITEC_HOST"); hostOld != "" {
		apiPrefix = hostOld
		resp.Diagnostics.AddWarning(
			"Environment variable HUMANITEC_HOST has been deprecated",
			"Environment variable HUMANITEC_HOST has been deprecated "+
				"please use HUMANITEC_API_PREFIX instead to set your api prefix to the terraform provider.")
	}

	if os.Getenv("HUMANITEC_API_PREFIX") != "" {
		apiPrefix = os.Getenv("HUMANITEC_API_PREFIX")
	}

	if apiPrefix == "" {
		apiPrefix = humanitec.DefaultAPIHost
	}

	if os.Getenv("HUMANITEC_ORG") != "" {
		orgID = os.Getenv("HUMANITEC_ORG")
	}

	if os.Getenv("HUMANITEC_TOKEN") != "" {
		token = os.Getenv("HUMANITEC_TOKEN")
	}

	// Check configuration data, which should take precedence over
	// environment variable data and config file, if found.

	if data.Host.ValueString() != "" {
		apiPrefix = data.Host.ValueString()
		resp.Diagnostics.AddWarning(
			"Attribute host has been deprecated",
			"Attribute hostT has been deprecated "+
				"please use api_prefix instead to set your api prefix to the terraform provider.")
	}
	if !data.APIPrefix.IsNull() {
		apiPrefix = data.APIPrefix.ValueString()
	}

	if !data.OrgID.IsNull() {
		orgID = data.OrgID.ValueString()
	}

	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddError(
			"Missing API Token Configuration",
			"While configuring the provider, the API token was not found in "+
				"the HUMANITEC_TOKEN environment variable or provider "+
				"configuration block token attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	}

	if orgID == "" {
		resp.Diagnostics.AddError(
			"Missing API Org ID Configuration",
			"While configuring the provider, the API token was not found in "+
				"the HUMANITEC_ORG environment variable or provider "+
				"configuration block org_id attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	}

	baseTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if data.DisableSSLCertificateVerification.ValueBool() {
		baseTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	doer := &http.Client{
		Timeout:   time.Minute,
		Transport: retryhttp.New(retryhttp.WithTransport(baseTransport)),
	}
	client, err := NewHumanitecClient(apiPrefix, token, p.version, doer)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Humanitec client", err.Error())
	}

	if resp.Diagnostics.HasError() {
		return
	}

	sourcedata := &HumanitecData{
		Client: client,
		OrgID:  orgID,
	}

	resp.DataSourceData = sourcedata
	resp.ResourceData = sourcedata
}

func (p *HumanitecProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewResourceAccountResource,
		NewResourceAgent,
		NewResourceApplication,
		NewResourceApplicationUser,
		NewResourceArtefactVersion,
		NewResourceDefinitionCriteriaResource,
		NewResourceDefinitionResource,
		NewResourceEnvironment,
		NewResourceEnvironmentType,
		NewResourceEnvironmentTypeUser,
		NewResourceKey,
		NewResourcePipeline,
		NewResourcePipelineCriteria,
		NewResourceRegistry,
		NewResourceResourceClass,
		NewResourceResourceType,
		NewResourceResourceDriver,
		NewResourceRule,
		NewResourceSecretStore,
		NewResourceServiceUserToken,
		NewResourceValue,
		NewResourceUser,
		NewResourceWebhook,
		NewResourceWorkloadProfileChartVersion,
		NewResourceWorkloadProfile,
		NewResourceUserGroup,
	}
}

func (p *HumanitecProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSourceIPRangesDataSource,
		NewUsersDataSource,
		NewUserGroupDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HumanitecProvider{
			version: version,
		}
	}
}
