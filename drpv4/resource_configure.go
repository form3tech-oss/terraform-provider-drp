package drpv4

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func providerData(cfg any, diags *diag.Diagnostics) *Config {
	if cfg == nil {
		diags.AddError(
			"Provider not configured",
			"Expected a configured provider before this resource is used.",
		)
		return nil
	}
	c, ok := cfg.(*Config)
	if !ok {
		diags.AddError(
			"Invalid provider configuration",
			fmt.Sprintf("Expected *Config, got %T", cfg),
		)
		return nil
	}
	return c
}

func configureResourceClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) *Config {
	// ValidateResourceConfig invokes Resource.Configure before ConfigureProvider has run in the
	// same graph walk, so ProviderData may still be nil. Skip wiring the API client until a
	// later Configure call (e.g. plan/apply) when the provider has finished configuring.
	if req.ProviderData == nil {
		return nil
	}
	return providerData(req.ProviderData, &resp.Diagnostics)
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.HasSuffix(s, "Not Found") || strings.HasSuffix(s, "Unable to get machine")
}
