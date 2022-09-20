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
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func NewIntegrationsServerTopology(in *models.IntegrationsServerTopologyElement) (*Topology, error) {
	var top Topology

	top.InstanceConfigurationId.Value = in.InstanceConfigurationID

	if in.Size != nil {
		top.Size.Value = util.MemoryToState(*in.Size.Value)
		top.SizeResource.Value = *in.Size.Resource
	}

	top.ZoneCount.Value = int64(in.ZoneCount)

	return &top, nil
}

func NewIntegrationsServerTopologies(in []*models.IntegrationsServerTopologyElement) ([]Topology, error) {
	if len(in) == 0 {
		return nil, nil
	}

	tops := make([]Topology, 0, len(in))
	for _, model := range in {
		if model.Size == nil || model.Size.Value == nil || *model.Size.Value == 0 {
			continue
		}

		top, err := NewIntegrationsServerTopology(model)
		if err != nil {
			return nil, err
		}

		tops = append(tops, *top)
	}

	return tops, nil
}
