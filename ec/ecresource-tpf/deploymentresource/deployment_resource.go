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
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/elastic/cloud-sdk-go/pkg/api"

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/planmodifier"
	"github.com/elastic/terraform-provider-ec/ec/internal/validators"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithGetSchema = &Resource{}
var _ resource.ResourceWithMetadata = &Resource{}

// var _ resource.ResourceWithImportState = &Resource{}

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = Resource{}
// These constants are only used to determine whether or not a dedicated
// tier of masters or ingest (coordinating) nodes are set.
const (
	dataTierRolePrefix   = "data_"
	ingestDataTierRole   = "ingest"
	masterDataTierRole   = "master"
	autodetect           = "autodetect"
	growAndShrink        = "grow_and_shrink"
	rollingGrowAndShrink = "rolling_grow_and_shrink"
	rollingAll           = "rolling_all"
)

// List of update strategies availables.
var strategiesList = []string{
	autodetect, growAndShrink, rollingGrowAndShrink, rollingAll,
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_deployment"
}

func (r *Resource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Elastic Cloud Deployment resource",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
				Optional: true,
			},
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
			"traffic_filter": {
				Type: types.SetType{
					ElemType: types.StringType,
				},
				Optional:    true,
				Description: "Optional list of traffic filters to apply to this deployment.",
			},
			"tags": {
				Description: "Optional map of deployment tags",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
		},

		Blocks: map[string]tfsdk.Block{
			"elasticsearch": {
				Description: "Required Elasticsearch resource definition",
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
						Computed:    true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifier.DefaultValue(types.String{Value: "main-elasticsearch"}),
						},
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

					"config": {
						NestingMode: tfsdk.BlockNestingModeList,
						MinItems:    0,
						MaxItems:    1,
						// TODO
						// DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
						Description: `Optional Elasticsearch settings which will be applied to all topologies unless overridden on the topology element`,
						Attributes: map[string]tfsdk.Attribute{
							"docker_image": {
								Type:        types.StringType,
								Description: "Optionally override the docker image the Elasticsearch nodes will use. Note that this field will only work for internal users only.",
								Optional:    true,
							},
							"plugins": {
								Type: types.SetType{
									ElemType: types.StringType,
								},
								Description: "List of Elasticsearch supported plugins, which vary from version to version. Check the Stack Pack version to see which plugins are supported for each version. This is currently only available from the UI and [ecctl](https://www.elastic.co/guide/en/ecctl/master/ecctl_stack_list.html)",
								Optional:    true,
							},
							"user_settings_json": {
								Type:        types.StringType,
								Description: `JSON-formatted user level "elasticsearch.yml" setting overrides`,
								Optional:    true,
							},
							"user_settings_override_json": {
								Type:        types.StringType,
								Description: `JSON-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
								Optional:    true,
							},
							"user_settings_yaml": {
								Type:        types.StringType,
								Description: `YAML-formatted user level "elasticsearch.yml" setting overrides`,
								Optional:    true,
							},
							"user_settings_override_yaml": {
								Type:        types.StringType,
								Description: `YAML-formatted admin (ECE) level "elasticsearch.yml" setting overrides`,
								Optional:    true,
							},
						},
					},

					"remote_cluster": {
						NestingMode: tfsdk.BlockNestingModeSet,
						MinItems:    0,
						Description: "Optional Elasticsearch remote clusters to configure for the Elasticsearch resource, can be set multiple times",
						Attributes: map[string]tfsdk.Attribute{
							"deployment_id": {
								Description: "Remote deployment ID",
								Type:        types.StringType,
								// TODO fix examples/deployment_css/deployment.tf#61
								// Validators:  []tfsdk.AttributeValidator{validators.Length(32, 32)},
								Required: true,
							},
							"alias": {
								Description: "Alias for this Cross Cluster Search binding",
								Type:        types.StringType,
								// TODO fix examples/deployment_css/deployment.tf#62
								// Validators:  []tfsdk.AttributeValidator{validators.NotEmpty()},
								Required: true,
							},
							"ref_id": {
								Description: `Remote elasticsearch "ref_id", it is best left to the default value`,
								Type:        types.StringType,
								Computed:    true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifier.DefaultValue(types.String{Value: "main-elasticsearch"}),
								},
								Optional: true,
							},
							"skip_unavailable": {
								Description: "If true, skip the cluster during search when disconnected",
								Type:        types.BoolType,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifier.DefaultValue(types.Bool{Value: false}),
								},
								Optional: true,
							},
						},
					},

					"snapshot_source": {
						NestingMode: tfsdk.BlockNestingModeList,
						Description: "Optional snapshot source settings. Restore data from a snapshot of another deployment.",
						MinItems:    0,
						MaxItems:    1,
						Attributes: map[string]tfsdk.Attribute{
							"source_elasticsearch_cluster_id": {
								Description: "ID of the Elasticsearch cluster that will be used as the source of the snapshot",
								Type:        types.StringType,
								Required:    true,
							},
							"snapshot_name": {
								Description: "Name of the snapshot to restore. Use '__latest_success__' to get the most recent successful snapshot.",
								Type:        types.StringType,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifier.DefaultValue(types.String{Value: "__latest_success__"}),
								},
								Optional: true,
							},
						},
					},

					"extension": {
						NestingMode: tfsdk.BlockNestingModeSet,
						Description: "Optional Elasticsearch extensions such as custom bundles or plugins.",
						MinItems:    0,
						Attributes: map[string]tfsdk.Attribute{
							"name": {
								Description: "Extension name.",
								Type:        types.StringType,
								Required:    true,
							},
							"type": {
								Description: "Extension type, only `bundle` or `plugin` are supported.",
								Type:        types.StringType,
								Required:    true,
								Validators:  []tfsdk.AttributeValidator{validators.OneOf([]string{"bundle", "plugin"})},
							},
							"version": {
								Description: "Elasticsearch compatibility version. Bundles should specify major or minor versions with wildcards, such as `7.*` or `*` but **plugins must use full version notation down to the patch level**, such as `7.10.1` and wildcards are not allowed.",
								Type:        types.StringType,
								Required:    true,
							},
							"url": {
								Description: "Bundle or plugin URL, the extension URL can be obtained from the `ec_deployment_extension.<name>.url` attribute or the API and cannot be a random HTTP address that is hosted elsewhere.",
								Type:        types.StringType,
								Required:    true,
							},
						},
					},

					"trust_account": {
						NestingMode: tfsdk.BlockNestingModeSet,
						Description: "Optional Elasticsearch account trust settings.",
						MinItems:    0,
						Attributes: map[string]tfsdk.Attribute{
							"account_id": {
								Description: "The ID of the Account.",
								Type:        types.StringType,
								Required:    true,
							},
							"trust_all": {
								Description: "If true, all clusters in this account will by default be trusted and the `trust_allowlist` is ignored.",
								Type:        types.BoolType,
								Required:    true,
							},
							"trust_allowlist": {
								Description: "The list of clusters to trust. Only used when `trust_all` is false.",
								Type: types.SetType{
									ElemType: types.StringType,
								},
								Optional: true,
							},
						},
					},

					"trust_external": {
						NestingMode: tfsdk.BlockNestingModeSet,
						Description: "Optional Elasticsearch external trust settings.",
						MinItems:    0,
						Attributes: map[string]tfsdk.Attribute{
							"relationship_id": {
								Description: "The ID of the external trust relationship.",
								Type:        types.StringType,
								Required:    true,
							},
							"trust_all": {
								Description: "If true, all clusters in this account will by default be trusted and the `trust_allowlist` is ignored.",
								Type:        types.BoolType,
								Required:    true,
							},
							"trust_allowlist": {
								Description: "The list of clusters to trust. Only used when `trust_all` is false.",
								Type: types.SetType{
									ElemType: types.StringType,
								},
								Optional: true,
							},
						},
					},

					"strategy": {
						NestingMode: tfsdk.BlockNestingModeList,
						Description: "Configuration strategy settings.",
						MinItems:    0,
						MaxItems:    1,
						Attributes: map[string]tfsdk.Attribute{
							"type": {
								Description: "Configuration strategy type " + strings.Join(strategiesList, ", "),
								Type:        types.StringType,
								Required:    true,
								Validators:  []tfsdk.AttributeValidator{validators.OneOf(strategiesList)},
								// TODO
								// changes on this setting do not change the plan.
								// DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// 	return true
								// },
							},
						},
					},
				},
			},

			"kibana": {
				NestingMode: tfsdk.BlockNestingModeList,
				Description: "Optional Kibana resource definition",
				MinItems:    0,
				MaxItems:    1,
				Attributes: map[string]tfsdk.Attribute{
					"elasticsearch_cluster_ref_id": {
						Type: types.StringType,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifier.DefaultValue(types.String{Value: "main-elasticsearch"}),
						},
						Computed: true,
						Optional: true,
					},
					"ref_id": {
						Type: types.StringType,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifier.DefaultValue(types.String{Value: "main-kibana"}),
						},
						Computed: true,
						Optional: true,
					},
					"resource_id": {
						Type:     types.StringType,
						Computed: true,
					},
					"region": {
						Type:     types.StringType,
						Computed: true,
					},
					"http_endpoint": {
						Type:     types.StringType,
						Computed: true,
					},
					"https_endpoint": {
						Type:     types.StringType,
						Computed: true,
					},
				},
				Blocks: map[string]tfsdk.Block{
					"topology": {
						NestingMode: tfsdk.BlockNestingModeList,
						MinItems:    0,
						Attributes: map[string]tfsdk.Attribute{
							"instance_configuration_id": {
								Type:     types.StringType,
								Optional: true,
								Computed: true,
							},
							"size": {
								Type:     types.StringType,
								Computed: true,
								Optional: true,
							},
							"size_resource": {
								Type:        types.StringType,
								Description: `Optional size type, defaults to "memory".`,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifier.DefaultValue(types.String{Value: "memory"}),
								},
								Computed: true,
								Optional: true,
							},
							"zone_count": {
								Type:     types.Int64Type,
								Computed: true,
								Optional: true,
							},
						},
					},
					"config": {
						NestingMode: tfsdk.BlockNestingModeList,
						MinItems:    0,
						MaxItems:    1,
						// TODO
						// DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
						Description: `Optionally define the Kibana configuration options for the Kibana Server`,
						Attributes: map[string]tfsdk.Attribute{
							"docker_image": {
								Type:        types.StringType,
								Description: "Optionally override the docker image the Kibana nodes will use. Note that this field will only work for internal users only.",
								Optional:    true,
							},
							"user_settings_json": {
								Type:        types.StringType,
								Description: `An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
								Optional:    true,
							},
							"user_settings_override_json": {
								Type:        types.StringType,
								Description: `An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
								Optional:    true,
							},
							"user_settings_yaml": {
								Type:        types.StringType,
								Description: `An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
								Optional:    true,
							},
							"user_settings_override_yaml": {
								Type:        types.StringType,
								Description: `An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
								Optional:    true,
							},
						},
					},
				},
			},

			"apm": {
				NestingMode: tfsdk.BlockNestingModeList,
				Description: "Optional APM resource definition",
				MinItems:    0,
				MaxItems:    1,
				Attributes: map[string]tfsdk.Attribute{
					"elasticsearch_cluster_ref_id": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifier.DefaultValue(types.String{Value: "main-elasticsearch"}),
						},
					},
					"ref_id": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifier.DefaultValue(types.String{Value: "main-apm"}),
						},
					},
					"resource_id": {
						Type:     types.StringType,
						Computed: true,
					},
					"region": {
						Type:     types.StringType,
						Computed: true,
					},
					"http_endpoint": {
						Type:     types.StringType,
						Computed: true,
					},
					"https_endpoint": {
						Type:     types.StringType,
						Computed: true,
					},
				},

				Blocks: map[string]tfsdk.Block{
					"topology": {
						NestingMode: tfsdk.BlockNestingModeList,
						MinItems:    0,
						Attributes: map[string]tfsdk.Attribute{
							"instance_configuration_id": {
								Type:     types.StringType,
								Optional: true,
								Computed: true,
							},
							"size": {
								Type:     types.StringType,
								Computed: true,
								Optional: true,
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
								Type:     types.Int64Type,
								Computed: true,
								Optional: true,
							},
						},
					},
					"config": {
						NestingMode: tfsdk.BlockNestingModeList,
						MinItems:    0,
						MaxItems:    1,
						// TODO
						// DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
						Description: `Optionally define the Apm configuration options for the APM Server`,
						Attributes: map[string]tfsdk.Attribute{
							"docker_image": {
								Type:        types.StringType,
								Description: "Optionally override the docker image the APM nodes will use. Note that this field will only work for internal users only.",
								Optional:    true,
							},
							"debug_enabled": {
								Type:        types.BoolType,
								Description: `Optionally enable debug mode for APM servers - defaults to false`,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifier.DefaultValue(types.Bool{Value: false}),
								},
							},
							"user_settings_json": {
								Type:        types.StringType,
								Description: `An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
								Optional:    true,
							},
							"user_settings_override_json": {
								Type:        types.StringType,
								Description: `An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
								Optional:    true,
							},
							"user_settings_yaml": {
								Type:        types.StringType,
								Description: `An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
								Optional:    true,
							},
							"user_settings_override_yaml": {
								Type:        types.StringType,
								Description: `An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
								Optional:    true,
							},
						},
					},
				},
			},

			"integrations_server": {
				NestingMode: tfsdk.BlockNestingModeList,
				Description: "Optional Integrations Server resource definition",
				MinItems:    0,
				MaxItems:    1,
				Attributes: map[string]tfsdk.Attribute{
					"elasticsearch_cluster_ref_id": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifier.DefaultValue(types.String{Value: "main-elasticsearch"}),
						},
					},
					"ref_id": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifier.DefaultValue(types.String{Value: "main-integrations_server"}),
						},
					},
					"resource_id": {
						Type:     types.StringType,
						Computed: true,
					},
					"region": {
						Type:     types.StringType,
						Computed: true,
					},
					"http_endpoint": {
						Type:     types.StringType,
						Computed: true,
					},
					"https_endpoint": {
						Type:     types.StringType,
						Computed: true,
					},
				},
				Blocks: map[string]tfsdk.Block{
					"topology": {
						NestingMode: tfsdk.BlockNestingModeList,
						MinItems:    0,
						Attributes: map[string]tfsdk.Attribute{
							"instance_configuration_id": {
								Type:     types.StringType,
								Optional: true,
								Computed: true,
							},
							"size": {
								Type:     types.StringType,
								Computed: true,
								Optional: true,
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
								Type:     types.Int64Type,
								Computed: true,
								Optional: true,
							},
						},
					},

					"config": {
						NestingMode: tfsdk.BlockNestingModeList,
						MinItems:    0,
						MaxItems:    1,
						// TODO
						// DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
						Description: `Optionally define the IntegrationsServer configuration options for the IntegrationsServer Server`,
						Attributes: map[string]tfsdk.Attribute{
							"docker_image": {
								Type:        types.StringType,
								Description: "Optionally override the docker image the IntegrationsServer nodes will use. Note that this field will only work for internal users only.",
								Optional:    true,
							},
							// IntegrationsServer System Settings
							"debug_enabled": {
								Type:        types.BoolType,
								Description: `Optionally enable debug mode for IntegrationsServer servers - defaults to false`,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifier.DefaultValue(types.Bool{Value: false}),
								},
							},
							"user_settings_json": {
								Type:        types.StringType,
								Description: `An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
								Optional:    true,
							},
							"user_settings_override_json": {
								Type:        types.StringType,
								Description: `An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
								Optional:    true,
							},
							"user_settings_yaml": {
								Type:        types.StringType,
								Description: `An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
								Optional:    true,
							},
							"user_settings_override_yaml": {
								Type:        types.StringType,
								Description: `An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
								Optional:    true,
							},
						},
					},
				},
			},

			"enterprise_search": {
				NestingMode: tfsdk.BlockNestingModeList,
				Description: "Optional Enterprise Search resource definition",
				MinItems:    0,
				MaxItems:    1,
				Attributes: map[string]tfsdk.Attribute{
					"elasticsearch_cluster_ref_id": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifier.DefaultValue(types.String{Value: "main-elasticsearch"}),
						},
					},
					"ref_id": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifier.DefaultValue(types.String{Value: "main-enterprise_search"}),
						},
					},
					"resource_id": {
						Type:     types.StringType,
						Computed: true,
					},
					"region": {
						Type:     types.StringType,
						Computed: true,
					},
					"http_endpoint": {
						Type:     types.StringType,
						Computed: true,
					},
					"https_endpoint": {
						Type:     types.StringType,
						Computed: true,
					},
				},
				Blocks: map[string]tfsdk.Block{
					"topology": {
						NestingMode: tfsdk.BlockNestingModeList,
						MinItems:    0,
						Attributes: map[string]tfsdk.Attribute{
							"instance_configuration_id": {
								Type:     types.StringType,
								Optional: true,
								Computed: true,
							},
							"size": {
								Type:     types.StringType,
								Computed: true,
								Optional: true,
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
								Type:     types.Int64Type,
								Computed: true,
								Optional: true,
							},
							"node_type_appserver": {
								Type:     types.BoolType,
								Computed: true,
							},
							"node_type_connector": {
								Type:     types.BoolType,
								Computed: true,
							},
							"node_type_worker": {
								Type:     types.BoolType,
								Computed: true,
							},
						},
					},
					"config": {
						NestingMode: tfsdk.BlockNestingModeList,
						MinItems:    0,
						MaxItems:    1,
						// TODO
						// DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
						Description: `Optionally define the Enterprise Search configuration options for the Enterprise Search Server`,
						Attributes: map[string]tfsdk.Attribute{
							"docker_image": {
								Type:        types.StringType,
								Description: "Optionally override the docker image the Enterprise Search nodes will use. Note that this field will only work for internal users only.",
								Optional:    true,
							},
							"user_settings_json": {
								Type:        types.StringType,
								Description: `An arbitrary JSON object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_yaml' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (This field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
								Optional:    true,
							},
							"user_settings_override_json": {
								Type:        types.StringType,
								Description: `An arbitrary JSON object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_yaml' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
								Optional:    true,
							},
							"user_settings_yaml": {
								Type:        types.StringType,
								Description: `An arbitrary YAML object allowing ECE admins owners to set clusters' parameters (only one of this and 'user_settings_override_json' is allowed), ie in addition to the documented 'system_settings'. (This field together with 'system_settings' and 'user_settings*' defines the total set of resource settings)`,
								Optional:    true,
							},
							"user_settings_override_yaml": {
								Type:        types.StringType,
								Description: `An arbitrary YAML object allowing (non-admin) cluster owners to set their parameters (only one of this and 'user_settings_json' is allowed), provided they are on the whitelist ('user_settings_whitelist') and not on the blacklist ('user_settings_blacklist'). (These field together with 'user_settings_override*' and 'system_settings' defines the total set of resource settings)`,
								Optional:    true,
							},
						},
					},
				},
			},

			"observability": {
				NestingMode: tfsdk.BlockNestingModeList,
				Description: "Optional observability settings. Ship logs and metrics to a dedicated deployment.",
				MinItems:    0,
				MaxItems:    1,
				Attributes: map[string]tfsdk.Attribute{
					"deployment_id": {
						Type:     types.StringType,
						Required: true,
						// TODO
						// DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
						// 	// The terraform config can contain 'self' as a deployment target
						// 	// However the API will return the actual deployment-id.
						// 	// This overrides 'self' with the deployment-id so the diff will work correctly.
						// 	var deploymentID = d.Id()
						// 	var mappedOldValue = mapSelfToDeploymentID(oldValue, deploymentID)
						// 	var mappedNewValue = mapSelfToDeploymentID(newValue, deploymentID)

						// 	return mappedOldValue == mappedNewValue
						// },
					},
					"ref_id": {
						Type:     types.StringType,
						Computed: true,
						Optional: true,
					},
					"logs": {
						Type:     types.BoolType,
						Optional: true,
						Computed: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifier.DefaultValue(types.Bool{Value: true}),
						},
					},
					"metrics": {
						Type:     types.BoolType,
						Optional: true,
						Computed: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifier.DefaultValue(types.Bool{Value: true}),
						},
					},
				},
			},
		},
	}, nil
}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client, diags := internal.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
}

type Resource struct {
	client *api.API
}

func (r Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured API Client",
			"Expected configured API client. Please report this issue to the provider developers.",
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

	deploymentResource, errors := Create(ctx, r.client, &cfg, &plan)

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

func (r Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

func (r Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

func (r Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

// func (r Resource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
// 	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
// }
