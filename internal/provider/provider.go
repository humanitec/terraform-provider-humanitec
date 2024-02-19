package provider

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Host  types.String `tfsdk:"host"`
	OrgID types.String `tfsdk:"org_id"`
	Token types.String `tfsdk:"token"`

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
			"host": schema.StringAttribute{
				MarkdownDescription: "Humanitec API host (or using the `HUMANITEC_HOST` environment variable)",
				Optional:            true,
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
		},
	}
}

func (p *HumanitecProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Check environment variables
	host := os.Getenv("HUMANITEC_HOST")
	if host == "" {
		host = humanitec.DefaultAPIHost
	}

	orgID := os.Getenv("HUMANITEC_ORG")
	token := os.Getenv("HUMANITEC_TOKEN")

	var data HumanitecProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check configuration data, which should take precedence over
	// environment variable data, if found.
	if data.Host.ValueString() != "" {
		host = data.Host.ValueString()
	}
	if data.OrgID.ValueString() != "" {
		orgID = data.OrgID.ValueString()
	}
	if data.Token.ValueString() != "" {
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

	var doer *http.Client
	if data.DisableSSLCertificateVerification.ValueBool() {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		doer = &http.Client{Transport: tr}
	} else {
		doer = &http.Client{}
	}

	client, err := NewHumanitecClient(host, token, p.version, doer)
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
		NewResourceEnvironmentType,
		NewResourceEnvironmentTypeUser,
		NewResourcePipeline,
		NewResourcePipelineCriteria,
		NewResourceRegistry,
		NewResourceResourceDriver,
		NewResourceRule,
		NewResourceSecretStore,
		NewResourceValue,
		NewResourceWebhook,
	}
}

func (p *HumanitecProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSourceIPRangesDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HumanitecProvider{
			version: version,
		}
	}
}
