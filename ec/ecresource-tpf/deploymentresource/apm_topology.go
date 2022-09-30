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

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Topology struct {
	InstanceConfigurationId *string `tfsdk:"instance_configuration_id"`
	Size                    *string `tfsdk:"size"`
	SizeResource            *string `tfsdk:"size_resource"`
	ZoneCount               int     `tfsdk:"zone_count"`
}

func readApmTopology(in *models.ApmTopologyElement) (*Topology, error) {
	var top Topology

	if in.InstanceConfigurationID != "" {
		top.InstanceConfigurationId = &in.InstanceConfigurationID
	}

	if in.Size != nil {
		top.Size = ec.String(util.MemoryToState(*in.Size.Value))
		top.SizeResource = ec.String(*in.Size.Resource)
	}

	top.ZoneCount = int(in.ZoneCount)

	return &top, nil
}

type Topologies []Topology

func readApmTopologies(in []*models.ApmTopologyElement) (Topologies, error) {
	topologies := make([]Topology, 0, len(in))

	for _, model := range in {
		if model.Size == nil || model.Size.Value == nil || *model.Size.Value == 0 {
			continue
		}

		topology, err := readApmTopology(model)
		if err != nil {
			return nil, nil
		}

		topologies = append(topologies, *topology)
	}

	// return tfsdk.ValueFrom(ctx, tops, apmTopology().Type(), topologies)
	return topologies, nil
}

type ApmTopologiesTF types.List

func (tops ApmTopologiesTF) Payload(ctx context.Context, planModels []*models.ApmTopologyElement) ([]*models.ApmTopologyElement, diag.Diagnostics) {
	if len(tops.Elems) == 0 {
		return defaultApmTopology(planModels), nil
	}

	payloads := make([]*models.ApmTopologyElement, 0, len(tops.Elems))

	planModels = defaultApmTopology(planModels)

	for i, elem := range tops.Elems {
		var topology TopologyTF
		if diags := tfsdk.ValueAs(ctx, elem, &topology); diags.HasError() {
			return nil, diags
		}
		icID := topology.InstanceConfigurationId.Value

		// When a topology element is set but no instance_configuration_id
		// is set, then obtain the instance_configuration_id from the topology
		// element.
		if icID == "" && i < len(planModels) {
			icID = planModels[i].InstanceConfigurationID
		}

		size, err := converters.ParseTopologySize(topology.Size, topology.SizeResource)

		var diags diag.Diagnostics
		if err != nil {
			diags.AddError("size parsing error", err.Error())
			return nil, diags
		}

		topologyElem, err := matchApmTopology(icID, planModels)
		if err != nil {
			diags.AddError("cannot match topology element", err.Error())
			return nil, diags
		}

		if size != nil {
			topologyElem.Size = size
		}

		if topology.ZoneCount.Value > 0 {
			topologyElem.ZoneCount = int32(topology.ZoneCount.Value)
		}

		payloads = append(payloads, topologyElem)
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
