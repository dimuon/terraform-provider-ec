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
	"sort"
	"strconv"
	"strings"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTopology struct {
	Id                      types.String                      `tfsdk:"id"`
	InstanceConfigurationId types.String                      `tfsdk:"instance_configuration_id"`
	Size                    types.String                      `tfsdk:"size"`
	SizeResource            types.String                      `tfsdk:"size_resource"`
	ZoneCount               types.Int64                       `tfsdk:"zone_count"`
	NodeTypeData            types.String                      `tfsdk:"node_type_data"`
	NodeTypeMaster          types.String                      `tfsdk:"node_type_master"`
	NodeTypeIngest          types.String                      `tfsdk:"node_type_ingest"`
	NodeTypeMl              types.String                      `tfsdk:"node_type_ml"`
	NodeRoles               []string                          `tfsdk:"node_roles"`
	Autoscaling             ElasticsearchTopologyAutoscalings `tfsdk:"autoscaling"`
	Config                  ElasticsearchConfigs              `tfsdk:"config"`
}

type ElasticsearchTopologies []*ElasticsearchTopology

func NewElasticsearchTopologies(in []*models.ElasticsearchClusterTopologyElement, autoscaling bool) (ElasticsearchTopologies, error) {
	if len(in) == 0 {
		return nil, nil
	}

	tops := make([]*ElasticsearchTopology, 0, len(in))

	for _, model := range in {
		if !isPotentiallySizedTopology(model, autoscaling) {
			continue
		}
		top, err := NewElasticsearchTopology(model)
		if err != nil {
			return nil, err
		}
		tops = append(tops, top)
	}

	sort.SliceStable(tops, func(i, j int) bool {
		a := (tops)[i]
		b := (tops)[j]
		return a.Id.Value < b.Id.Value
	})

	return tops, nil
}

func (tops ElasticsearchTopologies) Payload(planTopologies []*models.ElasticsearchClusterTopologyElement) ([]*models.ElasticsearchClusterTopologyElement, error) {
	payload := planTopologies

	for _, topology := range tops {

		topologyID := topology.Id.Value

		size, err := converters.ParseTopologySize(topology.Size, topology.SizeResource)
		if err != nil {
			return nil, err
		}

		elem, err := matchEsTopologyID(topologyID, planTopologies)
		if err != nil {
			return nil, fmt.Errorf("elasticsearch topology %s: %w", topologyID, err)
		}
		if size != nil {
			elem.Size = size
		}

		if topology.ZoneCount.Value > 0 {
			elem.ZoneCount = int32(topology.ZoneCount.Value)
		}

		if err := topology.ParseLegacyNodeType(elem.NodeType); err != nil {
			return nil, err
		}

		if len(topology.NodeRoles) > 0 {
			elem.NodeRoles = topology.NodeRoles
			elem.NodeType = nil
		}

		topology.Autoscaling.Payload(topologyID, elem)

		if elem.Elasticsearch, err = topology.Config.Payload(elem.Elasticsearch); err != nil {
			return nil, err
		}
	}

	return payload, nil
}

func NewElasticsearchTopology(topology *models.ElasticsearchClusterTopologyElement) (*ElasticsearchTopology, error) {
	var top ElasticsearchTopology

	top.Id.Value = topology.ID

	if topology.InstanceConfigurationID != "" {
		top.InstanceConfigurationId.Value = topology.InstanceConfigurationID
	}

	if topology.Size != nil {
		top.Size.Value = util.MemoryToState(*topology.Size.Value)
		top.SizeResource.Value = *topology.Size.Resource
	}

	top.ZoneCount.Value = int64(topology.ZoneCount)

	if nt := topology.NodeType; nt != nil {
		if nt.Data != nil {
			top.NodeTypeData.Value = strconv.FormatBool(*nt.Data)
		}

		if nt.Ingest != nil {
			top.NodeTypeIngest.Value = strconv.FormatBool(*nt.Ingest)
		}

		if nt.Master != nil {
			top.NodeTypeMaster.Value = strconv.FormatBool(*nt.Master)
		}

		if nt.Ml != nil {
			top.NodeTypeMl.Value = strconv.FormatBool(*nt.Ml)
		}
	}

	if len(topology.NodeRoles) > 0 {
		top.NodeRoles = append(top.NodeRoles, topology.NodeRoles...)
	}

	var err error
	if top.Autoscaling, err = NewElasticsearchTopologyAutoscalings(topology); err != nil {
		return &top, err
	}

	if top.Config, err = NewElasticsearchConfigs(topology.Elasticsearch); err != nil {
		return &top, err
	}

	return &top, nil
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
