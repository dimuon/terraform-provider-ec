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

package v2

import (
	"context"
	"fmt"

	"github.com/blang/semver"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deptemplateapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/esremoteclustersapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/apm/v2"
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/deployment/v1"
	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/elasticsearch/v2"
	v2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/elasticsearch/v2"
	enterprisesearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/enterprisesearch/v2"
	integrationsserverv2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/integrationsserver/v2"
	kibanav2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/kibana/v2"
	observabilityv2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/observability/v2"
	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Elasticsearch         types.Object `tfsdk:"elasticsearch"`
	Kibana                types.Object `tfsdk:"kibana"`
	Apm                   types.Object `tfsdk:"apm"`
	IntegrationsServer    types.Object `tfsdk:"integrations_server"`
	EnterpriseSearch      types.Object `tfsdk:"enterprise_search"`
	Observability         types.Object `tfsdk:"observability"`
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
	Elasticsearch         *elasticsearchv2.Elasticsearch           `tfsdk:"elasticsearch"`
	Kibana                *kibanav2.Kibana                         `tfsdk:"kibana"`
	Apm                   *apmv2.Apm                               `tfsdk:"apm"`
	IntegrationsServer    *integrationsserverv2.IntegrationsServer `tfsdk:"integrations_server"`
	EnterpriseSearch      *enterprisesearchv2.EnterpriseSearch     `tfsdk:"enterprise_search"`
	Observability         *observabilityv2.Observability           `tfsdk:"observability"`
}

// Nullify Elasticsearch topologies that are not specified in plan
// TODO: do it only for topologies that have zero size in response
func (dep *Deployment) NullifyNotUsedEsTopologies(ctx context.Context, esPlanObj types.Object) diag.Diagnostics {
	if esPlanObj.IsNull() || esPlanObj.IsUnknown() {
		return nil
	}

	var esPlan *v2.ElasticsearchTF

	if diags := tfsdk.ValueAs(ctx, esPlanObj, &esPlan); diags.HasError() {
		return diags
	}

	if esPlan == nil {
		return nil
	}

	if dep.Elasticsearch == nil {
		return nil
	}

	if esPlan.HotContentTier.IsNull() {
		dep.Elasticsearch.HotTier = nil
	}

	if esPlan.WarmTier.IsNull() {
		dep.Elasticsearch.WarmTier = nil
	}

	if esPlan.ColdTier.IsNull() {
		dep.Elasticsearch.ColdTier = nil
	}

	if esPlan.FrozenTier.IsNull() {
		dep.Elasticsearch.FrozenTier = nil
	}

	if esPlan.MlTier.IsNull() {
		dep.Elasticsearch.MlTier = nil
	}

	if esPlan.MasterTier.IsNull() {
		dep.Elasticsearch.MasterTier = nil
	}

	if esPlan.CoordinatingTier.IsNull() {
		dep.Elasticsearch.CoordinatingTier = nil
	}

	return nil
}

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

	if len(res.Resources.Elasticsearch) > 0 {
		dep.Elasticsearch, err = elasticsearchv2.ReadElasticsearch(res.Resources.Elasticsearch[0], remotes)
		if err != nil {
			return nil, err
		}
	}

	if len(res.Resources.Kibana) > 0 {
		dep.Kibana, err = kibanav2.ReadKibana(res.Resources.Kibana[0])
		if err != nil {
			return nil, err
		}
	}

	if len(res.Resources.Apm) > 0 {
		if dep.Apm, err = apmv2.ReadApm(res.Resources.Apm[0]); err != nil {
			return nil, err
		}
	}

	if len(res.Resources.IntegrationsServer) > 0 {
		if dep.IntegrationsServer, err = integrationsserverv2.ReadIntegrationsServer(res.Resources.IntegrationsServer[0]); err != nil {
			return nil, err
		}
	}

	if len(res.Resources.EnterpriseSearch) > 0 {
		if dep.EnterpriseSearch, err = enterprisesearchv2.ReadEnterpriseSearch(res.Resources.EnterpriseSearch[0]); err != nil {
			return nil, err
		}
	}

	if dep.TrafficFilter, err = v1.ReadTrafficFilters(res.Settings); err != nil {
		return nil, err
	}

	if dep.Observability, err = observabilityv2.ReadObservability(res.Settings); err != nil {
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

	useNodeRoles, err := v1.CompatibleWithNodeRoles(version)
	if err != nil {
		diagsnostics.AddError("Deployment parse error", err.Error())
		return nil, diagsnostics
	}

	elasticsearchPayload, diags := elasticsearchv2.ElasticsearchPayload(ctx, dep.Elasticsearch, template, dtID, version, useNodeRoles, false)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if elasticsearchPayload != nil {
		result.Resources.Elasticsearch = []*models.ElasticsearchPayload{elasticsearchPayload}
	}

	kibanaPayload, diags := kibanav2.KibanaPayload(ctx, dep.Kibana, template)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if kibanaPayload != nil {
		result.Resources.Kibana = []*models.KibanaPayload{kibanaPayload}
	}

	apmPayload, diags := apmv2.ApmPayload(ctx, dep.Apm, template)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if apmPayload != nil {
		result.Resources.Apm = []*models.ApmPayload{apmPayload}
	}

	integrationsServerPayload, diags := integrationsserverv2.IntegrationsServerPayload(ctx, dep.IntegrationsServer, template)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if integrationsServerPayload != nil {
		result.Resources.IntegrationsServer = []*models.IntegrationsServerPayload{integrationsServerPayload}
	}

	enterpriseSearchPayload, diags := enterprisesearchv2.EnterpriseSearchesPayload(ctx, dep.EnterpriseSearch, template)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if enterpriseSearchPayload != nil {
		result.Resources.EnterpriseSearch = []*models.EnterpriseSearchPayload{enterpriseSearchPayload}
	}

	if diags := v1.TrafficFilterToModel(ctx, dep.TrafficFilter, &result); diags.HasError() {
		diagsnostics.Append(diags...)
	}

	observabilityPayload, diags := observabilityv2.ObservabilityPayload(ctx, dep.Observability, client)

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

	if dep.Observability == nil {
		return
	}

	if dep.Observability.DeploymentId == nil {
		return
	}

	if *dep.Observability.DeploymentId == dep.Id {
		*dep.Observability.DeploymentId = "self"
	}
}

func (dep *Deployment) SetCredentialsIfEmpty(state *DeploymentTF) {
	if state == nil {
		return
	}

	if dep.ElasticsearchPassword == "" && state.ElasticsearchPassword.Value != "" {
		dep.ElasticsearchPassword = state.ElasticsearchPassword.Value
	}

	if dep.ElasticsearchUsername == "" && state.ElasticsearchUsername.Value != "" {
		dep.ElasticsearchUsername = state.ElasticsearchUsername.Value
	}

	if (dep.ApmSecretToken == nil || *dep.ApmSecretToken == "") && state.ApmSecretToken.Value != "" {
		dep.ApmSecretToken = &state.ApmSecretToken.Value
	}
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

	useNodeRoles, err := v1.CompatibleWithNodeRoles(version)
	if err != nil {
		diagsnostics.AddError("Deployment parse error", err.Error())
		return nil, diagsnostics
	}

	convertLegacy, diags := plan.legacyToNodeRoles(ctx, curState)
	if diags.HasError() {
		return nil, diags
	}
	useNodeRoles = useNodeRoles && convertLegacy

	elasticsearchPayload, diags := elasticsearchv2.ElasticsearchPayload(ctx, plan.Elasticsearch, template, dtID, version, useNodeRoles, skipEStopologies)

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

	kibanaPayload, diags := kibanav2.KibanaPayload(ctx, plan.Kibana, template)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if kibanaPayload != nil {
		result.Resources.Kibana = append(result.Resources.Kibana, kibanaPayload)
	}

	apmPayload, diags := apmv2.ApmPayload(ctx, plan.Apm, template)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if apmPayload != nil {
		result.Resources.Apm = append(result.Resources.Apm, apmPayload)
	}

	integrationsServerPayload, diags := integrationsserverv2.IntegrationsServerPayload(ctx, plan.IntegrationsServer, template)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if integrationsServerPayload != nil {
		result.Resources.IntegrationsServer = append(result.Resources.IntegrationsServer, integrationsServerPayload)
	}

	enterpriseSearchPayload, diags := enterprisesearchv2.EnterpriseSearchesPayload(ctx, plan.EnterpriseSearch, template)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}

	if enterpriseSearchPayload != nil {
		result.Resources.EnterpriseSearch = append(result.Resources.EnterpriseSearch, enterpriseSearchPayload)
	}

	observabilityPayload, diags := observabilityv2.ObservabilityPayload(ctx, plan.Observability, client)
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
	if oldV.LT(v1.DataTiersVersion) && newV.GE(v1.DataTiersVersion) {
		return false, nil
	}

	if plan.Elasticsearch.IsNull() || plan.Elasticsearch.IsUnknown() {
		diags.AddError("Cannot migrate node types to node roles", "cannot find elasticsearch data")
		return false, diags
	}

	var es *elasticsearchv2.ElasticsearchTF

	if diags := tfsdk.ValueAs(ctx, plan.Elasticsearch, &es); diags.HasError() {
		return false, diags
	}

	if es == nil {
		diags.AddError("Cannot migrate node types to node roles", "cannot find elasticsearch data")
		return false, diags
	}

	// When any topology elements in the state have the node_type_*
	// properties set, the node_role field cannot be used, since
	// we'd be changing the version AND migrating over `node_role`s
	// which is not permitted by the API.

	for _, obj := range []types.Object{es.HotContentTier, es.CoordinatingTier, es.MasterTier, es.WarmTier, es.ColdTier, es.FrozenTier, es.MlTier} {
		if obj.IsNull() || obj.IsUnknown() {
			continue
		}

		topology, diags := elasticsearchv2.ObjectToTopology(ctx, obj)

		if diags.HasError() {
			return false, diags
		}

		if topology.HasNodeType() {
			return false, nil
		}
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

// func HandleRemoteClusters(ctx context.Context, client *api.API, newState, oldState DeploymentTF) diag.Diagnostics {
func HandleRemoteClusters(ctx context.Context, client *api.API, deploymentId string, esObj types.Object) diag.Diagnostics {
	remoteClusters, refId, diags := ElasticsearchRemoteClustersPayload(ctx, client, deploymentId, esObj)

	if diags.HasError() {
		return diags
	}

	if err := esremoteclustersapi.Update(esremoteclustersapi.UpdateParams{
		API:             client,
		DeploymentID:    deploymentId,
		RefID:           refId,
		RemoteResources: remoteClusters,
	}); err != nil {
		diags.AddError("cannot update remote clusters", err.Error())
		return diags
	}

	return nil
}

func ElasticsearchRemoteClustersPayload(ctx context.Context, client *api.API, deploymentId string, esObj types.Object) (*models.RemoteResources, string, diag.Diagnostics) {
	var es *v2.ElasticsearchTF

	diags := tfsdk.ValueAs(ctx, esObj, &es)

	if diags.HasError() {
		return nil, "", diags
	}

	remoteRes, diags := elasticsearchv2.ElasticsearchRemoteClustersPayload(ctx, es.RemoteCluster)
	if diags.HasError() {
		return nil, "", diags
	}

	return remoteRes, es.RefId.Value, nil
}
