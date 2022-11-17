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
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TopologyTF struct {
	InstanceConfigurationId types.String `tfsdk:"instance_configuration_id"`
	Size                    types.String `tfsdk:"size"`
	SizeResource            types.String `tfsdk:"size_resource"`
	ZoneCount               types.Int64  `tfsdk:"zone_count"`
}

type Topology struct {
	InstanceConfigurationId *string `tfsdk:"instance_configuration_id"`
	Size                    *string `tfsdk:"size"`
	SizeResource            *string `tfsdk:"size_resource"`
	ZoneCount               int     `tfsdk:"zone_count"`
}

type Topologies []Topology

func (topology TopologyTF) ApmTopologyPayload(ctx context.Context, planModels []*models.ApmTopologyElement, index int) (*models.ApmTopologyElement, diag.Diagnostics) {

	icID := topology.InstanceConfigurationId.Value

	// When a topology element is set but no instance_configuration_id
	// is set, then obtain the instance_configuration_id from the topology
	// element.
	if icID == "" && index < len(planModels) {
		icID = planModels[index].InstanceConfigurationID
	}

	size, err := converters.ParseTopologySizeTF(topology.Size, topology.SizeResource)

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

func (topology TopologyTF) IntegrationsServerTopologyPayload(ctx context.Context, planModels []*models.IntegrationsServerTopologyElement, index int) (*models.IntegrationsServerTopologyElement, diag.Diagnostics) {

	icID := topology.InstanceConfigurationId.Value

	// When a topology element is set but no instance_configuration_id
	// is set, then obtain the instance_configuration_id from the topology
	// element.
	if icID == "" && index < len(planModels) {
		icID = planModels[index].InstanceConfigurationID
	}

	var diags diag.Diagnostics

	size, err := converters.ParseTopologySizeTF(topology.Size, topology.SizeResource)
	if err != nil {
		diags.AddError("parse topology error", err.Error())
		return nil, diags
	}

	elem, err := matchIntegrationsServerTopology(icID, planModels)
	if err != nil {
		diags.AddError("integrations_server topology payload error", err.Error())
		return nil, diags
	}

	if size != nil {
		elem.Size = size
	}

	if topology.ZoneCount.Value > 0 {
		elem.ZoneCount = int32(topology.ZoneCount.Value)
	}

	return elem, nil
}

func matchIntegrationsServerTopology(id string, topologies []*models.IntegrationsServerTopologyElement) (*models.IntegrationsServerTopologyElement, error) {
	for _, t := range topologies {
		if t.InstanceConfigurationID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf(
		`invalid instance_configuration_id: "%s" doesn't match any of the deployment template instance configurations`,
		id,
	)
}
