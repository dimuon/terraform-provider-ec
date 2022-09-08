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

package deploymentresource

import (
	"context"

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tpfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tpfprovider.ResourceType = DeploymentResourceType{}
var _ resource.Resource = deploymentResource{}

// var _ resource.ResourceWithImportState = deploymentResource{}

type DeploymentResourceType struct{}

func (t DeploymentResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Elastic Cloud Deployment resource",

		Attributes: map[string]tfsdk.Attribute{
			"alias": {
				Type:     types.StringType,
				Computed: true,
			},
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
			"request_id": {
				Type:        types.StringType,
				Description: "Optional request_id to set on the create operation, only use when previous create attempts return with an error and a request_id is returned as part of the error",
				Optional:    true,
			},
			"elasticsearch_username": {
				Type:        types.StringType,
				Description: "Computed username obtained upon creating the Elasticsearch resource",
				Computed:    true,
			},
			"elasticsearch_password": {
				Type:        types.StringType,
				Description: "Computed password obtained upon creating the Elasticsearch resource",
				Computed:    true,
				Sensitive:   true,
			},
			"apm_secret_token": {
				Type:      types.StringType,
				Computed:  true,
				Sensitive: true,
			},
		},

		Blocks: map[string]tfsdk.Block{
			"elasticsearch": {
				NestingMode: tfsdk.BlockNestingModeList,
				MaxItems:    1,
				Attributes: map[string]tfsdk.Attribute{
					"autoscale": {
						Type:        types.BoolType,
						Description: `Enable or disable autoscaling. Defaults to the setting coming from the deployment template. Accepted values are "true" or "false".`,
						Computed:    true,
						Optional:    true,
					},
					"ref_id": {
						Type:        types.StringType,
						Description: "Optional ref_id to set on the Elasticsearch resource",
						Optional:    true,
						// PlanModifiers: []tfsdk.AttributePlanModifier{
						// 	planmodifier.DefaultValue(types.String{Value: "main-elasticsearch"}),
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
					"cloud_id": {
						Type:        types.StringType,
						Description: "The encoded Elasticsearch credentials to use in Beats or Logstash",
						Computed:    true,
					},
					"http_endpoint": {
						Type:        types.StringType,
						Description: "The Elasticsearch resource HTTP endpoint",
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
						Description: `Optional topology element which must be set once but can be set multiple times to compose complex topologies`,
						Attributes: map[string]tfsdk.Attribute{
							"id": {
								Type:        types.StringType,
								Description: `Required topology ID from the deployment template`,
								Required:    true,
							},
							"instance_configuration_id": {
								Type:        types.StringType,
								Description: `Computed Instance Configuration ID of the topology element`,
								Computed:    true,
							},
							"size": {
								Type:        types.StringType,
								Description: `Optional amount of memory per node in the "<size in GB>g" notation`,
								Computed:    true,
								Optional:    true,
							},
							"size_resource": {
								Type:        types.StringType,
								Description: `Optional size type, defaults to "memory".`,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifier.DefaultValue(types.String{Value: "memory"}),
								},
							},
							"zone_count": {
								Type:        types.StringType,
								Description: `Optional number of zones that the Elasticsearch cluster will span. This is used to set HA`,
								Computed:    true,
								Optional:    true,
							},
							"node_type_data": {
								Type:        types.StringType,
								Description: `The node type for the Elasticsearch Topology element (data node)`,
								Computed:    true,
								Optional:    true,
							},
							"node_type_master": {
								Type:        types.StringType,
								Description: `The node type for the Elasticsearch Topology element (master node)`,
								Computed:    true,
								Optional:    true,
							},
							"node_type_ingest": {
								Type:        types.StringType,
								Description: `The node type for the Elasticsearch Topology element (ingest node)`,
								Computed:    true,
								Optional:    true,
							},
							"node_type_ml": {
								Type:        types.StringType,
								Description: `The node type for the Elasticsearch Topology element (machine learning node)`,
								Computed:    true,
								Optional:    true,
							},
							"node_roles": {
								Type: types.SetType{
									ElemType: types.StringType,
								},
								Description: `The computed list of node roles for the current topology element`,
								Computed:    true,
							},
						},

						Blocks: map[string]tfsdk.Block{
							"autoscaling": {
								NestingMode: tfsdk.BlockNestingModeList,
								MinItems:    0,
								MaxItems:    1,
								Description: "Optional Elasticsearch autoscaling settings, such a maximum and minimum size and resources.",
								Attributes: map[string]tfsdk.Attribute{
									"max_size_resource": {
										Description: "Maximum resource type for the maximum autoscaling setting.",
										Type:        types.StringType,
										Optional:    true,
										Computed:    true,
									},
									"max_size": {
										Description: "Maximum size value for the maximum autoscaling setting.",
										Type:        types.StringType,
										Optional:    true,
										Computed:    true,
									},
									"min_size_resource": {
										Description: "Minimum resource type for the minimum autoscaling setting.",
										Type:        types.StringType,
										Optional:    true,
										Computed:    true,
									},
									"min_size": {
										Description: "Minimum size value for the minimum autoscaling setting.",
										Type:        types.StringType,
										Optional:    true,
										Computed:    true,
									},
									"policy_override_json": {
										Type:        types.StringType,
										Description: "Computed policy overrides set directly via the API or other clients.",
										Computed:    true,
									},
								},
							},

							"config": {
								NestingMode: tfsdk.BlockNestingModeList,
								MinItems:    0,
								MaxItems:    1,
								Description: `Computed read-only configuration to avoid unsetting plan settings from 'topology.elasticsearch'`,
								Attributes: map[string]tfsdk.Attribute{
									"plugins": {
										Type: types.SetType{
											ElemType: types.StringType,
										},
										Description: "List of Elasticsearch supported plugins, which vary from version to version. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html)",
										Computed:    true,
									},
									"user_settings_json": {
										Type:        types.StringType,
										Description: `JSON-formatted user level "elasticsearch.yml" setting overrides`,
										Computed:    true,
									},
									"user_settings_override_json": {
										Type:        types.StringType,
										Description: `JSON-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
										Computed:    true,
									},
									"user_settings_yaml": {
										Type:        types.StringType,
										Description: `YAML-formatted user level "elasticsearch.yml" setting overrides`,
										Computed:    true,
									},
									"user_settings_override_yaml": {
										Type:        types.StringType,
										Description: `YAML-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
										Computed:    true,
									},
								},
							},
						},
					},
				},

				Description: "Required Elasticsearch resource definition",
			},
		},
	}, nil
}

func (t DeploymentResourceType) NewResource(ctx context.Context, in tpfprovider.Provider) (resource.Resource, diag.Diagnostics) {
	p, diags := internal.ConvertProviderType(in)

	return &deploymentResource{
		provider: p,
	}, diags
}

type deploymentResource struct {
	provider internal.Provider
}

func (r deploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.provider.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var cfg DeploymentData
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan DeploymentData
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deploymentResource, errors := Create(ctx, r.provider.GetClient(), &cfg, &plan)

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
	var state DeploymentData

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
	var data DeploymentData

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
	var data DeploymentData

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
