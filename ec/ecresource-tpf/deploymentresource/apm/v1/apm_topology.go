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

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/topology/v1"
	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	minimumApmSize = 512
)

func ReadApmTopology(in *models.ApmTopologyElement) (*v1.Topology, error) {
	var top v1.Topology

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

func ReadApmTopologies(in []*models.ApmTopologyElement) (v1.Topologies, error) {
	topologies := make([]v1.Topology, 0, len(in))

	for _, model := range in {
		if model.Size == nil || model.Size.Value == nil || *model.Size.Value == 0 {
			continue
		}

		topology, err := ReadApmTopology(model)
		if err != nil {
			return nil, nil
		}

		topologies = append(topologies, *topology)
	}

	// return tfsdk.ValueFrom(ctx, tops, apmTopology().Type(), topologies)
	return topologies, nil
}

func ApmTopologiesPayload(ctx context.Context, planModels []*models.ApmTopologyElement, tops types.List) ([]*models.ApmTopologyElement, diag.Diagnostics) {
	if len(tops.Elems) == 0 {
		return DefaultApmTopology(planModels), nil
	}

	payloads := make([]*models.ApmTopologyElement, 0, len(tops.Elems))

	planModels = DefaultApmTopology(planModels)

	for i, elem := range tops.Elems {
		payload, diags := ApmTopologyPayload(ctx, planModels, i, elem)

		if diags.HasError() {
			return nil, diags
		}

		if payload != nil {
			payloads = append(payloads, payload)
		}
	}

	return payloads, nil
}

func ApmTopologyPayload(ctx context.Context, planModels []*models.ApmTopologyElement, index int, topObj attr.Value) (*models.ApmTopologyElement, diag.Diagnostics) {
	if topObj.IsNull() || topObj.IsUnknown() {
		return nil, nil
	}

	var topology v1.TopologyTF

	if diags := tfsdk.ValueAs(ctx, topObj, &topology); diags.HasError() {
		return nil, diags
	}

	icID := topology.InstanceConfigurationId.Value

	// When a topology element is set but no instance_configuration_id
	// is set, then obtain the instance_configuration_id from the topology
	// element.
	if icID == "" && index < len(planModels) {
		icID = planModels[index].InstanceConfigurationID
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

	return topologyElem, nil
}

// DefaultApmTopology iterates over all the templated topology elements and
// sets the size to the default when the template size is smaller than the
// deployment template default, the same is done on the ZoneCount.
func DefaultApmTopology(topology []*models.ApmTopologyElement) []*models.ApmTopologyElement {
	for _, t := range topology {
		if *t.Size.Value < minimumApmSize {
			t.Size.Value = ec.Int32(minimumApmSize)
		}
		if t.ZoneCount < utils.MinimumZoneCount {
			t.ZoneCount = utils.MinimumZoneCount
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
