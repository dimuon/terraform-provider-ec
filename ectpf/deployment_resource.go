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

package ectpf

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tpfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/elastic/terraform-provider-ec/ectpf/ecresource/deploymentresource"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tpfprovider.ResourceType = deploymentResourceType{}
var _ resource.Resource = deploymentResource{}

// var _ resource.ResourceWithImportState = deploymentResource{}

type deploymentResourceType struct{}

func (t deploymentResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Elastic Cloud Deployment resource",

		Attributes: map[string]tfsdk.Attribute{
			"version": {
				Type:        types.StringType,
				Description: "Required Elastic Stack version to use for all of the deployment resources",
				Required:    true,
			},
			"region": {
				Type:        types.StringType,
				Description: `Required ESS region where to create the deployment, for ECE environments "ece-region" must be set`,
				Required:    true,
			},
			"deployment_template_id": {
				Type:        types.StringType,
				Description: "Required Deployment Template identifier to create the deployment from",
				Required:    true,
			},
			"name": {
				Type:        types.StringType,
				Description: "Optional name for the deployment",
				Optional:    true,
			},
		},

		Blocks: map[string]tfsdk.Block{
			"elasticsearch": {
				NestingMode: tfsdk.BlockNestingModeList,
				MaxItems:    1,
				Attributes: map[string]tfsdk.Attribute{
					"autoscale": {
						Type:        types.StringType,
						Description: `Enable or disable autoscaling. Defaults to the setting coming from the deployment template. Accepted values are "true" or "false".`,
						Computed:    true,
						Optional:    true,
						// ValidateFunc: func(i interface{}, s string) ([]string, []error) {
						// 	if _, err := strconv.ParseBool(i.(string)); err != nil {
						// 		return nil, []error{
						// 			fmt.Errorf("failed parsing autoscale value: %w", err),
						// 		}
						// 	}
						// 	return nil, nil
						// },
					},

					"resource_id": {
						Type:        types.StringType,
						Description: "The Elasticsearch resource unique identifier",
						Computed:    true,
					},
					"region": {
						Type:        types.StringType,
						Description: "The Elasticsearch resource region",
						Computed:    true,
					},
					"https_endpoint": {
						Type:        types.StringType,
						Description: "The Elasticsearch resource HTTPs endpoint",
						Computed:    true,
					},
				},

				Blocks: map[string]tfsdk.Block{
					"topology": {
						NestingMode: tfsdk.BlockNestingModeList,
						MinItems:    0,
						Attributes: map[string]tfsdk.Attribute{
							"id": {
								Type:        types.StringType,
								Description: `Required topology ID from the deployment template`,
								Required:    true,
							},
						},
					},
				},

				Description: "Required Elasticsearch resource definition",
			},
		},
	}, nil
}

func (t deploymentResourceType) NewResource(ctx context.Context, in tpfprovider.Provider) (resource.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return deploymentResource{
		provider: provider,
	}, diags
}

type deploymentResource struct {
	provider scaffoldingProvider
}

func (r deploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var cfg deploymentresource.DeploymentData
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan deploymentresource.DeploymentData
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deploymentResource, errors := deploymentresource.Create(ctx, r.provider.client, &cfg, &plan)

	if len(errors) > 0 {
		for _, err := range errors {
			resp.Diagnostics.AddError(
				"Cannot create deployment resource",
				err.Error(),
			)
		}
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// example, err := d.provider.client.CreateExample(...)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	// data.Id = types.String{Value: "example-id"}

	// write logs using the tflog package
	// see https://pkg.go.dev/github.com/hashicorp/terraform-plugin-log/tflog
	// for more information
	tflog.Trace(ctx, "created a resource")

	diags = resp.State.Set(ctx, &deploymentResource)
	resp.Diagnostics.Append(diags...)
}

func (r deploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deploymentresource.DeploymentData

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// example, err := d.provider.client.ReadExample(...)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }
	//  r.provider.client

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r deploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data deploymentresource.DeploymentData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// example, err := d.provider.client.UpdateExample(...)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r deploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data deploymentresource.DeploymentData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// example, err := d.provider.client.DeleteExample(...)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

// func (r deploymentResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
// 	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
// }
