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
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/testutil"
)

func fileAsResponseBody(t *testing.T, name string) io.ReadCloser {
	t.Helper()
	f, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var buf = new(bytes.Buffer)
	if _, err := io.Copy(buf, f); err != nil {
		t.Fatal(err)
	}
	buf.WriteString("\n")

	return io.NopCloser(buf)
}

func Test_createRequest(t *testing.T) {

	sampleKibana := &Kibana{
		ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
		RefId:                     ec.String("main-kibana"),
		ResourceId:                ec.String(mock.ValidClusterID),
		Region:                    ec.String("us-east-1"),
		Topology: Topologies{
			{
				InstanceConfigurationId: ec.String("aws.kibana.r5d"),
				Size:                    ec.String("1g"),
				ZoneCount:               1,
			},
		},
	}

	sampleApm := &Apm{
		ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
		RefId:                     ec.String("main-apm"),
		ResourceId:                ec.String(mock.ValidClusterID),
		Region:                    ec.String("us-east-1"),
		Config: &ApmConfig{
			DebugEnabled: ec.Bool(false),
		},
		Topology: Topologies{
			{
				InstanceConfigurationId: ec.String("aws.apm.r5d"),
				Size:                    ec.String("0.5g"),
				ZoneCount:               1,
			},
		},
	}

	sampleEnterpriseSearch := &EnterpriseSearch{
		ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
		RefId:                     ec.String("main-enterprise_search"),
		ResourceId:                ec.String(mock.ValidClusterID),
		Region:                    ec.String("us-east-1"),
		Topology: EnterpriseSearchTopologies{
			{
				InstanceConfigurationId: ec.String("aws.enterprisesearch.m5d"),
				Size:                    ec.String("2g"),
				ZoneCount:               1,
				NodeTypeAppserver:       ec.Bool(true),
				NodeTypeConnector:       ec.Bool(true),
				NodeTypeWorker:          ec.Bool(true),
			},
		},
	}

	sampleObservability := &Observability{
		DeploymentId: ec.String(mock.ValidClusterID),
		RefId:        ec.String("main-elasticsearch"),
		Logs:         true,
		Metrics:      true,
	}

	sampleDeployment := Deployment{
		Alias:                "my-deployment",
		Name:                 "my_deployment_name",
		DeploymentTemplateId: "aws-hot-warm-v2",
		Region:               "us-east-1",
		Version:              "7.11.1",
		Elasticsearch: &Elasticsearch{
			RefId:      ec.String("main-elasticsearch"),
			ResourceId: ec.String(mock.ValidClusterID),
			Region:     ec.String("us-east-1"),
			Config: &ElasticsearchConfig{
				UserSettingsYaml:         ec.String("some.setting: value"),
				UserSettingsOverrideYaml: ec.String("some.setting: value2"),
				UserSettingsJson:         ec.String("{\"some.setting\":\"value\"}"),
				UserSettingsOverrideJson: ec.String("{\"some.setting\":\"value2\"}"),
			},
			Topology: ElasticsearchTopologies{
				{
					Id:   "hot_content",
					Size: ec.String("2g"),
					NodeRoles: []string{
						"master",
						"ingest",
						"remote_cluster_client",
						"data_hot",
						"transform",
						"data_content",
					},
					ZoneCount: 1,
				},
				{
					Id:   "warm",
					Size: ec.String("2g"),
					NodeRoles: []string{
						"data_warm",
						"remote_cluster_client",
					},
					ZoneCount: 1,
				},
			},
		},
		Kibana:           sampleKibana,
		Apm:              sampleApm,
		EnterpriseSearch: sampleEnterpriseSearch,
		Observability:    sampleObservability,
		TrafficFilter:    []string{"0.0.0.0/0", "192.168.10.0/24"},
	}

	sampleElasticsearch := &Elasticsearch{
		RefId:      ec.String("main-elasticsearch"),
		ResourceId: ec.String(mock.ValidClusterID),
		Region:     ec.String("us-east-1"),
		Config: &ElasticsearchConfig{
			UserSettingsYaml:         ec.String("some.setting: value"),
			UserSettingsOverrideYaml: ec.String("some.setting: value2"),
			UserSettingsJson:         ec.String("{\"some.setting\":\"value\"}"),
			UserSettingsOverrideJson: ec.String("{\"some.setting\":\"value2\"}"),
		},
		Topology: ElasticsearchTopologies{
			{
				Id:                      "hot_content",
				InstanceConfigurationId: ec.String("aws.data.highio.i3"),
				Size:                    ec.String("2g"),
				NodeTypeData:            ec.String("true"),
				NodeTypeIngest:          ec.String("true"),
				NodeTypeMaster:          ec.String("true"),
				NodeTypeMl:              ec.String("false"),
				ZoneCount:               1,
			},
		},
	}

	sampleLegacyDeployment := Deployment{
		Alias:                "my-deployment",
		Name:                 "my_deployment_name",
		DeploymentTemplateId: "aws-io-optimized-v2",
		Region:               "us-east-1",
		Version:              "7.7.0",
		Elasticsearch:        sampleElasticsearch,
		Kibana:               sampleKibana,
		Apm:                  sampleApm,
		EnterpriseSearch:     sampleEnterpriseSearch,
		Observability:        sampleObservability,
		TrafficFilter:        []string{"0.0.0.0/0", "192.168.10.0/24"},
	}

	ioOptimizedTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-aws-io-optimized-v2.json")
	}

	hotWarmTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-aws-hot-warm-v2.json")
	}

	ccsTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-aws-cross-cluster-search-v2.json")
	}

	emptyTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-empty.json")
	}

	type args struct {
		plan   Deployment
		client *api.API
	}
	tests := []struct {
		name  string
		args  args
		want  *models.DeploymentCreateRequest
		diags diag.Diagnostics
	}{
		{
			name: "parses the resources",
			args: args{
				plan: sampleDeployment,
				client: api.NewMock(
					mock.New200Response(hotWarmTpl()),
					mock.New200Response(
						mock.NewStructBody(models.DeploymentGetResponse{
							Healthy: ec.Bool(true),
							ID:      ec.String(mock.ValidClusterID),
							Resources: &models.DeploymentResources{
								Elasticsearch: []*models.ElasticsearchResourceInfo{{
									ID:    ec.String(mock.ValidClusterID),
									RefID: ec.String("main-elasticsearch"),
								}},
							},
						}),
					),
				),
			},
			want: &models.DeploymentCreateRequest{
				Name:  "my_deployment_name",
				Alias: "my-deployment",
				Settings: &models.DeploymentCreateSettings{
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
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, hotWarmTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version:                  "7.11.1",
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
								ID: ec.String("aws-hot-warm-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									NodeRoles: []string{
										"master",
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highstorage.d2",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "warm"},
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
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
		{
			name: "parses the legacy resources",
			args: args{
				plan: sampleLegacyDeployment,
				client: api.NewMock(
					mock.New200Response(ioOptimizedTpl()),
					mock.New200Response(
						mock.NewStructBody(models.DeploymentGetResponse{
							Healthy: ec.Bool(true),
							ID:      ec.String(mock.ValidClusterID),
							Resources: &models.DeploymentResources{
								Elasticsearch: []*models.ElasticsearchResourceInfo{{
									ID:    ec.String(mock.ValidClusterID),
									RefID: ec.String("main-elasticsearch"),
								}},
							},
						}),
					),
				),
			},
			want: &models.DeploymentCreateRequest{
				Name:  "my_deployment_name",
				Alias: "my-deployment",
				Settings: &models.DeploymentCreateSettings{
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
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
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
								ID:                      "hot_content",
								ZoneCount:               1,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(2048),
								},
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
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
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
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
		{
			name: "parses the resources with empty declarations (IO Optimized)",
			args: args{
				plan: Deployment{
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.7.0",
					Elasticsearch:        &Elasticsearch{},
					Kibana:               &Kibana{},
					Apm:                  &Apm{},
					EnterpriseSearch:     &EnterpriseSearch{},
					TrafficFilter:        []string{"0.0.0.0/0", "192.168.10.0/24"},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			// Ref ids are taken from template, not from defautls values in this test.
			// Defaults are processed by TF during config processing.
			want: &models.DeploymentCreateRequest{
				Name: "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{
					TrafficFilterSettings: &models.TrafficFilterSettings{
						Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
					},
				},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("es-ref-id"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.7.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(8192),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("es-ref-id"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("kibana-ref-id"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("es-ref-id"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("apm-ref-id"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("es-ref-id"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("enterprise_search-ref-id"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
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
		{
			name: "parses the resources with empty declarations (IO Optimized) with node_roles",
			args: args{
				plan: Deployment{
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.11.0",
					Elasticsearch:        &Elasticsearch{},
					Kibana:               &Kibana{},
					Apm:                  &Apm{},
					EnterpriseSearch:     &EnterpriseSearch{},
					TrafficFilter:        []string{"0.0.0.0/0", "192.168.10.0/24"},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name: "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{
					TrafficFilterSettings: &models.TrafficFilterSettings{
						Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
					},
				},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				// Ref ids are taken from template, not from defautls values in this test.
				// Defaults are processed by TF during config processing.
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("es-ref-id"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.11.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(8192),
								},
								NodeRoles: []string{
									"master",
									"ingest",
									"remote_cluster_client",
									"data_hot",
									"transform",
									"data_content",
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("es-ref-id"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("kibana-ref-id"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("es-ref-id"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("apm-ref-id"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("es-ref-id"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("enterprise_search-ref-id"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
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
		{
			name: "parses the resources with topology overrides (size)",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Alias:                "my-deployment",
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.7.0",
					Elasticsearch: &Elasticsearch{
						RefId: ec.String("main-elasticsearch"),
						Topology: ElasticsearchTopologies{
							{
								Id:   "hot_content",
								Size: ec.String("4g"),
							},
						},
					},
					Kibana: &Kibana{
						RefId:                     ec.String("main-kibana"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Topology: Topologies{
							{
								Size: ec.String("2g"),
							},
						},
					},
					Apm: &Apm{
						RefId:                     ec.String("main-apm"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Topology: Topologies{
							{
								Size: ec.String("1g"),
							},
						},
					},
					EnterpriseSearch: &EnterpriseSearch{
						RefId:                     ec.String("main-enterprise_search"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Topology: EnterpriseSearchTopologies{
							{
								Size: ec.String("4g"),
							},
						},
					},
					TrafficFilter: []string{"0.0.0.0/0", "192.168.10.0/24"},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:  "my_deployment_name",
				Alias: "my-deployment",
				Settings: &models.DeploymentCreateSettings{
					TrafficFilterSettings: &models.TrafficFilterSettings{
						Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
					},
				},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.7.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(4096),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
								ClusterTopology: []*models.KibanaClusterTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.kibana.r5d",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(2048),
										},
									},
								},
							},
						},
					},
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
								ClusterTopology: []*models.ApmTopologyElement{{
									ZoneCount:               1,
									InstanceConfigurationID: "aws.apm.r5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								}},
							},
						},
					},
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
										InstanceConfigurationID: "aws.enterprisesearch.m5d",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(4096),
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
		{
			name: "parses the resources with topology overrides (IC)",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Alias:                "my-deployment",
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.7.0",
					Elasticsearch: &Elasticsearch{
						RefId: ec.String("main-elasticsearch"),
						Topology: ElasticsearchTopologies{
							{
								Id: "hot_content",
							},
						},
					},
					Kibana: &Kibana{
						RefId:                     ec.String("main-kibana"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("aws.kibana.r5d"),
							},
						},
					},
					Apm: &Apm{
						RefId:                     ec.String("main-apm"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("aws.apm.r5d"),
							},
						},
					},
					EnterpriseSearch: &EnterpriseSearch{
						RefId:                     ec.String("main-enterprise_search"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Topology: EnterpriseSearchTopologies{
							{
								InstanceConfigurationId: ec.String("aws.enterprisesearch.m5d"),
							},
						},
					},
					TrafficFilter: []string{"0.0.0.0/0", "192.168.10.0/24"},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:  "my_deployment_name",
				Alias: "my-deployment",
				Settings: &models.DeploymentCreateSettings{
					TrafficFilterSettings: &models.TrafficFilterSettings{
						Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
					},
				},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.7.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(8192),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
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
		{
			name: "parses the resources with empty declarations (Hot Warm)",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-hot-warm-v2",
					Region:               "us-east-1",
					Version:              "7.9.2",
					Elasticsearch:        &Elasticsearch{},
					Kibana:               &Kibana{},
				},
				client: api.NewMock(mock.New200Response(hotWarmTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, hotWarmTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("es-ref-id"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
							Curation:                  nil,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Curation: nil,
								Version:  "7.9.2",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-hot-warm-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeType: &models.ElasticsearchNodeType{
										Data:   ec.Bool(true),
										Ingest: ec.Bool(true),
										Master: ec.Bool(true),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d2",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeType: &models.ElasticsearchNodeType{
										Data:   ec.Bool(true),
										Ingest: ec.Bool(true),
										Master: ec.Bool(false),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("es-ref-id"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("kibana-ref-id"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
		{
			name: "parses the resources with empty declarations (Hot Warm) with node_roles",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-hot-warm-v2",
					Region:               "us-east-1",
					Version:              "7.12.0",
					Elasticsearch:        &Elasticsearch{},
					Kibana:               &Kibana{},
				},
				client: api.NewMock(mock.New200Response(hotWarmTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, hotWarmTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("es-ref-id"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
							Curation:                  nil,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Curation: nil,
								Version:  "7.12.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-hot-warm-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"master",
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d2",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("es-ref-id"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("kibana-ref-id"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
		{
			name: "parses the resources with empty declarations (Hot Warm) with node_roles and extensions",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-hot-warm-v2",
					Region:               "us-east-1",
					Version:              "7.12.0",
					Elasticsearch: &Elasticsearch{
						RefId: ec.String("main-elasticsearch"),
						Extension: ElasticsearchExtensions{
							{
								Name:    "my-plugin",
								Type:    "plugin",
								Url:     "repo://12311234",
								Version: "7.7.0",
							},
							{
								Name:    "my-second-plugin",
								Type:    "plugin",
								Url:     "repo://12311235",
								Version: "7.7.0",
							},
							{
								Name:    "my-bundle",
								Type:    "bundle",
								Url:     "repo://1231122",
								Version: "7.7.0",
							},
							{
								Name:    "my-second-bundle",
								Type:    "bundle",
								Url:     "repo://1231123",
								Version: "7.7.0",
							},
						},
					},
				},
				client: api.NewMock(mock.New200Response(hotWarmTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, hotWarmTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.0",
								UserBundles: []*models.ElasticsearchUserBundle{
									{
										URL:                  ec.String("repo://1231122"),
										Name:                 ec.String("my-bundle"),
										ElasticsearchVersion: ec.String("7.7.0"),
									},
									{
										URL:                  ec.String("repo://1231123"),
										Name:                 ec.String("my-second-bundle"),
										ElasticsearchVersion: ec.String("7.7.0"),
									},
								},
								UserPlugins: []*models.ElasticsearchUserPlugin{
									{
										URL:                  ec.String("repo://12311234"),
										Name:                 ec.String("my-plugin"),
										ElasticsearchVersion: ec.String("7.7.0"),
									},
									{
										URL:                  ec.String("repo://12311235"),
										Name:                 ec.String("my-second-plugin"),
										ElasticsearchVersion: ec.String("7.7.0"),
									},
								},
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-hot-warm-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"master",
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d2",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					})},
				},
			},
		},
		{
			name: "deployment with autoscaling enabled",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.12.0",
					Elasticsearch: &Elasticsearch{
						RefId:     ec.String("main-elasticsearch"),
						Autoscale: ec.String("true"),
						Topology: ElasticsearchTopologies{
							{
								Id:   "cold",
								Size: ec.String("2g"),
							},
							{
								Id:   "hot_content",
								Size: ec.String("8g"),
							},
							{
								Id:   "warm",
								Size: ec.String("4g"),
							},
						},
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(true),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "cold",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"data_cold",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "cold",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(59392),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"master",
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					})},
				},
			},
		},
		{
			name: "deployment with autoscaling enabled and custom policies set",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.12.0",
					Elasticsearch: &Elasticsearch{
						RefId:     ec.String("main-elasticsearch"),
						Autoscale: ec.String("true"),
						Topology: ElasticsearchTopologies{
							{
								Id:   "cold",
								Size: ec.String("2g"),
							},
							{
								Id:   "hot_content",
								Size: ec.String("8g"),
								Autoscaling: ElasticsearchTopologyAutoscalings{
									{
										MaxSize: ec.String("232g"),
									},
								},
							},
							{
								Id:   "warm",
								Size: ec.String("4g"),
								Autoscaling: ElasticsearchTopologyAutoscalings{
									{
										MaxSize: ec.String("116g"),
									},
								},
							},
						},
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(true),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "cold",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"data_cold",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "cold",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(59392),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"master",
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(237568),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					})},
				},
			},
		},
		{
			name: "deployment with dedicated master and cold tiers",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.12.0",
					Elasticsearch: &Elasticsearch{
						RefId: ec.String("main-elasticsearch"),
						Topology: ElasticsearchTopologies{
							{
								Id:   "cold",
								Size: ec.String("2g"),
							},
							{
								Id:   "hot_content",
								Size: ec.String("8g"),
							},
							{
								Id:        "master",
								Size:      ec.String("1g"),
								ZoneCount: 3,
							},
							{
								Id:   "warm",
								Size: ec.String("4g"),
							},
						},
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "cold",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"data_cold",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "cold",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(59392),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "master",
									ZoneCount:               3,
									InstanceConfigurationID: "aws.master.r5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
									NodeRoles: []string{
										"master",
										"remote_cluster_client",
									},
									// Elasticsearch: &models.ElasticsearchConfiguration{},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					})},
				},
			},
		},
		{
			name: "deployment with dedicated coordinating and cold tiers",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.12.0",
					Elasticsearch: &Elasticsearch{
						RefId: ec.String("main-elasticsearch"),
						Topology: ElasticsearchTopologies{
							{
								Id:   "cold",
								Size: ec.String("2g"),
							},
							{
								Id:        "coordinating",
								Size:      ec.String("2g"),
								ZoneCount: 2,
							},
							{
								Id:   "hot_content",
								Size: ec.String("8g"),
							},
							{
								Id:   "warm",
								Size: ec.String("4g"),
							},
						},
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "cold",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"data_cold",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "cold",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(59392),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "coordinating",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.coordinating.m5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"ingest",
										"remote_cluster_client",
									},
									// Elasticsearch: &models.ElasticsearchConfiguration{},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
								},
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"master",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					})},
				},
			},
		},
		{
			name: "deployment with dedicated coordinating, master and cold tiers",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.12.0",
					Elasticsearch: &Elasticsearch{
						RefId: ec.String("main-elasticsearch"),
						Topology: ElasticsearchTopologies{
							{
								Id:   "cold",
								Size: ec.String("2g"),
							},
							{
								Id:        "coordinating",
								Size:      ec.String("2g"),
								ZoneCount: 2,
							},
							{
								Id:   "hot_content",
								Size: ec.String("8g"),
							},
							{
								Id:        "master",
								Size:      ec.String("1g"),
								ZoneCount: 3,
							},
							{
								Id:   "warm",
								Size: ec.String("4g"),
							},
						},
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "cold",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"data_cold",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "cold",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(59392),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "coordinating",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.coordinating.m5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"ingest",
										"remote_cluster_client",
									},
									// Elasticsearch: &models.ElasticsearchConfiguration{},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
								},
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "master",
									ZoneCount:               3,
									InstanceConfigurationID: "aws.master.r5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
									NodeRoles: []string{
										"master",
										"remote_cluster_client",
									},
									// Elasticsearch: &models.ElasticsearchConfiguration{},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					})},
				},
			},
		},
		{
			name: "deployment with docker_image overrides",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.14.1",
					Elasticsearch: &Elasticsearch{
						RefId: ec.String("main-elasticsearch"),
						Config: &ElasticsearchConfig{
							DockerImage: ec.String("docker.elastic.com/elasticsearch/container:7.14.1-hash"),
						},
						Autoscale: ec.String("false"),
						TrustAccount: ElasticsearchTrustAccounts{
							{
								AccountId: ec.String("ANID"),
								TrustAll:  ec.Bool(true),
							},
						},
						Topology: ElasticsearchTopologies{
							{
								Id:   "hot_content",
								Size: ec.String("8g"),
							},
						},
					},
					Kibana: &Kibana{
						RefId:                     ec.String("main-kibana"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Config: &KibanaConfig{
							DockerImage: ec.String("docker.elastic.com/kibana/container:7.14.1-hash"),
						},
					},
					Apm: &Apm{
						RefId:                     ec.String("main-apm"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Config: &ApmConfig{
							DockerImage: ec.String("docker.elastic.com/apm/container:7.14.1-hash"),
						},
					},
					EnterpriseSearch: &EnterpriseSearch{
						RefId:                     ec.String("main-enterprise_search"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Config: &EnterpriseSearchConfig{
							DockerImage: ec.String("docker.elastic.com/enterprise_search/container:7.14.1-hash"),
						},
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
							Trust: &models.ElasticsearchClusterTrustSettings{
								Accounts: []*models.AccountTrustRelationship{
									{
										AccountID: ec.String("ANID"),
										TrustAll:  ec.Bool(true),
									},
								},
							},
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version:     "7.14.1",
								DockerImage: "docker.elastic.com/elasticsearch/container:7.14.1-hash",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"master",
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					})},
					Apm: []*models.ApmPayload{{
						ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
						Plan: &models.ApmPlan{
							Apm: &models.ApmConfiguration{
								DockerImage: "docker.elastic.com/apm/container:7.14.1-hash",
								// SystemSettings: &models.ApmSystemSettings{
								// 	DebugEnabled: ec.Bool(false),
								// },
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
						RefID:  ec.String("main-apm"),
						Region: ec.String("us-east-1"),
					}},
					Kibana: []*models.KibanaPayload{{
						ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
						Plan: &models.KibanaClusterPlan{
							Kibana: &models.KibanaConfiguration{
								DockerImage: "docker.elastic.com/kibana/container:7.14.1-hash",
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
						RefID:  ec.String("main-kibana"),
						Region: ec.String("us-east-1"),
					}},
					EnterpriseSearch: []*models.EnterpriseSearchPayload{{
						ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
						Plan: &models.EnterpriseSearchPlan{
							EnterpriseSearch: &models.EnterpriseSearchConfiguration{
								DockerImage: "docker.elastic.com/enterprise_search/container:7.14.1-hash",
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
						RefID:  ec.String("main-enterprise_search"),
						Region: ec.String("us-east-1"),
					}},
				},
			},
		},
		{
			name: "deployment with trust settings set",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.12.0",
					Elasticsearch: &Elasticsearch{
						RefId:     ec.String("main-elasticsearch"),
						Autoscale: ec.String("false"),
						TrustAccount: ElasticsearchTrustAccounts{
							{
								AccountId: ec.String("ANID"),
								TrustAll:  ec.Bool(true),
							},
							{
								AccountId:      ec.String("anotherID"),
								TrustAll:       ec.Bool(false),
								TrustAllowlist: []string{"abc", "hij", "dfg"},
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
						Topology: ElasticsearchTopologies{
							{
								Id:   "cold",
								Size: ec.String("2g"),
							},
							{
								Id:   "hot_content",
								Size: ec.String("8g"),
								Autoscaling: ElasticsearchTopologyAutoscalings{
									{
										MaxSize: ec.String("232g"),
									},
								},
							},
							{
								Id:   "warm",
								Size: ec.String("4g"),
								Autoscaling: ElasticsearchTopologyAutoscalings{
									{
										MaxSize: ec.String("116g"),
									},
								},
							},
						},
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
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
											"abc", "hij", "dfg",
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
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.12.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "cold",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(2048),
									},
									NodeRoles: []string{
										"data_cold",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "cold",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(59392),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(8192),
									},
									NodeRoles: []string{
										"master",
										"ingest",
										"remote_cluster_client",
										"data_hot",
										"transform",
										"data_content",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(237568),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeRoles: []string{
										"data_warm",
										"remote_cluster_client",
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{
											"data": "warm",
										},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					})},
				},
			},
		},
		{
			name: "parses the resources with empty declarations (Cross Cluster Search)",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-cross-cluster-search-v2",
					Region:               "us-east-1",
					Version:              "7.9.2",
					Elasticsearch: &Elasticsearch{
						RefId: ec.String("main-elasticsearch"),
					},
					Kibana: &Kibana{
						RefId:                     ec.String("main-kibana"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
					},
				},
				client: api.NewMock(mock.New200Response(ccsTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ccsTpl(), false), &models.ElasticsearchPayload{
						Region:   ec.String("us-east-1"),
						RefID:    ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{},
						Plan: &models.ElasticsearchClusterPlan{
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.9.2",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-cross-cluster-search-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               1,
									InstanceConfigurationID: "aws.ccs.r5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
									NodeType: &models.ElasticsearchNodeType{
										Data:   ec.Bool(true),
										Ingest: ec.Bool(true),
										Master: ec.Bool(true),
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
								},
							},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
		{
			name: "parses the resources with tags",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.10.1",
					Elasticsearch: &Elasticsearch{
						RefId: ec.String("main-elasticsearch"),
						Topology: ElasticsearchTopologies{
							{
								Id:   "hot_content",
								Size: ec.String("8g"),
							},
						},
					},
					Tags: map[string]string{
						"aaa":         "bbb",
						"owner":       "elastic",
						"cost-center": "rnd",
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{Tags: []*models.MetadataItem{
					{Key: ec.String("aaa"), Value: ec.String("bbb")},
					{Key: ec.String("cost-center"), Value: ec.String("rnd")},
					{Key: ec.String("owner"), Value: ec.String("elastic")},
				}},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.10.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(8192),
								},
								NodeRoles: []string{
									"master",
									"ingest",
									"remote_cluster_client",
									"data_hot",
									"transform",
									"data_content",
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					})},
				},
			},
		},
		{
			name: "handles a snapshot_source block, leaving the strategy as is",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.10.1",
					Elasticsearch: &Elasticsearch{
						RefId: ec.String("main-elasticsearch"),
						SnapshotSource: ElasticsearchSnapshotSources{
							{
								SourceElasticsearchClusterId: "8c63b87af9e24ea49b8a4bfe550e5fe9",
							},
						},
						Topology: ElasticsearchTopologies{
							{
								Id:   "hot_content",
								Size: ec.String("8g"),
							},
						},
					},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentCreateRequest{
				Name:     "my_deployment_name",
				Settings: &models.DeploymentCreateSettings{},
				Metadata: &models.DeploymentCreateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentCreateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							Transient: &models.TransientElasticsearchPlanConfiguration{
								RestoreSnapshot: &models.RestoreSnapshotConfiguration{
									SourceClusterID: "8c63b87af9e24ea49b8a4bfe550e5fe9",
									SnapshotName:    ec.String(""),
								},
							},
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.10.1",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(8192),
								},
								NodeRoles: []string{
									"master",
									"ingest",
									"remote_cluster_client",
									"data_hot",
									"transform",
									"data_content",
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					})},
				},
			},
		},
		// This case we're using an empty deployment_template to ensure that
		// resources not present in the template cannot be expanded, receiving
		// an error instead.
		{
			name: "parses the resources with empty explicit declarations (Empty deployment template)",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.10.1",
					Elasticsearch:        &Elasticsearch{},
					Kibana:               &Kibana{},
					Apm:                  &Apm{},
					EnterpriseSearch:     &EnterpriseSearch{},
				},
				client: api.NewMock(mock.New200Response(emptyTpl())),
			},
			diags: func() diag.Diagnostics {
				var diags diag.Diagnostics
				diags.AddError("kibana payload error", "kibana specified but deployment template is not configured for it. Use a different template if you wish to add kibana")
				diags.AddError("apm payload error", "apm specified but deployment template is not configured for it. Use a different template if you wish to add apm")
				diags.AddError("enterprise_search payload error", "deployment template is not configured for enterprise_search. Use a different template if you wish to add enterprise_search")
				return diags
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &Resource{}
			schema, diags := resource.GetSchema(context.Background())
			assert.Nil(t, diags)

			var plan DeploymentTF
			diags = tfsdk.ValueFrom(context.Background(), &tt.args.plan, schema.Type(), &plan)
			assert.Nil(t, diags)

			got, diags := plan.CreateRequest(context.Background(), tt.args.client)
			if tt.diags != nil {
				assert.Equal(t, diags, tt.diags)
			} else {
				assert.Nil(t, diags)
				assert.NotNil(t, got)
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

func Test_updateResourceToModel(t *testing.T) {
	// deploymentRD := util.NewResourceData(t, util.ResDataParams{
	// 	ID:     mock.ValidClusterID,
	// 	State:  newSampleLegacyDeployment(),
	// 	Schema: newSchema(),
	// })
	var ioOptimizedTpl = func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-aws-io-optimized-v2.json")
	}
	// deploymentEmptyRD := util.NewResourceData(t, util.ResDataParams{
	// 	ID:     mock.ValidClusterID,
	// 	State:  newSampleDeploymentEmptyRD(),
	// 	Schema: newSchema(),
	// })
	// deploymentOverrideRd := util.NewResourceData(t, util.ResDataParams{
	// 	ID:     mock.ValidClusterID,
	// 	State:  newSampleDeploymentOverrides(),
	// 	Schema: newSchema(),
	// })

	hotWarmTpl := func() io.ReadCloser {
		return fileAsResponseBody(t, "testdata/template-aws-hot-warm-v2.json")
	}
	// deploymentHotWarm := util.NewResourceData(t, util.ResDataParams{
	// 	ID:     mock.ValidClusterID,
	// 	Schema: newSchema(),
	// 	State: map[string]interface{}{
	// 		"name":                   "my_deployment_name",
	// 		"deployment_template_id": "aws-hot-warm-v2",
	// 		"region":                 "us-east-1",
	// 		"version":                "7.9.2",
	// 		"elasticsearch":          []interface{}{map[string]interface{}{}},
	// 		"kibana":                 []interface{}{map[string]interface{}{}},
	// 	},
	// })

	// ccsTpl := func() io.ReadCloser {
	// 	return fileAsResponseBody(t, "testdata/template-aws-cross-cluster-search-v2.json")
	// }
	// deploymentEmptyRDWithTemplateChange := util.NewResourceData(t, util.ResDataParams{
	// 	ID:    mock.ValidClusterID,
	// 	State: newSampleLegacyDeployment(),
	// 	Change: map[string]interface{}{
	// 		"name":                   "my_deployment_name",
	// 		"deployment_template_id": "aws-cross-cluster-search-v2",
	// 		"region":                 "us-east-1",
	// 		"version":                "7.9.2",
	// 		"elasticsearch":          []interface{}{map[string]interface{}{}},
	// 		"kibana":                 []interface{}{map[string]interface{}{}},
	// 	},
	// 	Schema: newSchema(),
	// })

	// deploymentEmptyRDWithTemplateChangeWithDiffSize := util.NewResourceData(t, util.ResDataParams{
	// 	ID: mock.ValidClusterID,
	// 	State: map[string]interface{}{
	// 		"name":                   "my_deployment_name",
	// 		"deployment_template_id": "aws-io-optimized-v2",
	// 		"region":                 "us-east-1",
	// 		"version":                "7.9.2",
	// 		"elasticsearch": []interface{}{
	// 			map[string]interface{}{
	// 				"topology": []interface{}{map[string]interface{}{
	// 					"id":   "hot_content",
	// 					"size": "16g",
	// 				}},
	// 			},
	// 			map[string]interface{}{
	// 				"topology": []interface{}{map[string]interface{}{
	// 					"id":   "coordinating",
	// 					"size": "16g",
	// 				}},
	// 			},
	// 		},
	// 		"kibana": []interface{}{map[string]interface{}{
	// 			"topology": []interface{}{map[string]interface{}{
	// 				"size": "2g",
	// 			}},
	// 		}},
	// 		"apm": []interface{}{map[string]interface{}{
	// 			"topology": []interface{}{map[string]interface{}{
	// 				"size": "1g",
	// 			}},
	// 		}},
	// 		"enterprise_search": []interface{}{map[string]interface{}{
	// 			"topology": []interface{}{map[string]interface{}{
	// 				"size": "2g",
	// 			}},
	// 		}},
	// 	},
	// 	Change: map[string]interface{}{
	// 		"name":                   "my_deployment_name",
	// 		"deployment_template_id": "aws-cross-cluster-search-v2",
	// 		"region":                 "us-east-1",
	// 		"version":                "7.9.2",
	// 		"elasticsearch":          []interface{}{map[string]interface{}{}},
	// 		"kibana":                 []interface{}{map[string]interface{}{}},
	// 	},
	// 	Schema: newSchema(),
	// })

	// emptyTpl := func() io.ReadCloser {
	// 	return fileAsResponseBody(t, "testdata/template-empty.json")
	// }
	// deploymentChangeFromExplicitSizingToEmpty := util.NewResourceData(t, util.ResDataParams{
	// 	ID: mock.ValidClusterID,
	// 	State: map[string]interface{}{
	// 		"name":                   "my_deployment_name",
	// 		"deployment_template_id": "aws-io-optimized-v2",
	// 		"region":                 "us-east-1",
	// 		"version":                "7.9.2",
	// 		"elasticsearch": []interface{}{
	// 			map[string]interface{}{
	// 				"topology": []interface{}{map[string]interface{}{
	// 					"id":   "hot_content",
	// 					"size": "16g",
	// 				}},
	// 			},
	// 			map[string]interface{}{
	// 				"topology": []interface{}{map[string]interface{}{
	// 					"id":   "coordinating",
	// 					"size": "16g",
	// 				}},
	// 			},
	// 		},
	// 		"kibana": []interface{}{map[string]interface{}{
	// 			"topology": []interface{}{map[string]interface{}{
	// 				"size": "2g",
	// 			}},
	// 		}},
	// 		"apm": []interface{}{map[string]interface{}{
	// 			"topology": []interface{}{map[string]interface{}{
	// 				"size": "1g",
	// 			}},
	// 		}},
	// 		"enterprise_search": []interface{}{map[string]interface{}{
	// 			"topology": []interface{}{map[string]interface{}{
	// 				"size": "8g",
	// 			}},
	// 		}},
	// 	},
	// 	Change: map[string]interface{}{
	// 		"name":                   "my_deployment_name",
	// 		"deployment_template_id": "aws-io-optimized-v2",
	// 		"region":                 "us-east-1",
	// 		"version":                "7.9.2",
	// 		"elasticsearch":          []interface{}{map[string]interface{}{}},
	// 		"kibana":                 []interface{}{map[string]interface{}{}},
	// 		"apm":                    []interface{}{map[string]interface{}{}},
	// 		"enterprise_search":      []interface{}{map[string]interface{}{}},
	// 	},
	// 	Schema: newSchema(),
	// })

	// deploymentChangeToEmptyDT := util.NewResourceData(t, util.ResDataParams{
	// 	ID: mock.ValidClusterID,
	// 	State: map[string]interface{}{
	// 		"name":                   "my_deployment_name",
	// 		"deployment_template_id": "aws-io-optimized-v2",
	// 		"region":                 "us-east-1",
	// 		"version":                "7.9.2",
	// 		"elasticsearch": []interface{}{
	// 			map[string]interface{}{
	// 				"topology": []interface{}{map[string]interface{}{
	// 					"id":   "hot_content",
	// 					"size": "16g",
	// 				}},
	// 			},
	// 			map[string]interface{}{
	// 				"topology": []interface{}{map[string]interface{}{
	// 					"id":   "coordinating",
	// 					"size": "16g",
	// 				}},
	// 			},
	// 		},
	// 		"kibana": []interface{}{map[string]interface{}{
	// 			"topology": []interface{}{map[string]interface{}{
	// 				"size": "2g",
	// 			}},
	// 		}},
	// 		"apm": []interface{}{map[string]interface{}{
	// 			"topology": []interface{}{map[string]interface{}{
	// 				"size": "1g",
	// 			}},
	// 		}},
	// 		"enterprise_search": []interface{}{map[string]interface{}{
	// 			"topology": []interface{}{map[string]interface{}{
	// 				"size": "8g",
	// 			}},
	// 		}},
	// 	},
	// 	Change: map[string]interface{}{
	// 		"name":                   "my_deployment_name",
	// 		"deployment_template_id": "empty-deployment-template",
	// 		"region":                 "us-east-1",
	// 		"version":                "7.9.2",
	// 		"elasticsearch":          []interface{}{map[string]interface{}{}},
	// 		"kibana":                 []interface{}{map[string]interface{}{}},
	// 		"apm":                    []interface{}{map[string]interface{}{}},
	// 		"enterprise_search":      []interface{}{map[string]interface{}{}},
	// 	},
	// 	Schema: newSchema(),
	// })

	// deploymentWithTags := util.NewResourceData(t, util.ResDataParams{
	// 	ID: mock.ValidClusterID,
	// 	State: map[string]interface{}{
	// 		"name":                   "my_deployment_name",
	// 		"deployment_template_id": "aws-io-optimized-v2",
	// 		"region":                 "us-east-1",
	// 		"version":                "7.10.1",
	// 		"elasticsearch": []interface{}{
	// 			map[string]interface{}{
	// 				"version": "7.10.1",
	// 				"topology": []interface{}{map[string]interface{}{
	// 					"id":   "hot_content",
	// 					"size": "8g",
	// 				}},
	// 			},
	// 		},
	// 	},
	// 	Change: map[string]interface{}{
	// 		"name":                   "my_deployment_name",
	// 		"deployment_template_id": "aws-io-optimized-v2",
	// 		"region":                 "us-east-1",
	// 		"version":                "7.10.1",
	// 		"elasticsearch": []interface{}{
	// 			map[string]interface{}{
	// 				"version": "7.10.1",
	// 				"topology": []interface{}{map[string]interface{}{
	// 					"id":   "hot_content",
	// 					"size": "8g",
	// 				}},
	// 			},
	// 		},
	// 		"tags": map[string]interface{}{
	// 			"aaa":         "bbb",
	// 			"owner":       "elastic",
	// 			"cost-center": "rnd",
	// 		},
	// 	},
	// 	Schema: newSchema(),
	// })

	type args struct {
		plan   Deployment
		state  *Deployment
		client *api.API
	}
	tests := []struct {
		name  string
		args  args
		want  *models.DeploymentUpdateRequest
		diags diag.Diagnostics
	}{
		{
			name: "parses the resources",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Alias:                "my-deployment",
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.7.0",
					Elasticsearch: &Elasticsearch{
						RefId:      ec.String("main-elasticsearch"),
						ResourceId: ec.String(mock.ValidClusterID),
						Region:     ec.String("us-east-1"),
						Config: &ElasticsearchConfig{
							UserSettingsYaml:         ec.String("some.setting: value"),
							UserSettingsOverrideYaml: ec.String("some.setting: value2"),
							UserSettingsJson:         ec.String("{\"some.setting\":\"value\"}"),
							UserSettingsOverrideJson: ec.String("{\"some.setting\":\"value2\"}"),
						},
						Topology: ElasticsearchTopologies{
							{
								Id:                      "hot_content",
								InstanceConfigurationId: ec.String("aws.data.highio.i3"),
								Size:                    ec.String("2g"),
								NodeTypeData:            ec.String("true"),
								NodeTypeIngest:          ec.String("true"),
								NodeTypeMaster:          ec.String("true"),
								NodeTypeMl:              ec.String("false"),
								ZoneCount:               1,
							},
						},
					},
					Kibana: &Kibana{
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						RefId:                     ec.String("main-kibana"),
						ResourceId:                ec.String(mock.ValidClusterID),
						Region:                    ec.String("us-east-1"),
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("aws.kibana.r5d"),
								Size:                    ec.String("1g"),
								ZoneCount:               1,
							},
						},
					},
					Apm: &Apm{
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						RefId:                     ec.String("main-apm"),
						ResourceId:                ec.String(mock.ValidClusterID),
						Region:                    ec.String("us-east-1"),
						Config: &ApmConfig{
							DebugEnabled: ec.Bool(false),
						},
						Topology: Topologies{
							{
								InstanceConfigurationId: ec.String("aws.apm.r5d"),
								Size:                    ec.String("0.5g"),
								ZoneCount:               1,
							},
						},
					},
					EnterpriseSearch: &EnterpriseSearch{
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						RefId:                     ec.String("main-enterprise_search"),
						ResourceId:                ec.String(mock.ValidClusterID),
						Region:                    ec.String("us-east-1"),
						Topology: EnterpriseSearchTopologies{
							{
								InstanceConfigurationId: ec.String("aws.enterprisesearch.m5d"),
								Size:                    ec.String("2g"),
								ZoneCount:               1,
								NodeTypeAppserver:       ec.Bool(true),
								NodeTypeConnector:       ec.Bool(true),
								NodeTypeWorker:          ec.Bool(true),
							},
						},
					},
					Observability: &Observability{
						DeploymentId: ec.String(mock.ValidClusterID),
						RefId:        ec.String("main-elasticsearch"),
						Logs:         true,
						Metrics:      true,
					},
					TrafficFilter: []string{"0.0.0.0/0", "192.168.10.0/24"},
				},
				client: api.NewMock(
					mock.New200Response(ioOptimizedTpl()),
					mock.New200Response(
						mock.NewStructBody(models.DeploymentGetResponse{
							Healthy: ec.Bool(true),
							ID:      ec.String(mock.ValidClusterID),
							Resources: &models.DeploymentResources{
								Elasticsearch: []*models.ElasticsearchResourceInfo{{
									ID:    ec.String(mock.ValidClusterID),
									RefID: ec.String("main-elasticsearch"),
								}},
							},
						}),
					),
				),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        "my-deployment",
				PruneOrphans: ec.Bool(true),
				Settings: &models.DeploymentUpdateSettings{
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
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
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
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
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
		{
			name: "parses the resources with empty declarations",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Alias:                "my-deployment",
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.7.0",
					Elasticsearch:        &Elasticsearch{},
					Kibana:               &Kibana{},
					Apm:                  &Apm{},
					EnterpriseSearch:     &EnterpriseSearch{},
					TrafficFilter:        []string{"0.0.0.0/0", "192.168.10.0/24"},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        "my-deployment",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("es-ref-id"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.7.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(8192),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("es-ref-id"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("kibana-ref-id"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("es-ref-id"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("apm-ref-id"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
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
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("es-ref-id"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("enterprise_search-ref-id"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
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
		{
			name: "parses the resources with topology overrides",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Alias:                "my-deployment",
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-io-optimized-v2",
					Region:               "us-east-1",
					Version:              "7.7.0",
					Elasticsearch: &Elasticsearch{
						RefId: ec.String("main-elasticsearch"),
						Topology: ElasticsearchTopologies{
							{
								Id:   "hot_content",
								Size: ec.String("4g"),
							},
						},
					},
					Kibana: &Kibana{
						RefId:                     ec.String("main-kibana"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Topology: Topologies{
							{
								Size: ec.String("2g"),
							},
						},
					},
					Apm: &Apm{
						RefId:                     ec.String("main-apm"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Topology: Topologies{
							{
								Size: ec.String("1g"),
							},
						},
					},
					EnterpriseSearch: &EnterpriseSearch{
						RefId:                     ec.String("main-enterprise_search"),
						ElasticsearchClusterRefId: ec.String("main-elasticsearch"),
						Topology: EnterpriseSearchTopologies{
							{
								Size: ec.String("4g"),
							},
						},
					},
					TrafficFilter: []string{"0.0.0.0/0", "192.168.10.0/24"},
				},
				client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				Alias:        "my-deployment",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("main-elasticsearch"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version: "7.7.0",
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-io-optimized-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
								ID: "hot_content",
								Elasticsearch: &models.ElasticsearchConfiguration{
									NodeAttributes: map[string]string{"data": "hot"},
								},
								ZoneCount:               2,
								InstanceConfigurationID: "aws.data.highio.i3",
								Size: &models.TopologySize{
									Resource: ec.String("memory"),
									Value:    ec.Int32(4096),
								},
								NodeType: &models.ElasticsearchNodeType{
									Data:   ec.Bool(true),
									Ingest: ec.Bool(true),
									Master: ec.Bool(true),
								},
								TopologyElementControl: &models.TopologyElementControl{
									Min: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								},
								AutoscalingMax: &models.TopologySize{
									Value:    ec.Int32(118784),
									Resource: ec.String("memory"),
								},
							}},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-kibana"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
								ClusterTopology: []*models.KibanaClusterTopologyElement{
									{
										ZoneCount:               1,
										InstanceConfigurationID: "aws.kibana.r5d",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(2048),
										},
									},
								},
							},
						},
					},
					Apm: []*models.ApmPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-apm"),
							Plan: &models.ApmPlan{
								Apm: &models.ApmConfiguration{},
								ClusterTopology: []*models.ApmTopologyElement{{
									ZoneCount:               1,
									InstanceConfigurationID: "aws.apm.r5d",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(1024),
									},
								}},
							},
						},
					},
					EnterpriseSearch: []*models.EnterpriseSearchPayload{
						{
							ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("main-enterprise_search"),
							Plan: &models.EnterpriseSearchPlan{
								EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
								ClusterTopology: []*models.EnterpriseSearchTopologyElement{
									{
										ZoneCount:               2,
										InstanceConfigurationID: "aws.enterprisesearch.m5d",
										Size: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(4096),
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
		{
			name: "parses the resources with empty declarations (Hot Warm)",
			args: args{
				plan: Deployment{
					Id:                   mock.ValidClusterID,
					Name:                 "my_deployment_name",
					DeploymentTemplateId: "aws-hot-warm-v2",
					Region:               "us-east-1",
					Version:              "7.9.2",
					Elasticsearch:        &Elasticsearch{},
					Kibana:               &Kibana{},
				},
				client: api.NewMock(mock.New200Response(hotWarmTpl())),
			},
			want: &models.DeploymentUpdateRequest{
				Name:         "my_deployment_name",
				PruneOrphans: ec.Bool(true),
				Settings:     &models.DeploymentUpdateSettings{},
				Metadata: &models.DeploymentUpdateMetadata{
					Tags: []*models.MetadataItem{},
				},
				Resources: &models.DeploymentUpdateResources{
					Elasticsearch: []*models.ElasticsearchPayload{testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, hotWarmTpl(), false), &models.ElasticsearchPayload{
						Region: ec.String("us-east-1"),
						RefID:  ec.String("es-ref-id"),
						Settings: &models.ElasticsearchClusterSettings{
							DedicatedMastersThreshold: 6,
							Curation:                  nil,
						},
						Plan: &models.ElasticsearchClusterPlan{
							AutoscalingEnabled: ec.Bool(false),
							Elasticsearch: &models.ElasticsearchConfiguration{
								Version:  "7.9.2",
								Curation: nil,
							},
							DeploymentTemplate: &models.DeploymentTemplateReference{
								ID: ec.String("aws-hot-warm-v2"),
							},
							ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
								{
									ID:                      "hot_content",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highio.i3",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeType: &models.ElasticsearchNodeType{
										Data:   ec.Bool(true),
										Ingest: ec.Bool(true),
										Master: ec.Bool(true),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "hot"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(1024),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
								{
									ID:                      "warm",
									ZoneCount:               2,
									InstanceConfigurationID: "aws.data.highstorage.d2",
									Size: &models.TopologySize{
										Resource: ec.String("memory"),
										Value:    ec.Int32(4096),
									},
									NodeType: &models.ElasticsearchNodeType{
										Data:   ec.Bool(true),
										Ingest: ec.Bool(true),
										Master: ec.Bool(false),
									},
									Elasticsearch: &models.ElasticsearchConfiguration{
										NodeAttributes: map[string]string{"data": "warm"},
									},
									TopologyElementControl: &models.TopologyElementControl{
										Min: &models.TopologySize{
											Resource: ec.String("memory"),
											Value:    ec.Int32(0),
										},
									},
									AutoscalingMax: &models.TopologySize{
										Value:    ec.Int32(118784),
										Resource: ec.String("memory"),
									},
								},
							},
						},
					})},
					Kibana: []*models.KibanaPayload{
						{
							ElasticsearchClusterRefID: ec.String("es-ref-id"),
							Region:                    ec.String("us-east-1"),
							RefID:                     ec.String("kibana-ref-id"),
							Plan: &models.KibanaClusterPlan{
								Kibana: &models.KibanaConfiguration{},
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
		// {
		// 	name: "toplogy change from hot / warm to cross cluster search",
		// 	args: args{
		// 		d:      deploymentEmptyRDWithTemplateChange,
		// 		client: api.NewMock(mock.New200Response(ccsTpl())),
		// 	},
		// 	want: &models.DeploymentUpdateRequest{
		// 		Name:         "my_deployment_name",
		// 		Alias:        "my-deployment",
		// 		PruneOrphans: ec.Bool(true),
		// 		Settings: &models.DeploymentUpdateSettings{
		// 			Observability: &models.DeploymentObservabilitySettings{},
		// 		},
		// 		Metadata: &models.DeploymentUpdateMetadata{
		// 			Tags: []*models.MetadataItem{},
		// 		},
		// 		Resources: &models.DeploymentUpdateResources{
		// 			Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ccsTpl(), false), &models.ElasticsearchPayload{
		// 				Region:   ec.String("us-east-1"),
		// 				RefID:    ec.String("main-elasticsearch"),
		// 				Settings: &models.ElasticsearchClusterSettings{},
		// 				Plan: &models.ElasticsearchClusterPlan{
		// 					Elasticsearch: &models.ElasticsearchConfiguration{
		// 						Version: "7.9.2",
		// 					},
		// 					DeploymentTemplate: &models.DeploymentTemplateReference{
		// 						ID: ec.String("aws-cross-cluster-search-v2"),
		// 					},
		// 					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
		// 						ID:                      "hot_content",
		// 						ZoneCount:               1,
		// 						InstanceConfigurationID: "aws.ccs.r5d",
		// 						Size: &models.TopologySize{
		// 							Resource: ec.String("memory"),
		// 							Value:    ec.Int32(1024),
		// 						},
		// 						NodeType: &models.ElasticsearchNodeType{
		// 							Data:   ec.Bool(true),
		// 							Ingest: ec.Bool(true),
		// 							Master: ec.Bool(true),
		// 						},
		// 						TopologyElementControl: &models.TopologyElementControl{
		// 							Min: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(1024),
		// 							},
		// 						},
		// 					}},
		// 				},
		// 			}),
		// 			Kibana: []*models.KibanaPayload{{
		// 				ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
		// 				Region:                    ec.String("us-east-1"),
		// 				RefID:                     ec.String("main-kibana"),
		// 				Plan: &models.KibanaClusterPlan{
		// 					Kibana: &models.KibanaConfiguration{},
		// 					ClusterTopology: []*models.KibanaClusterTopologyElement{
		// 						{
		// 							ZoneCount:               1,
		// 							InstanceConfigurationID: "aws.kibana.r5d",
		// 							Size: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(1024),
		// 							},
		// 						},
		// 					},
		// 				},
		// 			}},
		// 		},
		// 	},
		// },
		// // The behavior of this change should be:
		// // * Resets the Elasticsearch topology: from 16g (due to unsetTopology call on DT change).
		// // * Keeps the kibana toplogy size to 2g even though the topology element has been removed (saved value persists).
		// // * Removes all other non present resources
		// {
		// 	name: "topology change with sizes not default from io optimized to cross cluster search",
		// 	args: args{
		// 		d:      deploymentEmptyRDWithTemplateChangeWithDiffSize,
		// 		client: api.NewMock(mock.New200Response(ccsTpl())),
		// 	},
		// 	want: &models.DeploymentUpdateRequest{
		// 		Name:         "my_deployment_name",
		// 		PruneOrphans: ec.Bool(true),
		// 		Settings:     &models.DeploymentUpdateSettings{},
		// 		Metadata: &models.DeploymentUpdateMetadata{
		// 			Tags: []*models.MetadataItem{},
		// 		},
		// 		Resources: &models.DeploymentUpdateResources{
		// 			Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ccsTpl(), false), &models.ElasticsearchPayload{
		// 				Region:   ec.String("us-east-1"),
		// 				RefID:    ec.String("main-elasticsearch"),
		// 				Settings: &models.ElasticsearchClusterSettings{},
		// 				Plan: &models.ElasticsearchClusterPlan{
		// 					Elasticsearch: &models.ElasticsearchConfiguration{
		// 						Version: "7.9.2",
		// 					},
		// 					DeploymentTemplate: &models.DeploymentTemplateReference{
		// 						ID: ec.String("aws-cross-cluster-search-v2"),
		// 					},
		// 					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
		// 						ID:                      "hot_content",
		// 						ZoneCount:               1,
		// 						InstanceConfigurationID: "aws.ccs.r5d",
		// 						Size: &models.TopologySize{
		// 							Resource: ec.String("memory"),
		// 							// This field's value is reset.
		// 							Value: ec.Int32(1024),
		// 						},
		// 						NodeType: &models.ElasticsearchNodeType{
		// 							Data:   ec.Bool(true),
		// 							Ingest: ec.Bool(true),
		// 							Master: ec.Bool(true),
		// 						},
		// 						TopologyElementControl: &models.TopologyElementControl{
		// 							Min: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(1024),
		// 							},
		// 						},
		// 					}},
		// 				},
		// 			}),
		// 			Kibana: []*models.KibanaPayload{{
		// 				ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
		// 				Region:                    ec.String("us-east-1"),
		// 				RefID:                     ec.String("main-kibana"),
		// 				Plan: &models.KibanaClusterPlan{
		// 					Kibana: &models.KibanaConfiguration{},
		// 					ClusterTopology: []*models.KibanaClusterTopologyElement{
		// 						{
		// 							ZoneCount:               1,
		// 							InstanceConfigurationID: "aws.kibana.r5d",
		// 							Size: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(2048),
		// 							},
		// 						},
		// 					},
		// 				},
		// 			}},
		// 		},
		// 	},
		// },
		// // The behavior of this change should be:
		// // * Keeps all topology sizes as they were defined (saved value persists).
		// {
		// 	name: "topology change with sizes not default from explicit value to empty",
		// 	args: args{
		// 		d:      deploymentChangeFromExplicitSizingToEmpty,
		// 		client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
		// 	},
		// 	want: &models.DeploymentUpdateRequest{
		// 		Name:         "my_deployment_name",
		// 		PruneOrphans: ec.Bool(true),
		// 		Settings:     &models.DeploymentUpdateSettings{},
		// 		Metadata: &models.DeploymentUpdateMetadata{
		// 			Tags: []*models.MetadataItem{},
		// 		},
		// 		Resources: &models.DeploymentUpdateResources{
		// 			Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
		// 				Region: ec.String("us-east-1"),
		// 				RefID:  ec.String("main-elasticsearch"),
		// 				Settings: &models.ElasticsearchClusterSettings{
		// 					DedicatedMastersThreshold: 6,
		// 				},
		// 				Plan: &models.ElasticsearchClusterPlan{
		// 					AutoscalingEnabled: ec.Bool(false),
		// 					Elasticsearch: &models.ElasticsearchConfiguration{
		// 						Version: "7.9.2",
		// 					},
		// 					DeploymentTemplate: &models.DeploymentTemplateReference{
		// 						ID: ec.String("aws-io-optimized-v2"),
		// 					},
		// 					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
		// 						ID: "hot_content",
		// 						Elasticsearch: &models.ElasticsearchConfiguration{
		// 							NodeAttributes: map[string]string{"data": "hot"},
		// 						},
		// 						ZoneCount:               2,
		// 						InstanceConfigurationID: "aws.data.highio.i3",
		// 						Size: &models.TopologySize{
		// 							Resource: ec.String("memory"),
		// 							Value:    ec.Int32(16384),
		// 						},
		// 						NodeType: &models.ElasticsearchNodeType{
		// 							Data:   ec.Bool(true),
		// 							Ingest: ec.Bool(true),
		// 							Master: ec.Bool(true),
		// 						},
		// 						TopologyElementControl: &models.TopologyElementControl{
		// 							Min: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(1024),
		// 							},
		// 						},
		// 						AutoscalingMax: &models.TopologySize{
		// 							Value:    ec.Int32(118784),
		// 							Resource: ec.String("memory"),
		// 						},
		// 					},
		// 					},
		// 				},
		// 			}),
		// 			Kibana: []*models.KibanaPayload{{
		// 				ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
		// 				Region:                    ec.String("us-east-1"),
		// 				RefID:                     ec.String("main-kibana"),
		// 				Plan: &models.KibanaClusterPlan{
		// 					Kibana: &models.KibanaConfiguration{},
		// 					ClusterTopology: []*models.KibanaClusterTopologyElement{
		// 						{
		// 							ZoneCount:               1,
		// 							InstanceConfigurationID: "aws.kibana.r5d",
		// 							Size: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(2048),
		// 							},
		// 						},
		// 					},
		// 				},
		// 			}},
		// 			Apm: []*models.ApmPayload{{
		// 				ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
		// 				Region:                    ec.String("us-east-1"),
		// 				RefID:                     ec.String("main-apm"),
		// 				Plan: &models.ApmPlan{
		// 					Apm: &models.ApmConfiguration{},
		// 					ClusterTopology: []*models.ApmTopologyElement{{
		// 						ZoneCount:               1,
		// 						InstanceConfigurationID: "aws.apm.r5d",
		// 						Size: &models.TopologySize{
		// 							Resource: ec.String("memory"),
		// 							Value:    ec.Int32(1024),
		// 						},
		// 					}},
		// 				},
		// 			}},
		// 			EnterpriseSearch: []*models.EnterpriseSearchPayload{{
		// 				ElasticsearchClusterRefID: ec.String("main-elasticsearch"),
		// 				Region:                    ec.String("us-east-1"),
		// 				RefID:                     ec.String("main-enterprise_search"),
		// 				Plan: &models.EnterpriseSearchPlan{
		// 					EnterpriseSearch: &models.EnterpriseSearchConfiguration{},
		// 					ClusterTopology: []*models.EnterpriseSearchTopologyElement{
		// 						{
		// 							ZoneCount:               2,
		// 							InstanceConfigurationID: "aws.enterprisesearch.m5d",
		// 							Size: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(8192),
		// 							},
		// 							NodeType: &models.EnterpriseSearchNodeTypes{
		// 								Appserver: ec.Bool(true),
		// 								Connector: ec.Bool(true),
		// 								Worker:    ec.Bool(true),
		// 							},
		// 						},
		// 					},
		// 				},
		// 			}},
		// 		},
		// 	},
		// },
		// {
		// 	name: "does not migrate node_type to node_role on version upgrade that's lower than 7.10.0",
		// 	args: args{
		// 		d: util.NewResourceData(t, util.ResDataParams{
		// 			ID: mock.ValidClusterID,
		// 			State: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.9.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"topology": []interface{}{map[string]interface{}{
		// 						"id":               "hot_content",
		// 						"size":             "16g",
		// 						"node_type_data":   "true",
		// 						"node_type_ingest": "true",
		// 						"node_type_master": "true",
		// 						"node_type_ml":     "false",
		// 					}},
		// 				}},
		// 			},
		// 			Change: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.11.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"topology": []interface{}{map[string]interface{}{
		// 						"id":               "hot_content",
		// 						"size":             "16g",
		// 						"node_type_data":   "true",
		// 						"node_type_ingest": "true",
		// 						"node_type_master": "true",
		// 						"node_type_ml":     "false",
		// 					}},
		// 				}},
		// 			},
		// 			Schema: newSchema(),
		// 		}),
		// 		client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
		// 	},
		// 	want: &models.DeploymentUpdateRequest{
		// 		Name:         "my_deployment_name",
		// 		PruneOrphans: ec.Bool(true),
		// 		Settings:     &models.DeploymentUpdateSettings{},
		// 		Metadata: &models.DeploymentUpdateMetadata{
		// 			Tags: []*models.MetadataItem{},
		// 		},
		// 		Resources: &models.DeploymentUpdateResources{
		// 			Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
		// 				Region: ec.String("us-east-1"),
		// 				RefID:  ec.String("main-elasticsearch"),
		// 				Settings: &models.ElasticsearchClusterSettings{
		// 					DedicatedMastersThreshold: 6,
		// 				},
		// 				Plan: &models.ElasticsearchClusterPlan{
		// 					AutoscalingEnabled: ec.Bool(false),
		// 					Elasticsearch: &models.ElasticsearchConfiguration{
		// 						Version: "7.11.1",
		// 					},
		// 					DeploymentTemplate: &models.DeploymentTemplateReference{
		// 						ID: ec.String("aws-io-optimized-v2"),
		// 					},
		// 					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
		// 						ID: "hot_content",
		// 						Elasticsearch: &models.ElasticsearchConfiguration{
		// 							NodeAttributes: map[string]string{"data": "hot"},
		// 						},
		// 						ZoneCount:               2,
		// 						InstanceConfigurationID: "aws.data.highio.i3",
		// 						Size: &models.TopologySize{
		// 							Resource: ec.String("memory"),
		// 							Value:    ec.Int32(16384),
		// 						},
		// 						NodeType: &models.ElasticsearchNodeType{
		// 							Data:   ec.Bool(true),
		// 							Ingest: ec.Bool(true),
		// 							Master: ec.Bool(true),
		// 							Ml:     ec.Bool(false),
		// 						},
		// 						TopologyElementControl: &models.TopologyElementControl{
		// 							Min: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(1024),
		// 							},
		// 						},
		// 						AutoscalingMax: &models.TopologySize{
		// 							Value:    ec.Int32(118784),
		// 							Resource: ec.String("memory"),
		// 						},
		// 					}},
		// 				},
		// 			}),
		// 		},
		// 	},
		// },
		// {
		// 	name: "does not migrate node_type to node_role on version upgrade that's higher than 7.10.0",
		// 	args: args{
		// 		d: util.NewResourceData(t, util.ResDataParams{
		// 			ID: mock.ValidClusterID,
		// 			State: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.10.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"topology": []interface{}{map[string]interface{}{
		// 						"id":               "hot_content",
		// 						"size":             "16g",
		// 						"node_type_data":   "true",
		// 						"node_type_ingest": "true",
		// 						"node_type_master": "true",
		// 						"node_type_ml":     "false",
		// 					}},
		// 				}},
		// 			},
		// 			Change: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.11.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"topology": []interface{}{map[string]interface{}{
		// 						"id":               "hot_content",
		// 						"size":             "16g",
		// 						"node_type_data":   "true",
		// 						"node_type_ingest": "true",
		// 						"node_type_master": "true",
		// 						"node_type_ml":     "false",
		// 					}},
		// 				}},
		// 			},
		// 			Schema: newSchema(),
		// 		}),
		// 		client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
		// 	},
		// 	want: &models.DeploymentUpdateRequest{
		// 		Name:         "my_deployment_name",
		// 		PruneOrphans: ec.Bool(true),
		// 		Settings:     &models.DeploymentUpdateSettings{},
		// 		Metadata: &models.DeploymentUpdateMetadata{
		// 			Tags: []*models.MetadataItem{},
		// 		},
		// 		Resources: &models.DeploymentUpdateResources{
		// 			Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), false), &models.ElasticsearchPayload{
		// 				Region: ec.String("us-east-1"),
		// 				RefID:  ec.String("main-elasticsearch"),
		// 				Settings: &models.ElasticsearchClusterSettings{
		// 					DedicatedMastersThreshold: 6,
		// 				},
		// 				Plan: &models.ElasticsearchClusterPlan{
		// 					AutoscalingEnabled: ec.Bool(false),
		// 					Elasticsearch: &models.ElasticsearchConfiguration{
		// 						Version: "7.11.1",
		// 					},
		// 					DeploymentTemplate: &models.DeploymentTemplateReference{
		// 						ID: ec.String("aws-io-optimized-v2"),
		// 					},
		// 					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
		// 						{
		// 							ID: "hot_content",
		// 							Elasticsearch: &models.ElasticsearchConfiguration{
		// 								NodeAttributes: map[string]string{"data": "hot"},
		// 							},
		// 							ZoneCount:               2,
		// 							InstanceConfigurationID: "aws.data.highio.i3",
		// 							Size: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(16384),
		// 							},
		// 							NodeType: &models.ElasticsearchNodeType{
		// 								Data:   ec.Bool(true),
		// 								Ingest: ec.Bool(true),
		// 								Master: ec.Bool(true),
		// 								Ml:     ec.Bool(false),
		// 							},
		// 							TopologyElementControl: &models.TopologyElementControl{
		// 								Min: &models.TopologySize{
		// 									Resource: ec.String("memory"),
		// 									Value:    ec.Int32(1024),
		// 								},
		// 							},
		// 							AutoscalingMax: &models.TopologySize{
		// 								Value:    ec.Int32(118784),
		// 								Resource: ec.String("memory"),
		// 							},
		// 						},
		// 					},
		// 				},
		// 			}),
		// 		},
		// 	},
		// },
		// {
		// 	name: "migrates node_type to node_role when the existing topology element size is updated",
		// 	args: args{
		// 		d: util.NewResourceData(t, util.ResDataParams{
		// 			ID: mock.ValidClusterID,
		// 			State: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.10.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"topology": []interface{}{map[string]interface{}{
		// 						"id":               "hot_content",
		// 						"size":             "16g",
		// 						"node_type_data":   "true",
		// 						"node_type_ingest": "true",
		// 						"node_type_master": "true",
		// 						"node_type_ml":     "false",
		// 					}},
		// 				}},
		// 			},
		// 			Change: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.10.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"topology": []interface{}{map[string]interface{}{
		// 						"id":               "hot_content",
		// 						"size":             "32g",
		// 						"node_type_data":   "true",
		// 						"node_type_ingest": "true",
		// 						"node_type_master": "true",
		// 						"node_type_ml":     "false",
		// 					}},
		// 				}},
		// 			},
		// 			Schema: newSchema(),
		// 		}),
		// 		client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
		// 	},
		// 	want: &models.DeploymentUpdateRequest{
		// 		Name:         "my_deployment_name",
		// 		PruneOrphans: ec.Bool(true),
		// 		Settings:     &models.DeploymentUpdateSettings{},
		// 		Metadata: &models.DeploymentUpdateMetadata{
		// 			Tags: []*models.MetadataItem{},
		// 		},
		// 		Resources: &models.DeploymentUpdateResources{
		// 			Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
		// 				Region: ec.String("us-east-1"),
		// 				RefID:  ec.String("main-elasticsearch"),
		// 				Settings: &models.ElasticsearchClusterSettings{
		// 					DedicatedMastersThreshold: 6,
		// 				},
		// 				Plan: &models.ElasticsearchClusterPlan{
		// 					AutoscalingEnabled: ec.Bool(false),
		// 					Elasticsearch: &models.ElasticsearchConfiguration{
		// 						Version: "7.10.1",
		// 					},
		// 					DeploymentTemplate: &models.DeploymentTemplateReference{
		// 						ID: ec.String("aws-io-optimized-v2"),
		// 					},
		// 					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
		// 						ID: "hot_content",
		// 						Elasticsearch: &models.ElasticsearchConfiguration{
		// 							NodeAttributes: map[string]string{"data": "hot"},
		// 						},
		// 						ZoneCount:               2,
		// 						InstanceConfigurationID: "aws.data.highio.i3",
		// 						Size: &models.TopologySize{
		// 							Resource: ec.String("memory"),
		// 							Value:    ec.Int32(32768),
		// 						},
		// 						NodeRoles: []string{
		// 							"master",
		// 							"ingest",
		// 							"remote_cluster_client",
		// 							"data_hot",
		// 							"transform",
		// 							"data_content",
		// 						},
		// 						TopologyElementControl: &models.TopologyElementControl{
		// 							Min: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(1024),
		// 							},
		// 						},
		// 						AutoscalingMax: &models.TopologySize{
		// 							Value:    ec.Int32(118784),
		// 							Resource: ec.String("memory"),
		// 						},
		// 					}},
		// 				},
		// 			}),
		// 		},
		// 	},
		// },
		// {
		// 	name: "migrates node_type to node_role when the existing topology element size is updated and adds warm tier",
		// 	args: args{
		// 		d: util.NewResourceData(t, util.ResDataParams{
		// 			ID: mock.ValidClusterID,
		// 			State: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.10.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"topology": []interface{}{map[string]interface{}{
		// 						"id":               "hot_content",
		// 						"size":             "16g",
		// 						"node_type_data":   "true",
		// 						"node_type_ingest": "true",
		// 						"node_type_master": "true",
		// 						"node_type_ml":     "false",
		// 					}},
		// 				}},
		// 			},
		// 			Change: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.10.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"topology": []interface{}{
		// 						map[string]interface{}{
		// 							"id":               "hot_content",
		// 							"size":             "16g",
		// 							"node_type_data":   "true",
		// 							"node_type_ingest": "true",
		// 							"node_type_master": "true",
		// 							"node_type_ml":     "false",
		// 						},
		// 						map[string]interface{}{
		// 							"id":   "warm",
		// 							"size": "8g",
		// 						},
		// 					},
		// 				}},
		// 			},
		// 			Schema: newSchema(),
		// 		}),
		// 		client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
		// 	},
		// 	want: &models.DeploymentUpdateRequest{
		// 		Name:         "my_deployment_name",
		// 		PruneOrphans: ec.Bool(true),
		// 		Settings:     &models.DeploymentUpdateSettings{},
		// 		Metadata: &models.DeploymentUpdateMetadata{
		// 			Tags: []*models.MetadataItem{},
		// 		},
		// 		Resources: &models.DeploymentUpdateResources{
		// 			Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
		// 				Region: ec.String("us-east-1"),
		// 				RefID:  ec.String("main-elasticsearch"),
		// 				Settings: &models.ElasticsearchClusterSettings{
		// 					DedicatedMastersThreshold: 6,
		// 				},
		// 				Plan: &models.ElasticsearchClusterPlan{
		// 					AutoscalingEnabled: ec.Bool(false),
		// 					Elasticsearch: &models.ElasticsearchConfiguration{
		// 						Version: "7.10.1",
		// 					},
		// 					DeploymentTemplate: &models.DeploymentTemplateReference{
		// 						ID: ec.String("aws-io-optimized-v2"),
		// 					},
		// 					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
		// 						{
		// 							ID: "hot_content",
		// 							Elasticsearch: &models.ElasticsearchConfiguration{
		// 								NodeAttributes: map[string]string{"data": "hot"},
		// 							},
		// 							ZoneCount:               2,
		// 							InstanceConfigurationID: "aws.data.highio.i3",
		// 							Size: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(16384),
		// 							},
		// 							NodeRoles: []string{
		// 								"master",
		// 								"ingest",
		// 								"remote_cluster_client",
		// 								"data_hot",
		// 								"transform",
		// 								"data_content",
		// 							},
		// 							TopologyElementControl: &models.TopologyElementControl{
		// 								Min: &models.TopologySize{
		// 									Resource: ec.String("memory"),
		// 									Value:    ec.Int32(1024),
		// 								},
		// 							},
		// 							AutoscalingMax: &models.TopologySize{
		// 								Value:    ec.Int32(118784),
		// 								Resource: ec.String("memory"),
		// 							},
		// 						},
		// 						{
		// 							ID: "warm",
		// 							Elasticsearch: &models.ElasticsearchConfiguration{
		// 								NodeAttributes: map[string]string{"data": "warm"},
		// 							},
		// 							ZoneCount:               2,
		// 							InstanceConfigurationID: "aws.data.highstorage.d3",
		// 							Size: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(8192),
		// 							},
		// 							NodeRoles: []string{
		// 								"data_warm",
		// 								"remote_cluster_client",
		// 							},
		// 							TopologyElementControl: &models.TopologyElementControl{
		// 								Min: &models.TopologySize{
		// 									Resource: ec.String("memory"),
		// 									Value:    ec.Int32(0),
		// 								},
		// 							},
		// 							AutoscalingMax: &models.TopologySize{
		// 								Value:    ec.Int32(118784),
		// 								Resource: ec.String("memory"),
		// 							},
		// 						},
		// 					},
		// 				},
		// 			}),
		// 		},
		// 	},
		// },
		// {
		// 	name: "enables autoscaling with the default policies",
		// 	args: args{
		// 		d: util.NewResourceData(t, util.ResDataParams{
		// 			ID: mock.ValidClusterID,
		// 			State: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.12.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"topology": []interface{}{map[string]interface{}{
		// 						"id":   "hot_content",
		// 						"size": "16g",
		// 					}},
		// 				}},
		// 			},
		// 			Change: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.12.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"autoscale": "true",
		// 					"topology": []interface{}{
		// 						map[string]interface{}{
		// 							"id":   "hot_content",
		// 							"size": "16g",
		// 						},
		// 						map[string]interface{}{
		// 							"id":   "warm",
		// 							"size": "8g",
		// 						},
		// 					},
		// 				}},
		// 			},
		// 			Schema: newSchema(),
		// 		}),
		// 		client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
		// 	},
		// 	want: &models.DeploymentUpdateRequest{
		// 		Name:         "my_deployment_name",
		// 		PruneOrphans: ec.Bool(true),
		// 		Settings:     &models.DeploymentUpdateSettings{},
		// 		Metadata: &models.DeploymentUpdateMetadata{
		// 			Tags: []*models.MetadataItem{},
		// 		},
		// 		Resources: &models.DeploymentUpdateResources{
		// 			Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
		// 				Region: ec.String("us-east-1"),
		// 				RefID:  ec.String("main-elasticsearch"),
		// 				Settings: &models.ElasticsearchClusterSettings{
		// 					DedicatedMastersThreshold: 6,
		// 				},
		// 				Plan: &models.ElasticsearchClusterPlan{
		// 					AutoscalingEnabled: ec.Bool(true),
		// 					Elasticsearch: &models.ElasticsearchConfiguration{
		// 						Version: "7.12.1",
		// 					},
		// 					DeploymentTemplate: &models.DeploymentTemplateReference{
		// 						ID: ec.String("aws-io-optimized-v2"),
		// 					},
		// 					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
		// 						{
		// 							ID: "hot_content",
		// 							Elasticsearch: &models.ElasticsearchConfiguration{
		// 								NodeAttributes: map[string]string{"data": "hot"},
		// 							},
		// 							ZoneCount:               2,
		// 							InstanceConfigurationID: "aws.data.highio.i3",
		// 							Size: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(16384),
		// 							},
		// 							NodeRoles: []string{
		// 								"master",
		// 								"ingest",
		// 								"remote_cluster_client",
		// 								"data_hot",
		// 								"transform",
		// 								"data_content",
		// 							},
		// 							TopologyElementControl: &models.TopologyElementControl{
		// 								Min: &models.TopologySize{
		// 									Resource: ec.String("memory"),
		// 									Value:    ec.Int32(1024),
		// 								},
		// 							},
		// 							AutoscalingMax: &models.TopologySize{
		// 								Value:    ec.Int32(118784),
		// 								Resource: ec.String("memory"),
		// 							},
		// 						},
		// 						{
		// 							ID: "warm",
		// 							Elasticsearch: &models.ElasticsearchConfiguration{
		// 								NodeAttributes: map[string]string{"data": "warm"},
		// 							},
		// 							ZoneCount:               2,
		// 							InstanceConfigurationID: "aws.data.highstorage.d3",
		// 							Size: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(8192),
		// 							},
		// 							NodeRoles: []string{
		// 								"data_warm",
		// 								"remote_cluster_client",
		// 							},
		// 							TopologyElementControl: &models.TopologyElementControl{
		// 								Min: &models.TopologySize{
		// 									Resource: ec.String("memory"),
		// 									Value:    ec.Int32(0),
		// 								},
		// 							},
		// 							AutoscalingMax: &models.TopologySize{
		// 								Value:    ec.Int32(118784),
		// 								Resource: ec.String("memory"),
		// 							},
		// 						},
		// 					},
		// 				},
		// 			}),
		// 		},
		// 	},
		// },
		// {
		// 	name: "updates topologies configuration",
		// 	args: args{
		// 		d: util.NewResourceData(t, util.ResDataParams{
		// 			ID: mock.ValidClusterID,
		// 			State: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.12.1",
		// 				"elasticsearch": []interface{}{
		// 					map[string]interface{}{
		// 						"topology": []interface{}{map[string]interface{}{
		// 							"id":         "hot_content",
		// 							"size":       "16g",
		// 							"zone_count": 3,
		// 							"config": []interface{}{map[string]interface{}{
		// 								"user_settings_yaml": "setting: true",
		// 							}},
		// 						}},
		// 					},
		// 					map[string]interface{}{
		// 						"topology": []interface{}{map[string]interface{}{
		// 							"id":         "master",
		// 							"size":       "1g",
		// 							"zone_count": 3,
		// 							"config": []interface{}{map[string]interface{}{
		// 								"user_settings_yaml": "setting: true",
		// 							}},
		// 						}},
		// 					},
		// 					map[string]interface{}{
		// 						"topology": []interface{}{map[string]interface{}{
		// 							"id":         "warm",
		// 							"size":       "8g",
		// 							"zone_count": 3,
		// 							"config": []interface{}{map[string]interface{}{
		// 								"user_settings_yaml": "setting: true",
		// 							}},
		// 						}},
		// 					},
		// 				},
		// 			},
		// 			Change: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.12.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"topology": []interface{}{
		// 						map[string]interface{}{
		// 							"id":         "hot_content",
		// 							"size":       "16g",
		// 							"zone_count": 3,
		// 							"config": []interface{}{map[string]interface{}{
		// 								"user_settings_yaml": "setting: false",
		// 							}},
		// 						},
		// 						map[string]interface{}{
		// 							"id":         "master",
		// 							"size":       "1g",
		// 							"zone_count": 3,
		// 							"config": []interface{}{map[string]interface{}{
		// 								"user_settings_yaml": "setting: false",
		// 							}},
		// 						},
		// 						map[string]interface{}{
		// 							"id":         "warm",
		// 							"size":       "8g",
		// 							"zone_count": 3,
		// 							"config": []interface{}{map[string]interface{}{
		// 								"user_settings_yaml": "setting: false",
		// 							}},
		// 						},
		// 					},
		// 				}},
		// 			},
		// 			Schema: newSchema(),
		// 		}),
		// 		client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
		// 	},
		// 	want: &models.DeploymentUpdateRequest{
		// 		Name:         "my_deployment_name",
		// 		PruneOrphans: ec.Bool(true),
		// 		Settings:     &models.DeploymentUpdateSettings{},
		// 		Metadata: &models.DeploymentUpdateMetadata{
		// 			Tags: []*models.MetadataItem{},
		// 		},
		// 		Resources: &models.DeploymentUpdateResources{
		// 			Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
		// 				Region: ec.String("us-east-1"),
		// 				RefID:  ec.String("main-elasticsearch"),
		// 				Settings: &models.ElasticsearchClusterSettings{
		// 					DedicatedMastersThreshold: 6,
		// 				},
		// 				Plan: &models.ElasticsearchClusterPlan{
		// 					AutoscalingEnabled: ec.Bool(false),
		// 					Elasticsearch: &models.ElasticsearchConfiguration{
		// 						Version: "7.12.1",
		// 					},
		// 					DeploymentTemplate: &models.DeploymentTemplateReference{
		// 						ID: ec.String("aws-io-optimized-v2"),
		// 					},
		// 					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{
		// 						{
		// 							ID: "hot_content",
		// 							Elasticsearch: &models.ElasticsearchConfiguration{
		// 								NodeAttributes:   map[string]string{"data": "hot"},
		// 								UserSettingsYaml: "setting: false",
		// 							},
		// 							ZoneCount:               3,
		// 							InstanceConfigurationID: "aws.data.highio.i3",
		// 							Size: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(16384),
		// 							},
		// 							NodeRoles: []string{
		// 								"ingest",
		// 								"remote_cluster_client",
		// 								"data_hot",
		// 								"transform",
		// 								"data_content",
		// 							},
		// 							TopologyElementControl: &models.TopologyElementControl{
		// 								Min: &models.TopologySize{
		// 									Resource: ec.String("memory"),
		// 									Value:    ec.Int32(1024),
		// 								},
		// 							},
		// 							AutoscalingMax: &models.TopologySize{
		// 								Value:    ec.Int32(118784),
		// 								Resource: ec.String("memory"),
		// 							},
		// 						},
		// 						{
		// 							ID: "master",
		// 							Elasticsearch: &models.ElasticsearchConfiguration{
		// 								UserSettingsYaml: "setting: false",
		// 							},
		// 							ZoneCount:               3,
		// 							InstanceConfigurationID: "aws.master.r5d",
		// 							Size: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(1024),
		// 							},
		// 							NodeRoles: []string{
		// 								"master",
		// 								"remote_cluster_client",
		// 							},
		// 							TopologyElementControl: &models.TopologyElementControl{
		// 								Min: &models.TopologySize{
		// 									Resource: ec.String("memory"),
		// 									Value:    ec.Int32(0),
		// 								},
		// 							},
		// 						},
		// 						{
		// 							ID: "warm",
		// 							Elasticsearch: &models.ElasticsearchConfiguration{
		// 								NodeAttributes:   map[string]string{"data": "warm"},
		// 								UserSettingsYaml: "setting: false",
		// 							},
		// 							ZoneCount:               3,
		// 							InstanceConfigurationID: "aws.data.highstorage.d3",
		// 							Size: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(8192),
		// 							},
		// 							NodeRoles: []string{
		// 								"data_warm",
		// 								"remote_cluster_client",
		// 							},
		// 							TopologyElementControl: &models.TopologyElementControl{
		// 								Min: &models.TopologySize{
		// 									Resource: ec.String("memory"),
		// 									Value:    ec.Int32(0),
		// 								},
		// 							},
		// 							AutoscalingMax: &models.TopologySize{
		// 								Value:    ec.Int32(118784),
		// 								Resource: ec.String("memory"),
		// 							},
		// 						},
		// 					},
		// 				},
		// 			}),
		// 		},
		// 	},
		// },
		// {
		// 	name: "parses the resources with tags",
		// 	args: args{
		// 		d:      deploymentWithTags,
		// 		client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
		// 	},
		// 	want: &models.DeploymentUpdateRequest{
		// 		Name:         "my_deployment_name",
		// 		PruneOrphans: ec.Bool(true),
		// 		Settings:     &models.DeploymentUpdateSettings{},
		// 		Metadata: &models.DeploymentUpdateMetadata{Tags: []*models.MetadataItem{
		// 			{Key: ec.String("aaa"), Value: ec.String("bbb")},
		// 			{Key: ec.String("cost-center"), Value: ec.String("rnd")},
		// 			{Key: ec.String("owner"), Value: ec.String("elastic")},
		// 		}},
		// 		Resources: &models.DeploymentUpdateResources{
		// 			Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
		// 				Region: ec.String("us-east-1"),
		// 				RefID:  ec.String("main-elasticsearch"),
		// 				Settings: &models.ElasticsearchClusterSettings{
		// 					DedicatedMastersThreshold: 6,
		// 				},
		// 				Plan: &models.ElasticsearchClusterPlan{
		// 					AutoscalingEnabled: ec.Bool(false),
		// 					Elasticsearch: &models.ElasticsearchConfiguration{
		// 						Version: "7.10.1",
		// 					},
		// 					DeploymentTemplate: &models.DeploymentTemplateReference{
		// 						ID: ec.String("aws-io-optimized-v2"),
		// 					},
		// 					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
		// 						ID: "hot_content",
		// 						Elasticsearch: &models.ElasticsearchConfiguration{
		// 							NodeAttributes: map[string]string{"data": "hot"},
		// 						},
		// 						ZoneCount:               2,
		// 						InstanceConfigurationID: "aws.data.highio.i3",
		// 						Size: &models.TopologySize{
		// 							Resource: ec.String("memory"),
		// 							Value:    ec.Int32(8192),
		// 						},
		// 						NodeRoles: []string{
		// 							"master",
		// 							"ingest",
		// 							"remote_cluster_client",
		// 							"data_hot",
		// 							"transform",
		// 							"data_content",
		// 						},
		// 						TopologyElementControl: &models.TopologyElementControl{
		// 							Min: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(1024),
		// 							},
		// 						},
		// 						AutoscalingMax: &models.TopologySize{
		// 							Value:    ec.Int32(118784),
		// 							Resource: ec.String("memory"),
		// 						},
		// 					}},
		// 				},
		// 			}),
		// 		},
		// 	},
		// },
		// {
		// 	name: "handles a snapshot_source block adding Strategy: partial",
		// 	args: args{
		// 		d: util.NewResourceData(t, util.ResDataParams{
		// 			ID: mock.ValidClusterID,
		// 			State: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.10.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"version": "7.10.1",
		// 					"topology": []interface{}{map[string]interface{}{
		// 						"id":   "hot_content",
		// 						"size": "8g",
		// 					}},
		// 				}},
		// 			},
		// 			Change: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.10.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"version": "7.10.1",
		// 					"snapshot_source": []interface{}{map[string]interface{}{
		// 						"source_elasticsearch_cluster_id": "8c63b87af9e24ea49b8a4bfe550e5fe9",
		// 					}},
		// 					"topology": []interface{}{map[string]interface{}{
		// 						"id":   "hot_content",
		// 						"size": "8g",
		// 					}},
		// 				}},
		// 			},
		// 			Schema: newSchema(),
		// 		}),
		// 		client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
		// 	},
		// 	want: &models.DeploymentUpdateRequest{
		// 		Name:         "my_deployment_name",
		// 		PruneOrphans: ec.Bool(true),
		// 		Settings:     &models.DeploymentUpdateSettings{},
		// 		Metadata: &models.DeploymentUpdateMetadata{
		// 			Tags: []*models.MetadataItem{},
		// 		},
		// 		Resources: &models.DeploymentUpdateResources{
		// 			Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
		// 				Region: ec.String("us-east-1"),
		// 				RefID:  ec.String("main-elasticsearch"),
		// 				Settings: &models.ElasticsearchClusterSettings{
		// 					DedicatedMastersThreshold: 6,
		// 				},
		// 				Plan: &models.ElasticsearchClusterPlan{
		// 					Transient: &models.TransientElasticsearchPlanConfiguration{
		// 						RestoreSnapshot: &models.RestoreSnapshotConfiguration{
		// 							SourceClusterID: "8c63b87af9e24ea49b8a4bfe550e5fe9",
		// 							SnapshotName:    ec.String("__latest_success__"),
		// 							Strategy:        "partial",
		// 						},
		// 					},
		// 					AutoscalingEnabled: ec.Bool(false),
		// 					Elasticsearch: &models.ElasticsearchConfiguration{
		// 						Version: "7.10.1",
		// 					},
		// 					DeploymentTemplate: &models.DeploymentTemplateReference{
		// 						ID: ec.String("aws-io-optimized-v2"),
		// 					},
		// 					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
		// 						ID: "hot_content",
		// 						Elasticsearch: &models.ElasticsearchConfiguration{
		// 							NodeAttributes: map[string]string{"data": "hot"},
		// 						},
		// 						ZoneCount:               2,
		// 						InstanceConfigurationID: "aws.data.highio.i3",
		// 						Size: &models.TopologySize{
		// 							Resource: ec.String("memory"),
		// 							Value:    ec.Int32(8192),
		// 						},
		// 						NodeRoles: []string{
		// 							"master",
		// 							"ingest",
		// 							"remote_cluster_client",
		// 							"data_hot",
		// 							"transform",
		// 							"data_content",
		// 						},
		// 						TopologyElementControl: &models.TopologyElementControl{
		// 							Min: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(1024),
		// 							},
		// 						},
		// 						AutoscalingMax: &models.TopologySize{
		// 							Value:    ec.Int32(118784),
		// 							Resource: ec.String("memory"),
		// 						},
		// 					}},
		// 				},
		// 			}),
		// 		},
		// 	},
		// },
		// {
		// 	name: "handles empty Elasticsearch empty config block",
		// 	args: args{
		// 		d: util.NewResourceData(t, util.ResDataParams{
		// 			ID: mock.ValidClusterID,
		// 			State: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.10.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"version": "7.10.1",
		// 					"topology": []interface{}{map[string]interface{}{
		// 						"id":     "hot_content",
		// 						"size":   "8g",
		// 						"config": []interface{}{},
		// 					}},
		// 				}},
		// 			},
		// 			Schema: newSchema(),
		// 		}),
		// 		client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
		// 	},
		// 	want: &models.DeploymentUpdateRequest{
		// 		Name:         "my_deployment_name",
		// 		PruneOrphans: ec.Bool(true),
		// 		Settings:     &models.DeploymentUpdateSettings{},
		// 		Metadata: &models.DeploymentUpdateMetadata{
		// 			Tags: []*models.MetadataItem{},
		// 		},
		// 		Resources: &models.DeploymentUpdateResources{
		// 			Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
		// 				Region: ec.String("us-east-1"),
		// 				RefID:  ec.String("main-elasticsearch"),
		// 				Settings: &models.ElasticsearchClusterSettings{
		// 					DedicatedMastersThreshold: 6,
		// 				},
		// 				Plan: &models.ElasticsearchClusterPlan{
		// 					AutoscalingEnabled: ec.Bool(false),
		// 					Elasticsearch: &models.ElasticsearchConfiguration{
		// 						Version: "7.10.1",
		// 					},
		// 					DeploymentTemplate: &models.DeploymentTemplateReference{
		// 						ID: ec.String("aws-io-optimized-v2"),
		// 					},
		// 					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
		// 						ID: "hot_content",
		// 						Elasticsearch: &models.ElasticsearchConfiguration{
		// 							NodeAttributes: map[string]string{"data": "hot"},
		// 						},
		// 						ZoneCount:               2,
		// 						InstanceConfigurationID: "aws.data.highio.i3",
		// 						Size: &models.TopologySize{
		// 							Resource: ec.String("memory"),
		// 							Value:    ec.Int32(8192),
		// 						},
		// 						NodeRoles: []string{
		// 							"master",
		// 							"ingest",
		// 							"remote_cluster_client",
		// 							"data_hot",
		// 							"transform",
		// 							"data_content",
		// 						},
		// 						TopologyElementControl: &models.TopologyElementControl{
		// 							Min: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(1024),
		// 							},
		// 						},
		// 						AutoscalingMax: &models.TopologySize{
		// 							Value:    ec.Int32(118784),
		// 							Resource: ec.String("memory"),
		// 						},
		// 					}},
		// 				},
		// 			}),
		// 		},
		// 	},
		// },
		// {
		// 	name: "handles Elasticsearch with topology.config block",
		// 	args: args{
		// 		d: util.NewResourceData(t, util.ResDataParams{
		// 			ID: mock.ValidClusterID,
		// 			State: map[string]interface{}{
		// 				"name":                   "my_deployment_name",
		// 				"deployment_template_id": "aws-io-optimized-v2",
		// 				"region":                 "us-east-1",
		// 				"version":                "7.10.1",
		// 				"elasticsearch": []interface{}{map[string]interface{}{
		// 					"version": "7.10.1",
		// 					"config":  []interface{}{},
		// 					"topology": []interface{}{map[string]interface{}{
		// 						"id":   "hot_content",
		// 						"size": "8g",
		// 						"config": []interface{}{map[string]interface{}{
		// 							"user_settings_yaml": "setting: true",
		// 						}},
		// 					}},
		// 				}},
		// 			},
		// 			Schema: newSchema(),
		// 		}),
		// 		client: api.NewMock(mock.New200Response(ioOptimizedTpl())),
		// 	},
		// 	want: &models.DeploymentUpdateRequest{
		// 		Name:         "my_deployment_name",
		// 		PruneOrphans: ec.Bool(true),
		// 		Settings:     &models.DeploymentUpdateSettings{},
		// 		Metadata: &models.DeploymentUpdateMetadata{
		// 			Tags: []*models.MetadataItem{},
		// 		},
		// 		Resources: &models.DeploymentUpdateResources{
		// 			Elasticsearch: enrichWithEmptyTopologies(readerToESPayload(t, ioOptimizedTpl(), true), &models.ElasticsearchPayload{
		// 				Region: ec.String("us-east-1"),
		// 				RefID:  ec.String("main-elasticsearch"),
		// 				Settings: &models.ElasticsearchClusterSettings{
		// 					DedicatedMastersThreshold: 6,
		// 				},
		// 				Plan: &models.ElasticsearchClusterPlan{
		// 					AutoscalingEnabled: ec.Bool(false),
		// 					Elasticsearch: &models.ElasticsearchConfiguration{
		// 						Version: "7.10.1",
		// 					},
		// 					DeploymentTemplate: &models.DeploymentTemplateReference{
		// 						ID: ec.String("aws-io-optimized-v2"),
		// 					},
		// 					ClusterTopology: []*models.ElasticsearchClusterTopologyElement{{
		// 						ID: "hot_content",
		// 						Elasticsearch: &models.ElasticsearchConfiguration{
		// 							NodeAttributes:   map[string]string{"data": "hot"},
		// 							UserSettingsYaml: "setting: true",
		// 						},
		// 						ZoneCount:               2,
		// 						InstanceConfigurationID: "aws.data.highio.i3",
		// 						Size: &models.TopologySize{
		// 							Resource: ec.String("memory"),
		// 							Value:    ec.Int32(8192),
		// 						},
		// 						NodeRoles: []string{
		// 							"master",
		// 							"ingest",
		// 							"remote_cluster_client",
		// 							"data_hot",
		// 							"transform",
		// 							"data_content",
		// 						},
		// 						TopologyElementControl: &models.TopologyElementControl{
		// 							Min: &models.TopologySize{
		// 								Resource: ec.String("memory"),
		// 								Value:    ec.Int32(1024),
		// 							},
		// 						},
		// 						AutoscalingMax: &models.TopologySize{
		// 							Value:    ec.Int32(118784),
		// 							Resource: ec.String("memory"),
		// 						},
		// 					}},
		// 				},
		// 			}),
		// 		},
		// 	},
		// },
		// {
		// 	name: "topology change with invalid resources returns an error",
		// 	args: args{
		// 		d:      deploymentChangeToEmptyDT,
		// 		client: api.NewMock(mock.New200Response(emptyTpl())),
		// 	},
		// 	err: multierror.NewPrefixed("invalid configuration",
		// 		errors.New("kibana specified but deployment template is not configured for it. Use a different template if you wish to add kibana"),
		// 		errors.New("apm specified but deployment template is not configured for it. Use a different template if you wish to add apm"),
		// 		errors.New("enterprise_search specified but deployment template is not configured for it. Use a different template if you wish to add enterprise_search"),
		// 	),
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &Resource{}
			schema, diags := resource.GetSchema(context.Background())
			assert.Nil(t, diags)

			var plan DeploymentTF
			diags = tfsdk.ValueFrom(context.Background(), &tt.args.plan, schema.Type(), &plan)
			assert.Nil(t, diags)

			state := tt.args.state
			if state == nil {
				state = &tt.args.plan
			}

			var stateTF DeploymentTF

			diags = tfsdk.ValueFrom(context.Background(), state, schema.Type(), &stateTF)
			assert.Nil(t, diags)

			got, diags := plan.UpdateRequest(context.Background(), tt.args.client, stateTF)
			if tt.diags != nil {
				assert.Equal(t, diags, tt.diags)
			} else {
				assert.Nil(t, diags)
				assert.NotNil(t, got)
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

func Test_ensurePartialSnapshotStrategy(t *testing.T) {
	type args struct {
		es *models.ElasticsearchPayload
	}
	tests := []struct {
		name string
		args args
		want *models.ElasticsearchPayload
	}{
		{
			name: "ignores resources with no transient block",
			args: args{es: &models.ElasticsearchPayload{
				Plan: &models.ElasticsearchClusterPlan{},
			}},
			want: &models.ElasticsearchPayload{
				Plan: &models.ElasticsearchClusterPlan{},
			},
		},
		{
			name: "ignores resources with no transient.snapshot block",
			args: args{es: &models.ElasticsearchPayload{
				Plan: &models.ElasticsearchClusterPlan{
					Transient: &models.TransientElasticsearchPlanConfiguration{},
				},
			}},
			want: &models.ElasticsearchPayload{
				Plan: &models.ElasticsearchClusterPlan{
					Transient: &models.TransientElasticsearchPlanConfiguration{},
				},
			},
		},
		{
			name: "Sets strategy to partial",
			args: args{es: &models.ElasticsearchPayload{
				Plan: &models.ElasticsearchClusterPlan{
					Transient: &models.TransientElasticsearchPlanConfiguration{
						RestoreSnapshot: &models.RestoreSnapshotConfiguration{
							SourceClusterID: "some",
						},
					},
				},
			}},
			want: &models.ElasticsearchPayload{
				Plan: &models.ElasticsearchClusterPlan{
					Transient: &models.TransientElasticsearchPlanConfiguration{
						RestoreSnapshot: &models.RestoreSnapshotConfiguration{
							SourceClusterID: "some",
							Strategy:        "partial",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ensurePartialSnapshotStrategy(tt.args.es)
			assert.Equal(t, tt.want, tt.args.es)
		})
	}
}
