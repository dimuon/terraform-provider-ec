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
	"fmt"

	"github.com/blang/semver"
	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deptemplateapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
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
	Elasticsearch         types.List   `tfsdk:"elasticsearch"`
	Kibana                types.List   `tfsdk:"kibana"`
	Apm                   types.List   `tfsdk:"apm"`
	IntegrationsServer    types.List   `tfsdk:"integrations_server"`
	EnterpriseSearch      types.List   `tfsdk:"enterprise_search"`
	Observability         types.List   `tfsdk:"observability"`
}

type Deployment struct {
	Id                    string              `tfsdk:"id"`
	Alias                 string              `tfsdk:"alias"`
	Version               string              `tfsdk:"version"`
	Region                string              `tfsdk:"region"`
	DeploymentTemplateId  string              `tfsdk:"deployment_template_id"`
	Name                  string              `tfsdk:"name"`
	RequestId             string              `tfsdk:"request_id"`
	ElasticsearchUsername string              `tfsdk:"elasticsearch_username"`
	ElasticsearchPassword string              `tfsdk:"elasticsearch_password"`
	ApmSecretToken        string              `tfsdk:"apm_secret_token"`
	TrafficFilter         []string            `tfsdk:"traffic_filter"`
	Tags                  map[string]string   `tfsdk:"tags"`
	Elasticsearch         Elasticsearches     `tfsdk:"elasticsearch"`
	Kibana                Kibanas             `tfsdk:"kibana"`
	Apm                   Apms                `tfsdk:"apm"`
	IntegrationsServer    IntegrationsServers `tfsdk:"integrations_server"`
	EnterpriseSearch      EnterpriseSearches  `tfsdk:"enterprise_search"`
	Observability         Observabilities     `tfsdk:"observability"`
}

var (
	dataTiersVersion = semver.MustParse("7.10.0")
)

func missingField(field string) error {
	return fmt.Errorf("server response doesn't contain deployment '%s'", field)
}

func readDeployment(res *models.DeploymentGetResponse, remotes *models.RemoteResources, depRes []*models.DeploymentResource) (*Deployment, error) {
	var dep Deployment

	if res.ID == nil {
		return nil, missingField("ID")
	}
	dep.Id = *res.ID

	dep.Alias = res.Alias

	if res.Name == nil {
		return nil, missingField("Name")
	}
	dep.Name = *res.Name

	dep.Tags = converters.TagsToMap(res.Metadata.Tags)

	if res.Resources == nil {
		return nil, nil
	}

	templateID, err := getDeploymentTemplateID(res.Resources)
	if err != nil {
		return nil, err
	}

	dep.DeploymentTemplateId = templateID

	dep.Region = getRegion(res.Resources)

	// We're reconciling the version and storing the lowest version of any
	// of the deployment resources. This ensures that if an upgrade fails,
	// the state version will be lower than the desired version, making
	// retries possible. Once more resource types are added, the function
	// needs to be modified to check those as well.
	version, err := getLowestVersion(res.Resources)
	if err != nil {
		// This code path is highly unlikely, but we're bubbling up the
		// error in case one of the versions isn't parseable by semver.
		return nil, fmt.Errorf("failed reading deployment: %w", err)
	}
	dep.Version = version

	if dep.Elasticsearch, err = readElasticsearches(res.Resources.Elasticsearch, remotes); err != nil {
		return nil, err
	}

	if dep.Kibana, err = readKibanas(res.Resources.Kibana); err != nil {
		return nil, err
	}

	if dep.Apm, err = readApms(res.Resources.Apm); err != nil {
		return nil, err
	}

	if dep.IntegrationsServer, err = readIntegrationsServers(res.Resources.IntegrationsServer); err != nil {
		return nil, err
	}

	if dep.EnterpriseSearch, err = readEnterpriseSearches(res.Resources.EnterpriseSearch); err != nil {
		return nil, err
	}

	if dep.EnterpriseSearch, err = readEnterpriseSearches(res.Resources.EnterpriseSearch); err != nil {
		return nil, err
	}

	if dep.TrafficFilter, err = readTrafficFilters(res.Settings); err != nil {
		return nil, err
	}

	if dep.Observability, err = ReadObservability(res.Settings); err != nil {
		return nil, err
	}

	if err := dep.parseCredentials(depRes); err != nil {
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

	useNodeRoles, err := compatibleWithNodeRoles(version)
	if err != nil {
		diagsnostics.AddError("Deployment parse error", err.Error())
		return nil, diagsnostics
	}

	esRes, diags := elasticsearchPayload(ctx, template, dtID, version, useNodeRoles, dep.Elasticsearch, false)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}
	result.Resources.Elasticsearch = append(result.Resources.Elasticsearch, esRes...)

	kibanaRes, diags := kibanaPayload(ctx, template, dep.Kibana)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}
	result.Resources.Kibana = append(result.Resources.Kibana, kibanaRes...)

	apms, diags := apmsPayload(ctx, template, dep.Apm)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}
	result.Resources.Apm = append(result.Resources.Apm, apms...)

	integrationsServerRes, diags := integrationsServerPayload(ctx, template, dep.IntegrationsServer)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}
	result.Resources.IntegrationsServer = append(result.Resources.IntegrationsServer, integrationsServerRes...)

	enterpriseSearchRes, diags := enterpriseSearchesPayload(ctx, template, dep.EnterpriseSearch)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}
	result.Resources.EnterpriseSearch = append(result.Resources.EnterpriseSearch, enterpriseSearchRes...)

	if diags := trafficFilterToModel(ctx, dep.TrafficFilter, &result); diags.HasError() {
		diagsnostics.Append(diags...)
	}

	observability, diags := observabilityPayload(ctx, client, dep.Observability)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}
	result.Settings.Observability = observability

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
			dep.ApmSecretToken = res.SecretToken
		}
	}

	return nil
}

func readTrafficFilters(in *models.DeploymentSettings) ([]string, error) {
	if in == nil || in.TrafficFilterSettings == nil || len(in.TrafficFilterSettings.Rulesets) == 0 {
		return nil, nil
	}

	var rules []string

	return append(rules, in.TrafficFilterSettings.Rulesets...), nil
}

func compatibleWithNodeRoles(version string) (bool, error) {
	deploymentVersion, err := semver.Parse(version)
	if err != nil {
		return false, fmt.Errorf("failed to parse Elasticsearch version: %w", err)
	}

	return deploymentVersion.GE(dataTiersVersion), nil
}

// trafficFilterToModel expands the flattened "traffic_filter" settings to
// a DeploymentCreateRequest.
func trafficFilterToModel(ctx context.Context, set types.Set, req *models.DeploymentCreateRequest) diag.Diagnostics {
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
	template, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 dtID,
		Region:                     plan.Region.Value,
		HideInstanceConfigurations: true,
	})

	var diagsnostics diag.Diagnostics

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

	useNodeRoles, err := compatibleWithNodeRoles(version)
	if err != nil {
		diagsnostics.AddError("Deployment parse error", err.Error())
		return nil, diagsnostics
	}

	convertLegacy, diags := plan.legacyToNodeRoles(ctx, curState)
	if diags.HasError() {
		return nil, diags
	}
	useNodeRoles = useNodeRoles && convertLegacy

	esRes, diags := elasticsearchPayload(ctx, template, dtID, version, useNodeRoles, plan.Elasticsearch, skipEStopologies)

	if diags.HasError() {
		diagsnostics.Append(diags...)
	}
	result.Resources.Elasticsearch = append(result.Resources.Elasticsearch, esRes...)

	// if the restore snapshot operation has been specified, the snapshot restore
	// can't be full once the cluster has been created, so the Strategy must be set
	// to "partial".
	ensurePartialSnapshotStrategy(esRes)

	kibanaRes, diags := kibanaPayload(ctx, template, plan.Kibana)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}
	result.Resources.Kibana = append(result.Resources.Kibana, kibanaRes...)

	apms, diags := apmsPayload(ctx, template, plan.Apm)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}
	result.Resources.Apm = append(result.Resources.Apm, apms...)

	integrationsServerRes, diags := integrationsServerPayload(ctx, template, plan.IntegrationsServer)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}
	result.Resources.IntegrationsServer = append(result.Resources.IntegrationsServer, integrationsServerRes...)

	enterpriseSearchRes, diags := enterpriseSearchesPayload(ctx, template, plan.EnterpriseSearch)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}
	result.Resources.EnterpriseSearch = append(result.Resources.EnterpriseSearch, enterpriseSearchRes...)

	observability, diags := observabilityPayload(ctx, client, plan.Observability)
	if diags.HasError() {
		diagsnostics.Append(diags...)
	}
	result.Settings.Observability = observability

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

	return &result, nil
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
	if oldV.LT(dataTiersVersion) && newV.GE(dataTiersVersion) {
		return false, nil
	}

	// When any topology elements in the state have the node_type_*
	// properties set, the node_role field cannot be used, since
	// we'd be changing the version AND migrating over `node_role`s
	// which is not permitted by the API.
	var hasNodeTypeSet bool
	var es ElasticsearchTF
	if diags := tfsdk.ValueAs(ctx, plan.Elasticsearch.Elems[0], &es); diags.HasError() {
		return false, diags
	}

	var esTopologies []ElasticsearchTopologyTF
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

func ensurePartialSnapshotStrategy(ess []*models.ElasticsearchPayload) {
	for _, es := range ess {
		transient := es.Plan.Transient
		if transient == nil || transient.RestoreSnapshot == nil {
			continue
		}
		transient.RestoreSnapshot.Strategy = "partial"
	}
}
