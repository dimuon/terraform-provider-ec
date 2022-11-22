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

package deploymentresource_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource"
	elasticsearchv1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/elasticsearch/v1"
)

func Test_upgradeElasticsearchStrategyV1ToElasticsearchStrategyV2(t *testing.T) {
	tests := []struct {
		name string
		in   attr.Value
		want types.String
	}{
		{
			name: "Empty list",
			in:   types.List{},
			want: types.String{Null: true},
		},
		{
			name: "Empty object",
			in:   types.List{Elems: []attr.Value{types.Object{}}},
			want: types.String{Null: true},
		},
		{
			name: "With value",
			in:   types.List{Elems: []attr.Value{types.Object{Attrs: map[string]attr.Value{"type": types.String{Value: "foo"}}}}},
			want: types.String{Value: "foo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deploymentresource.ElasticsearchStrategyV1ToV2(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_upgradeElasticsearchV1ToV2(t *testing.T) {
	in := types.List{
		Elems: []attr.Value{
			types.Object{
				Attrs: map[string]attr.Value{
					"autoscale":       types.String{Value: "false"},
					"cloud_id":        types.String{Value: "terraform_acc_rr488z9lbx:dXMtZWFzdC0xLmF3cy5mb3VuZC5pbzo0NDMkZGU3YTdhODk2NjMyNDcxZThjNDQ1ZjRjNmUyZGI1ZmMk"},
					"config":          types.List{Elems: []attr.Value{}, ElemType: elasticsearchv1.ElasticsearchConfigSchema().Attributes.Type().(types.ListType).ElemType},
					"extension":       types.Set{Elems: []attr.Value{}, ElemType: elasticsearchv1.ElasticsearchExtensionSchema().Attributes.Type().(types.SetType).ElemType},
					"http_endpoint":   types.String{Value: "http://de7a7a896632471e8c445f4c6e2db5fc.us-east-1.aws.found.io:9200"},
					"https_endpoint":  types.String{Value: "https://de7a7a896632471e8c445f4c6e2db5fc.us-east-1.aws.found.io:443"},
					"ref_id":          types.String{Value: "main-elasticsearch"},
					"region":          types.String{Value: "us-east-1"},
					"remote_cluster":  types.Set{Elems: []attr.Value{}, ElemType: elasticsearchv1.ElasticsearchRemoteClusterSchema().Attributes.Type().(types.SetType).ElemType},
					"resource_id":     types.String{Value: "de7a7a896632471e8c445f4c6e2db5fc"},
					"snapshot_source": types.List{Elems: []attr.Value{}, ElemType: elasticsearchv1.ElasticsearchSnapshotSourceSchema().Attributes.Type().(types.ListType).ElemType},
					"strategy": types.List{
						Null:     true,
						Elems:    ([]attr.Value)(nil),
						ElemType: elasticsearchv1.ElasticsearchStrategySchema().Attributes.Type().(types.ListType).ElemType,
					},
					"topology": types.List{
						Elems: []attr.Value{
							types.Object{
								Attrs: map[string]attr.Value{
									"autoscaling": types.List{
										Elems: []attr.Value{
											types.Object{
												Attrs: map[string]attr.Value{
													"max_size":             types.String{Value: "4g"},
													"max_size_resource":    types.String{Value: "memory"},
													"min_size":             types.String{Value: ""},
													"min_size_resource":    types.String{Value: ""},
													"policy_override_json": types.String{Value: ""},
												},
												AttrTypes: elasticsearchv1.ElasticsearchTopologyAutoscalingSchema().Attributes.Type().(types.ListType).ElementType().(types.ObjectType).AttrTypes,
											},
										},
										ElemType: elasticsearchv1.ElasticsearchTopologyAutoscalingSchema().Attributes.Type().(types.ListType).ElemType,
									},
									"config":                    types.List{Elems: []attr.Value{}, ElemType: elasticsearchv1.ElasticsearchTopologyConfigSchema().Attributes.Type().(types.ListType).ElemType},
									"id":                        types.String{Value: "hot_content"},
									"instance_configuration_id": types.String{Value: "aws.data.highio.i3"},
									"node_roles": types.Set{
										Elems: []attr.Value{
											types.String{Value: "data_content"},
											types.String{Value: "data_hot"},
											types.String{Value: "ingest"},
											types.String{Value: "master"},
											types.String{Value: "remote_cluster_client"},
											types.String{Value: "transform"},
										},
										ElemType: types.StringType,
									},
									"node_type_data":   types.String{Value: ""},
									"node_type_ingest": types.String{Value: ""},
									"node_type_master": types.String{Value: ""},
									"node_type_ml":     types.String{Value: ""},
									"size":             types.String{Value: "1g"},
									"size_resource":    types.String{Value: "memory"},
									"zone_count":       types.Int64{Value: 2},
								},
								AttrTypes: elasticsearchv1.ElasticsearchTopologySchema().Attributes.Type().(types.ListType).ElementType().(types.ObjectType).AttrTypes,
							},
						},
						ElemType: elasticsearchv1.ElasticsearchTopologySchema().Attributes.Type().(types.ListType).ElemType,
					},
					"trust_account": types.Set{
						Elems: []attr.Value{
							types.Object{
								Attrs: map[string]attr.Value{
									"account_id":      types.String{Value: "4184822784"},
									"trust_all":       types.Bool{Value: true},
									"trust_allowlist": types.Set{Elems: []attr.Value{}, ElemType: types.StringType},
								},
								AttrTypes: elasticsearchv1.ElasticsearchTrustAccountSchema().Attributes.Type().(types.SetType).ElemType.(types.ObjectType).AttrTypes,
							},
						},
						ElemType: elasticsearchv1.ElasticsearchTrustAccountSchema().Attributes.Type().(types.SetType).ElemType,
					},
					"trust_external": types.Set{Elems: []attr.Value{}, ElemType: elasticsearchv1.ElasticsearchTrustExternalSchema().Attributes.Type().(types.SetType).ElemType},
				},
				AttrTypes: elasticsearchv1.ElasticsearchSchema().Attributes.Type().(types.ListType).ElementType().(types.ObjectType).AttrTypes,
			},
		},
		ElemType: elasticsearchv1.ElasticsearchSchema().Attributes.Type().(types.ListType).ElemType,
	}

	inOrig := types.List{
		Unknown: false,
		Null:    false,
		Elems: []attr.Value{
			types.Object{
				Unknown: false,
				Null:    false,
				Attrs: map[string]attr.Value{
					"autoscale": types.String{
						Unknown: false,
						Null:    false,
						Value:   "false",
					},
					"cloud_id": types.String{
						Unknown: false,
						Null:    false,
						Value:   "terraform_acc_rr488z9lbx:dXMtZWFzdC0xLmF3cy5mb3VuZC5pbzo0NDMkZGU3YTdhODk2NjMyNDcxZThjNDQ1ZjRjNmUyZGI1ZmMk",
					},
					"config": types.List{
						Unknown: false,
						Null:    false,
						Elems:   []attr.Value{},
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"docker_image": types.StringType,
								"plugins": types.SetType{
									ElemType: types.StringType,
								},
								"user_settings_json":          types.StringType,
								"user_settings_override_json": types.StringType,
								"user_settings_override_yaml": types.StringType,
								"user_settings_yaml":          types.StringType,
							},
						},
					},
					"extension": types.Set{
						Unknown: false,
						Null:    false,
						Elems:   []attr.Value{},
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"name":    types.StringType,
								"type":    types.StringType,
								"url":     types.StringType,
								"version": types.StringType,
							},
						},
					},
					"http_endpoint": types.String{
						Unknown: false,
						Null:    false,
						Value:   "http://de7a7a896632471e8c445f4c6e2db5fc.us-east-1.aws.found.io:9200",
					},
					"https_endpoint": types.String{
						Unknown: false,
						Null:    false,
						Value:   "https://de7a7a896632471e8c445f4c6e2db5fc.us-east-1.aws.found.io:443",
					},
					"ref_id": types.String{
						Unknown: false,
						Null:    false,
						Value:   "main-elasticsearch",
					},
					"region": types.String{
						Unknown: false,
						Null:    false,
						Value:   "us-east-1",
					},
					"remote_cluster": types.Set{
						Unknown: false,
						Null:    false,
						Elems:   []attr.Value{},
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"alias":            types.StringType,
								"deployment_id":    types.StringType,
								"ref_id":           types.StringType,
								"skip_unavailable": types.BoolType,
							},
						},
					},
					"resource_id": types.String{
						Unknown: false,
						Null:    false,
						Value:   "de7a7a896632471e8c445f4c6e2db5fc",
					},
					"snapshot_source": types.List{
						Unknown: false,
						Null:    false,
						Elems:   []attr.Value{},
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"snapshot_name":                   types.StringType,
								"source_elasticsearch_cluster_id": types.StringType,
							},
						},
					},
					"strategy": types.List{
						Unknown: false,
						Null:    true,
						Elems:   ([]attr.Value)(nil),
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"type": types.StringType,
							},
						},
					},
					"topology": types.List{
						Unknown: false,
						Null:    false,
						Elems: []attr.Value{
							types.Object{
								Unknown: false,
								Null:    false,
								Attrs: map[string]attr.Value{
									"autoscaling": types.List{
										Unknown: false,
										Null:    false,
										Elems: []attr.Value{
											types.Object{
												Unknown: false,
												Null:    false,
												Attrs: map[string]attr.Value{
													"max_size": types.String{
														Unknown: false,
														Null:    false,
														Value:   "4g",
													},
													"max_size_resource": types.String{
														Unknown: false,
														Null:    false,
														Value:   "memory",
													},
													"min_size": types.String{
														Unknown: false,
														Null:    false,
														Value:   "",
													},
													"min_size_resource": types.String{
														Unknown: false,
														Null:    false,
														Value:   "",
													},
													"policy_override_json": types.String{
														Unknown: false,
														Null:    false,
														Value:   "",
													},
												},
												AttrTypes: map[string]attr.Type{
													"max_size":             types.StringType,
													"max_size_resource":    types.StringType,
													"min_size":             types.StringType,
													"min_size_resource":    types.StringType,
													"policy_override_json": types.StringType,
												},
											},
										},
										ElemType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"max_size":             types.StringType,
												"max_size_resource":    types.StringType,
												"min_size":             types.StringType,
												"min_size_resource":    types.StringType,
												"policy_override_json": types.StringType,
											},
										},
									},
									"config": types.List{
										Unknown: false,
										Null:    false,
										Elems:   []attr.Value{},
										ElemType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"plugins": types.SetType{
													ElemType: types.StringType,
												},
												"user_settings_json":          types.StringType,
												"user_settings_override_json": types.StringType,
												"user_settings_override_yaml": types.StringType,
												"user_settings_yaml":          types.StringType,
											},
										},
									},
									"id": types.String{
										Unknown: false,
										Null:    false,
										Value:   "hot_content",
									},
									"instance_configuration_id": types.String{
										Unknown: false,
										Null:    false,
										Value:   "aws.data.highio.i3",
									},
									"node_roles": types.Set{
										Unknown: false,
										Null:    false,
										Elems: []attr.Value{
											types.String{
												Unknown: false,
												Null:    false,
												Value:   "data_content",
											},
											types.String{
												Unknown: false,
												Null:    false,
												Value:   "data_hot",
											},
											types.String{
												Unknown: false,
												Null:    false,
												Value:   "ingest",
											},
											types.String{
												Unknown: false,
												Null:    false,
												Value:   "master",
											},
											types.String{
												Unknown: false,
												Null:    false,
												Value:   "remote_cluster_client",
											},
											types.String{
												Unknown: false,
												Null:    false,
												Value:   "transform",
											},
										},
										ElemType: types.StringType,
									},
									"node_type_data": types.String{
										Unknown: false,
										Null:    false,
										Value:   "",
									},
									"node_type_ingest": types.String{
										Unknown: false,
										Null:    false,
										Value:   "",
									},
									"node_type_master": types.String{
										Unknown: false,
										Null:    false,
										Value:   "",
									},
									"node_type_ml": types.String{
										Unknown: false,
										Null:    false,
										Value:   "",
									},
									"size": types.String{
										Unknown: false,
										Null:    false,
										Value:   "1g",
									},
									"size_resource": types.String{
										Unknown: false,
										Null:    false,
										Value:   "memory",
									},
									"zone_count": types.Int64{
										Unknown: false,
										Null:    false,
										Value:   2,
									},
								},
								AttrTypes: map[string]attr.Type{
									"autoscaling": types.ListType{
										ElemType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"max_size":             types.StringType,
												"max_size_resource":    types.StringType,
												"min_size":             types.StringType,
												"min_size_resource":    types.StringType,
												"policy_override_json": types.StringType,
											},
										},
									},
									"config": types.ListType{
										ElemType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"plugins": types.SetType{
													ElemType: types.StringType,
												},
												"user_settings_json":          types.StringType,
												"user_settings_override_json": types.StringType,
												"user_settings_override_yaml": types.StringType,
												"user_settings_yaml":          types.StringType,
											},
										},
									},
									"id":                        types.StringType,
									"instance_configuration_id": types.StringType,
									"node_roles": types.SetType{
										ElemType: types.StringType,
									},
									"node_type_data":   types.StringType,
									"node_type_ingest": types.StringType,
									"node_type_master": types.StringType,
									"node_type_ml":     types.StringType,
									"size":             types.StringType,
									"size_resource":    types.StringType,
									"zone_count":       types.Int64Type,
								},
							},
						},
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"autoscaling": types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"max_size":             types.StringType,
											"max_size_resource":    types.StringType,
											"min_size":             types.StringType,
											"min_size_resource":    types.StringType,
											"policy_override_json": types.StringType,
										},
									},
								},
								"config": types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"plugins": types.SetType{
												ElemType: types.StringType,
											},
											"user_settings_json":          types.StringType,
											"user_settings_override_json": types.StringType,
											"user_settings_override_yaml": types.StringType,
											"user_settings_yaml":          types.StringType,
										},
									},
								},
								"id":                        types.StringType,
								"instance_configuration_id": types.StringType,
								"node_roles": types.SetType{
									ElemType: types.StringType,
								},
								"node_type_data":   types.StringType,
								"node_type_ingest": types.StringType,
								"node_type_master": types.StringType,
								"node_type_ml":     types.StringType,
								"size":             types.StringType,
								"size_resource":    types.StringType,
								"zone_count":       types.Int64Type,
							},
						},
					},
					"trust_account": types.Set{
						Unknown: false,
						Null:    false,
						Elems: []attr.Value{
							types.Object{
								Unknown: false,
								Null:    false,
								Attrs: map[string]attr.Value{
									"account_id": types.String{
										Unknown: false,
										Null:    false,
										Value:   "4184822784",
									},
									"trust_all": types.Bool{
										Unknown: false,
										Null:    false,
										Value:   true,
									},
									"trust_allowlist": types.Set{
										Unknown:  false,
										Null:     false,
										Elems:    []attr.Value{},
										ElemType: types.StringType,
									},
								},
								AttrTypes: map[string]attr.Type{
									"account_id": types.StringType,
									"trust_all":  types.BoolType,
									"trust_allowlist": types.SetType{
										ElemType: types.StringType,
									},
								},
							},
						},
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"account_id": types.StringType,
								"trust_all":  types.BoolType,
								"trust_allowlist": types.SetType{
									ElemType: types.StringType,
								},
							},
						},
					},
					"trust_external": types.Set{
						Unknown: false,
						Null:    false,
						Elems:   []attr.Value{},
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"relationship_id": types.StringType,
								"trust_all":       types.BoolType,
								"trust_allowlist": types.SetType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
				AttrTypes: map[string]attr.Type{
					"autoscale": types.StringType,
					"cloud_id":  types.StringType,
					"config": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"docker_image": types.StringType,
								"plugins": types.SetType{
									ElemType: types.StringType,
								},
								"user_settings_json":          types.StringType,
								"user_settings_override_json": types.StringType,
								"user_settings_override_yaml": types.StringType,
								"user_settings_yaml":          types.StringType,
							},
						},
					},
					"extension": types.SetType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"name":    types.StringType,
								"type":    types.StringType,
								"url":     types.StringType,
								"version": types.StringType,
							},
						},
					},
					"http_endpoint":  types.StringType,
					"https_endpoint": types.StringType,
					"ref_id":         types.StringType,
					"region":         types.StringType,
					"remote_cluster": types.SetType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"alias":            types.StringType,
								"deployment_id":    types.StringType,
								"ref_id":           types.StringType,
								"skip_unavailable": types.BoolType,
							},
						},
					},
					"resource_id": types.StringType,
					"snapshot_source": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"snapshot_name":                   types.StringType,
								"source_elasticsearch_cluster_id": types.StringType,
							},
						},
					},
					"strategy": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"type": types.StringType,
							},
						},
					},
					"topology": types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"autoscaling": types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"max_size":             types.StringType,
											"max_size_resource":    types.StringType,
											"min_size":             types.StringType,
											"min_size_resource":    types.StringType,
											"policy_override_json": types.StringType,
										},
									},
								},
								"config": types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"plugins": types.SetType{
												ElemType: types.StringType,
											},
											"user_settings_json":          types.StringType,
											"user_settings_override_json": types.StringType,
											"user_settings_override_yaml": types.StringType,
											"user_settings_yaml":          types.StringType,
										},
									},
								},
								"id":                        types.StringType,
								"instance_configuration_id": types.StringType,
								"node_roles": types.SetType{
									ElemType: types.StringType,
								},
								"node_type_data":   types.StringType,
								"node_type_ingest": types.StringType,
								"node_type_master": types.StringType,
								"node_type_ml":     types.StringType,
								"size":             types.StringType,
								"size_resource":    types.StringType,
								"zone_count":       types.Int64Type,
							},
						},
					},
					"trust_account": types.SetType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"account_id": types.StringType,
								"trust_all":  types.BoolType,
								"trust_allowlist": types.SetType{
									ElemType: types.StringType,
								},
							},
						},
					},
					"trust_external": types.SetType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"relationship_id": types.StringType,
								"trust_all":       types.BoolType,
								"trust_allowlist": types.SetType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
		},
		ElemType: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"autoscale": types.StringType,
				"cloud_id":  types.StringType,
				"config": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"docker_image": types.StringType,
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
				},
				"extension": types.SetType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"name":    types.StringType,
							"type":    types.StringType,
							"url":     types.StringType,
							"version": types.StringType,
						},
					},
				},
				"http_endpoint":  types.StringType,
				"https_endpoint": types.StringType,
				"ref_id":         types.StringType,
				"region":         types.StringType,
				"remote_cluster": types.SetType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"alias":            types.StringType,
							"deployment_id":    types.StringType,
							"ref_id":           types.StringType,
							"skip_unavailable": types.BoolType,
						},
					},
				},
				"resource_id": types.StringType,
				"snapshot_source": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"snapshot_name":                   types.StringType,
							"source_elasticsearch_cluster_id": types.StringType,
						},
					},
				},
				"strategy": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type": types.StringType,
						},
					},
				},
				"topology": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"autoscaling": types.ListType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"max_size":             types.StringType,
										"max_size_resource":    types.StringType,
										"min_size":             types.StringType,
										"min_size_resource":    types.StringType,
										"policy_override_json": types.StringType,
									},
								},
							},
							"config": types.ListType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"plugins": types.SetType{
											ElemType: types.StringType,
										},
										"user_settings_json":          types.StringType,
										"user_settings_override_json": types.StringType,
										"user_settings_override_yaml": types.StringType,
										"user_settings_yaml":          types.StringType,
									},
								},
							},
							"id":                        types.StringType,
							"instance_configuration_id": types.StringType,
							"node_roles": types.SetType{
								ElemType: types.StringType,
							},
							"node_type_data":   types.StringType,
							"node_type_ingest": types.StringType,
							"node_type_master": types.StringType,
							"node_type_ml":     types.StringType,
							"size":             types.StringType,
							"size_resource":    types.StringType,
							"zone_count":       types.Int64Type,
						},
					},
				},
				"trust_account": types.SetType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"account_id": types.StringType,
							"trust_all":  types.BoolType,
							"trust_allowlist": types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
				"trust_external": types.SetType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"relationship_id": types.StringType,
							"trust_all":       types.BoolType,
							"trust_allowlist": types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
		},
	}

	want := types.Object{
		Unknown: false,
		Null:    false,
		Attrs: map[string]attr.Value{
			"autoscale": types.String{
				Unknown: false,
				Null:    false,
				Value:   "false",
			},
			"cloud_id": types.String{
				Unknown: false,
				Null:    false,
				Value:   "terraform_acc_rr488z9lbx:dXMtZWFzdC0xLmF3cy5mb3VuZC5pbzo0NDMkZGU3YTdhODk2NjMyNDcxZThjNDQ1ZjRjNmUyZGI1ZmMk",
			},
			"cold_tier": types.Object{
				Unknown: false,
				Null:    true,
				Attrs:   (map[string]attr.Value)(nil),
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
			"config": types.Object{
				Unknown: false,
				Null:    true,
				Attrs:   (map[string]attr.Value)(nil),
				AttrTypes: map[string]attr.Type{
					"docker_image": types.StringType,
					"plugins": types.SetType{
						ElemType: types.StringType,
					},
					"user_settings_json":          types.StringType,
					"user_settings_override_json": types.StringType,
					"user_settings_override_yaml": types.StringType,
					"user_settings_yaml":          types.StringType,
				},
			},
			"coordinating_tier": types.Object{
				Unknown: false,
				Null:    true,
				Attrs:   (map[string]attr.Value)(nil),
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
			"extension": types.Set{
				Unknown: false,
				Null:    false,
				Elems:   []attr.Value{},
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":    types.StringType,
						"type":    types.StringType,
						"url":     types.StringType,
						"version": types.StringType,
					},
				},
			},
			"frozen_tier": types.Object{
				Unknown: false,
				Null:    true,
				Attrs:   (map[string]attr.Value)(nil),
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
			"hot_content_tier": types.Object{
				Unknown: false,
				Null:    true,
				Attrs:   (map[string]attr.Value)(nil),
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
			"http_endpoint": types.String{
				Unknown: false,
				Null:    false,
				Value:   "http://de7a7a896632471e8c445f4c6e2db5fc.us-east-1.aws.found.io:9200",
			},
			"https_endpoint": types.String{
				Unknown: false,
				Null:    false,
				Value:   "https://de7a7a896632471e8c445f4c6e2db5fc.us-east-1.aws.found.io:443",
			},
			"master_tier": types.Object{
				Unknown: false,
				Null:    true,
				Attrs:   (map[string]attr.Value)(nil),
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
			"ml_tier": types.Object{
				Unknown: false,
				Null:    true,
				Attrs:   (map[string]attr.Value)(nil),
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
			"ref_id": types.String{
				Unknown: false,
				Null:    false,
				Value:   "main-elasticsearch",
			},
			"region": types.String{
				Unknown: false,
				Null:    false,
				Value:   "us-east-1",
			},
			"remote_cluster": types.Set{
				Unknown: false,
				Null:    false,
				Elems:   []attr.Value{},
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"alias":            types.StringType,
						"deployment_id":    types.StringType,
						"ref_id":           types.StringType,
						"skip_unavailable": types.BoolType,
					},
				},
			},
			"resource_id": types.String{
				Unknown: false,
				Null:    false,
				Value:   "de7a7a896632471e8c445f4c6e2db5fc",
			},
			"snapshot_source": types.Object{
				Unknown: false,
				Null:    true,
				Attrs:   (map[string]attr.Value)(nil),
				AttrTypes: map[string]attr.Type{
					"snapshot_name":                   types.StringType,
					"source_elasticsearch_cluster_id": types.StringType,
				},
			},
			"strategy": types.String{
				Unknown: false,
				Null:    true,
				Value:   "",
			},
			"trust_account": types.Set{
				Unknown: false,
				Null:    false,
				Elems: []attr.Value{
					types.Object{
						Unknown: false,
						Null:    false,
						Attrs: map[string]attr.Value{
							"account_id": types.String{
								Unknown: false,
								Null:    false,
								Value:   "4184822784",
							},
							"trust_all": types.Bool{
								Unknown: false,
								Null:    false,
								Value:   true,
							},
							"trust_allowlist": types.Set{
								Unknown:  false,
								Null:     false,
								Elems:    []attr.Value{},
								ElemType: types.StringType,
							},
						},
						AttrTypes: map[string]attr.Type{
							"account_id": types.StringType,
							"trust_all":  types.BoolType,
							"trust_allowlist": types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"account_id": types.StringType,
						"trust_all":  types.BoolType,
						"trust_allowlist": types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			"trust_external": types.Set{
				Unknown: false,
				Null:    false,
				Elems:   []attr.Value{},
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"relationship_id": types.StringType,
						"trust_all":       types.BoolType,
						"trust_allowlist": types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			"warm_tier": types.Object{
				Unknown: false,
				Null:    true,
				Attrs:   (map[string]attr.Value)(nil),
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
		},
		AttrTypes: map[string]attr.Type{
			"autoscale": types.StringType,
			"cloud_id":  types.StringType,
			"cold_tier": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
			"config": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"docker_image": types.StringType,
					"plugins": types.SetType{
						ElemType: types.StringType,
					},
					"user_settings_json":          types.StringType,
					"user_settings_override_json": types.StringType,
					"user_settings_override_yaml": types.StringType,
					"user_settings_yaml":          types.StringType,
				},
			},
			"coordinating_tier": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
			"extension": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":    types.StringType,
						"type":    types.StringType,
						"url":     types.StringType,
						"version": types.StringType,
					},
				},
			},
			"frozen_tier": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
			"hot_content_tier": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
			"http_endpoint":  types.StringType,
			"https_endpoint": types.StringType,
			"master_tier": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
			"ml_tier": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
			"ref_id": types.StringType,
			"region": types.StringType,
			"remote_cluster": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"alias":            types.StringType,
						"deployment_id":    types.StringType,
						"ref_id":           types.StringType,
						"skip_unavailable": types.BoolType,
					},
				},
			},
			"resource_id": types.StringType,
			"snapshot_source": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"snapshot_name":                   types.StringType,
					"source_elasticsearch_cluster_id": types.StringType,
				},
			},
			"strategy": types.StringType,
			"trust_account": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"account_id": types.StringType,
						"trust_all":  types.BoolType,
						"trust_allowlist": types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			"trust_external": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"relationship_id": types.StringType,
						"trust_all":       types.BoolType,
						"trust_allowlist": types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			"warm_tier": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"autoscaling": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"max_size":             types.StringType,
							"max_size_resource":    types.StringType,
							"min_size":             types.StringType,
							"min_size_resource":    types.StringType,
							"policy_override_json": types.StringType,
						},
					},
					"config": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"plugins": types.SetType{
								ElemType: types.StringType,
							},
							"user_settings_json":          types.StringType,
							"user_settings_override_json": types.StringType,
							"user_settings_override_yaml": types.StringType,
							"user_settings_yaml":          types.StringType,
						},
					},
					"instance_configuration_id": types.StringType,
					"node_roles": types.SetType{
						ElemType: types.StringType,
					},
					"node_type_data":   types.StringType,
					"node_type_ingest": types.StringType,
					"node_type_master": types.StringType,
					"node_type_ml":     types.StringType,
					"size":             types.StringType,
					"size_resource":    types.StringType,
					"zone_count":       types.Int64Type,
				},
			},
		},
	}
	got := deploymentresource.ElasticsearchV1ToV2(in)

	assert.Equal(t, inOrig, in)
	assert.Equal(t, want, got)
}
