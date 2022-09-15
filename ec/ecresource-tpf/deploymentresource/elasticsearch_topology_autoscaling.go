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
	"encoding/json"
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTopologyAutoscalings []ElasticsearchTopologyAutoscaling

func (autos *ElasticsearchTopologyAutoscalings) fromModel(in *models.ElasticsearchClusterTopologyElement) error {
	var auto ElasticsearchTopologyAutoscaling
	auto.fromModel(in)

	*autos = nil

	if auto != (ElasticsearchTopologyAutoscaling{}) {
		*autos = []ElasticsearchTopologyAutoscaling{auto}
	}

	return nil
}

type ElasticsearchTopologyAutoscaling struct {
	MaxSizeResource    types.String `tfsdk:"max_size_resource"`
	MaxSize            types.String `tfsdk:"max_size"`
	MinSizeResource    types.String `tfsdk:"min_size_resource"`
	MinSize            types.String `tfsdk:"min_size"`
	PolicyOverrideJson types.String `tfsdk:"policy_override_json"`
}

func (a *ElasticsearchTopologyAutoscaling) fromModel(topology *models.ElasticsearchClusterTopologyElement) error {
	if ascale := topology.AutoscalingMax; ascale != nil {
		a.MaxSizeResource.Value = *ascale.Resource
		a.MaxSize.Value = util.MemoryToState(*ascale.Value)
	}

	if ascale := topology.AutoscalingMin; ascale != nil {
		a.MinSizeResource.Value = *ascale.Resource
		a.MinSize.Value = util.MemoryToState(*ascale.Value)
	}

	if topology.AutoscalingPolicyOverrideJSON != nil {
		b, err := json.Marshal(topology.AutoscalingPolicyOverrideJSON)
		if err != nil {
			return fmt.Errorf(
				"elasticsearch topology %s: unable to persist policy_override_json: %w",
				topology.ID, err,
			)
		}
		a.PolicyOverrideJson.Value = string(b)
	}

	return nil
}
