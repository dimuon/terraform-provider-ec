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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DeploymentData struct {
	Id                    types.String           `tfsdk:"id"`
	Alias                 types.String           `tfsdk:"alias"`
	Version               types.String           `tfsdk:"version"`
	Region                types.String           `tfsdk:"region"`
	DeploymentTemplateId  types.String           `tfsdk:"deployment_template_id"`
	Name                  types.String           `tfsdk:"name"`
	RequestId             types.String           `tfsdk:"request_id"`
	ElasticsearchUsername types.String           `tfsdk:"elasticsearch_username"`
	ElasticsearchPassword types.String           `tfsdk:"elasticsearch_password"`
	ApmSecretToken        types.String           `tfsdk:"apm_secret_token"`
	TrafficFilter         types.Set              `tfsdk:"traffic_filter"`
	Tags                  types.Map              `tfsdk:"tags"`
	Elasticsearch         ElasticsearchData      `tfsdk:"elasticsearch"`
	Kibana                KibanaData             `tfsdk:"kibana"`
	Apm                   ApmData                `tfsdk:"apm"`
	IntegrationsServer    IntegrationsServerData `tfsdk:"integrations_server"`
	EnterpriseSearch      EnterpriseSearchData   `tfsdk:"enterprise_search"`
	Observability         ObservabilityData      `tfsdk:"observability"`
}

type ElasticsearchData struct {
	Autoscale      types.String                    `tfsdk:"autoscale"`
	RefId          types.String                    `tfsdk:"ref_id"`
	ResourceId     types.String                    `tfsdk:"resource_id"`
	Region         types.String                    `tfsdk:"region"`
	CloudID        types.String                    `tfsdk:"cloud_id"`
	HttpEndpoint   types.String                    `tfsdk:"http_endpoint"`
	HttpsEndpoint  types.String                    `tfsdk:"https_endpoint"`
	Topology       ElasticsearchTopologyData       `tfsdk:"topology"`
	Config         ElasticsearchConfigData         `tfsdk:"config"`
	RemoteCluster  ElasticsearchRemoteClusterData  `tfsdk:"remote_cluster"`
	SnapshotSource ElasticsearchSnapshotSourceData `tfsdk:"snapshot_source"`
	Extension      ElasticsearchExtensionData      `tfsdk:"extension"`
	TrustAccount   ElasticsearchTrustAccountData   `tfsdk:"trust_account"`
	TrustExternal  ElasticsearchTrustExternalData  `tfsdk:"trust_external"`
	Strategy       ElasticsearchStrategyData       `tfsdk:"strategy"`
}

type ElasticsearchTopologyData struct {
	Id                      types.String                    `tfsdk:"id"`
	InstanceConfigurationId types.String                    `tfsdk:"instance_configuration_id"`
	Size                    types.String                    `tfsdk:"size"`
	SizeResource            types.String                    `tfsdk:"size_resource"`
	ZoneCount               types.Int64                     `tfsdk:"zone_count"`
	NodeTypeData            types.String                    `tfsdk:"node_type_data"`
	NodeTypeMaster          types.String                    `tfsdk:"node_type_master"`
	NodeTypeIngest          types.String                    `tfsdk:"node_type_ingest"`
	NodeTypeMl              types.String                    `tfsdk:"node_type_ml"`
	NodeRoles               types.Set                       `tfsdk:"node_roles"`
	Autoscaling             ElasticsearchAutoscalingData    `tfsdk:"autoscaling"`
	Config                  ElasticsearchTopologyConfigData `tfsdk:"config"`
}

type ElasticsearchAutoscalingData struct {
	MaxSizeResource    types.String `tfsdk:"max_size_resource"`
	MaxSize            types.String `tfsdk:"max_size"`
	MinSizeResource    types.String `tfsdk:"min_size_resource"`
	MinSize            types.String `tfsdk:"min_size"`
	PolicyOverrideJson types.String `tfsdk:"policy_override_json"`
}

type ElasticsearchTopologyConfigData struct {
	Plugins                  types.String `tfsdk:"plugins"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type ElasticsearchConfigData struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	Plugins                  types.Set    `tfsdk:"plugins"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type ElasticsearchRemoteClusterData struct {
	DeploymentId    types.String `tfsdk:"deployment_id"`
	Alias           types.String `tfsdk:"alias"`
	RefId           types.String `tfsdk:"ref_id"`
	SkipUnavailable types.String `tfsdk:"skip_unavailable"`
}

type ElasticsearchSnapshotSourceData struct {
	SourceElasticsearchClusterId types.String `tfsdk:"source_elasticsearch_cluster_id"`
	SnapshotName                 types.String `tfsdk:"snapshot_name"`
}

type ElasticsearchExtensionData struct {
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Version types.String `tfsdk:"version"`
	Url     types.String `tfsdk:"url"`
}

type ElasticsearchTrustAccountData struct {
	AccountId      types.String `tfsdk:"account_id"`
	TrustAll       types.Bool   `tfsdk:"trust_all"`
	TrustAllowlist types.Set    `tfsdk:"trust_allowlist"`
}

type ElasticsearchTrustExternalData struct {
	RelationshipId types.String `tfsdk:"relationship_id"`
	TrustAll       types.String `tfsdk:"trust_all"`
	TrustAllowlist types.String `tfsdk:"trust_allowlist"`
}

type ElasticsearchStrategyData struct {
	Type types.String `tfsdk:"type"`
}

type KibanaData struct {
	ElasticsearchClusterRefId types.String       `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String       `tfsdk:"ref_id"`
	ResourceId                types.String       `tfsdk:"resource_id"`
	Region                    types.String       `tfsdk:"region"`
	HttpEndpoint              types.String       `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String       `tfsdk:"https_endpoint"`
	Topology                  KibanaTopologyData `tfsdk:"topology"`
	Config                    KibanaConfigData   `tfsdk:"config"`
}

type KibanaTopologyData struct {
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
}

type KibanaConfigData struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type ApmData struct {
	ElasticsearchClusterRefId types.String    `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String    `tfsdk:"ref_id"`
	ResourceId                types.String    `tfsdk:"resource_id"`
	Region                    types.String    `tfsdk:"region"`
	HttpEndpoint              types.String    `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String    `tfsdk:"https_endpoint"`
	Topology                  ApmTopologyData `tfsdk:"topology"`
	Config                    ApmConfigData   `tfsdk:"config"`
}

type ApmTopologyData struct {
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
}

type ApmConfigData struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	DebugEnabled             types.Bool   `tfsdk:"debug_enabled"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type IntegrationsServerData struct {
	ElasticsearchClusterRefId types.String                   `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String                   `tfsdk:"ref_id"`
	ResourceId                types.String                   `tfsdk:"resource_id"`
	Region                    types.String                   `tfsdk:"region"`
	HttpEndpoint              types.String                   `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String                   `tfsdk:"https_endpoint"`
	Topology                  IntegrationsServerTopologyData `tfsdk:"topology"`
	Config                    IntegrationsServerConfigData   `tfsdk:"config"`
}

type IntegrationsServerTopologyData struct {
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
}

type IntegrationsServerConfigData struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	DebugEnabled             types.Bool   `tfsdk:"debug_enabled"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type EnterpriseSearchData struct {
	ElasticsearchClusterRefId types.String                 `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String                 `tfsdk:"ref_id"`
	ResourceId                types.String                 `tfsdk:"resource_id"`
	Region                    types.String                 `tfsdk:"region"`
	HttpEndpoint              types.String                 `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String                 `tfsdk:"https_endpoint"`
	Topology                  EnterpriseSearchTopologyData `tfsdk:"topology"`
	Config                    EnterpriseSearchConfigData   `tfsdk:"config"`
}

type EnterpriseSearchTopologyData struct {
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
	NodeTypeAppserver       types.Bool   `tfsdk:"node_type_appserver"`
	NodeTypeConnector       types.Bool   `tfsdk:"node_type_connector"`
	NodeTypeWorker          types.Bool   `tfsdk:"node_type_worker"`
}

type EnterpriseSearchConfigData struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type ObservabilityData struct {
	DeploymentId types.String `tfsdk:"deployment_id"`
	RefId        types.String `tfsdk:"ref_id"`
	Logs         types.Bool   `tfsdk:"logs"`
	Metrics      types.Bool   `tfsdk:"metrics"`
}
