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
		{
			name: "flattens an aws plan (io-optimized)",
			args: args{res: testutil.OpenDeploymentGet(t, "testdata/deployment-aws-io-optimized.json")},
			want: Deployment{
				Id:                   "123365f2805e46808d40849b1c0b266b",
				Alias:                "my-deployment",
				Name:                 "up2d",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "aws-eu-central-1",
				Version:              "7.9.2",
				Elasticsearch: Elasticsearches{
					{
						RefId:         ec.String("main-elasticsearch"),
						ResourceId:    ec.String("1239f7ee7196439ba2d105319ac5eba7"),
						Region:        ec.String("aws-eu-central-1"),
						Autoscale:     ec.String("false"),
						CloudID:       ec.String("up2d:someCloudID"),
						HttpEndpoint:  ec.String("http://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9200"),
						HttpsEndpoint: ec.String("https://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9243"),
						Topology: ElasticsearchTopologies{
							{
								Id:                      "hot_content",
								InstanceConfigurationId: ec.String("aws.data.highio.i3"),
								Size:                    ec.String("8g"),
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
						ResourceId:                ec.String("123dcfda06254ca789eb287e8b73ff4c"),
						Region:                    ec.String("aws-eu-central-1"),
						HttpEndpoint:              ec.String("http://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9200"),
						HttpsEndpoint:             ec.String("https://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9243"),
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
						ResourceId:                ec.String("12328579b3bf40c8b58c1a0ed5a4bd8b"),
						Region:                    ec.String("aws-eu-central-1"),
						HttpEndpoint:              ec.String("http://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:80"),
						HttpsEndpoint:             ec.String("https://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:443"),
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
			},
		},
		{
			name: "flattens an aws plan with extensions (io-optimized)",
			args: args{
				res: testutil.OpenDeploymentGet(t, "testdata/deployment-aws-io-optimized-extension.json"),
			},
			want: Deployment{
				Id:                   "123365f2805e46808d40849b1c0b266b",
				Alias:                "my-deployment",
				Name:                 "up2d",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "aws-eu-central-1",
				Version:              "7.9.2",
				Elasticsearch: Elasticsearches{
					{
						RefId:         ec.String("main-elasticsearch"),
						ResourceId:    ec.String("1239f7ee7196439ba2d105319ac5eba7"),
						Region:        ec.String("aws-eu-central-1"),
						Autoscale:     ec.String("false"),
						CloudID:       ec.String("up2d:someCloudID"),
						HttpEndpoint:  ec.String("http://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9200"),
						HttpsEndpoint: ec.String("https://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9243"),
						Topology: ElasticsearchTopologies{
							{
								Id:                      "hot_content",
								InstanceConfigurationId: ec.String("aws.data.highio.i3"),
								Size:                    ec.String("8g"),
								SizeResource:            ec.String("memory"),
								NodeTypeData:            ec.String("true"),
								NodeTypeIngest:          ec.String("true"),
								NodeTypeMaster:          ec.String("true"),
								NodeTypeMl:              ec.String("false"),
								ZoneCount:               2,
							},
						},
						Extension: ElasticsearchExtensions{
							{
								Name:    "custom-bundle",
								Version: "7.9.2",
								Url:     "http://12345",
								Type:    "bundle",
							},
							{
								Name:    "custom-bundle2",
								Version: "7.9.2",
								Url:     "http://123456",
								Type:    "bundle",
							},
							{
								Name:    "custom-plugin",
								Version: "7.9.2",
								Url:     "http://12345",
								Type:    "plugin",
							},
							{
								Name:    "custom-plugin2",
								Version: "7.9.2",
								Url:     "http://123456",
								Type:    "plugin",
							},
						},
					},
				},
				Kibana: Kibanas{
					{
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						RefId:                     ec.String("main-kibana"),
						ResourceId:                ec.String("123dcfda06254ca789eb287e8b73ff4c"),
						Region:                    ec.String("aws-eu-central-1"),
						HttpEndpoint:              ec.String("http://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9200"),
						HttpsEndpoint:             ec.String("https://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9243"),
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
						ResourceId:                ec.String("12328579b3bf40c8b58c1a0ed5a4bd8b"),
						Region:                    ec.String("aws-eu-central-1"),
						HttpEndpoint:              ec.String("http://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:80"),
						HttpsEndpoint:             ec.String("https://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:443"),
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
			},
		},
		{
			name: "flattens an aws plan with trusts",
			args: args{
				res: &models.DeploymentGetResponse{
					ID:    ec.String("123b7b540dfc967a7a649c18e2fce4ed"),
					Alias: "OH",
					Name:  ec.String("up2d"),
					Resources: &models.DeploymentResources{
						Elasticsearch: []*models.ElasticsearchResourceInfo{{
							RefID:  ec.String("main-elasticsearch"),
							Region: ec.String("aws-eu-central-1"),
							Info: &models.ElasticsearchClusterInfo{
								Status: ec.String("running"),
								PlanInfo: &models.ElasticsearchClusterPlansInfo{
									Current: &models.ElasticsearchClusterPlanInfo{
										Plan: &models.ElasticsearchClusterPlan{
											DeploymentTemplate: &models.DeploymentTemplateReference{
												ID: ec.String("aws-io-optimized-v2"),
											},
											Elasticsearch: &models.ElasticsearchConfiguration{
												Version: "7.13.1",
											},
											ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
												ID: "hot_content",
												Size: &models.TopologySize{
													Value:    ec.Int32(4096),
													Resource: ec.String("memory"),
												},
											}},
										},
									},
								},
								Settings: &models.ElasticsearchClusterSettings{
									Trust: &models.ElasticsearchClusterTrustSettings{
										Accounts: []*models.AccountTrustRelationship{
											{
												AccountID: ec.String("ANID"),
												TrustAll:  ec.Bool(true),
											},
											{
												AccountID: ec.String("anotherID"),
												TrustAll:  ec.Bool(false),
												TrustAllowlist: []string{
													"abc", "dfg", "hij",
												},
											},
										},
										External: []*models.ExternalTrustRelationship{
											{
												TrustRelationshipID: ec.String("external_id"),
												TrustAll:            ec.Bool(true),
											},
											{
												TrustRelationshipID: ec.String("another_external_id"),
												TrustAll:            ec.Bool(false),
												TrustAllowlist: []string{
													"abc", "dfg",
												},
											},
										},
									},
								},
							},
						}},
					},
				},
			},
			want: Deployment{
				Id:                   "123b7b540dfc967a7a649c18e2fce4ed",
				Alias:                "OH",
				Name:                 "up2d",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "aws-eu-central-1",
				Version:              "7.13.1",
				Elasticsearch: Elasticsearches{
					{
						RefId:  ec.String("main-elasticsearch"),
						Region: ec.String("aws-eu-central-1"),
						Topology: ElasticsearchTopologies{
							{
								Id:           "hot_content",
								Size:         ec.String("4g"),
								SizeResource: ec.String("memory"),
							},
						},
						TrustAccount: ElasticsearchTrustAccounts{
							{
								AccountId: ec.String("ANID"),
								TrustAll:  ec.Bool(true),
							},
							{
								AccountId:      ec.String("anotherID"),
								TrustAll:       ec.Bool(false),
								TrustAllowlist: []string{"abc", "dfg", "hij"},
							},
						},
						TrustExternal: ElasticsearchTrustExternals{
							{
								RelationshipId: ec.String("external_id"),
								TrustAll:       ec.Bool(true),
							},
							{
								RelationshipId: ec.String("another_external_id"),
								TrustAll:       ec.Bool(false),
								TrustAllowlist: []string{"abc", "dfg"},
							},
						},
					},
				},
			},
		},
		{
			name: "flattens an aws plan with topology.config set",
			args: args{
				res: &models.DeploymentGetResponse{
					ID:    ec.String("123b7b540dfc967a7a649c18e2fce4ed"),
					Alias: "OH",
					Name:  ec.String("up2d"),
					Resources: &models.DeploymentResources{
						Elasticsearch: []*models.ElasticsearchResourceInfo{{
							RefID:  ec.String("main-elasticsearch"),
							Region: ec.String("aws-eu-central-1"),
							Info: &models.ElasticsearchClusterInfo{
								Status: ec.String("running"),
								PlanInfo: &models.ElasticsearchClusterPlansInfo{
									Current: &models.ElasticsearchClusterPlanInfo{
										Plan: &models.ElasticsearchClusterPlan{
											DeploymentTemplate: &models.DeploymentTemplateReference{
												ID: ec.String("aws-io-optimized-v2"),
											},
											Elasticsearch: &models.ElasticsearchConfiguration{
												Version: "7.13.1",
											},
											ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
												ID: "hot_content",
												Size: &models.TopologySize{
													Value:    ec.Int32(4096),
													Resource: ec.String("memory"),
												},
												Elasticsearch: &models.ElasticsearchConfiguration{
													UserSettingsYaml: "a.setting: true",
												},
											}},
										},
									},
								},
								Settings: &models.ElasticsearchClusterSettings{},
							},
						}},
					},
				},
			},
			want: Deployment{
				Id:                   "123b7b540dfc967a7a649c18e2fce4ed",
				Alias:                "OH",
				Name:                 "up2d",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "aws-eu-central-1",
				Version:              "7.13.1",
				Elasticsearch: Elasticsearches{
					{
						RefId:  ec.String("main-elasticsearch"),
						Region: ec.String("aws-eu-central-1"),
						Topology: ElasticsearchTopologies{
							{
								Id:           "hot_content",
								Size:         ec.String("4g"),
								SizeResource: ec.String("memory"),
								Config: ElasticsearchTopologyConfigs{
									{
										UserSettingsYaml: ec.String("a.setting: true"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "flattens an plan with config.docker_image set",
			args: args{
				res: &models.DeploymentGetResponse{
					ID:    ec.String("123b7b540dfc967a7a649c18e2fce4ed"),
					Alias: "OH",
					Name:  ec.String("up2d"),
					Resources: &models.DeploymentResources{
						Elasticsearch: []*models.ElasticsearchResourceInfo{{
							RefID:  ec.String("main-elasticsearch"),
							Region: ec.String("aws-eu-central-1"),
							Info: &models.ElasticsearchClusterInfo{
								Status: ec.String("running"),
								PlanInfo: &models.ElasticsearchClusterPlansInfo{
									Current: &models.ElasticsearchClusterPlanInfo{
										Plan: &models.ElasticsearchClusterPlan{
											DeploymentTemplate: &models.DeploymentTemplateReference{
												ID: ec.String("aws-io-optimized-v2"),
											},
											Elasticsearch: &models.ElasticsearchConfiguration{
												Version:     "7.14.1",
												DockerImage: "docker.elastic.com/elasticsearch/cloud:7.14.1-hash",
											},
											ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
												ID: "hot_content",
												Size: &models.TopologySize{
													Value:    ec.Int32(4096),
													Resource: ec.String("memory"),
												},
												Elasticsearch: &models.ElasticsearchConfiguration{
													UserSettingsYaml: "a.setting: true",
												},
												ZoneCount: 1,
											}},
										},
									},
								},
								Settings: &models.ElasticsearchClusterSettings{},
							},
						}},
						Apm: []*models.ApmResourceInfo{{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							RefID:                     ec.String("main-apm"),
							Region:                    ec.String("aws-eu-central-1"),
							Info: &models.ApmInfo{
								Status: ec.String("running"),
								PlanInfo: &models.ApmPlansInfo{Current: &models.ApmPlanInfo{
									Plan: &models.ApmPlan{
										Apm: &models.ApmConfiguration{
											Version:     "7.14.1",
											DockerImage: "docker.elastic.com/apm/cloud:7.14.1-hash",
											SystemSettings: &models.ApmSystemSettings{
												DebugEnabled: ec.Bool(false),
											},
										},
										ClusterTopology: []*models.ApmTopologyElement{{
											InstanceConfigurationID: "aws.apm.r5d",
											Size: &models.TopologySize{
												Resource: ec.String("memory"),
												Value:    ec.Int32(512),
											},
											ZoneCount: 1,
										}},
									},
								}},
							},
						}},
						Kibana: []*models.KibanaResourceInfo{{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							RefID:                     ec.String("main-kibana"),
							Region:                    ec.String("aws-eu-central-1"),
							Info: &models.KibanaClusterInfo{
								Status: ec.String("running"),
								PlanInfo: &models.KibanaClusterPlansInfo{Current: &models.KibanaClusterPlanInfo{
									Plan: &models.KibanaClusterPlan{
										Kibana: &models.KibanaConfiguration{
											Version:     "7.14.1",
											DockerImage: "docker.elastic.com/kibana/cloud:7.14.1-hash",
										},
										ClusterTopology: []*models.KibanaClusterTopologyElement{{
											InstanceConfigurationID: "aws.kibana.r5d",
											Size: &models.TopologySize{
												Resource: ec.String("memory"),
												Value:    ec.Int32(1024),
											},
											ZoneCount: 1,
										}},
									},
								}},
							},
						}},
						EnterpriseSearch: []*models.EnterpriseSearchResourceInfo{{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							RefID:                     ec.String("main-enterprise_search"),
							Region:                    ec.String("aws-eu-central-1"),
							Info: &models.EnterpriseSearchInfo{
								Status: ec.String("running"),
								PlanInfo: &models.EnterpriseSearchPlansInfo{Current: &models.EnterpriseSearchPlanInfo{
									Plan: &models.EnterpriseSearchPlan{
										EnterpriseSearch: &models.EnterpriseSearchConfiguration{
											Version:     "7.14.1",
											DockerImage: "docker.elastic.com/enterprise_search/cloud:7.14.1-hash",
										},
										ClusterTopology: []*models.EnterpriseSearchTopologyElement{{
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
											ZoneCount: 2,
										}},
									},
								}},
							},
						}},
					},
				},
			},
			want: Deployment{
				Id:                   "123b7b540dfc967a7a649c18e2fce4ed",
				Alias:                "OH",
				Name:                 "up2d",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "aws-eu-central-1",
				Version:              "7.14.1",
				Elasticsearch: Elasticsearches{
					{
						RefId:  ec.String("main-elasticsearch"),
						Region: ec.String("aws-eu-central-1"),
						Config: ElasticsearchConfigs{
							{
								DockerImage: ec.String("docker.elastic.com/elasticsearch/cloud:7.14.1-hash"),
							},
						},
						Topology: ElasticsearchTopologies{
							{
								Id:           "hot_content",
								Size:         ec.String("4g"),
								SizeResource: ec.String("memory"),
								ZoneCount:    1,
								Config: ElasticsearchTopologyConfigs{
									{
										UserSettingsYaml: ec.String("a.setting: true"),
									},
								},
							},
						},
					},
				},
				Kibana: Kibanas{
					{
						RefId:                     ec.String("main-kibana"),
						Region:                    ec.String("aws-eu-central-1"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Config: KibanaConfigs{
							{
								DockerImage: ec.String("docker.elastic.com/kibana/cloud:7.14.1-hash"),
							},
						},
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
						RefId:                     ec.String("main-apm"),
						Region:                    ec.String("aws-eu-central-1"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Config: ApmConfigs{
							{
								DockerImage:  ec.String("docker.elastic.com/apm/cloud:7.14.1-hash"),
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
						RefId:                     ec.String("main-enterprise_search"),
						Region:                    ec.String("aws-eu-central-1"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Config: EnterpriseSearchConfigs{
							{
								DockerImage: ec.String("docker.elastic.com/enterprise_search/cloud:7.14.1-hash"),
							},
						},
						Topology: EnterpriseSearchTopologies{
							{
								InstanceConfigurationId: ec.String("aws.enterprisesearch.m5d"),
								Size:                    ec.String("2g"),
								SizeResource:            ec.String("memory"),
								ZoneCount:               2,
								NodeTypeAppserver:       ec.Bool(true),
								NodeTypeConnector:       ec.Bool(true),
								NodeTypeWorker:          ec.Bool(true),
							},
						},
					},
				},
			},
		},
		{
			name: "flattens an aws plan (io-optimized) with tags",
			args: args{res: testutil.OpenDeploymentGet(t, "testdata/deployment-aws-io-optimized-tags.json")},
			want: Deployment{
				Id:                   "123365f2805e46808d40849b1c0b266b",
				Alias:                "my-deployment",
				Name:                 "up2d",
				DeploymentTemplateId: "aws-io-optimized-v2",
				Region:               "aws-eu-central-1",
				Version:              "7.9.2",
				Tags: map[string]string{
					"aaa":   "bbb",
					"cost":  "rnd",
					"owner": "elastic",
				},
				Elasticsearch: Elasticsearches{
					{
						RefId:         ec.String("main-elasticsearch"),
						ResourceId:    ec.String("1239f7ee7196439ba2d105319ac5eba7"),
						Region:        ec.String("aws-eu-central-1"),
						Autoscale:     ec.String("false"),
						CloudID:       ec.String("up2d:someCloudID"),
						HttpEndpoint:  ec.String("http://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9200"),
						HttpsEndpoint: ec.String("https://1239f7ee7196439ba2d105319ac5eba7.eu-central-1.aws.cloud.es.io:9243"),
						Topology: ElasticsearchTopologies{
							{
								Id:                      "hot_content",
								InstanceConfigurationId: ec.String("aws.data.highio.i3"),
								Size:                    ec.String("8g"),
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
						ResourceId:                ec.String("123dcfda06254ca789eb287e8b73ff4c"),
						Region:                    ec.String("aws-eu-central-1"),
						HttpEndpoint:              ec.String("http://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9200"),
						HttpsEndpoint:             ec.String("https://123dcfda06254ca789eb287e8b73ff4c.eu-central-1.aws.cloud.es.io:9243"),
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
						ResourceId:                ec.String("12328579b3bf40c8b58c1a0ed5a4bd8b"),
						Region:                    ec.String("aws-eu-central-1"),
						HttpEndpoint:              ec.String("http://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:80"),
						HttpsEndpoint:             ec.String("https://12328579b3bf40c8b58c1a0ed5a4bd8b.apm.eu-central-1.aws.cloud.es.io:443"),
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
			},
		},
		{
			name: "flattens a gcp plan (io-optimized)",
			args: args{res: testutil.OpenDeploymentGet(t, "testdata/deployment-gcp-io-optimized.json")},
			want: Deployment{
				Id:                   "1239e402d6df471ea374bd68e3f91cc5",
				Alias:                "my-deployment",
				Name:                 "up2d",
				DeploymentTemplateId: "gcp-io-optimized",
				Region:               "gcp-asia-east1",
				Version:              "7.9.2",
				Elasticsearch: Elasticsearches{
					{
						RefId:         ec.String("main-elasticsearch"),
						ResourceId:    ec.String("123695e76d914005bf90b717e668ad4b"),
						Region:        ec.String("gcp-asia-east1"),
						Autoscale:     ec.String("false"),
						CloudID:       ec.String("up2d:someCloudID"),
						HttpEndpoint:  ec.String("http://123695e76d914005bf90b717e668ad4b.asia-east1.gcp.elastic-cloud.com:9200"),
						HttpsEndpoint: ec.String("https://123695e76d914005bf90b717e668ad4b.asia-east1.gcp.elastic-cloud.com:9243"),
						Topology: ElasticsearchTopologies{
							{
								Id:                      "hot_content",
								InstanceConfigurationId: ec.String("gcp.data.highio.1"),
								Size:                    ec.String("8g"),
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
						ResourceId:                ec.String("12365046781e4d729a07df64fe67c8c6"),
						Region:                    ec.String("gcp-asia-east1"),
						HttpEndpoint:              ec.String("http://12365046781e4d729a07df64fe67c8c6.asia-east1.gcp.elastic-cloud.com:9200"),
						HttpsEndpoint:             ec.String("https://12365046781e4d729a07df64fe67c8c6.asia-east1.gcp.elastic-cloud.com:9243"),
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("gcp.kibana.1"),
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
						ResourceId:                ec.String("12307c6c304949b8a9f3682b80900879"),
						Region:                    ec.String("gcp-asia-east1"),
						HttpEndpoint:              ec.String("http://12307c6c304949b8a9f3682b80900879.apm.asia-east1.gcp.elastic-cloud.com:80"),
						HttpsEndpoint:             ec.String("https://12307c6c304949b8a9f3682b80900879.apm.asia-east1.gcp.elastic-cloud.com:443"),
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("gcp.apm.1"),
								Size:                    ec.String("0.5g"),
								SizeResource:            ec.String("memory"),
								ZoneCount:               1,
							},
						},
					},
				},
			},
		},
		{
			name: "flattens a gcp plan with autoscale set (io-optimized)",
			args: args{res: testutil.OpenDeploymentGet(t, "testdata/deployment-gcp-io-optimized-autoscale.json")},
			want: Deployment{
				Id:                   "1239e402d6df471ea374bd68e3f91cc5",
				Alias:                "",
				Name:                 "up2d",
				DeploymentTemplateId: "gcp-io-optimized",
				Region:               "gcp-asia-east1",
				Version:              "7.9.2",
				Elasticsearch: Elasticsearches{
					{
						RefId:         ec.String("main-elasticsearch"),
						ResourceId:    ec.String("123695e76d914005bf90b717e668ad4b"),
						Region:        ec.String("gcp-asia-east1"),
						Autoscale:     ec.String("true"),
						CloudID:       ec.String("up2d:someCloudID"),
						HttpEndpoint:  ec.String("http://123695e76d914005bf90b717e668ad4b.asia-east1.gcp.elastic-cloud.com:9200"),
						HttpsEndpoint: ec.String("https://123695e76d914005bf90b717e668ad4b.asia-east1.gcp.elastic-cloud.com:9243"),
						Topology: ElasticsearchTopologies{
							{
								Id:                      "hot_content",
								InstanceConfigurationId: ec.String("gcp.data.highio.1"),
								Size:                    ec.String("8g"),
								SizeResource:            ec.String("memory"),
								NodeTypeData:            ec.String("true"),
								NodeTypeIngest:          ec.String("true"),
								NodeTypeMaster:          ec.String("true"),
								NodeTypeMl:              ec.String("false"),
								ZoneCount:               2,
								Autoscaling: ElasticsearchTopologyAutoscalings{
									{
										MaxSize:            ec.String("29g"),
										MaxSizeResource:    ec.String("memory"),
										PolicyOverrideJson: ec.String(`{"proactive_storage":{"forecast_window":"3 h"}}`),
									},
								},
							},
							{
								Id:                      "ml",
								InstanceConfigurationId: ec.String("gcp.ml.1"),
								Size:                    ec.String("1g"),
								SizeResource:            ec.String("memory"),
								NodeTypeData:            ec.String("false"),
								NodeTypeIngest:          ec.String("false"),
								NodeTypeMaster:          ec.String("false"),
								NodeTypeMl:              ec.String("true"),
								ZoneCount:               1,
								Autoscaling: ElasticsearchTopologyAutoscalings{
									{
										MaxSize:         ec.String("30g"),
										MaxSizeResource: ec.String("memory"),
										MinSize:         ec.String("1g"),
										MinSizeResource: ec.String("memory"),
									},
								},
							},
						},
					},
				},
				Kibana: Kibanas{
					{
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						RefId:                     ec.String("main-kibana"),
						ResourceId:                ec.String("12365046781e4d729a07df64fe67c8c6"),
						Region:                    ec.String("gcp-asia-east1"),
						HttpEndpoint:              ec.String("http://12365046781e4d729a07df64fe67c8c6.asia-east1.gcp.elastic-cloud.com:9200"),
						HttpsEndpoint:             ec.String("https://12365046781e4d729a07df64fe67c8c6.asia-east1.gcp.elastic-cloud.com:9243"),
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("gcp.kibana.1"),
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
						ResourceId:                ec.String("12307c6c304949b8a9f3682b80900879"),
						Region:                    ec.String("gcp-asia-east1"),
						HttpEndpoint:              ec.String("http://12307c6c304949b8a9f3682b80900879.apm.asia-east1.gcp.elastic-cloud.com:80"),
						HttpsEndpoint:             ec.String("https://12307c6c304949b8a9f3682b80900879.apm.asia-east1.gcp.elastic-cloud.com:443"),
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("gcp.apm.1"),
								Size:                    ec.String("0.5g"),
								SizeResource:            ec.String("memory"),
								ZoneCount:               1,
							},
						},
					},
				},
			},
		},
		{
			name: "flattens a gcp plan (hot-warm)",
			args: args{res: testutil.OpenDeploymentGet(t, "testdata/deployment-gcp-hot-warm.json")},
			want: Deployment{
				Id:                   "123d148423864552aa57b59929d4bf4d",
				Name:                 "up2d-hot-warm",
				DeploymentTemplateId: "gcp-hot-warm",
				Region:               "gcp-us-central1",
				Version:              "7.9.2",
				Elasticsearch: Elasticsearches{
					{
						RefId:         ec.String("main-elasticsearch"),
						ResourceId:    ec.String("123e837db6ee4391bb74887be35a7a91"),
						Region:        ec.String("gcp-us-central1"),
						Autoscale:     ec.String("false"),
						CloudID:       ec.String("up2d-hot-warm:someCloudID"),
						HttpEndpoint:  ec.String("http://123e837db6ee4391bb74887be35a7a91.us-central1.gcp.cloud.es.io:9200"),
						HttpsEndpoint: ec.String("https://123e837db6ee4391bb74887be35a7a91.us-central1.gcp.cloud.es.io:9243"),
						Topology: ElasticsearchTopologies{
							{
								Id:                      "hot_content",
								InstanceConfigurationId: ec.String("gcp.data.highio.1"),
								Size:                    ec.String("4g"),
								SizeResource:            ec.String("memory"),
								NodeTypeData:            ec.String("true"),
								NodeTypeIngest:          ec.String("true"),
								NodeTypeMaster:          ec.String("true"),
								NodeTypeMl:              ec.String("false"),
								ZoneCount:               2,
							},
							{
								Id:                      "warm",
								InstanceConfigurationId: ec.String("gcp.data.highstorage.1"),
								Size:                    ec.String("4g"),
								SizeResource:            ec.String("memory"),
								NodeTypeData:            ec.String("true"),
								NodeTypeIngest:          ec.String("true"),
								NodeTypeMaster:          ec.String("false"),
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
						ResourceId:                ec.String("12372cc60d284e7e96b95ad14727c23d"),
						Region:                    ec.String("gcp-us-central1"),
						HttpEndpoint:              ec.String("http://12372cc60d284e7e96b95ad14727c23d.us-central1.gcp.cloud.es.io:9200"),
						HttpsEndpoint:             ec.String("https://12372cc60d284e7e96b95ad14727c23d.us-central1.gcp.cloud.es.io:9243"),
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("gcp.kibana.1"),
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
						ResourceId:                ec.String("1234b68b0b9347f1b49b1e01b33bf4a4"),
						Region:                    ec.String("gcp-us-central1"),
						HttpEndpoint:              ec.String("http://1234b68b0b9347f1b49b1e01b33bf4a4.apm.us-central1.gcp.cloud.es.io:80"),
						HttpsEndpoint:             ec.String("https://1234b68b0b9347f1b49b1e01b33bf4a4.apm.us-central1.gcp.cloud.es.io:443"),
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("gcp.apm.1"),
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
