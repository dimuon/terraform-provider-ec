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
	"sort"
	"strconv"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTopologyTF struct {
	Id                      types.String `tfsdk:"id"`
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
	NodeTypeData            types.String `tfsdk:"node_type_data"`
	NodeTypeMaster          types.String `tfsdk:"node_type_master"`
	NodeTypeIngest          types.String `tfsdk:"node_type_ingest"`
	NodeTypeMl              types.String `tfsdk:"node_type_ml"`
	NodeRoles               types.Set    `tfsdk:"node_roles"`
	Autoscaling             types.List   `tfsdk:"autoscaling"`
	Config                  types.List   `tfsdk:"config"`
}

type ElasticsearchTopology struct {
	Id                      string                            `tfsdk:"id"`
	InstanceConfigurationId *string                           `tfsdk:"instance_configuration_id"`
	Size                    *string                           `tfsdk:"size"`
	SizeResource            *string                           `tfsdk:"size_resource"`
	ZoneCount               int                               `tfsdk:"zone_count"`
	NodeTypeData            *string                           `tfsdk:"node_type_data"`
	NodeTypeMaster          *string                           `tfsdk:"node_type_master"`
	NodeTypeIngest          *string                           `tfsdk:"node_type_ingest"`
	NodeTypeMl              *string                           `tfsdk:"node_type_ml"`
	NodeRoles               []string                          `tfsdk:"node_roles"`
	Autoscaling             ElasticsearchTopologyAutoscalings `tfsdk:"autoscaling"`
	Config                  ElasticsearchTopologyConfigs      `tfsdk:"config"`
}

type ElasticsearchTopologies []ElasticsearchTopology

func ReadElasticsearchTopologies(in *models.ElasticsearchClusterPlan) (ElasticsearchTopologies, error) {
	if len(in.ClusterTopology) == 0 {
		return nil, nil
	}

	tops := make([]ElasticsearchTopology, 0, len(in.ClusterTopology))

	for _, model := range in.ClusterTopology {
		if !IsPotentiallySizedTopology(model, in.AutoscalingEnabled != nil && *in.AutoscalingEnabled) {
			continue
		}

		topology, err := ReadElasticsearchTopology(model)
		if err != nil {
			return nil, err
		}
		tops = append(tops, *topology)
	}

	sort.SliceStable(tops, func(i, j int) bool {
		a := (tops)[i]
		b := (tops)[j]
		return a.Id < b.Id
	})

	return tops, nil
}

func ElasticsearchTopologiesPayload(ctx context.Context, tops types.List, planTopologies []*models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, elem := range tops.Elems {
		var topology ElasticsearchTopologyTF

		ds := tfsdk.ValueAs(ctx, elem, &topology)
		diags = append(diags, ds...)

		if ds.HasError() {
			continue
		}

		diags = append(diags, topology.Payload(ctx, planTopologies)...)
	}

	return diags
}

func (topology ElasticsearchTopologyTF) Payload(ctx context.Context, planTopologies []*models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	var diags diag.Diagnostics

	topologyID := topology.Id.Value

	topologyElem, err := MatchEsTopologyID(topologyID, planTopologies)
	if err != nil {
		diags.AddError("topology matching error", err.Error())
		return diags
	}

	size, err := converters.ParseTopologySize(topology.Size, topology.SizeResource)
	if err != nil {
		diags.AddError("size parsing error", err.Error())
	}

	if size != nil {
		topologyElem.Size = size
	}

	if topology.ZoneCount.Value > 0 {
		topologyElem.ZoneCount = int32(topology.ZoneCount.Value)
	}

	if err := topology.ParseLegacyNodeType(topologyElem.NodeType); err != nil {
		diags.AddError("topology legacy node type error", err.Error())
	}

	var nodeRoles []string
	ds := topology.NodeRoles.ElementsAs(ctx, &nodeRoles, true)
	diags.Append(ds...)

	if !ds.HasError() && len(nodeRoles) > 0 {
		topologyElem.NodeRoles = nodeRoles
		topologyElem.NodeType = nil
	}

	diags.Append(ElasticsearchTopologyAutoscalingsPayload(ctx, topology.Autoscaling, topologyID, topologyElem)...)

	topologyElem.Elasticsearch, ds = ElasticsearchTopologyConfigsPayload(ctx, topology.Config, topologyElem.Elasticsearch)

	diags = append(diags, ds...)

	return diags
}

func ReadElasticsearchTopology(model *models.ElasticsearchClusterTopologyElement) (*ElasticsearchTopology, error) {
	var topology ElasticsearchTopology

	topology.Id = model.ID

	if model.InstanceConfigurationID != "" {
		topology.InstanceConfigurationId = &model.InstanceConfigurationID
	}

	if model.Size != nil {
		topology.Size = ec.String(util.MemoryToState(*model.Size.Value))
		topology.SizeResource = model.Size.Resource
	}

	topology.ZoneCount = int(model.ZoneCount)

	if nt := model.NodeType; nt != nil {
		if nt.Data != nil {
			topology.NodeTypeData = ec.String(strconv.FormatBool(*nt.Data))
		}

		if nt.Ingest != nil {
			topology.NodeTypeIngest = ec.String(strconv.FormatBool(*nt.Ingest))
		}

		if nt.Master != nil {
			topology.NodeTypeMaster = ec.String(strconv.FormatBool(*nt.Master))
		}

		if nt.Ml != nil {
			topology.NodeTypeMl = ec.String(strconv.FormatBool(*nt.Ml))
		}
	}

	topology.NodeRoles = model.NodeRoles

	autoscaling, err := ReadElasticsearchTopologyAutoscalings(model)
	if err != nil {
		return nil, err
	}
	topology.Autoscaling = autoscaling

	config, err := ReadElasticsearchTopologyConfigs(model.Elasticsearch)
	if err != nil {
		return nil, err
	}
	topology.Config = config

	return &topology, nil
}

func IsPotentiallySizedTopology(topology *models.ElasticsearchClusterTopologyElement, isAutoscaling bool) bool {
	currentlySized := topology.Size != nil && topology.Size.Value != nil && *topology.Size.Value > 0
	canBeSized := isAutoscaling && topology.AutoscalingMax != nil && topology.AutoscalingMax.Value != nil && *topology.AutoscalingMax.Value > 0

	return currentlySized || canBeSized
}

func MatchEsTopologyID(id string, topologies []*models.ElasticsearchClusterTopologyElement) (*models.ElasticsearchClusterTopologyElement, error) {
	for _, t := range topologies {
		if t.ID == id {
			return t, nil
		}
	}

	topIDs := TopologyIDs(topologies)
	for i, id := range topIDs {
		topIDs[i] = "\"" + id + "\""
	}

	return nil, fmt.Errorf(`invalid id: valid topology IDs are %s`,
		strings.Join(topIDs, ", "),
	)
}

func TopologyIDs(topologies []*models.ElasticsearchClusterTopologyElement) []string {
	var result []string

	for _, topology := range topologies {
		result = append(result, topology.ID)
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func (topology *ElasticsearchTopologyTF) ParseLegacyNodeType(nodeType *models.ElasticsearchNodeType) error {
	if nodeType == nil {
		return nil
	}

	if topology.NodeTypeData.Value != "" {
		nt, err := strconv.ParseBool(topology.NodeTypeData.Value)
		if err != nil {
			return fmt.Errorf("failed parsing node_type_data value: %w", err)
		}
		nodeType.Data = &nt
	}

	if topology.NodeTypeMaster.Value != "" {
		nt, err := strconv.ParseBool(topology.NodeTypeMaster.Value)
		if err != nil {
			return fmt.Errorf("failed parsing node_type_master value: %w", err)
		}
		nodeType.Master = &nt
	}

	if topology.NodeTypeIngest.Value != "" {
		nt, err := strconv.ParseBool(topology.NodeTypeIngest.Value)
		if err != nil {
			return fmt.Errorf("failed parsing node_type_ingest value: %w", err)
		}
		nodeType.Ingest = &nt
	}

	if topology.NodeTypeMl.Value != "" {
		nt, err := strconv.ParseBool(topology.NodeTypeMl.Value)
		if err != nil {
			return fmt.Errorf("failed parsing node_type_ml value: %w", err)
		}
		nodeType.Ml = &nt
	}

	return nil
}
