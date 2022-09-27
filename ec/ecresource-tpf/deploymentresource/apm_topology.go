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

func NewApmTopology(in *models.ApmTopologyElement) (*Topology, error) {
	var top Topology

	top.InstanceConfigurationId = types.String{Value: in.InstanceConfigurationID}

	if in.Size != nil {
		top.Size = types.String{Value: util.MemoryToState(*in.Size.Value)}
		top.SizeResource = types.String{Value: *in.Size.Resource}
	}

	top.ZoneCount = types.Int64{Value: int64(in.ZoneCount)}

	return &top, nil
}

type ApmTopologies []*Topology

func NewApmTopologies(in []*models.ApmTopologyElement) (ApmTopologies, error) {
	if len(in) == 0 {
		return nil, nil
	}

	tops := make([]*Topology, 0, len(in))
	for _, model := range in {
		if model.Size == nil || model.Size.Value == nil || *model.Size.Value == 0 {
			continue
		}

		top, err := NewApmTopology(model)
		if err != nil {
			return nil, err
		}

		tops = append(tops, top)
	}

	return tops, nil
}

func (tops ApmTopologies) Payload(planModels []*models.ApmTopologyElement) ([]*models.ApmTopologyElement, error) {
	if len(tops) == 0 {
		return defaultApmTopology(planModels), nil
	}

	payloads := make([]*models.ApmTopologyElement, 0, len(tops))

	planModels = defaultApmTopology(planModels)

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

		if err != nil {
			return nil, err
		}

		elem, err := matchApmTopology(icID, planModels)
		if err != nil {
			return nil, err
		}
		if size != nil {
			elem.Size = size
		}

		if topology.ZoneCount.Value > 0 {
			elem.ZoneCount = int32(topology.ZoneCount.Value)
		}

		payloads = append(payloads, elem)
	}

	return payloads, nil
}

// defaultApmTopology iterates over all the templated topology elements and
// sets the size to the default when the template size is smaller than the
// deployment template default, the same is done on the ZoneCount.
func defaultApmTopology(topology []*models.ApmTopologyElement) []*models.ApmTopologyElement {
	for _, t := range topology {
		if *t.Size.Value < minimumApmSize {
			t.Size.Value = ec.Int32(minimumApmSize)
		}
		if t.ZoneCount < minimumZoneCount {
			t.ZoneCount = minimumZoneCount
		}
	}

	return topology
}

func matchApmTopology(id string, topologies []*models.ApmTopologyElement) (*models.ApmTopologyElement, error) {
	for _, t := range topologies {
		if t.InstanceConfigurationID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf(
		`apm topology: invalid instance_configuration_id: "%s" doesn't match any of the deployment template instance configurations`,
		id,
	)
}
