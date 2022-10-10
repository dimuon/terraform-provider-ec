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
	"fmt"
	"io"
	"net/url"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	r "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/mock"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	provider "github.com/elastic/terraform-provider-ec/ec"
	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource"
	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/testutil"
)

func Test_createDeploymentWithEmptyFields(t *testing.T) {
	// traffic_filter = ["0.0.0.0/0", "192.168.10.0/24"]
	requestId := "cuchxqanal0g8rmx9ljog7qrrpd68iitulaz2mrch1vuuihetgo5ge3f6555vn4s"
	deploymentWithDefaultsIoOptimized := fmt.Sprintf(`
		resource "ec_deployment" "empty-declarations-IO-Optimized" {
			request_id = "%s"
			name = "my_deployment_name"
			deployment_template_id = "aws-io-optimized-v2"
			region = "us-east-1"
			version = "7.7.0"
			elasticsearch {}
		}`,
		requestId,
	)

	deploymentId := mock.ValidClusterID

	expectedDeployment := deploymentresource.Deployment{
		Id: deploymentId,
	}

	templateFileName := "testdata/template-aws-io-optimized-v2.json"

	// models.DeploymentGetResponse{
	// 		Healthy: ec.Bool(true),
	// 		ID:      ec.String(mock.ValidClusterID),
	// 		Resources: &models.DeploymentResources{
	// 			Elasticsearch: []*models.ElasticsearchResourceInfo{{
	// 				ID:    ec.String(mock.ValidClusterID),
	// 				RefID: ec.String("main-elasticsearch"),
	// 			}},
	// 		},
	// 	}
	deploymentResources := &models.DeploymentResources{
		Elasticsearch: []*models.ElasticsearchResourceInfo{
			{
				ID:     &mock.ValidClusterID,
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
									Version: "7.7.0",
									// UserSettingsYaml:         `some.setting: value`,
									// UserSettingsOverrideYaml: `some.setting: value2`,
									// UserSettingsJSON: map[string]interface{}{
									// 	"some.setting": "value",
									// },
									// UserSettingsOverrideJSON: map[string]interface{}{
									// 	"some.setting": "value2",
									// },
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
	}

	expectedRequest := &models.DeploymentCreateRequest{
		Name:     "my_deployment_name",
		Settings: &models.DeploymentCreateSettings{
			// TrafficFilterSettings: &models.TrafficFilterSettings{
			// 	Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
			// },
		},
		Metadata: &models.DeploymentCreateMetadata{
			Tags: []*models.MetadataItem{},
		},
		Resources: &models.DeploymentCreateResources{
			Elasticsearch: testutil.EnrichWithEmptyTopologies(testutil.ReaderToESPayload(t, readTestData(t, templateFileName), false), &models.ElasticsearchPayload{
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
			}),
		},
	}

	createResponse := &models.DeploymentCreateResponse{
		ID:      &deploymentId,
		Created: ec.Bool(true),
		Name:    ec.String("my_deployment_name"),
		Resources: []*models.DeploymentResource{
			// "kind": "elasticsearch",
			// "cloud_id": "my_deployment_name:dXMtZWFzdC0xLmF3cy5mb3VuZC5pbzo0NDMkMDU4OWRkYjNhY2VlNDY0MWI5NTgzMzAyMmJmMDRkMmIk",
			// "region": "us-east-1",
			// "ref_id": "main-elasticsearch",
			// "credentials": {
			// 	"username": "elastic",
			// 	"password": "password"
			// },
			// "id": "0589ddb3acee4641b95833022bf04d2b"
			{
				ID:      &mock.ValidClusterID,
				Kind:    ec.String("elasticsearch"),
				CloudID: "some cloud id",
				Region:  ec.String("us-east-1"),
				RefID:   ec.String("main-elasticsearch"),
				Credentials: &models.ClusterCredentials{
					Username: ec.String("elastic"),
					Password: ec.String("password"),
				},
			},
		},
	}

	r.UnitTest(t, r.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactoriesWithMockClient(
			api.NewDebugMock(
				os.Stdout,
				getTemplate(t, templateFileName),
				createDeployment(t, expectedRequest, createResponse, requestId),
				// updateRemoteClusters(t, expectedRequest, createResponse, requestId),
				readDeployment(t, &expectedDeployment, deploymentResources),
				readDeployment(t, &expectedDeployment, deploymentResources),
				readDeployment(t, &expectedDeployment, deploymentResources),
				readRemoteClusters(t),
				// readDeployment(t, &expectedDeployment, deploymentResources),
			),
		),
		Steps: []r.TestStep{
			{ // Create resource
				Config: deploymentWithDefaultsIoOptimized,
				// ExpectNonEmptyPlan:        true,
				PreventPostDestroyRefresh: true,
				// Check:  checkResource1(),
			},
		},
	})
}

func checkResource1() r.TestCheckFunc {
	resource := "ec_deployment_extension.my_extension"
	return r.ComposeAggregateTestCheckFunc(
		r.TestCheckResourceAttr(resource, "id", "someid"),
		r.TestCheckResourceAttr(resource, "name", "My extension"),
		r.TestCheckResourceAttr(resource, "description", "Some description"),
		r.TestCheckResourceAttr(resource, "version", "*"),
		r.TestCheckResourceAttr(resource, "extension_type", "bundle"),
	)
}

func getTemplate(t *testing.T, filename string) mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultReadMockHeaders,
			Method: "GET",
			Path:   "/api/v1/deployments/templates/aws-io-optimized-v2",
			Query:  url.Values{"region": {"us-east-1"}, "show_instance_configurations": {"false"}},
		},
		readTestData(t, filename),
	)
}

func readTestData(t *testing.T, filename string) io.ReadCloser {
	f, err := os.Open(filename)
	assert.Nil(t, err)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func createDeployment(t *testing.T, expectedRequest *models.DeploymentCreateRequest, response *models.DeploymentCreateResponse, requestId string) mock.Response {
	return mock.New201ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "POST",
			Path:   "/api/v1/deployments",
			Query:  url.Values{"request_id": {requestId}},
			// Body:   modelToBody(t, expectedRequest),
			Body: mock.NewStructBody(expectedRequest),
		},
		// modelToBody(t, response),
		mock.NewStructBody(response),
	)
}

func updateRemoteClusters(t *testing.T, expectedRequest *models.DeploymentCreateRequest, response *models.DeploymentCreateResponse, requestId string) mock.Response {
	return mock.New200ResponseAssertion(
		&mock.RequestAssertion{
			Host:   api.DefaultMockHost,
			Header: api.DefaultWriteMockHeaders,
			Method: "PUT",
			Path:   "/api/v1/deployments/320b7b540dfc967a7a649c18e2fce4ed/elasticsearch/main-elasticsearch/remote-clusters",
			Query:  url.Values{"request_id": {requestId}},
			// Body:   modelToBody(t, expectedRequest),
			Body: mock.NewStructBody(expectedRequest),
		},
		// modelToBody(t, response),
		mock.NewStructBody(response),
	)
}

// func modelToBody(t *testing.T, model any) io.ReadCloser {
// 	bytes, err := json.Marshal(model)
// 	assert.Nil(t, err)
// 	return mock.NewStringBody(string(bytes) + "\n")
// }

func readDeployment(t *testing.T, expectedDeployment *deploymentresource.Deployment, deploymentResources *models.DeploymentResources) mock.Response {

	return mock.New200StructResponse(&models.DeploymentGetResponse{
		ID:       &mock.ValidClusterID,
		Alias:    "my-deployment",
		Name:     ec.String("my_deployment_name"),
		Settings: &models.DeploymentSettings{
			// TrafficFilterSettings: &models.TrafficFilterSettings{
			// 	Rulesets: []string{"0.0.0.0/0", "192.168.10.0/24"},
			// },
			// Observability: &models.DeploymentObservabilitySettings{
			// 	Logging: &models.DeploymentLoggingSettings{
			// 		Destination: &models.ObservabilityAbsoluteDeployment{
			// 			DeploymentID: &mock.ValidClusterID,
			// 			RefID:        "main-elasticsearch",
			// 		},
			// 	},
			// 	Metrics: &models.DeploymentMetricsSettings{
			// 		Destination: &models.ObservabilityAbsoluteDeployment{
			// 			DeploymentID: &mock.ValidClusterID,
			// 			RefID:        "main-elasticsearch",
			// 		},
			// 	},
			// },
		},
		Resources: deploymentResources,
		Metadata:  &models.DeploymentMetadata{},
	})

	// return mock.New200ResponseAssertion(
	// 	&mock.RequestAssertion{
	// 		Header: api.DefaultReadMockHeaders,
	// 		Method: "GET",
	// 		Host:   api.DefaultMockHost,
	// 		Path:   fmt.Sprintf("/api/v1/deployments/%s", expectedDeployment.Id),
	// 		Query:  url.Values{"convert_legacy_plans": {"false"}, "show_metadata": {"true"}, "show_plan_defaults": {"true"}, "show_plan_history": {"false"}, "show_plan_logs": {"false"}, "show_plans": {"true"}, "show_settings": {"true"}, "show_system_alerts": {"5"}},
	// 	},
	// 	mock.NewStructBody(expectedDeployment),
	// )
}

func readRemoteClusters(t *testing.T) mock.Response {

	return mock.New200StructResponse(
		&models.RemoteResources{Resources: []*models.RemoteResourceRef{}},
	)
}

// func lastModified() strfmt.DateTime {
// 	lastModified, _ := strfmt.ParseDateTime("2021-01-07T22:13:42.999Z")
// 	return lastModified
// }

func protoV6ProviderFactoriesWithMockClient(client *api.API) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"ec": func() (tfprotov6.ProviderServer, error) {
			return providerserver.NewProtocol6(provider.ProviderWithClient(client, "unit-tests"))(), nil
		},
	}
}
