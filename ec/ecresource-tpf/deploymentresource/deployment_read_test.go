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
	"testing"

	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_readDeployment(t *testing.T) {
	type args struct {
		res     *models.DeploymentGetResponse
		remotes models.RemoteResources
	}
	tests := []struct {
		name string
		args args
		want Deployment
		err  error
	}{
		{
			name: "flattens deployment resources",
			want: Deployment{
				Id:                   mock.ValidClusterID,
				Alias:                "my-deployment",
				Name:                 "my_deployment_name",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "us-east-1",
				Version:              "7.7.0",
				Elasticsearch: Elasticsearches{
					{
						RefId:      ec.String("main-elasticsearch"),
						ResourceId: &mock.ValidClusterID,
						Region:     ec.String("us-east-1"),
						Config: ElasticsearchConfigs{
							{
								UserSettingsYaml:         ec.String("some.setting: value"),
								UserSettingsOverrideYaml: ec.String("some.setting: value2"),
								UserSettingsJson:         ec.String("{\"some.setting\":\"value\"}"),
								UserSettingsOverrideJson: ec.String("{\"some.setting\":\"value2\"}"),
							},
						},
						Topology: ElasticsearchTopologies{
							{
								Id:                      "hot_content",
								InstanceConfigurationId: ec.String("aws.data.highio.i3"),
								Size:                    ec.String("2g"),
								SizeResource:            ec.String("memory"),
								NodeTypeData:            ec.String("true"),
								NodeTypeIngest:          ec.String("true"),
								NodeTypeMaster:          ec.String("true"),
								NodeTypeMl:              ec.String("false"),
								ZoneCount:               1,
							},
						},
					},
				},
				Kibana: Kibanas{
					{
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						RefId:                     ec.String("main-kibana"),
						ResourceId:                ec.String(mock.ValidClusterID),
						Region:                    ec.String("us-east-1"),
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("aws.kibana.r5d"),
								Size:                    ec.String("1g"),
								SizeResource:            ec.String("memory"),
								ZoneCount:               1,
							},
						},
					},
				},
				Apm: Apms{
					{
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						RefId:                     ec.String("main-apm"),
						ResourceId:                ec.String(mock.ValidClusterID),
						Region:                    ec.String("us-east-1"),
						Config: ApmConfigs{
							{
								DebugEnabled: ec.Bool(false),
							},
						},
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("aws.apm.r5d"),
								Size:                    ec.String("0.5g"),
								SizeResource:            ec.String("memory"),
								ZoneCount:               1,
							},
						},
					},
				},
				EnterpriseSearch: EnterpriseSearches{
					{
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						RefId:                     ec.String("main-enterprise_search"),
						ResourceId:                ec.String(mock.ValidClusterID),
						Region:                    ec.String("us-east-1"),
						Topology: EnterpriseSearchTopologies{
							{
								InstanceConfigurationId: ec.String("aws.enterprisesearch.m5d"),
								Size:                    ec.String("2g"),
								SizeResource:            ec.String("memory"),
								ZoneCount:               1,
								NodeTypeAppserver:       ec.Bool(true),
								NodeTypeConnector:       ec.Bool(true),
								NodeTypeWorker:          ec.Bool(true),
							},
						},
					},
				},
				Observability: Observabilities{
					{
						DeploymentId: ec.String(mock.ValidClusterID),
						RefId:        ec.String("main-elasticsearch"),
						Logs:         true,
						Metrics:      true,
					},
				},
				TrafficFilter: []string{"0.0.0.0/0", "192.168.10.0/24"},
			},
			args: args{
				res: &models.DeploymentGetResponse{
					ID:    &mock.ValidClusterID,
					Alias: "my-deployment",
					Name:  ec.String("my_deployment_name"),
					Settings: &models.DeploymentSettings{
						TrafficFilterSettings: &models.TrafficFilterSettings{
							Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
						},
						Observability: &models.DeploymentObservabilitySettings{
							Logging: &models.DeploymentLoggingSettings{
								Destination: &models.ObservabilityAbsoluteDeployment{
									DeploymentID: &mock.ValidClusterID,
									RefID:        "main-elasticsearch",
								},
							},
							Metrics: &models.DeploymentMetricsSettings{
								Destination: &models.ObservabilityAbsoluteDeployment{
									DeploymentID: &mock.ValidClusterID,
									RefID:        "main-elasticsearch",
								},
							},
						},
					},
					Resources: &models.DeploymentResources{
						Elasticsearch: []*models.ElasticsearchResourceInfo{
							{
								Region: ec.String("us-east-1"),
								RefID:  ec.String("main-elasticsearch"),
								Info: &models.ElasticsearchClusterInfo{
									Status:      ec.String("started"),
									ClusterID:   &mock.ValidClusterID,
									ClusterName: ec.String("some-name"),
									Region:      "us-east-1",
									ElasticsearchMonitoringInfo: &models.ElasticsearchMonitoringInfo{
										DestinationClusterIds: []string{"some"},
									},
									PlanInfo: &models.ElasticsearchClusterPlansInfo{
										Current: &models.ElasticsearchClusterPlanInfo{
											Plan: &models.ElasticsearchClusterPlan{
												Elasticsearch: &models.ElasticsearchConfiguration{
													Version:                  "7.7.0",
													UserSettingsYaml:         `some.setting: value`,
													UserSettingsOverrideYaml: `some.setting: value2`,
													UserSettingsJSON: map[string]interface{}{
														"some.setting": "value",
													},
													UserSettingsOverrideJSON: map[string]interface{}{
														"some.setting": "value2",
													},
												},
												DeploymentTemplate: &models.DeploymentTemplateReference{
													ID: ec.String("aws-io-optimized-v2"),
												},
												ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
													ID: "hot_content",
													Elasticsearch: &models.ElasticsearchConfiguration{
														NodeAttributes: map[string]string{"data": "hot"},
													},
													ZoneCount:               1,
													InstanceConfigurationID: "aws.data.highio.i3",
													Size: &models.TopologySize{
														Resource: ec.String("memory"),
														Value:    ec.Int32(2048),
													},
													NodeType: &models.ElasticsearchNodeType{
														Data:   ec.Bool(true),
														Ingest: ec.Bool(true),
														Master: ec.Bool(true),
														Ml:     ec.Bool(false),
													},
													TopologyElementControl: &models.TopologyElementControl{
														Min: &models.TopologySize{
															Resource: ec.String("memory"),
															Value:    ec.Int32(1024),
														},
													},
												}},
											},
										},
									},
								},
							},
						},
						Kibana: []*models.KibanaResourceInfo{
							{
								Region:                    ec.String("us-east-1"),
								RefID:                     ec.String("main-kibana"),
								ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
								Info: &models.KibanaClusterInfo{
									Status:      ec.String("started"),
									ClusterID:   &mock.ValidClusterID,
									ClusterName: ec.String("some-kibana-name"),
									Region:      "us-east-1",
									PlanInfo: &models.KibanaClusterPlansInfo{
										Current: &models.KibanaClusterPlanInfo{
											Plan: &models.KibanaClusterPlan{
												Kibana: &models.KibanaConfiguration{
													Version: "7.7.0",
												},
												ClusterTopology: []*models.KibanaClusterTopologyElement{
													{
														ZoneCount:               1,
														InstanceConfigurationID: "aws.kibana.r5d",
														Size: &models.TopologySize{
															Resource: ec.String("memory"),
															Value:    ec.Int32(1024),
														},
													},
												},
											},
										},
									},
								},
							},
						},
						Apm: []*models.ApmResourceInfo{{
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Info: &models.ApmInfo{
								Status: ec.String("started"),
								ID:     &mock.ValidClusterID,
								Name:   ec.String("some-apm-name"),
								Region: "us-east-1",
								PlanInfo: &models.ApmPlansInfo{
									Current: &models.ApmPlanInfo{
										Plan: &models.ApmPlan{
											Apm: &models.ApmConfiguration{
												Version: "7.7.0",
												SystemSettings: &models.ApmSystemSettings{
													DebugEnabled: ec.Bool(false),
												},
											},
											ClusterTopology: []*models.ApmTopologyElement{{
												ZoneCount:               1,
												InstanceConfigurationID: "aws.apm.r5d",
												Size: &models.TopologySize{
													Resource: ec.String("memory"),
													Value:    ec.Int32(512),
												},
											}},
										},
									},
								},
							},
						}},
						EnterpriseSearch: []*models.EnterpriseSearchResourceInfo{
							{
								Region:                    ec.String("us-east-1"),
								RefID:                     ec.String("main-enterprise_search"),
								ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
								Info: &models.EnterpriseSearchInfo{
									Status: ec.String("started"),
									ID:     &mock.ValidClusterID,
									Name:   ec.String("some-enterprise_search-name"),
									Region: "us-east-1",
									PlanInfo: &models.EnterpriseSearchPlansInfo{
										Current: &models.EnterpriseSearchPlanInfo{
											Plan: &models.EnterpriseSearchPlan{
												EnterpriseSearch: &models.EnterpriseSearchConfiguration{
													Version: "7.7.0",
												},
												ClusterTopology: []*models.EnterpriseSearchTopologyElement{
													{
														ZoneCount:               1,
														InstanceConfigurationID: "aws.enterprisesearch.m5d",
														Size: &models.TopologySize{
															Resource: ec.String("memory"),
															Value:    ec.Int32(2048),
														},
														NodeType: &models.EnterpriseSearchNodeTypes{
															Appserver: ec.Bool(true),
															Connector: ec.Bool(true),
															Worker:    ec.Bool(true),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "sets the global version to the lesser version",
			args: args{
				res: &models.DeploymentGetResponse{
					ID:    &mock.ValidClusterID,
					Alias: "my-deployment",
					Name:  ec.String("my_deployment_name"),
					Settings: &models.DeploymentSettings{
						TrafficFilterSettings: &models.TrafficFilterSettings{
							Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
						},
					},
					Resources: &models.DeploymentResources{
						Elasticsearch: []*models.ElasticsearchResourceInfo{
							{
								Region: ec.String("us-east-1"),
								RefID:  ec.String("main-elasticsearch"),
								Info: &models.ElasticsearchClusterInfo{
									Status:      ec.String("started"),
									ClusterID:   &mock.ValidClusterID,
									ClusterName: ec.String("some-name"),
									Region:      "us-east-1",
									ElasticsearchMonitoringInfo: &models.ElasticsearchMonitoringInfo{
										DestinationClusterIds: []string{"some"},
									},
									PlanInfo: &models.ElasticsearchClusterPlansInfo{
										Current: &models.ElasticsearchClusterPlanInfo{
											Plan: &models.ElasticsearchClusterPlan{
												Elasticsearch: &models.ElasticsearchConfiguration{
													Version:                  "7.7.0",
													UserSettingsYaml:         `some.setting: value`,
													UserSettingsOverrideYaml: `some.setting: value2`,
													UserSettingsJSON: map[string]interface{}{
														"some.setting": "value",
													},
													UserSettingsOverrideJSON: map[string]interface{}{
														"some.setting": "value2",
													},
												},
												DeploymentTemplate: &models.DeploymentTemplateReference{
													ID: ec.String("aws-io-optimized-v2"),
												},
												ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
													ID: "hot_content",
													Elasticsearch: &models.ElasticsearchConfiguration{
														NodeAttributes: map[string]string{"data": "hot"},
													},
													ZoneCount:               1,
													InstanceConfigurationID: "aws.data.highio.i3",
													Size: &models.TopologySize{
														Resource: ec.String("memory"),
														Value:    ec.Int32(2048),
													},
													NodeType: &models.ElasticsearchNodeType{
														Data:   ec.Bool(true),
														Ingest: ec.Bool(true),
														Master: ec.Bool(true),
														Ml:     ec.Bool(false),
													},
													TopologyElementControl: &models.TopologyElementControl{
														Min: &models.TopologySize{
															Resource: ec.String("memory"),
															Value:    ec.Int32(1024),
														},
													},
												}},
											},
										},
									},
								},
							},
						},
						Kibana: []*models.KibanaResourceInfo{
							{
								Region:                    ec.String("us-east-1"),
								RefID:                     ec.String("main-kibana"),
								ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
								Info: &models.KibanaClusterInfo{
									Status:      ec.String("started"),
									ClusterID:   &mock.ValidClusterID,
									ClusterName: ec.String("some-kibana-name"),
									Region:      "us-east-1",
									PlanInfo: &models.KibanaClusterPlansInfo{
										Current: &models.KibanaClusterPlanInfo{
											Plan: &models.KibanaClusterPlan{
												Kibana: &models.KibanaConfiguration{
													Version: "7.6.2",
												},
												ClusterTopology: []*models.KibanaClusterTopologyElement{
													{
														ZoneCount:               1,
														InstanceConfigurationID: "aws.kibana.r5d",
														Size: &models.TopologySize{
															Resource: ec.String("memory"),
															Value:    ec.Int32(1024),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: Deployment{
				Id:                   mock.ValidClusterID,
				Alias:                "my-deployment",
				Name:                 "my_deployment_name",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "us-east-1",
				Version:              "7.6.2",
				Elasticsearch: Elasticsearches{
					{
						RefId:      ec.String("main-elasticsearch"),
						ResourceId: &mock.ValidClusterID,
						Region:     ec.String("us-east-1"),
						Config: ElasticsearchConfigs{
							{
								UserSettingsYaml:         ec.String("some.setting: value"),
								UserSettingsOverrideYaml: ec.String("some.setting: value2"),
								UserSettingsJson:         ec.String("{\"some.setting\":\"value\"}"),
								UserSettingsOverrideJson: ec.String("{\"some.setting\":\"value2\"}"),
							},
						},
						Topology: ElasticsearchTopologies{
							{
								Id:                      "hot_content",
								InstanceConfigurationId: ec.String("aws.data.highio.i3"),
								Size:                    ec.String("2g"),
								SizeResource:            ec.String("memory"),
								NodeTypeData:            ec.String("true"),
								NodeTypeIngest:          ec.String("true"),
								NodeTypeMaster:          ec.String("true"),
								NodeTypeMl:              ec.String("false"),
								ZoneCount:               1,
							},
						},
					},
				},
				Kibana: Kibanas{
					{
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						RefId:                     ec.String("main-kibana"),
						ResourceId:                ec.String(mock.ValidClusterID),
						Region:                    ec.String("us-east-1"),
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("aws.kibana.r5d"),
								Size:                    ec.String("1g"),
								SizeResource:            ec.String("memory"),
								ZoneCount:               1,
							},
						},
					},
				},
				TrafficFilter: []string{"0.0.0.0/0", "192.168.10.0/24"},
			},
		},
		{
			name: "flattens an azure plan (io-optimized)",
			args: args{
				res: testutil.OpenDeploymentGet(t, "testdata/deployment-azure-io-optimized.json"),
			},
			want: Deployment{
				Id:                   "123e79d8109c4a0790b0b333110bf715",
				Alias:                "my-deployment",
				Name:                 "up2d",
				DeploymentTemplateId: "azure-io-optimized",
				Region:               "azure-eastus2",
				Version:              "7.9.2",
				Elasticsearch: Elasticsearches{
					{
						RefId:         ec.String("main-elasticsearch"),
						ResourceId:    ec.String("1238f19957874af69306787dca662154"),
						Region:        ec.String("azure-eastus2"),
						Autoscale:     ec.String("false"),
						CloudID:       ec.String("up2d:somecloudID"),
						HttpEndpoint:  ec.String("http://1238f19957874af69306787dca662154.eastus2.azure.elastic-cloud.com:9200"),
						HttpsEndpoint: ec.String("https://1238f19957874af69306787dca662154.eastus2.azure.elastic-cloud.com:9243"),
						Topology: ElasticsearchTopologies{
							{
								Id:                      "hot_content",
								InstanceConfigurationId: ec.String("azure.data.highio.l32sv2"),
								Size:                    ec.String("4g"),
								SizeResource:            ec.String("memory"),
								NodeTypeData:            ec.String("true"),
								NodeTypeIngest:          ec.String("true"),
								NodeTypeMaster:          ec.String("true"),
								NodeTypeMl:              ec.String("false"),
								ZoneCount:               2,
							},
						},
					},
				},
				Kibana: Kibanas{
					{
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						RefId:                     ec.String("main-kibana"),
						ResourceId:                ec.String("1235cd4a4c7f464bbcfd795f3638b769"),
						Region:                    ec.String("azure-eastus2"),
						HttpEndpoint:              ec.String("http://1235cd4a4c7f464bbcfd795f3638b769.eastus2.azure.elastic-cloud.com:9200"),
						HttpsEndpoint:             ec.String("https://1235cd4a4c7f464bbcfd795f3638b769.eastus2.azure.elastic-cloud.com:9243"),
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("azure.kibana.e32sv3"),
								Size:                    ec.String("1g"),
								SizeResource:            ec.String("memory"),
								ZoneCount:               1,
							},
						},
					},
				},
				Apm: Apms{
					{
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						RefId:                     ec.String("main-apm"),
						ResourceId:                ec.String("1235d8c911b74dd6a03c2a7b37fd68ab"),
						Region:                    ec.String("azure-eastus2"),
						HttpEndpoint:              ec.String("http://1235d8c911b74dd6a03c2a7b37fd68ab.apm.eastus2.azure.elastic-cloud.com:9200"),
						HttpsEndpoint:             ec.String("https://1235d8c911b74dd6a03c2a7b37fd68ab.apm.eastus2.azure.elastic-cloud.com:443"),
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("azure.apm.e32sv3"),
								Size:                    ec.String("0.5g"),
								SizeResource:            ec.String("memory"),
								ZoneCount:               1,
							},
						},
					},
				},
			},
		},
		// {
		// 	name: "flattens an aws plan (io-optimized)",
		// 	args: args{d: awsIOOptimizedRD, res: awsIOOptimizedRes},
		// 	want: wantAwsIOOptimizedDeployment,
		// },
		// {
		// 	name: "flattens an aws plan with extensions (io-optimized)",
		// 	args: args{
		// 		d:   awsIOOptimizedExtensionRD,
		// 		res: openDeploymentGet(t, "testdata/deployment-aws-io-optimized-extension.json"),
		// 	},
		// 	want: util.NewResourceData(t, util.ResDataParams{
		// 		ID: mock.ValidClusterID,
		// 		State: map[string]interface{}{
		// 			"alias":                  "my-deployment",
		// 			"deployment_template_id": "aws-io-optimized-v2",
		// 			"id":                     "123b7b540dfc967a7a649c18e2fce4ed",
		// 			"name":                   "up2d",
		// 			"region":                 "aws-eu-central-1",
		// 			"version":                "7.9.2",
		// 			"apm": []interface{}{map[string]interface{}{
		// 				"elasticsearch_cluster_ref_id": "main-elasticsearch",
		// 				"ref_id":                       "main-apm",
		// 				"region":                       "aws-eu-central-1",
		// 				"resource_id":                  "12328579b3bf40c8b58c1a0ed5a4bd8b",
		// 				"version":                      "7.9.2",
		// 				"http_endpoint":                "http://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:80",
		// 				"https_endpoint":               "https://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:443",
		// 				"topology": []interface{}{map[string]interface{}{
		// 					"instance_configuration_id": "aws.apm.r5d",
		// 					"size":                      "0.5g",
		// 					"size_resource":             "memory",
		// 					"zone_count":                1,
		// 				}},
		// 			}},
		// 			"elasticsearch": []interface{}{map[string]interface{}{
		// 				"autoscale": "false",
		// 				"cloud_id":  "up2d:someCloudID",
		// 				"extension": []interface{}{
		// 					map[string]interface{}{
		// 						"name":    "custom-bundle",
		// 						"version": "7.9.2",
		// 						"url":     "http://12345",
		// 						"type":    "bundle",
		// 					},
		// 					map[string]interface{}{
		// 						"name":    "custom-bundle2",
		// 						"version": "7.9.2",
		// 						"url":     "http://123456",
		// 						"type":    "bundle",
		// 					},
		// 					map[string]interface{}{
		// 						"name":    "custom-plugin",
		// 						"version": "7.9.2",
		// 						"url":     "http://12345",
		// 						"type":    "plugin",
		// 					},
		// 					map[string]interface{}{
		// 						"name":    "custom-plugin2",
		// 						"version": "7.9.2",
		// 						"url":     "http://123456",
		// 						"type":    "plugin",
		// 					},
		// 				},
		// 				"http_endpoint":  "http://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9200",
		// 				"https_endpoint": "https://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9243",
		// 				"ref_id":         "main-elasticsearch",
		// 				"region":         "aws-eu-central-1",
		// 				"resource_id":    "1239f7ee7196439ba2d105319ac5eba7",
		// 				"topology": []interface{}{map[string]interface{}{
		// 					"id":                        "hot_content",
		// 					"instance_configuration_id": "aws.data.highio.i3",
		// 					"node_type_data":            "true",
		// 					"node_type_ingest":          "true",
		// 					"node_type_master":          "true",
		// 					"node_type_ml":              "false",
		// 					"size":                      "8g",
		// 					"size_resource":             "memory",
		// 					"zone_count":                2,
		// 				}},
		// 			}},
		// 			"kibana": []interface{}{map[string]interface{}{
		// 				"elasticsearch_cluster_ref_id": "main-elasticsearch",
		// 				"ref_id":                       "main-kibana",
		// 				"region":                       "aws-eu-central-1",
		// 				"resource_id":                  "123dcfda06254ca789eb287e8b73ff4c",
		// 				"version":                      "7.9.2",
		// 				"http_endpoint":                "http://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9200",
		// 				"https_endpoint":               "https://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9243",
		// 				"topology": []interface{}{map[string]interface{}{
		// 					"instance_configuration_id": "aws.kibana.r5d",
		// 					"size":                      "1g",
		// 					"size_resource":             "memory",
		// 					"zone_count":                1,
		// 				}},
		// 			}},
		// 		},
		// 		Schema: newSchema(),
		// 	}),
		// },
		// {
		// 	name: "flattens an aws plan with trusts",
		// 	args: args{
		// 		d: newDeploymentRD(t, "123b7b540dfc967a7a649c18e2fce4ed", nil),
		// 		res: &models.DeploymentGetResponse{
		// 			ID:    ec.String("123b7b540dfc967a7a649c18e2fce4ed"),
		// 			Alias: "OH",
		// 			Name:  ec.String("up2d"),
		// 			Resources: &models.DeploymentResources{
		// 				Elasticsearch: []*models.ElasticsearchResourceInfo{{
		// 					RefID:  ec.String("main-elasticsearch"),
		// 					Region: ec.String("aws-eu-central-1"),
		// 					Info: &models.ElasticsearchClusterInfo{
		// 						Status: ec.String("running"),
		// 						PlanInfo: &models.ElasticsearchClusterPlansInfo{
		// 							Current: &models.ElasticsearchClusterPlanInfo{
		// 								Plan: &models.ElasticsearchClusterPlan{
		// 									DeploymentTemplate: &models.DeploymentTemplateReference{
		// 										ID: ec.String("aws-io-optimized-v2"),
		// 									},
		// 									Elasticsearch: &models.ElasticsearchConfiguration{
		// 										Version: "7.13.1",
		// 									},
		// 									ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
		// 										ID: "hot_content",
		// 										Size: &models.TopologySize{
		// 											Value:    ec.Int32(4096),
		// 											Resource: ec.String("memory"),
		// 										},
		// 									}},
		// 								},
		// 							},
		// 						},
		// 						Settings: &models.ElasticsearchClusterSettings{
		// 							Trust: &models.ElasticsearchClusterTrustSettings{
		// 								Accounts: []*models.AccountTrustRelationship{
		// 									{
		// 										AccountID: ec.String("ANID"),
		// 										TrustAll:  ec.Bool(true),
		// 									},
		// 									{
		// 										AccountID: ec.String("anotherID"),
		// 										TrustAll:  ec.Bool(false),
		// 										TrustAllowlist: []string{
		// 											"abc", "dfg", "hij",
		// 										},
		// 									},
		// 								},
		// 								External: []*models.ExternalTrustRelationship{
		// 									{
		// 										TrustRelationshipID: ec.String("external_id"),
		// 										TrustAll:            ec.Bool(true),
		// 									},
		// 									{
		// 										TrustRelationshipID: ec.String("another_external_id"),
		// 										TrustAll:            ec.Bool(false),
		// 										TrustAllowlist: []string{
		// 											"abc", "dfg",
		// 										},
		// 									},
		// 								},
		// 							},
		// 						},
		// 					},
		// 				}},
		// 			},
		// 		},
		// 	},
		// 	want: util.NewResourceData(t, util.ResDataParams{
		// 		ID: "123b7b540dfc967a7a649c18e2fce4ed",
		// 		State: map[string]interface{}{
		// 			"alias":                  "OH",
		// 			"deployment_template_id": "aws-io-optimized-v2",
		// 			"id":                     "123b7b540dfc967a7a649c18e2fce4ed",
		// 			"name":                   "up2d",
		// 			"region":                 "aws-eu-central-1",
		// 			"version":                "7.13.1",
		// 			"elasticsearch": []interface{}{map[string]interface{}{
		// 				"region": "aws-eu-central-1",
		// 				"ref_id": "main-elasticsearch",
		// 				"topology": []interface{}{map[string]interface{}{
		// 					"id":            "hot_content",
		// 					"size":          "4g",
		// 					"size_resource": "memory",
		// 				}},
		// 				"trust_account": []interface{}{
		// 					map[string]interface{}{
		// 						"account_id": "ANID",
		// 						"trust_all":  "true",
		// 					},
		// 					map[string]interface{}{
		// 						"account_id": "anotherID",
		// 						"trust_all":  "false",
		// 						"trust_allowlist": []interface{}{
		// 							"abc", "hij", "dfg",
		// 						},
		// 					},
		// 				},
		// 				"trust_external": []interface{}{
		// 					map[string]interface{}{
		// 						"relationship_id": "another_external_id",
		// 						"trust_all":       "false",
		// 						"trust_allowlist": []interface{}{
		// 							"abc", "dfg",
		// 						},
		// 					},
		// 					map[string]interface{}{
		// 						"relationship_id": "external_id",
		// 						"trust_all":       "true",
		// 					},
		// 				},
		// 			}},
		// 		},
		// 		Schema: newSchema(),
		// 	}),
		// },
		// {
		// 	name: "flattens an aws plan with topology.config set",
		// 	args: args{
		// 		d: newDeploymentRD(t, "123b7b540dfc967a7a649c18e2fce4ed", nil),
		// 		res: &models.DeploymentGetResponse{
		// 			ID:    ec.String("123b7b540dfc967a7a649c18e2fce4ed"),
		// 			Alias: "OH",
		// 			Name:  ec.String("up2d"),
		// 			Resources: &models.DeploymentResources{
		// 				Elasticsearch: []*models.ElasticsearchResourceInfo{{
		// 					RefID:  ec.String("main-elasticsearch"),
		// 					Region: ec.String("aws-eu-central-1"),
		// 					Info: &models.ElasticsearchClusterInfo{
		// 						Status: ec.String("running"),
		// 						PlanInfo: &models.ElasticsearchClusterPlansInfo{
		// 							Current: &models.ElasticsearchClusterPlanInfo{
		// 								Plan: &models.ElasticsearchClusterPlan{
		// 									DeploymentTemplate: &models.DeploymentTemplateReference{
		// 										ID: ec.String("aws-io-optimized-v2"),
		// 									},
		// 									Elasticsearch: &models.ElasticsearchConfiguration{
		// 										Version: "7.13.1",
		// 									},
		// 									ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
		// 										ID: "hot_content",
		// 										Size: &models.TopologySize{
		// 											Value:    ec.Int32(4096),
		// 											Resource: ec.String("memory"),
		// 										},
		// 										Elasticsearch: &models.ElasticsearchConfiguration{
		// 											UserSettingsYaml: "a.setting: true",
		// 										},
		// 									}},
		// 								},
		// 							},
		// 						},
		// 						Settings: &models.ElasticsearchClusterSettings{},
		// 					},
		// 				}},
		// 			},
		// 		},
		// 	},
		// 	want: util.NewResourceData(t, util.ResDataParams{
		// 		ID: "123b7b540dfc967a7a649c18e2fce4ed",
		// 		State: map[string]interface{}{
		// 			"alias":                  "OH",
		// 			"deployment_template_id": "aws-io-optimized-v2",
		// 			"id":                     "123b7b540dfc967a7a649c18e2fce4ed",
		// 			"name":                   "up2d",
		// 			"region":                 "aws-eu-central-1",
		// 			"version":                "7.13.1",
		// 			"elasticsearch": []interface{}{map[string]interface{}{
		// 				"region": "aws-eu-central-1",
		// 				"ref_id": "main-elasticsearch",
		// 				"topology": []interface{}{map[string]interface{}{
		// 					"id":            "hot_content",
		// 					"size":          "4g",
		// 					"size_resource": "memory",
		// 					"config": []interface{}{map[string]interface{}{
		// 						"user_settings_yaml": "a.setting: true",
		// 					}},
		// 				}},
		// 			}},
		// 		},
		// 		Schema: newSchema(),
		// 	}),
		// },
		// {
		// 	name: "flattens an plan with config.docker_image set",
		// 	args: args{
		// 		d: newDeploymentRD(t, "123b7b540dfc967a7a649c18e2fce4ed", nil),
		// 		res: &models.DeploymentGetResponse{
		// 			ID:    ec.String("123b7b540dfc967a7a649c18e2fce4ed"),
		// 			Alias: "OH",
		// 			Name:  ec.String("up2d"),
		// 			Resources: &models.DeploymentResources{
		// 				Elasticsearch: []*models.ElasticsearchResourceInfo{{
		// 					RefID:  ec.String("main-elasticsearch"),
		// 					Region: ec.String("aws-eu-central-1"),
		// 					Info: &models.ElasticsearchClusterInfo{
		// 						Status: ec.String("running"),
		// 						PlanInfo: &models.ElasticsearchClusterPlansInfo{
		// 							Current: &models.ElasticsearchClusterPlanInfo{
		// 								Plan: &models.ElasticsearchClusterPlan{
		// 									DeploymentTemplate: &models.DeploymentTemplateReference{
		// 										ID: ec.String("aws-io-optimized-v2"),
		// 									},
		// 									Elasticsearch: &models.ElasticsearchConfiguration{
		// 										Version:     "7.14.1",
		// 										DockerImage: "docker.elastic.com/elasticsearch/cloud:7.14.1-hash",
		// 									},
		// 									ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
		// 										ID: "hot_content",
		// 										Size: &models.TopologySize{
		// 											Value:    ec.Int32(4096),
		// 											Resource: ec.String("memory"),
		// 										},
		// 										Elasticsearch: &models.ElasticsearchConfiguration{
		// 											UserSettingsYaml: "a.setting: true",
		// 										},
		// 										ZoneCount: 1,
		// 									}},
		// 								},
		// 							},
		// 						},
		// 						Settings: &models.ElasticsearchClusterSettings{},
		// 					},
		// 				}},
		// 				Apm: []*models.ApmResourceInfo{{
		// 					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
		// 					RefID:                     ec.String("main-apm"),
		// 					Region:                    ec.String("aws-eu-central-1"),
		// 					Info: &models.ApmInfo{
		// 						Status: ec.String("running"),
		// 						PlanInfo: &models.ApmPlansInfo{Current: &models.ApmPlanInfo{
		// 							Plan: &models.ApmPlan{
		// 								Apm: &models.ApmConfiguration{
		// 									Version:     "7.14.1",
		// 									DockerImage: "docker.elastic.com/apm/cloud:7.14.1-hash",
		// 									SystemSettings: &models.ApmSystemSettings{
		// 										DebugEnabled: ec.Bool(false),
		// 									},
		// 								},
		// 								ClusterTopology: []*models.ApmTopologyElement{{
		// 									InstanceConfigurationID: "aws.apm.r5d",
		// 									Size: &models.TopologySize{
		// 										Resource: ec.String("memory"),
		// 										Value:    ec.Int32(512),
		// 									},
		// 									ZoneCount: 1,
		// 								}},
		// 							},
		// 						}},
		// 					},
		// 				}},
		// 				Kibana: []*models.KibanaResourceInfo{{
		// 					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
		// 					RefID:                     ec.String("main-kibana"),
		// 					Region:                    ec.String("aws-eu-central-1"),
		// 					Info: &models.KibanaClusterInfo{
		// 						Status: ec.String("running"),
		// 						PlanInfo: &models.KibanaClusterPlansInfo{Current: &models.KibanaClusterPlanInfo{
		// 							Plan: &models.KibanaClusterPlan{
		// 								Kibana: &models.KibanaConfiguration{
		// 									Version:     "7.14.1",
		// 									DockerImage: "docker.elastic.com/kibana/cloud:7.14.1-hash",
		// 								},
		// 								ClusterTopology: []*models.KibanaClusterTopologyElement{{
		// 									InstanceConfigurationID: "aws.kibana.r5d",
		// 									Size: &models.TopologySize{
		// 										Resource: ec.String("memory"),
		// 										Value:    ec.Int32(1024),
		// 									},
		// 									ZoneCount: 1,
		// 								}},
		// 							},
		// 						}},
		// 					},
		// 				}},
		// 				EnterpriseSearch: []*models.EnterpriseSearchResourceInfo{{
		// 					ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
		// 					RefID:                     ec.String("main-enterprise_search"),
		// 					Region:                    ec.String("aws-eu-central-1"),
		// 					Info: &models.EnterpriseSearchInfo{
		// 						Status: ec.String("running"),
		// 						PlanInfo: &models.EnterpriseSearchPlansInfo{Current: &models.EnterpriseSearchPlanInfo{
		// 							Plan: &models.EnterpriseSearchPlan{
		// 								EnterpriseSearch: &models.EnterpriseSearchConfiguration{
		// 									Version:     "7.14.1",
		// 									DockerImage: "docker.elastic.com/enterprise_search/cloud:7.14.1-hash",
		// 								},
		// 								ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
		// 									InstanceConfigurationID: "aws.enterprisesearch.m5d",
		// 									Size: &models.TopologySize{
		// 										Resource: ec.String("memory"),
		// 										Value:    ec.Int32(2048),
		// 									},
		// 									NodeType: &models.EnterpriseSearchNodeTypes{
		// 										Appserver: ec.Bool(true),
		// 										Connector: ec.Bool(true),
		// 										Worker:    ec.Bool(true),
		// 									},
		// 									ZoneCount: 2,
		// 								}},
		// 							},
		// 						}},
		// 					},
		// 				}},
		// 			},
		// 		},
		// 	},
		// 	want: util.NewResourceData(t, util.ResDataParams{
		// 		ID: "123b7b540dfc967a7a649c18e2fce4ed",
		// 		State: map[string]interface{}{
		// 			"alias":                  "OH",
		// 			"deployment_template_id": "aws-io-optimized-v2",
		// 			"id":                     "123b7b540dfc967a7a649c18e2fce4ed",
		// 			"name":                   "up2d",
		// 			"region":                 "aws-eu-central-1",
		// 			"version":                "7.14.1",
		// 			"elasticsearch": []interface{}{map[string]interface{}{
		// 				"region": "aws-eu-central-1",
		// 				"ref_id": "main-elasticsearch",
		// 				"config": []interface{}{map[string]interface{}{
		// 					"docker_image": "docker.elastic.com/elasticsearch/cloud:7.14.1-hash",
		// 				}},
		// 				"topology": []interface{}{map[string]interface{}{
		// 					"id":            "hot_content",
		// 					"size":          "4g",
		// 					"size_resource": "memory",
		// 					"zone_count":    1,
		// 					"config": []interface{}{map[string]interface{}{
		// 						"user_settings_yaml": "a.setting: true",
		// 					}},
		// 				}},
		// 			}},
		// 			"kibana": []interface{}{map[string]interface{}{
		// 				"region": "aws-eu-central-1",
		// 				"ref_id": "main-kibana",
		// 				"config": []interface{}{map[string]interface{}{
		// 					"docker_image": "docker.elastic.com/kibana/cloud:7.14.1-hash",
		// 				}},
		// 				"topology": []interface{}{map[string]interface{}{
		// 					"instance_configuration_id": "aws.kibana.r5d",
		// 					"size":                      "1g",
		// 					"size_resource":             "memory",
		// 					"zone_count":                1,
		// 				}},
		// 			}},
		// 			"apm": []interface{}{map[string]interface{}{
		// 				"region": "aws-eu-central-1",
		// 				"ref_id": "main-apm",
		// 				"config": []interface{}{map[string]interface{}{
		// 					"docker_image": "docker.elastic.com/apm/cloud:7.14.1-hash",
		// 				}},
		// 				"topology": []interface{}{map[string]interface{}{
		// 					"instance_configuration_id": "aws.apm.r5d",
		// 					"size":                      "0.5g",
		// 					"size_resource":             "memory",
		// 					"zone_count":                1,
		// 				}},
		// 			}},
		// 			"enterprise_search": []interface{}{map[string]interface{}{
		// 				"region": "aws-eu-central-1",
		// 				"ref_id": "main-enterprise_search",
		// 				"config": []interface{}{map[string]interface{}{
		// 					"docker_image": "docker.elastic.com/enterprise_search/cloud:7.14.1-hash",
		// 				}},
		// 				"topology": []interface{}{map[string]interface{}{
		// 					"instance_configuration_id": "aws.enterprisesearch.m5d",
		// 					"size":                      "2g",
		// 					"size_resource":             "memory",
		// 					"zone_count":                2,
		// 					"node_type_appserver":       "true",
		// 					"node_type_connector":       "true",
		// 					"node_type_worker":          "true",
		// 				}},
		// 			}},
		// 		},
		// 		Schema: newSchema(),
		// 	}),
		// },
		// {
		// 	name: "flattens an aws plan (io-optimized) with tags",
		// 	args: args{d: awsIOOptimizedTagsRD, res: awsIOOptimizedTagsRes},
		// 	want: wantAwsIOOptimizedDeploymentTags,
		// },
		// {
		// 	name: "flattens a gcp plan (io-optimized)",
		// 	args: args{d: gcpIOOptimizedRD, res: gcpIOOptimizedRes},
		// 	want: wantGcpIOOptimizedDeployment,
		// },
		// {
		// 	name: "flattens a gcp plan with autoscale set (io-optimized)",
		// 	args: args{d: gcpIOOptimizedRD, res: gcpIOOptimizedAutoscaleRes},
		// 	want: wantGcpIOOptAutoscale,
		// },
		// {
		// 	name: "flattens a gcp plan (hot-warm)",
		// 	args: args{d: gcpHotWarmRD, res: gcpHotWarmRes},
		// 	want: wantGcpHotWarmDeployment,
		// },
		// {
		// 	name: "flattens a gcp plan (hot-warm) with node_roles",
		// 	args: args{d: gcpHotWarmNodeRolesRD, res: gcpHotWarmNodeRolesRes},
		// 	want: wantGcpHotWarmNodeRolesDeployment,
		// },
		// {
		// 	name: "flattens an aws plan (Cross Cluster Search)",
		// 	args: args{d: awsCCSRD, res: awsCCSRes, remotes: argCCSRemotes},
		// 	want: wantAWSCCSDeployment,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep, err := readDeployment(tt.args.res, &tt.args.remotes, nil)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, dep)
				assert.Equal(t, tt.want, *dep)
			}
		})
	}
}
