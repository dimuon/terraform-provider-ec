// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// New provider during the migration to Terraform Plugin Framework.
// It should host all data sources and resources that migrated to TPF,
// while non-migrated data sources and resouces should be handled by `ec.Provider()`.
// Once all data sources and resources are migrated,
// the provider shall become the only one provider in the repo.

package ectpf

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tpfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/api"
)

const (
	eceOnlyText      = "Available only when targeting ECE Installations or Elasticsearch Service Private"
	saasRequiredText = "The only valid authentication mechanism for the Elasticsearch Service"

	endpointDesc     = "Endpoint where the terraform provider will point to. Defaults to \"%s\"."
	insecureDesc     = "Allow the provider to skip TLS validation on its outgoing HTTP calls."
	timeoutDesc      = "Timeout used for individual HTTP calls. Defaults to \"1m\"."
	verboseDesc      = "When set, a \"request.log\" file will be written with all outgoing HTTP requests. Defaults to \"false\"."
	verboseCredsDesc = "When set with verbose, the contents of the Authorization header will not be redacted. Defaults to \"false\"."
)

var (
	apikeyDesc   = fmt.Sprint("API Key to use for API authentication. ", saasRequiredText, ".")
	usernameDesc = fmt.Sprint("Username to use for API authentication. ", eceOnlyText, ".")
	passwordDesc = fmt.Sprint("Password to use for API authentication. ", eceOnlyText, ".")

	validURLSchemes = []string{"http", "https"}
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tpfprovider.Provider = &scaffoldingProvider{}

// provider satisfies the tfsdk.Provider interface and usually is included
// with all Resource and DataSource implementations.
type scaffoldingProvider struct {
	// client can contain the upstream provider SDK or HTTP client used to
	// communicate with the upstream service. Resource and DataSource
	// implementations can then make calls using this client.
	//
	// TODO: If appropriate, implement upstream provider SDK or HTTP client.
	// client vendorsdk.ExampleClient

	// configured is set to true at the end of the Configure method.
	// This can be used in Resource and DataSource implementations to verify
	// that the provider was previously configured.
	configured bool

	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string

	client *api.API
}

// providerData can be used to store data from the Terraform configuration.
type providerData struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Apikey   types.String `tfsdk:"apikey"`
	Insecure types.Bool   `tfsdk:"insecure"`
	Timeout  types.String `tfsdk:"timeout"`
	Verbose  types.Bool   `tfsdk:"verbose"`
}

func (p *scaffoldingProvider) Configure(ctx context.Context, req tpfprovider.ConfigureRequest, resp *tpfprovider.ConfigureResponse) {
	var data providerData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := newAPIConfig(&data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot configure Elastic API",
			err.Error(),
		)
		return
	}

	client, err := api.NewAPI(cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot initialise Elastic API client from the configuration",
			err.Error(),
		)
		return
	}

	p.client = client

	p.configured = true
}

func (p *scaffoldingProvider) GetResources(ctx context.Context) (map[string]tpfprovider.ResourceType, diag.Diagnostics) {
	return map[string]tpfprovider.ResourceType{
		"ec_deployment": deploymentResourceType{},
	}, nil
}

func (p *scaffoldingProvider) GetDataSources(ctx context.Context) (map[string]tpfprovider.DataSourceType, diag.Diagnostics) {
	return map[string]tpfprovider.DataSourceType{}, nil
}

// type endpointPlanModifier struct{}

// func (*endpointPlanModifier) Description(context.Context) string {
// 	return ""
// }

// func (*endpointPlanModifier) MarkdownDescription(context.Context) string {
// 	return ""
// }

// func (*endpointPlanModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
// 	if req.AttributeConfig.IsNull() {
// 		v, err := MultiEnvDefaultFunc([]string{"EC_ENDPOINT", "EC_HOST"}, api.ESSEndpoint)
// 		if err != nil {
// 			resp.Diagnostics.AddError(
// 				`Cannot read endpoint from default environment variables "EC_ENDPOINT", "EC_HOST"`,
// 				err.Error(),
// 			)
// 			return
// 		}
// 		resp.AttributePlan = types.String{Value: v}
// 	}
// }

func MultiEnvDefaultFunc(ks []string, def string) (string, error) {
	for _, k := range ks {
		if v := os.Getenv(k); v != "" {
			return v, nil
		}
	}
	return def, nil
}

func (p *scaffoldingProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"endpoint": {
				Description: fmt.Sprintf(endpointDesc, api.ESSEndpoint),
				Type:        types.StringType,
				Optional:    true,
			},
			"apikey": {
				Description: apikeyDesc,
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
			},
			"username": {
				Description: usernameDesc,
				Type:        types.StringType,
				Optional:    true,
			},
			"password": {
				Description: passwordDesc,
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
			},
			"insecure": {
				Description: "Allow the provider to skip TLS validation on its outgoing HTTP calls.",
				Type:        types.BoolType,
				Optional:    true,
			},
			"timeout": {
				Description: timeoutDesc,
				Type:        types.StringType,
				Optional:    true,
			},
			"verbose": {
				Description: verboseDesc,
				Type:        types.BoolType,
				Optional:    true,
			},
			"verbose_credentials": {
				Description: verboseCredsDesc,
				Type:        types.BoolType,
				Optional:    true,
			},
			"verbose_file": {
				Description: timeoutDesc,
				Type:        types.StringType,
				Optional:    true,
			},
		},
	}, nil
}

func New(version string) func() tpfprovider.Provider {
	return func() tpfprovider.Provider {
		return &scaffoldingProvider{
			version: version,
		}
	}
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in tpfprovider.Provider) (scaffoldingProvider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*scaffoldingProvider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return scaffoldingProvider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return scaffoldingProvider{}, diags
	}

	return *p, diags
}
