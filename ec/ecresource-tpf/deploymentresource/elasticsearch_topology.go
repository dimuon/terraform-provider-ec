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
	"sort"
	"strconv"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTopology struct {
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

type ElasticsearchTopologies types.List

func (topologies ElasticsearchTopologies) Read(ctx context.Context, in []*models.ElasticsearchClusterTopologyElement, autoscaling bool) diag.Diagnostics {
	if len(in) == 0 {
		return nil
	}

	tops := make([]ElasticsearchTopology, 0, len(in))

	for _, model := range in {
		if !isPotentiallySizedTopology(model, autoscaling) {
			continue
		}
		var top ElasticsearchTopology
		diags := top.Read(ctx, model)
		if diags.HasError() {
			return diags
		}
		tops = append(tops, top)
	}

	sort.SliceStable(tops, func(i, j int) bool {
		a := (tops)[i]
		b := (tops)[j]
		return a.Id.Value < b.Id.Value
	})

	return tfsdk.ValueFrom(ctx, tops, elasticsearchTopology().FrameworkType(), &topologies)
}

func (tops ElasticsearchTopologies) Payload(ctx context.Context, planTopologies []*models.ElasticsearchClusterTopologyElement) ([]*models.ElasticsearchClusterTopologyElement, diag.Diagnostics) {
	payload := planTopologies

	for _, elem := range tops.Elems {

		var topology ElasticsearchTopology

		diags := tfsdk.ValueAs(ctx, elem, &topology)
		if diags.HasError() {
			return nil, diags
		}

		topologyID := topology.Id.Value

		size, err := converters.ParseTopologySize(topology.Size, topology.SizeResource)
		if err != nil {
			diags.AddError("size parsing error", err.Error())
			return nil, diags
		}

		topologyElem, err := matchEsTopologyID(topologyID, planTopologies)
		if err != nil {
			diags.AddError("topology matching error", fmt.Errorf("id %s: %w", topologyID, err).Error())
			return nil, diags
		}
		if size != nil {
			topologyElem.Size = size
		}

		if topology.ZoneCount.Value > 0 {
			topologyElem.ZoneCount = int32(topology.ZoneCount.Value)
		}

		if err := topology.ParseLegacyNodeType(topologyElem.NodeType); err != nil {
			diags.AddError("topology legacy node type error", err.Error())
			return nil, diags
		}

		if len(topology.NodeRoles.Elems) > 0 {
			topologyElem.NodeRoles = make([]string, 0, len(topology.NodeRoles.Elems))
			for _, nd := range topology.NodeRoles.Elems {
				topologyElem.NodeRoles = append(topologyElem.NodeRoles, nd.(types.String).Value)
			}
			topologyElem.NodeType = nil
		}

		ElasticsearchTopologyAutoscalings(topology.Autoscaling).Payload(ctx, topologyID, topologyElem)

		topologyElem.Elasticsearch, diags = ElasticsearchConfigs(topology.Config).Payload(ctx, topologyElem.Elasticsearch)
		if diags.HasError() {
			return nil, diags
		}
	}

	return payload, nil
}

func (topology *ElasticsearchTopology) Read(ctx context.Context, model *models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	topology.Id = types.String{Value: model.ID}

	if model.InstanceConfigurationID != "" {
		topology.InstanceConfigurationId = types.String{Value: model.InstanceConfigurationID}
	}

	if model.Size != nil {
		topology.Size = types.String{Value: util.MemoryToState(*model.Size.Value)}
		topology.SizeResource = types.String{Value: *model.Size.Resource}
	}

	topology.ZoneCount = types.Int64{Value: int64(model.ZoneCount)}

	if nt := model.NodeType; nt != nil {
		if nt.Data != nil {
			topology.NodeTypeData = types.String{Value: strconv.FormatBool(*nt.Data)}
		}

		if nt.Ingest != nil {
			topology.NodeTypeIngest = types.String{Value: strconv.FormatBool(*nt.Ingest)}
		}

		if nt.Master != nil {
			topology.NodeTypeMaster = types.String{Value: strconv.FormatBool(*nt.Master)}
		}

		if nt.Ml != nil {
			topology.NodeTypeMl = types.String{Value: strconv.FormatBool(*nt.Ml)}
		}
	}

	if len(model.NodeRoles) > 0 {
		topology.NodeRoles.Elems = make([]attr.Value, 0, len(model.NodeRoles))
		for _, nd := range model.NodeRoles {
			topology.NodeRoles.Elems = append(topology.NodeRoles.Elems, types.String{Value: nd})
		}
	}

	if diags := ElasticsearchTopologyAutoscalings(topology.Autoscaling).Read(ctx, model); diags.HasError() {
		return diags
	}

	if diags := ElasticsearchConfigs(topology.Config).Read(ctx, model.Elasticsearch); diags.HasError() {
		return diags
	}

	return nil
}

func isPotentiallySizedTopology(topology *models.ElasticsearchClusterTopologyElement, isAutoscaling bool) bool {
	currentlySized := topology.Size != nil && topology.Size.Value != nil && *topology.Size.Value > 0
	canBeSized := isAutoscaling && topology.AutoscalingMax != nil && topology.AutoscalingMax.Value != nil && *topology.AutoscalingMax.Value > 0

	return currentlySized || canBeSized
}

func matchEsTopologyID(id string, topologies []*models.ElasticsearchClusterTopologyElement) (*models.ElasticsearchClusterTopologyElement, error) {
	for _, t := range topologies {
		if t.ID == id {
			return t, nil
		}
	}

	topIDs := topologyIDs(topologies)
	for i, id := range topIDs {
		topIDs[i] = "\"" + id + "\""
	}

	return nil, fmt.Errorf(`invalid id: valid topology IDs are %s`,
		strings.Join(topIDs, ", "),
	)
}

func topologyIDs(topologies []*models.ElasticsearchClusterTopologyElement) []string {
	var result []string

	for _, topology := range topologies {
		result = append(result, topology.ID)
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func (topology *ElasticsearchTopology) ParseLegacyNodeType(nodeType *models.ElasticsearchNodeType) error {
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
