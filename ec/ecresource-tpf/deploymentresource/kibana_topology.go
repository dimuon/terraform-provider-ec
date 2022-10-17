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

type TopologyTF struct {
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
}

func readKibanaTopology(in *models.KibanaClusterTopologyElement) (*Topology, error) {
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

type KibanaTopologiesTF []*TopologyTF

func readKibanaTopologies(in []*models.KibanaClusterTopologyElement) (Topologies, error) {
	if len(in) == 0 {
		return nil, nil
	}

	tops := make(Topologies, 0, len(in))
	for _, model := range in {
		if model.Size == nil || model.Size.Value == nil || *model.Size.Value == 0 {
			continue
		}

		top, err := readKibanaTopology(model)
		if err != nil {
			return nil, err
		}

		tops = append(tops, *top)
	}

	return tops, nil
}

func kibanaTopologyPayload(ctx context.Context, planModels []*models.KibanaClusterTopologyElement, tops *types.List) ([]*models.KibanaClusterTopologyElement, diag.Diagnostics) {
	if len(tops.Elems) == 0 {
		return defaultKibanaTopology(planModels), nil
	}

	planModels = defaultKibanaTopology(planModels)

	var payloads = make([]*models.KibanaClusterTopologyElement, 0, len(tops.Elems))

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

		elem, err := matchKibanaTopology(icID, planModels)
		if err != nil {
			diags.AddError("kibana topology payload error", err.Error())
			return nil, diags
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
// sets the size to the default when the template size is greater than the
// local terraform default, the same is done on the ZoneCount.
func defaultKibanaTopology(topology []*models.KibanaClusterTopologyElement) []*models.KibanaClusterTopologyElement {
	for _, t := range topology {
		if *t.Size.Value > minimumKibanaSize {
			t.Size.Value = ec.Int32(minimumKibanaSize)
		}
		if t.ZoneCount > minimumZoneCount {
			t.ZoneCount = minimumZoneCount
		}
	}

	return topology
}

func matchKibanaTopology(id string, topologies []*models.KibanaClusterTopologyElement) (*models.KibanaClusterTopologyElement, error) {
	for _, t := range topologies {
		if t.InstanceConfigurationID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf(
		`kibana topology: invalid instance_configuration_id: "%s" doesn't match any of the deployment template instance configurations`,
		id,
	)
}
