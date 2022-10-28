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
	"strconv"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/elasticsearch/v1"
	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

type ElasticsearchTF struct {
	Autoscale        types.String `tfsdk:"autoscale"`
	RefId            types.String `tfsdk:"ref_id"`
	ResourceId       types.String `tfsdk:"resource_id"`
	Region           types.String `tfsdk:"region"`
	CloudID          types.String `tfsdk:"cloud_id"`
	HttpEndpoint     types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint    types.String `tfsdk:"https_endpoint"`
	HotContentTier   types.Object `tfsdk:"hot_content_tier"`
	CoordinatingTier types.Object `tfsdk:"coordinating_tier"`
	MasterTier       types.Object `tfsdk:"master_tier"`
	WarmTier         types.Object `tfsdk:"warm_tier"`
	ColdTier         types.Object `tfsdk:"cold_tier"`
	FrozenTier       types.Object `tfsdk:"frozen_tier"`
	MlTier           types.Object `tfsdk:"ml_tier"`
	Config           types.List   `tfsdk:"config"`
	RemoteCluster    types.Set    `tfsdk:"remote_cluster"`
	SnapshotSource   types.List   `tfsdk:"snapshot_source"`
	Extension        types.Set    `tfsdk:"extension"`
	TrustAccount     types.Set    `tfsdk:"trust_account"`
	TrustExternal    types.Set    `tfsdk:"trust_external"`
	Strategy         types.List   `tfsdk:"strategy"`
}

type Elasticsearch struct {
	Autoscale        *string                         `tfsdk:"autoscale"`
	RefId            *string                         `tfsdk:"ref_id"`
	ResourceId       *string                         `tfsdk:"resource_id"`
	Region           *string                         `tfsdk:"region"`
	CloudID          *string                         `tfsdk:"cloud_id"`
	HttpEndpoint     *string                         `tfsdk:"http_endpoint"`
	HttpsEndpoint    *string                         `tfsdk:"https_endpoint"`
	HotContentTier   *v1.ElasticsearchTopology       `tfsdk:"hot_content_tier"`
	CoordinatingTier *v1.ElasticsearchTopology       `tfsdk:"coordinating_tier"`
	MasterTier       *v1.ElasticsearchTopology       `tfsdk:"master_tier"`
	WarmTier         *v1.ElasticsearchTopology       `tfsdk:"warm_tier"`
	ColdTier         *v1.ElasticsearchTopology       `tfsdk:"cold_tier"`
	FrozenTier       *v1.ElasticsearchTopology       `tfsdk:"frozen_tier"`
	MlTielr          *v1.ElasticsearchTopology       `tfsdk:"ml_tier"`
	Config           v1.ElasticsearchConfigs         `tfsdk:"config"`
	RemoteCluster    v1.ElasticsearchRemoteClusters  `tfsdk:"remote_cluster"`
	SnapshotSource   v1.ElasticsearchSnapshotSources `tfsdk:"snapshot_source"`
	Extension        v1.ElasticsearchExtensions      `tfsdk:"extension"`
	TrustAccount     v1.ElasticsearchTrustAccounts   `tfsdk:"trust_account"`
	TrustExternal    v1.ElasticsearchTrustExternals  `tfsdk:"trust_external"`
	Strategy         v1.ElasticsearchStrategiesV1    `tfsdk:"strategy"`
}

type Elasticsearches []Elasticsearch

func ReadElasticsearches(in []*models.ElasticsearchResourceInfo, remotes *models.RemoteResources) (Elasticsearches, error) {
	for _, model := range in {
		if util.IsCurrentEsPlanEmpty(model) || utils.IsEsResourceStopped(model) {
			continue
		}
		es, err := ReadElasticsearch(model, remotes)
		if err != nil {
			return nil, err
		}
		return Elasticsearches{*es}, nil
	}

	return nil, nil
}

func ElasticsearchPayload(ctx context.Context, list types.List, template *models.DeploymentTemplateInfoV2, dtID, version string, useNodeRoles bool, skipTopologies bool) (*models.ElasticsearchPayload, diag.Diagnostics) {
	var es *ElasticsearchTF

	if diags := utils.GetFirst(ctx, list, &es); diags.HasError() {
		return nil, diags
	}

	if es == nil {
		var diags diag.Diagnostics
		diags.AddError("Elasticsearch payload error", "cannot find elasticsearch data")
		return nil, diags
	}

	templatePayload := enrichElasticsearchTemplate(esResource(template), dtID, version, useNodeRoles)

	payload, diags := es.Payload(ctx, templatePayload, skipTopologies)
	if diags.HasError() {
		return nil, diags
	}

	return payload, nil
}

func ReadElasticsearch(in *models.ElasticsearchResourceInfo, remotes *models.RemoteResources) (*Elasticsearch, error) {
	var es Elasticsearch

	if util.IsCurrentEsPlanEmpty(in) || utils.IsEsResourceStopped(in) {
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

	topologies, err := v1.ReadElasticsearchTopologies(plan)
	if err != nil {
		return nil, err
	}
	es.setTopology(topologies)

	if plan.AutoscalingEnabled != nil {
		es.Autoscale = ec.String(strconv.FormatBool(*plan.AutoscalingEnabled))
	}

	if meta := in.Info.Metadata; meta != nil && meta.CloudID != "" {
		es.CloudID = &meta.CloudID
	}

	es.HttpEndpoint, es.HttpsEndpoint = converters.ExtractEndpoints(in.Info.Metadata)

	es.Config, err = v1.ReadElasticsearchConfig(plan.Elasticsearch)
	if err != nil {
		return nil, err
	}

	clusters, err := v1.ReadElasticsearchRemoteClusters(remotes.Resources)
	if err != nil {
		return nil, err
	}
	es.RemoteCluster = clusters

	extensions, err := v1.ReadElasticsearchExtensions(plan.Elasticsearch)
	if err != nil {
		return nil, err
	}
	es.Extension = extensions

	accounts, err := v1.ReadElasticsearchTrustAccounts(in.Info.Settings)
	if err != nil {
		return nil, err
	}
	es.TrustAccount = accounts

	externals, err := v1.ReadElasticsearchTrustExternals(in.Info.Settings)
	if err != nil {
		return nil, err
	}
	es.TrustExternal = externals

	return &es, nil
}

func (es *ElasticsearchTF) Payload(ctx context.Context, res *models.ElasticsearchPayload, skipTopologies bool) (*models.ElasticsearchPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !es.RefId.IsNull() {
		res.RefID = &es.RefId.Value
	}

	if es.Region.Value != "" {
		res.Region = &es.Region.Value
	}

	// Unsetting the curation properties is since they're deprecated since
	// >= 6.6.0 which is when ILM is introduced in Elasticsearch.
	unsetElasticsearchCuration(res)

	var ds diag.Diagnostics

	if !skipTopologies {
		diags.Append(es.topologiesPayload(ctx, res.Plan.ClusterTopology)...)
	}

	// Fixes the node_roles field to remove the dedicated tier roles from the
	// list when these are set as a dedicated tier as a topology element.
	updateNodeRolesOnDedicatedTiers(res.Plan.ClusterTopology)

	res.Plan.Elasticsearch, ds = v1.ElasticsearchConfigPayload(ctx, es.Config, res.Plan.Elasticsearch)
	diags = append(diags, ds...)

	diags.Append(v1.ElasticsearchSnapshotSourcePayload(ctx, es.SnapshotSource, res.Plan)...)

	diags.Append(v1.ElasticsearchExtensionPayload(ctx, es.Extension, res.Plan.Elasticsearch)...)

	if es.Autoscale.Value != "" {
		autoscaleBool, err := strconv.ParseBool(es.Autoscale.Value)
		if err != nil {
			diags.AddError("failed parsing autoscale value", err.Error())
		} else {
			res.Plan.AutoscalingEnabled = &autoscaleBool
		}
	}

	res.Settings, ds = v1.ElasticsearchTrustAccountPayload(ctx, es.TrustAccount, res.Settings)
	diags.Append(ds...)

	res.Settings, ds = v1.ElasticsearchTrustExternalPayload(ctx, es.TrustExternal, res.Settings)
	diags.Append(ds...)

	diags.Append(v1.ElasticsearchStrategyPayload(ctx, es.Strategy, res.Plan)...)

	return res, diags
}

func (es *ElasticsearchTF) topologiesPayload(ctx context.Context, topologies []*models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(topologyPayload(ctx, es.HotContentTier, topologies)...)
	diags.Append(topologyPayload(ctx, es.CoordinatingTier, topologies)...)
	diags.Append(topologyPayload(ctx, es.MasterTier, topologies)...)
	diags.Append(topologyPayload(ctx, es.WarmTier, topologies)...)
	diags.Append(topologyPayload(ctx, es.ColdTier, topologies)...)
	diags.Append(topologyPayload(ctx, es.FrozenTier, topologies)...)
	diags.Append(topologyPayload(ctx, es.MlTier, topologies)...)

	return diags
}

func topologyPayload(ctx context.Context, topologyObj types.Object, topologies []*models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	var diags diag.Diagnostics

	if !topologyObj.IsNull() && !topologyObj.IsUnknown() {
		var topology v1.ElasticsearchTopologyTF

		ds := tfsdk.ValueAs(ctx, topologyObj, &topology)
		diags.Append(ds...)

		if !ds.HasError() {
			diags.Append(topology.Payload(ctx, topologies)...)
		}
	}

	return diags
}

func enrichElasticsearchTemplate(tpl *models.ElasticsearchPayload, templateId, version string, useNodeRoles bool) *models.ElasticsearchPayload {
	if tpl.Plan.DeploymentTemplate == nil {
		tpl.Plan.DeploymentTemplate = &models.DeploymentTemplateReference{}
	}

	if tpl.Plan.DeploymentTemplate.ID == nil || *tpl.Plan.DeploymentTemplate.ID == "" {
		tpl.Plan.DeploymentTemplate.ID = ec.String(templateId)
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
	if res == nil || len(res.DeploymentTemplate.Resources.Elasticsearch) == 0 {
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

func (es *Elasticsearch) setTopology(topologies v1.ElasticsearchTopologies) {
	set := topologies.Set()

	for id, topology := range set {
		switch id {
		case "hot_tier":
			es.HotContentTier = &topology
		case "coordinating":
			es.CoordinatingTier = &topology
		case "master":
			es.MasterTier = &topology
		case "warm":
			es.WarmTier = &topology
		case "cold":
			es.ColdTier = &topology
		case "frozen":
			es.FrozenTier = &topology
		case "ml":
			es.MlTielr = &topology
		}
	}
}
