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
	"strconv"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

type ElasticsearchTF struct {
	Autoscale      types.String `tfsdk:"autoscale"`
	RefId          types.String `tfsdk:"ref_id"`
	ResourceId     types.String `tfsdk:"resource_id"`
	Region         types.String `tfsdk:"region"`
	CloudID        types.String `tfsdk:"cloud_id"`
	HttpEndpoint   types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint  types.String `tfsdk:"https_endpoint"`
	Topology       types.List   `tfsdk:"topology"`
	Config         types.List   `tfsdk:"config"`
	RemoteCluster  types.Set    `tfsdk:"remote_cluster"`
	SnapshotSource types.List   `tfsdk:"snapshot_source"`
	Extension      types.Set    `tfsdk:"extension"`
	TrustAccount   types.Set    `tfsdk:"trust_account"`
	TrustExternal  types.Set    `tfsdk:"trust_external"`
	Strategy       types.List   `tfsdk:"strategy"`
}

type Elasticsearch struct {
	Autoscale      *string                      `tfsdk:"autoscale"`
	RefId          *string                      `tfsdk:"ref_id"`
	ResourceId     *string                      `tfsdk:"resource_id"`
	Region         *string                      `tfsdk:"region"`
	CloudID        *string                      `tfsdk:"cloud_id"`
	HttpEndpoint   *string                      `tfsdk:"http_endpoint"`
	HttpsEndpoint  *string                      `tfsdk:"https_endpoint"`
	Topology       ElasticsearchTopologies      `tfsdk:"topology"`
	Config         ElasticsearchConfigs         `tfsdk:"config"`
	RemoteCluster  ElasticsearchRemoteClusters  `tfsdk:"remote_cluster"`
	SnapshotSource ElasticsearchSnapshotSources `tfsdk:"snapshot_source"`
	Extension      ElasticsearchExtensions      `tfsdk:"extension"`
	TrustAccount   ElasticsearchTrustAccounts   `tfsdk:"trust_account"`
	TrustExternal  ElasticsearchTrustExternals  `tfsdk:"trust_external"`
	Strategy       ElasticsearchStrategies      `tfsdk:"strategy"`
}

type ElasticsearchesTF []*ElasticsearchTF

type Elasticsearches []Elasticsearch

func readElasticsearches(in []*models.ElasticsearchResourceInfo, remotes *models.RemoteResources) (Elasticsearches, error) {
	if len(in) == 0 {
		return nil, nil
	}

	ess := make([]Elasticsearch, 0, len(in))

	for _, model := range in {
		if util.IsCurrentEsPlanEmpty(model) || isEsResourceStopped(model) {
			continue
		}
		es, err := readElasticsearch(model, remotes)
		if err != nil {
			return nil, err
		}
		ess = append(ess, *es)
	}

	return ess, nil
}

func elasticsearchPayload(ctx context.Context, template *models.DeploymentTemplateInfoV2, dtID, version string, useNodeRoles bool, ess types.List, skipTopologies bool) ([]*models.ElasticsearchPayload, diag.Diagnostics) {
	if len(ess.Elems) == 0 {
		return nil, nil
	}

	templatePayload := enrichElasticsearchTemplate(esResource(template), dtID, version, useNodeRoles)

	payloads := make([]*models.ElasticsearchPayload, 0, len(ess.Elems))

	for _, elem := range ess.Elems {
		var es ElasticsearchTF
		if diags := tfsdk.ValueAs(ctx, elem, &es); diags.HasError() {
			return nil, diags
		}
		payload, diags := es.Payload(ctx, templatePayload, skipTopologies)
		if diags.HasError() {
			return nil, diags
		}
		payloads = append(payloads, payload)
	}

	return payloads, nil
}

func readElasticsearch(in *models.ElasticsearchResourceInfo, remotes *models.RemoteResources) (*Elasticsearch, error) {
	var es Elasticsearch

	if util.IsCurrentEsPlanEmpty(in) || isEsResourceStopped(in) {
		return &es, nil
	}

	if in.Info.ClusterID != nil && *in.Info.ClusterID != "" {
		es.ResourceId = in.Info.ClusterID
	}

	if in.RefID != nil && *in.RefID != "" {
		es.RefId = in.RefID
	}

	if in.Region != nil {
		es.Region = in.Region
	}

	plan := in.Info.PlanInfo.Current.Plan
	var err error

	topologies, err := readElasticsearchTopologies(plan.ClusterTopology, plan.AutoscalingEnabled != nil && *plan.AutoscalingEnabled)
	if err != nil {
		return nil, err
	}
	es.Topology = topologies

	if plan.AutoscalingEnabled != nil {
		es.Autoscale = ec.String(strconv.FormatBool(*plan.AutoscalingEnabled))
	}

	if meta := in.Info.Metadata; meta != nil && meta.CloudID != "" {
		es.CloudID = &meta.CloudID
	}

	es.HttpEndpoint, es.HttpsEndpoint = converters.ExtractEndpoints(in.Info.Metadata)

	configs, err := readElasticsearchConfigs(plan.Elasticsearch)
	if err != nil {
		return nil, err
	}
	es.Config = configs

	clusters, err := readElasticsearchRemoteClusters(remotes.Resources)
	if err != nil {
		return nil, err
	}
	es.RemoteCluster = clusters

	extensions, err := readElasticsearchExtensions(plan.Elasticsearch)
	if err != nil {
		return nil, err
	}
	es.Extension = extensions

	accounts, err := readElasticsearchTrustAccounts(in.Info.Settings.Trust)
	if err != nil {
		return nil, err
	}
	es.TrustAccount = accounts

	externals, err := readElasticsearchTrustExternals(in.Info.Settings.Trust)
	if err != nil {
		return nil, err
	}
	es.TrustExternal = externals

	return &es, nil
}

func (es *ElasticsearchTF) Payload(ctx context.Context, res *models.ElasticsearchPayload, skipTopologies bool) (*models.ElasticsearchPayload, diag.Diagnostics) {
	if !es.RefId.IsNull() {
		res.RefID = &es.RefId.Value
	}

	if es.Region.Value != "" {
		res.Region = &es.Region.Value
	}

	// Unsetting the curation properties is since they're deprecated since
	// >= 6.6.0 which is when ILM is introduced in Elasticsearch.
	unsetElasticsearchCuration(res)

	var diags diag.Diagnostics

	if !skipTopologies {
		res.Plan.ClusterTopology, diags = ElasticsearchTopologiesTF(es.Topology).Payload(ctx, res.Plan.ClusterTopology)
		if diags.HasError() {
			return nil, diags
		}
	}

	// Fixes the node_roles field to remove the dedicated tier roles from the
	// list when these are set as a dedicated tier as a topology element.
	updateNodeRolesOnDedicatedTiers(res.Plan.ClusterTopology)

	res.Plan.Elasticsearch, diags = ElasticsearchConfigsTF(es.Config).Payload(ctx, res.Plan.Elasticsearch)
	if diags.HasError() {
		return nil, diags
	}

	res.Plan.Transient, diags = ElasticsearchSnapshotSourcesTF(es.SnapshotSource).Payload(ctx)
	if diags.HasError() {
		return nil, diags
	}

	if diags := ElasticsearchExtensionsTF(es.Extension).Payload(ctx, res.Plan.Elasticsearch); diags.HasError() {
		return nil, diags
	}

	if es.Autoscale.Value != "" {
		autoscaleBool, err := strconv.ParseBool(es.Autoscale.Value)
		if err != nil {
			diags.AddError("failed parsing autoscale value", err.Error())
			return nil, diags
		}
		res.Plan.AutoscalingEnabled = &autoscaleBool
	}

	settings, diags := ElasticsearchTrustAccountsTF(es.TrustAccount).Payload(ctx, res.Settings)
	if diags.HasError() {
		return nil, diags
	}
	if settings != nil {
		res.Settings = settings
	}

	settings, diags = ElasticsearchTrustExternalsTF(es.TrustExternal).Payload(ctx, res.Settings)
	if diags.HasError() {
		return nil, diags
	}
	if settings != nil {
		res.Settings = settings
	}

	res.Plan.Transient, diags = ElasticsearchStrategiesTF(es.Strategy).Payload(ctx, res.Plan.Transient)

	if diags.HasError() {
		return nil, diags
	}

	return res, nil
}

func enrichElasticsearchTemplate(tpl *models.ElasticsearchPayload, dt, version string, useNodeRoles bool) *models.ElasticsearchPayload {
	if tpl.Plan.DeploymentTemplate == nil {
		tpl.Plan.DeploymentTemplate = &models.DeploymentTemplateReference{}
	}

	if tpl.Plan.DeploymentTemplate.ID == nil || *tpl.Plan.DeploymentTemplate.ID == "" {
		tpl.Plan.DeploymentTemplate.ID = ec.String(dt)
	}

	if tpl.Plan.Elasticsearch.Version == "" {
		tpl.Plan.Elasticsearch.Version = version
	}

	for _, topology := range tpl.Plan.ClusterTopology {
		if useNodeRoles {
			topology.NodeType = nil
			continue
		}
		topology.NodeRoles = nil
	}

	return tpl
}

func esResource(res *models.DeploymentTemplateInfoV2) *models.ElasticsearchPayload {
	if len(res.DeploymentTemplate.Resources.Elasticsearch) == 0 {
		return &models.ElasticsearchPayload{
			Plan: &models.ElasticsearchClusterPlan{
				Elasticsearch: &models.ElasticsearchConfiguration{},
			},
			Settings: &models.ElasticsearchClusterSettings{},
		}
	}
	return res.DeploymentTemplate.Resources.Elasticsearch[0]
}

func unsetElasticsearchCuration(payload *models.ElasticsearchPayload) {
	if payload.Plan.Elasticsearch != nil {
		payload.Plan.Elasticsearch.Curation = nil
	}

	if payload.Settings != nil {
		payload.Settings.Curation = nil
	}
}

func updateNodeRolesOnDedicatedTiers(topologies []*models.ElasticsearchClusterTopologyElement) {
	dataTier, hasMasterTier, hasIngestTier := dedicatedTopoogies(topologies)
	// This case is not very likely since all deployments will have a data tier.
	// It's here because the code path is technically possible and it's better
	// than a straight panic.
	if dataTier == nil {
		return
	}

	if hasIngestTier {
		dataTier.NodeRoles = removeItemFromSlice(
			dataTier.NodeRoles, ingestDataTierRole,
		)
	}
	if hasMasterTier {
		dataTier.NodeRoles = removeItemFromSlice(
			dataTier.NodeRoles, masterDataTierRole,
		)
	}
}

func removeItemFromSlice(slice []string, item string) []string {
	var hasItem bool
	var itemIndex int
	for i, str := range slice {
		if str == item {
			hasItem = true
			itemIndex = i
		}
	}
	if hasItem {
		copy(slice[itemIndex:], slice[itemIndex+1:])
		return slice[:len(slice)-1]
	}
	return slice
}

func dedicatedTopoogies(topologies []*models.ElasticsearchClusterTopologyElement) (dataTier *models.ElasticsearchClusterTopologyElement, hasMasterTier, hasIngestTier bool) {
	for _, topology := range topologies {
		var hasSomeDataRole bool
		var hasMasterRole bool
		var hasIngestRole bool
		for _, role := range topology.NodeRoles {
			sizeNonZero := *topology.Size.Value > 0
			if strings.HasPrefix(role, dataTierRolePrefix) && sizeNonZero {
				hasSomeDataRole = true
			}
			if role == ingestDataTierRole && sizeNonZero {
				hasIngestRole = true
			}
			if role == masterDataTierRole && sizeNonZero {
				hasMasterRole = true
			}
		}

		if !hasSomeDataRole && hasMasterRole {
			hasMasterTier = true
		}

		if !hasSomeDataRole && hasIngestRole {
			hasIngestTier = true
		}

		if hasSomeDataRole && hasMasterRole {
			dataTier = topology
		}
	}

	return dataTier, hasMasterTier, hasIngestTier
}
