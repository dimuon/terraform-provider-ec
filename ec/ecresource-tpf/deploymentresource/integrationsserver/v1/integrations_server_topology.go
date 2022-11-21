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

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	topologyv1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/topology/v1"
	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	minimumIntegrationsServerSize = 1024
)

func ReadIntegrationsServerTopology(in *models.IntegrationsServerTopologyElement) (*topologyv1.Topology, error) {
	var top topologyv1.Topology

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

func ReadIntegrationsServerTopologies(in []*models.IntegrationsServerTopologyElement) (topologyv1.Topologies, error) {
	if len(in) == 0 {
		return nil, nil
	}

	tops := make(topologyv1.Topologies, 0, len(in))
	for _, model := range in {
		if model.Size == nil || model.Size.Value == nil || *model.Size.Value == 0 {
			continue
		}

		top, err := ReadIntegrationsServerTopology(model)
		if err != nil {
			return nil, err
		}

		tops = append(tops, *top)
	}

	return tops, nil
}

func IntegrationsServerTopologiesPayload(ctx context.Context, planModels []*models.IntegrationsServerTopologyElement, tops *types.List) ([]*models.IntegrationsServerTopologyElement, diag.Diagnostics) {
	if len(tops.Elems) == 0 {
		return DefaultIntegrationsServerTopology(planModels), nil
	}

	planModels = DefaultIntegrationsServerTopology(planModels)

	payloads := make([]*models.IntegrationsServerTopologyElement, 0, len(tops.Elems))

	for i, elem := range tops.Elems {
		payload, diags := IntegrationsServerTopologyPayload(ctx, planModels, i, elem)

		if diags.HasError() {
			return nil, diags
		}

		if payload != nil {
			payloads = append(payloads, payload)
		}
	}

	return payloads, nil
}

// DefaultIntegrationsServerTopology iterates over all the templated topology elements and
// sets the size to the default when the template size is smaller than the
// deployment template default, the same is done on the ZoneCount.
func DefaultIntegrationsServerTopology(topology []*models.IntegrationsServerTopologyElement) []*models.IntegrationsServerTopologyElement {
	for _, t := range topology {
		if *t.Size.Value < minimumIntegrationsServerSize {
			t.Size.Value = ec.Int32(minimumIntegrationsServerSize)
		}
		if t.ZoneCount < utils.MinimumZoneCount {
			t.ZoneCount = utils.MinimumZoneCount
		}
	}

	return topology
}

func IntegrationsServerTopologyPayload(ctx context.Context, planModels []*models.IntegrationsServerTopologyElement, index int, topObj attr.Value) (*models.IntegrationsServerTopologyElement, diag.Diagnostics) {
	if topObj.IsNull() || topObj.IsUnknown() {
		return nil, nil
	}

	var top topologyv1.TopologyTF

	if diags := tfsdk.ValueAs(ctx, topObj, &top); diags.HasError() {
		return nil, diags
	}

	return top.IntegrationsServerTopologyPayload(ctx, planModels, index)
}
