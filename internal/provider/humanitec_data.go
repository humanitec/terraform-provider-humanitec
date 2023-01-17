package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/humanitec/humanitec-go-autogen/client"
)

type HumanitecData struct {
	Client *client.ClientWithResponses
	OrgID  string

	fetchDriversMu sync.Mutex
	driversByType  map[string]*client.DriverDefinitionResponse

	fetchTypesMu sync.Mutex
	typesByType  map[string]*client.ResourceTypeResponse
}

func (h *HumanitecData) fetchResourceDrivers(ctx context.Context) (map[string]*client.DriverDefinitionResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	h.fetchDriversMu.Lock()
	defer h.fetchDriversMu.Unlock()

	if h.driversByType != nil {
		return h.driversByType, diags
	}

	httpResp, err := h.Client.GetOrgsOrgIdResourcesDriversWithResponse(ctx, h.OrgID)
	if err != nil {
		diags.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to get resource drivers, got error: %s", err))
		return nil, diags
	}

	if httpResp.StatusCode() != 200 {
		diags.AddError(HUM_API_ERR, fmt.Sprintf("Unable to get resource drivers, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return nil, diags
	}

	if httpResp.JSON200 == nil {
		diags.AddError(HUM_API_ERR, fmt.Sprintf("Unable to get resource drivers, missing body, body: %s", httpResp.Body))
		return nil, diags
	}

	driversByType := map[string]*client.DriverDefinitionResponse{}
	for _, d := range *httpResp.JSON200 {
		d := d
		driversByType[fmt.Sprintf("%s/%s", d.OrgId, d.Id)] = &d
	}

	h.driversByType = driversByType

	return driversByType, diags
}

func (h *HumanitecData) fetchResourceTypes(ctx context.Context) (map[string]*client.ResourceTypeResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	h.fetchTypesMu.Lock()
	defer h.fetchTypesMu.Unlock()

	if h.typesByType != nil {
		return h.typesByType, diags
	}

	httpResp, err := h.Client.GetOrgsOrgIdResourcesTypesWithResponse(ctx, h.OrgID)
	if err != nil {
		diags.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to get resource types, got error: %s", err))
		return nil, diags
	}

	if httpResp.StatusCode() != 200 {
		diags.AddError(HUM_API_ERR, fmt.Sprintf("Unable to get resource types, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return nil, diags
	}

	if httpResp.JSON200 == nil {
		diags.AddError(HUM_API_ERR, fmt.Sprintf("Unable to get resource types, missing body, body: %s", httpResp.Body))
		return nil, diags
	}

	typesByType := map[string]*client.ResourceTypeResponse{}
	for _, d := range *httpResp.JSON200 {
		d := d
		typesByType[d.Type] = &d
	}

	h.typesByType = typesByType

	return typesByType, diags
}

func (h *HumanitecData) driverByDriverType(ctx context.Context, driverType string) (*client.DriverDefinitionResponse, diag.Diagnostics) {
	driversByType, diags := h.fetchResourceDrivers(ctx)
	if diags.HasError() {
		return nil, diags
	}

	driver, ok := driversByType[driverType]
	if !ok {
		diags.AddError(HUM_INPUT_ERR, fmt.Sprintf("No resource driver found for type: %s", driverType))
		return nil, diags
	}

	return driver, diags
}

func (h *HumanitecData) resourceByType(ctx context.Context, resourceType string) (*client.ResourceTypeResponse, diag.Diagnostics) {
	resourcesByType, diags := h.fetchResourceTypes(ctx)
	if diags.HasError() {
		return nil, diags
	}

	resource, ok := resourcesByType[resourceType]
	if !ok {
		diags.AddError(HUM_INPUT_ERR, fmt.Sprintf("No resource type found for type: %s", resourceType))
		return nil, diags
	}

	return resource, diags
}

func (h *HumanitecData) DriverInputSchemaByDriverTypeOrType(ctx context.Context, driverType, resourceType string) (map[string]interface{}, diag.Diagnostics) {
	// The static driver has no input schema and matches the output schema of the resource type
	if driverType == "humanitec/static" {
		resource, diags := h.resourceByType(ctx, resourceType)
		if diags.HasError() {
			return nil, diags
		}

		return resource.OutputsSchema, diags
	}

	driver, diags := h.driverByDriverType(ctx, driverType)
	if diags.HasError() {
		return nil, diags
	}

	return driver.InputsSchema, diags
}
