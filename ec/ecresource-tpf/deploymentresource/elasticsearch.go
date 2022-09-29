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
	"strconv"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

type Elasticsearch struct {
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

type Elasticsearches []*Elasticsearch

func NewElasticsearches(in []*models.ElasticsearchResourceInfo, remotes *models.RemoteResources) (Elasticsearches, error) {
	if len(in) == 0 {
		return nil, nil
	}

	ess := make([]*Elasticsearch, 0, len(in))

	for _, model := range in {
		if util.IsCurrentEsPlanEmpty(model) || isEsResourceStopped(model) {
			continue
		}
		var es *Elasticsearch
		var err error
		if es, err = NewElasticsearch(model, remotes); err != nil {
			return nil, err
		}
		ess = append(ess, es)
	}

	return ess, nil
}

func (ess Elasticsearches) Payload(template *models.DeploymentTemplateInfoV2, dtID, version string, useNodeRoles bool) ([]*models.ElasticsearchPayload, error) {
	ctx := context.TODO()
	if len(ess) == 0 {
		return nil, nil
	}

	templatePayload := enrichElasticsearchTemplate(esResource(template), dtID, version, useNodeRoles)

	payloads := make([]*models.ElasticsearchPayload, 0, len(ess))

	for _, es := range ess {

		payload, diags := es.Payload(ctx, templatePayload)
		if diags.HasError() {
			return nil, fmt.Errorf("error: %v", diags)
		}
		payloads = append(payloads, payload)
	}

	return payloads, nil
}

func NewElasticsearch(in *models.ElasticsearchResourceInfo, remotes *models.RemoteResources) (*Elasticsearch, error) {
	ctx := context.TODO()
	var es Elasticsearch

	if util.IsCurrentEsPlanEmpty(in) || isEsResourceStopped(in) {
		return &es, nil
	}

	if in.Info.ClusterID != nil && *in.Info.ClusterID != "" {
		es.ResourceId = types.String{Value: *in.Info.ClusterID}
	}

	if in.RefID != nil && *in.RefID != "" {
		es.RefId = types.String{Value: *in.RefID}
	}

	if in.Region != nil {
		es.Region = types.String{Value: *in.Region}
	}

	plan := in.Info.PlanInfo.Current.Plan
	var err error

	if diags := ElasticsearchTopologies(es.Topology).Read(ctx, plan.ClusterTopology, plan.AutoscalingEnabled != nil && *plan.AutoscalingEnabled); diags.HasError() {
		return &es, err
	}

	if plan.AutoscalingEnabled != nil {
		es.Autoscale = types.String{Value: strconv.FormatBool(*plan.AutoscalingEnabled)}
	}

	if meta := in.Info.Metadata; meta != nil && meta.CloudID != "" {
		es.CloudID = types.String{Value: meta.CloudID}
	}

	es.HttpEndpoint, es.HttpsEndpoint = converters.ExtractEndpointsTF(in.Info.Metadata)

	if diags := ElasticsearchConfigs(es.Config).Read(ctx, plan.Elasticsearch); diags.HasError() {
		return nil, err
	}

	if diags := ElasticsearchRemoteClusters(es.RemoteCluster).Read(ctx, remotes.Resources); diags.HasError() {
		return nil, err
	}

	if diags := ElasticsearchExtensions(es.Extension).Read(ctx, plan.Elasticsearch); diags.HasError() {
		return nil, err
	}

	if diags := ElasticsearchTrustAccounts(es.TrustAccount).Read(ctx, in.Info.Settings.Trust); diags.HasError() {
		return nil, err
	}

	if diags := ElasticsearchTrustExternals(es.TrustExternal).Read(ctx, in.Info.Settings.Trust); diags.HasError() {
		return nil, err
	}

	return &es, nil
}

func (es *Elasticsearch) Payload(ctx context.Context, res *models.ElasticsearchPayload) (*models.ElasticsearchPayload, diag.Diagnostics) {
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

	res.Plan.ClusterTopology, diags = ElasticsearchTopologies(es.Topology).Payload(ctx, res.Plan.ClusterTopology)
	if diags.HasError() {
		return nil, diags
	}

	// Fixes the node_roles field to remove the dedicated tier roles from the
	// list when these are set as a dedicated tier as a topology element.
	updateNodeRolesOnDedicatedTiers(res.Plan.ClusterTopology)

	res.Plan.Elasticsearch, diags = ElasticsearchConfigs(es.Config).Payload(ctx, res.Plan.Elasticsearch)
	if diags.HasError() {
		return nil, diags
	}

	res.Plan.Transient, diags = ElasticsearchSnapshotSources(es.SnapshotSource).Payload(ctx)
	if diags.HasError() {
		return nil, diags
	}

	if diags := ElasticsearchExtensions(es.Extension).Payload(ctx, res.Plan.Elasticsearch); diags.HasError() {
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

	settings, diags := ElasticsearchTrustAccounts(es.TrustAccount).Payload(ctx, res.Settings)
	if diags.HasError() {
		return nil, diags
	}
	if settings != nil {
		res.Settings = settings
	}

	settings, diags = ElasticsearchTrustExternals(es.TrustExternal).Payload(ctx, res.Settings)
	if diags.HasError() {
		return nil, diags
	}
	if settings != nil {
		res.Settings = settings
	}

	res.Plan.Transient, diags = ElasticsearchStrategies(es.Strategy).Payload(ctx, res.Plan.Transient)

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
