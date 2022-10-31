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

package v1

import (
	"context"
	"fmt"

	"github.com/blang/semver"
	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deptemplateapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/esremoteclustersapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	apmv1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/apm/v1"
	elasticsearchv1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/elasticsearch/v1"
	enterprisesearchv1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/enterprisesearch/v1"
	integrationsserverv1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/integrationsserver/v1"
	kibanav1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/kibana/v1"
	observabilityv1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/observability/v1"
	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/utils"
)

type DeploymentTF struct {
	Id                    types.String `tfsdk:"id"`
	Alias                 types.String `tfsdk:"alias"`
	Version               types.String `tfsdk:"version"`
	Region                types.String `tfsdk:"region"`
	DeploymentTemplateId  types.String `tfsdk:"deployment_template_id"`
	Name                  types.String `tfsdk:"name"`
	RequestId             types.String `tfsdk:"request_id"`
	ElasticsearchUsername types.String `tfsdk:"elasticsearch_username"`
	ElasticsearchPassword types.String `tfsdk:"elasticsearch_password"`
	ApmSecretToken        types.String `tfsdk:"apm_secret_token"`
	TrafficFilter         types.Set    `tfsdk:"traffic_filter"`
	Tags                  types.Map    `tfsdk:"tags"`
	Elasticsearch         types.List   `tfsdk:"elasticsearch"`
	Kibana                types.List   `tfsdk:"kibana"`
	Apm                   types.List   `tfsdk:"apm"`
	IntegrationsServer    types.List   `tfsdk:"integrations_server"`
	EnterpriseSearch      types.List   `tfsdk:"enterprise_search"`
	Observability         types.List   `tfsdk:"observability"`
}

type Deployment struct {
	Id                    string                                   `tfsdk:"id"`
	Alias                 string                                   `tfsdk:"alias"`
	Version               string                                   `tfsdk:"version"`
	Region                string                                   `tfsdk:"region"`
	DeploymentTemplateId  string                                   `tfsdk:"deployment_template_id"`
	Name                  string                                   `tfsdk:"name"`
	RequestId             string                                   `tfsdk:"request_id"`
	ElasticsearchUsername string                                   `tfsdk:"elasticsearch_username"`
	ElasticsearchPassword string                                   `tfsdk:"elasticsearch_password"`
	ApmSecretToken        *string                                  `tfsdk:"apm_secret_token"`
	TrafficFilter         []string                                 `tfsdk:"traffic_filter"`
	Tags                  map[string]string                        `tfsdk:"tags"`
	Elasticsearch         elasticsearchv1.Elasticsearches          `tfsdk:"elasticsearch"`
	Kibana                kibanav1.Kibanas                         `tfsdk:"kibana"`
	Apm                   apmv1.Apms                               `tfsdk:"apm"`
	IntegrationsServer    integrationsserverv1.IntegrationsServers `tfsdk:"integrations_server"`
	EnterpriseSearch      enterprisesearchv1.EnterpriseSearches    `tfsdk:"enterprise_search"`
	Observability         observabilityv1.Observabilities          `tfsdk:"observability"`
}

var (
	DataTiersVersion = semver.MustParse("7.10.0")
)

func ReadDeployment(res *models.DeploymentGetResponse, remotes *models.RemoteResources, deploymentResources []*models.DeploymentResource) (*Deployment, error) {
	var dep Deployment

	if res.ID == nil {
		return nil, utils.MissingField("ID")
	}
	dep.Id = *res.ID

	dep.Alias = res.Alias

	if res.Name == nil {
		return nil, utils.MissingField("Name")
	}
	dep.Name = *res.Name

	if res.Metadata != nil {
		dep.Tags = converters.TagsToMap(res.Metadata.Tags)
	}

	if res.Resources == nil {
		return nil, nil
	}

	templateID, err := utils.GetDeploymentTemplateID(res.Resources)
	if err != nil {
		return nil, err
	}

	dep.DeploymentTemplateId = templateID

	dep.Region = utils.GetRegion(res.Resources)

	// We're reconciling the version and storing the lowest version of any
	// of the deployment resources. This ensures that if an upgrade fails,
	// the state version will be lower than the desired version, making
	// retries possible. Once more resource types are added, the function
	// needs to be modified to check those as well.
	version, err := utils.GetLowestVersion(res.Resources)
	if err != nil {
		// This code path is highly unlikely, but we're bubbling up the
		// error in case one of the versions isn't parseable by semver.
		return nil, fmt.Errorf("failed reading deployment: %w", err)
	}
	dep.Version = version

	dep.Elasticsearch, err = elasticsearchv1.ReadElasticsearches(res.Resources.Elasticsearch, remotes)
	if err != nil {
		return nil, err
	}

	dep.Kibana, err = kibanav1.ReadKibanas(res.Resources.Kibana)
	if err != nil {
		return nil, err
	}

	if dep.Apm, err = apmv1.ReadApms(res.Resources.Apm); err != nil {
		return nil, err
	}

	if dep.IntegrationsServer, err = integrationsserverv1.ReadIntegrationsServers(res.Resources.IntegrationsServer); err != nil {
		return nil, err
	}

	if dep.EnterpriseSearch, err = enterprisesearchv1.ReadEnterpriseSearches(res.Resources.EnterpriseSearch); err != nil {
		return nil, err
	}

	if dep.TrafficFilter, err = ReadTrafficFilters(res.Settings); err != nil {
		return nil, err
	}

	if dep.Observability, err = observabilityv1.ReadObservabilities(res.Settings); err != nil {
		return nil, err
	}

	if err := dep.parseCredentials(deploymentResources); err != nil {
		return nil, err
	}

	return &dep, nil
}

func (dep DeploymentTF) CreateRequest(ctx context.Context, client *api.API) (*models.DeploymentCreateRequest, diag.Diagnostics) {
	var result = models.DeploymentCreateRequest{
		Name:      dep.Name.Value,
		Alias:     dep.Alias.Value,
		Resources: &models.DeploymentCreateResources{},
		Settings:  &models.DeploymentCreateSettings{},
		Metadata:  &models.DeploymentCreateMetadata{},
	}

	dtID := dep.DeploymentTemplateId.Value
	version := dep.Version.Value

	var diagsnostics diag.Diagnostics

	template, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 dtID,
		Region:                     dep.Region.Value,
		HideInstanceConfigurations: true,
	})
	if err != nil {
		diagsnostics.AddError("Deployment template get error", err.Error())
		return nil, diagsnostics
	}

	useNodeRoles, err := CompatibleWithNodeRoles(version)
	if err != nil {
		diagsnostics.AddError("Deployment parse error", err.Error())
		return nil, diagsnostics
	}

	elasticsearchPayload, diags := elasticsearchv1.ElasticsearchPayload(ctx, dep.Elasticsearch, template, dtID, version, useNodeRoles, false)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if elasticsearchPayload != nil {
		result.Resources.Elasticsearch = []*models.ElasticsearchPayload{elasticsearchPayload}
	}

	kibanaPayload, diags := kibanav1.KibanaPayload(ctx, dep.Kibana, template)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if kibanaPayload != nil {
		result.Resources.Kibana = []*models.KibanaPayload{kibanaPayload}
	}

	apmPayload, diags := apmv1.ApmPayload(ctx, dep.Apm, template)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if apmPayload != nil {
		result.Resources.Apm = []*models.ApmPayload{apmPayload}
	}

	integrationsServerPayload, diags := integrationsserverv1.IntegrationsServerPayload(ctx, dep.IntegrationsServer, template)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if integrationsServerPayload != nil {
		result.Resources.IntegrationsServer = []*models.IntegrationsServerPayload{integrationsServerPayload}
	}

	enterpriseSearchPayload, diags := enterprisesearchv1.EnterpriseSearchesPayload(ctx, dep.EnterpriseSearch, template)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if enterpriseSearchPayload != nil {
		result.Resources.EnterpriseSearch = []*models.EnterpriseSearchPayload{enterpriseSearchPayload}
	}

	if diags := TrafficFilterToModel(ctx, dep.TrafficFilter, &result); diags.HasError() {
		diagsnostics.Append(diags...)
	}

	observabilityPayload, diags := observabilityv1.ObservabilityPayload(ctx, dep.Observability, client)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	result.Settings.Observability = observabilityPayload

	result.Metadata.Tags, diags = converters.TFmapToTags(ctx, dep.Tags)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	return &result, diagsnostics
}

// parseCredentials parses the Create or Update response Resources populating
// credential settings in the Terraform state if the keys are found, currently
// populates the following credentials in plain text:
// * Elasticsearch username and Password
func (dep *Deployment) parseCredentials(resources []*models.DeploymentResource) error {
	for _, res := range resources {

		if creds := res.Credentials; creds != nil {
			if creds.Username != nil && *creds.Username != "" {
				dep.ElasticsearchUsername = *creds.Username
			}

			if creds.Password != nil && *creds.Password != "" {
				dep.ElasticsearchPassword = *creds.Password
			}
		}

		if res.SecretToken != "" {
			dep.ApmSecretToken = &res.SecretToken
		}
	}

	return nil
}

func (dep *Deployment) ProcessSelfInObservability() {

	if len(dep.Observability) == 0 {
		return
	}

	if dep.Observability[0].DeploymentId == nil {
		return
	}

	if *dep.Observability[0].DeploymentId == dep.Id {
		*dep.Observability[0].DeploymentId = "self"
	}
}

func (dep *Deployment) SetCredentialsIfEmpty(current DeploymentTF) {

	if dep.ElasticsearchPassword == "" && current.ElasticsearchPassword.Value != "" {
		dep.ElasticsearchPassword = current.ElasticsearchPassword.Value
	}

	if dep.ElasticsearchUsername == "" && current.ElasticsearchUsername.Value != "" {
		dep.ElasticsearchUsername = current.ElasticsearchUsername.Value
	}

	if (dep.ApmSecretToken == nil || *dep.ApmSecretToken == "") && current.ApmSecretToken.Value != "" {
		dep.ApmSecretToken = &current.ApmSecretToken.Value
	}
}

func ReadTrafficFilters(in *models.DeploymentSettings) ([]string, error) {
	if in == nil || in.TrafficFilterSettings == nil || len(in.TrafficFilterSettings.Rulesets) == 0 {
		return nil, nil
	}

	var rules []string

	return append(rules, in.TrafficFilterSettings.Rulesets...), nil
}

func CompatibleWithNodeRoles(version string) (bool, error) {
	deploymentVersion, err := semver.Parse(version)
	if err != nil {
		return false, fmt.Errorf("failed to parse Elasticsearch version: %w", err)
	}

	return deploymentVersion.GE(DataTiersVersion), nil
}

// TrafficFilterToModel expands the flattened "traffic_filter" settings to
// a DeploymentCreateRequest.
func TrafficFilterToModel(ctx context.Context, set types.Set, req *models.DeploymentCreateRequest) diag.Diagnostics {
	if len(set.Elems) == 0 || req == nil {
		return nil
	}

	if req.Settings == nil {
		req.Settings = &models.DeploymentCreateSettings{}
	}

	if req.Settings.TrafficFilterSettings == nil {
		req.Settings.TrafficFilterSettings = &models.TrafficFilterSettings{}
	}

	var rulesets []string
	if diags := tfsdk.ValueAs(ctx, set, &rulesets); diags.HasError() {
		return diags
	}

	req.Settings.TrafficFilterSettings.Rulesets = append(
		req.Settings.TrafficFilterSettings.Rulesets,
		rulesets...,
	)

	return nil
}

func (plan DeploymentTF) UpdateRequest(ctx context.Context, client *api.API, curState DeploymentTF) (*models.DeploymentUpdateRequest, diag.Diagnostics) {
	var result = models.DeploymentUpdateRequest{
		Name:         plan.Name.Value,
		Alias:        plan.Alias.Value,
		PruneOrphans: ec.Bool(true),
		Resources:    &models.DeploymentUpdateResources{},
		Settings:     &models.DeploymentUpdateSettings{},
		Metadata:     &models.DeploymentUpdateMetadata{},
	}

	dtID := plan.DeploymentTemplateId.Value
	version := plan.Version.Value

	var diagsnostics diag.Diagnostics

	template, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 dtID,
		Region:                     plan.Region.Value,
		HideInstanceConfigurations: true,
	})
	if err != nil {
		diagsnostics.AddError("Deployment template get error", err.Error())
		return nil, diagsnostics
	}

	// When the deployment template is changed, we need to skip the missing
	// resource topologies to account for a new instance_configuration_id and
	// a different default value.
	skipEStopologies := plan.DeploymentTemplateId.Value != "" && plan.DeploymentTemplateId.Value != curState.DeploymentTemplateId.Value && curState.DeploymentTemplateId.Value != ""
	// If the deployment_template_id is changed, then we skip updating the
	// Elasticsearch topology to account for the case where the
	// instance_configuration_id changes, i.e. Hot / Warm, etc.
	// This might not be necessary going forward as we move to
	// tiered Elasticsearch nodes.

	useNodeRoles, err := CompatibleWithNodeRoles(version)
	if err != nil {
		diagsnostics.AddError("Deployment parse error", err.Error())
		return nil, diagsnostics
	}

	convertLegacy, diags := plan.legacyToNodeRoles(ctx, curState)
	if diags.HasError() {
		return nil, diags
	}
	useNodeRoles = useNodeRoles && convertLegacy

	elasticsearchPayload, diags := elasticsearchv1.ElasticsearchPayload(ctx, plan.Elasticsearch, template, dtID, version, useNodeRoles, skipEStopologies)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if elasticsearchPayload != nil {
		// if the restore snapshot operation has been specified, the snapshot restore
		// can't be full once the cluster has been created, so the Strategy must be set
		// to "partial".
		ensurePartialSnapshotStrategy(elasticsearchPayload)

		result.Resources.Elasticsearch = append(result.Resources.Elasticsearch, elasticsearchPayload)
	}

	kibanaPayload, diags := kibanav1.KibanaPayload(ctx, plan.Kibana, template)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if kibanaPayload != nil {
		result.Resources.Kibana = append(result.Resources.Kibana, kibanaPayload)
	}

	apmPayload, diags := apmv1.ApmPayload(ctx, plan.Apm, template)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if apmPayload != nil {
		result.Resources.Apm = append(result.Resources.Apm, apmPayload)
	}

	integrationsServerPayload, diags := integrationsserverv1.IntegrationsServerPayload(ctx, plan.IntegrationsServer, template)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if integrationsServerPayload != nil {
		result.Resources.IntegrationsServer = append(result.Resources.IntegrationsServer, integrationsServerPayload)
	}

	enterpriseSearchPayload, diags := enterprisesearchv1.EnterpriseSearchesPayload(ctx, plan.EnterpriseSearch, template)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if enterpriseSearchPayload != nil {
		result.Resources.EnterpriseSearch = append(result.Resources.EnterpriseSearch, enterpriseSearchPayload)
	}

	observabilityPayload, diags := observabilityv1.ObservabilityPayload(ctx, plan.Observability, client)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}
	result.Settings.Observability = observabilityPayload

	// In order to stop shipping logs and metrics, an empty Observability
	// object must be passed, as opposed to a nil object when creating a
	// deployment without observability settings.
	if plan.Observability.IsNull() && !curState.Observability.IsNull() {
		result.Settings.Observability = &models.DeploymentObservabilitySettings{}
	}

	result.Metadata.Tags, diags = converters.TFmapToTags(ctx, plan.Tags)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	return &result, diagsnostics
}

// legacyToNodeRoles returns true when the legacy  "node_type_*" should be
// migrated over to node_roles. Which will be true when:
// * The version field doesn't change.
// * The version field changes but:
//   - The Elasticsearch.0.toplogy doesn't have any node_type_* set.
func (plan DeploymentTF) legacyToNodeRoles(ctx context.Context, curState DeploymentTF) (bool, diag.Diagnostics) {
	if curState.Version.Value == "" || curState.Version.Value == plan.Version.Value {
		return true, nil
	}

	// If the previous version is empty, node_roles should be used.
	if curState.Version.Value == "" {
		return true, nil
	}

	var diags diag.Diagnostics
	oldV, err := semver.Parse(curState.Version.Value)
	if err != nil {
		diags.AddError("failed to parse previous Elasticsearch version", err.Error())
		return false, diags
	}
	newV, err := semver.Parse(plan.Version.Value)
	if err != nil {
		diags.AddError("failed to parse new Elasticsearch version", err.Error())
		return false, diags
	}

	// if the version change moves from non-node_roles to one
	// that supports node roles, do not migrate on that step.
	if oldV.LT(DataTiersVersion) && newV.GE(DataTiersVersion) {
		return false, nil
	}

	// When any topology elements in the state have the node_type_*
	// properties set, the node_role field cannot be used, since
	// we'd be changing the version AND migrating over `node_role`s
	// which is not permitted by the API.
	var hasNodeTypeSet bool

	var es *elasticsearchv1.ElasticsearchTF

	if diags := utils.GetFirst(ctx, plan.Elasticsearch, &es); diags.HasError() {
		return false, diags
	}

	if es == nil {
		diags.AddError("Cannot migrate node types to node roles", "cannot find elasticsearch data")
		return false, diags
	}

	var esTopologies []elasticsearchv1.ElasticsearchTopologyTF
	if diags := es.Topology.ElementsAs(ctx, &esTopologies, true); diags.HasError() {
		return false, diags
	}

	for _, topology := range esTopologies {
		hasNodeTypeSet = topology.NodeTypeData.Value != "" ||
			topology.NodeTypeIngest.Value != "" ||
			topology.NodeTypeMaster.Value != "" ||
			topology.NodeTypeMl.Value != ""
	}

	if hasNodeTypeSet {
		return false, nil
	}

	return true, nil
}

func ensurePartialSnapshotStrategy(es *models.ElasticsearchPayload) {
	transient := es.Plan.Transient
	if transient == nil || transient.RestoreSnapshot == nil {
		return
	}
	transient.RestoreSnapshot.Strategy = "partial"
}

func HandleRemoteClusters(ctx context.Context, client *api.API, plan, state DeploymentTF) diag.Diagnostics {
	if plan.Elasticsearch.Equal(state.Elasticsearch) {
		return nil
	}

	var es *elasticsearchv1.ElasticsearchTF

	var diags diag.Diagnostics

	if diags = utils.GetFirst(ctx, plan.Elasticsearch, &es); diags.HasError() {
		return diags
	}

	if es == nil {
		return nil
	}

	if len(es.RemoteCluster.Elems) == 0 {
		return nil
	}

	remoteRes, diags := elasticsearchv1.ElasticsearchRemoteClustersPayload(ctx, es.RemoteCluster)
	if diags.HasError() {
		return diags
	}

	if err := esremoteclustersapi.Update(esremoteclustersapi.UpdateParams{
		API:             client,
		DeploymentID:    plan.Id.Value,
		RefID:           es.RefId.Value,
		RemoteResources: remoteRes,
	}); err != nil {
		diags.AddError("cannot update remote clusters", err.Error())
		return diags
	}

	return nil
}
