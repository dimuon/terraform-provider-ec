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
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/flatteners"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Deployment struct {
	Id                    types.String         `tfsdk:"id"`
	Alias                 types.String         `tfsdk:"alias"`
	Version               types.String         `tfsdk:"version"`
	Region                types.String         `tfsdk:"region"`
	DeploymentTemplateId  types.String         `tfsdk:"deployment_template_id"`
	Name                  types.String         `tfsdk:"name"`
	RequestId             types.String         `tfsdk:"request_id"`
	ElasticsearchUsername types.String         `tfsdk:"elasticsearch_username"`
	ElasticsearchPassword types.String         `tfsdk:"elasticsearch_password"`
	ApmSecretToken        types.String         `tfsdk:"apm_secret_token"`
	TrafficFilter         []string             `tfsdk:"traffic_filter"`
	Tags                  types.Map            `tfsdk:"tags"`
	Elasticsearch         Elasticsearches      `tfsdk:"elasticsearch"`
	Kibana                []Kibana             `tfsdk:"kibana"`
	Apm                   []Apm                `tfsdk:"apm"`
	IntegrationsServer    []IntegrationsServer `tfsdk:"integrations_server"`
	EnterpriseSearch      []EnterpriseSearch   `tfsdk:"enterprise_search"`
	Observability         []Observability      `tfsdk:"observability"`
}

type Elasticsearches []Elasticsearch

func (dep *Deployment) fromModel(res *models.DeploymentGetResponse, remotes *models.RemoteResources) error {
	if res.Name == nil {
		return fmt.Errorf("server response doesn't contain name")
	}
	dep.Name.Value = *res.Name

	dep.Alias.Value = res.Alias

	if res.Metadata != nil && len(res.Metadata.Tags) >= 0 {
		dep.Tags = flatteners.FlattenTags(res.Metadata.Tags)
	}

	if res.Resources != nil {
		var err error

		dep.DeploymentTemplateId.Value, err = getDeploymentTemplateID(res.Resources)
		if err != nil {
			return err
		}

		dep.Region.Value = getRegion(res.Resources)

		// We're reconciling the version and storing the lowest version of any
		// of the deployment resources. This ensures that if an upgrade fails,
		// the state version will be lower than the desired version, making
		// retries possible. Once more resource types are added, the function
		// needs to be modified to check those as well.
		dep.Version.Value, err = getLowestVersion(res.Resources)
		if err != nil {
			// This code path is highly unlikely, but we're bubbling up the
			// error in case one of the versions isn't parseable by semver.
			return fmt.Errorf("failed reading deployment: %w", err)
		}

		if err := dep.Elasticsearch.fromModel(res.Resources.Elasticsearch, remotes); err != nil {
			return err
		}
	}

	return nil
}

func (ess *Elasticsearches) fromModel(in []*models.ElasticsearchResourceInfo, remotes *models.RemoteResources) error {
	if len(in) == 0 {
		return nil
	}

	if *ess == nil {
		*ess = make([]Elasticsearch, 0, len(in))
	}

	for _, model := range in {
		if util.IsCurrentEsPlanEmpty(model) || isEsResourceStopped(model) {
			continue
		}
		var es Elasticsearch
		if err := es.fromModel(model, remotes); err != nil {
			return err
		}
		*ess = append(*ess, es)
	}

	return nil
}

type ElasticsearchSnapshotSource struct {
	SourceElasticsearchClusterId types.String `tfsdk:"source_elasticsearch_cluster_id"`
	SnapshotName                 types.String `tfsdk:"snapshot_name"`
}

type ElasticsearchTrustAccount struct {
	AccountId      types.String `tfsdk:"account_id"`
	TrustAll       types.Bool   `tfsdk:"trust_all"`
	TrustAllowlist []string     `tfsdk:"trust_allowlist"`
}

type ElasticsearchTrustExternal struct {
	RelationshipId types.String `tfsdk:"relationship_id"`
	TrustAll       types.String `tfsdk:"trust_all"`
	TrustAllowlist types.String `tfsdk:"trust_allowlist"`
}

type ElasticsearchStrategy struct {
	Type types.String `tfsdk:"type"`
}

type Kibana struct {
	ElasticsearchClusterRefId types.String   `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String   `tfsdk:"ref_id"`
	ResourceId                types.String   `tfsdk:"resource_id"`
	Region                    types.String   `tfsdk:"region"`
	HttpEndpoint              types.String   `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String   `tfsdk:"https_endpoint"`
	Topology                  KibanaTopology `tfsdk:"topology"`
	Config                    KibanaConfig   `tfsdk:"config"`
}

type KibanaTopology struct {
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
}

type KibanaConfig struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type Apm struct {
	ElasticsearchClusterRefId types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String `tfsdk:"ref_id"`
	ResourceId                types.String `tfsdk:"resource_id"`
	Region                    types.String `tfsdk:"region"`
	HttpEndpoint              types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String `tfsdk:"https_endpoint"`
	Topology                  ApmTopology  `tfsdk:"topology"`
	Config                    ApmConfig    `tfsdk:"config"`
}

type ApmTopology struct {
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
}

type ApmConfig struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	DebugEnabled             types.Bool   `tfsdk:"debug_enabled"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type IntegrationsServer struct {
	ElasticsearchClusterRefId types.String               `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String               `tfsdk:"ref_id"`
	ResourceId                types.String               `tfsdk:"resource_id"`
	Region                    types.String               `tfsdk:"region"`
	HttpEndpoint              types.String               `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String               `tfsdk:"https_endpoint"`
	Topology                  IntegrationsServerTopology `tfsdk:"topology"`
	Config                    IntegrationsServerConfig   `tfsdk:"config"`
}

type IntegrationsServerTopology struct {
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
}

type IntegrationsServerConfig struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	DebugEnabled             types.Bool   `tfsdk:"debug_enabled"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type EnterpriseSearch struct {
	ElasticsearchClusterRefId types.String             `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String             `tfsdk:"ref_id"`
	ResourceId                types.String             `tfsdk:"resource_id"`
	Region                    types.String             `tfsdk:"region"`
	HttpEndpoint              types.String             `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String             `tfsdk:"https_endpoint"`
	Topology                  EnterpriseSearchTopology `tfsdk:"topology"`
	Config                    EnterpriseSearchConfig   `tfsdk:"config"`
}

type EnterpriseSearchTopology struct {
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
	NodeTypeAppserver       types.Bool   `tfsdk:"node_type_appserver"`
	NodeTypeConnector       types.Bool   `tfsdk:"node_type_connector"`
	NodeTypeWorker          types.Bool   `tfsdk:"node_type_worker"`
}

type EnterpriseSearchConfig struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type Observability struct {
	DeploymentId types.String `tfsdk:"deployment_id"`
	RefId        types.String `tfsdk:"ref_id"`
	Logs         types.Bool   `tfsdk:"logs"`
	Metrics      types.Bool   `tfsdk:"metrics"`
}
