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

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	minimumKibanaSize             = 1024
	minimumApmSize                = 512
	minimumEnterpriseSearchSize   = 2048
	minimumIntegrationsServerSize = 1024

	minimumZoneCount = 1
)

type EnterpriseSearchTopology struct {
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
	NodeTypeAppserver       types.Bool   `tfsdk:"node_type_appserver"`
	NodeTypeConnector       types.Bool   `tfsdk:"node_type_connector"`
	NodeTypeWorker          types.Bool   `tfsdk:"node_type_worker"`
}

func NewEnterpriseSearchTopology(in *models.EnterpriseSearchTopologyElement) (*EnterpriseSearchTopology, error) {
	var top EnterpriseSearchTopology

	top.InstanceConfigurationId.Value = in.InstanceConfigurationID

	if in.Size != nil {
		top.Size.Value = util.MemoryToState(*in.Size.Value)
		top.SizeResource.Value = *in.Size.Resource
	}

	if nt := in.NodeType; nt != nil {
		if nt.Appserver != nil {
			top.NodeTypeAppserver.Value = *nt.Appserver
		}

		if nt.Connector != nil {
			top.NodeTypeConnector.Value = *nt.Connector
		}

		if nt.Worker != nil {
			top.NodeTypeWorker.Value = *nt.Worker
		}
	}

	top.ZoneCount.Value = int64(in.ZoneCount)

	return &top, nil
}

type EnterpriseSearchTopologies []*EnterpriseSearchTopology

func NewEnterpriseSearchTopologies(in []*models.EnterpriseSearchTopologyElement) ([]*EnterpriseSearchTopology, error) {
	if len(in) == 0 {
		return nil, nil
	}

	tops := make([]*EnterpriseSearchTopology, 0, len(in))
	for _, model := range in {
		if model.Size == nil || model.Size.Value == nil || *model.Size.Value == 0 {
			continue
		}

		top, err := NewEnterpriseSearchTopology(model)
		if err != nil {
			return nil, err
		}

		tops = append(tops, top)
	}

	return tops, nil
}

func (tops EnterpriseSearchTopologies) Payload(planModels []*models.EnterpriseSearchTopologyElement) ([]*models.EnterpriseSearchTopologyElement, error) {
	if len(tops) == 0 {
		return defaultEssTopology(planModels), nil
	}

	planModels = defaultEssTopology(planModels)

	res := make([]*models.EnterpriseSearchTopologyElement, 0, len(tops))

	for i, topology := range tops {
		icID := topology.InstanceConfigurationId.Value

		// When a topology element is set but no instance_configuration_id
		// is set, then obtain the instance_configuration_id from the topology
		// element.
		if icID == "" && i < len(planModels) {
			icID = planModels[i].InstanceConfigurationID
		}

		size, err := converters.ParseTopologySize(topology.Size, topology.SizeResource)
		if err != nil {
			return nil, err
		}

		// Since Enterprise Search is not enabled by default in the template,
		// if the size == nil, it means that the size hasn't been specified in
		// the definition.
		if size == nil {
			size = &models.TopologySize{
				Resource: ec.String("memory"),
				Value:    ec.Int32(minimumEnterpriseSearchSize),
			}
		}

		elem, err := matchEssTopology(icID, planModels)
		if err != nil {
			return nil, err
		}
		if size != nil {
			elem.Size = size
		}

		if topology.ZoneCount.Value > 0 {
			elem.ZoneCount = int32(topology.ZoneCount.Value)
		}

		res = append(res, elem)
	}

	return res, nil
}

// defaultApmTopology iterates over all the templated topology elements and
// sets the size to the default when the template size is smaller than the
// deployment template default, the same is done on the ZoneCount.
func defaultEssTopology(topology []*models.EnterpriseSearchTopologyElement) []*models.EnterpriseSearchTopologyElement {
	for _, t := range topology {
		if *t.Size.Value < minimumEnterpriseSearchSize || *t.Size.Value == 0 {
			t.Size.Value = ec.Int32(minimumEnterpriseSearchSize)
		}
		if t.ZoneCount < minimumZoneCount {
			t.ZoneCount = minimumZoneCount
		}
	}

	return topology
}

func matchEssTopology(id string, topologies []*models.EnterpriseSearchTopologyElement) (*models.EnterpriseSearchTopologyElement, error) {
	for _, t := range topologies {
		if t.InstanceConfigurationID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf(
		`enterprise_search topology: invalid instance_configuration_id: "%s" doesn't match any of the deployment template instance configurations`,
		id,
	)
}
