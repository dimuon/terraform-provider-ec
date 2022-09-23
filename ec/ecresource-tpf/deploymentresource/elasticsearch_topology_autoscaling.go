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
	"reflect"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deploymentsize"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTopologyAutoscalings []*ElasticsearchTopologyAutoscaling

func NewElasticsearchTopologyAutoscalings(in *models.ElasticsearchClusterTopologyElement) (ElasticsearchTopologyAutoscalings, error) {
	auto, err := NewElasticsearchTopologyAutoscaling(in)
	if err != nil {
		return nil, err
	}

	if *auto != (ElasticsearchTopologyAutoscaling{}) {
		return []*ElasticsearchTopologyAutoscaling{auto}, nil
	}

	return nil, nil
}

func (autos ElasticsearchTopologyAutoscalings) Payload(topologyID string, elem *models.ElasticsearchClusterTopologyElement) error {
	if len(autos) == 0 {
		return nil
	}

	// it should be only one element if any
	autoscale := autos[0]

	if elem.AutoscalingMax == nil {
		elem.AutoscalingMax = new(models.TopologySize)
	}

	if elem.AutoscalingMin == nil {
		elem.AutoscalingMin = new(models.TopologySize)
	}

	err := autoscale.ExpandAutoscalingDimension(elem.AutoscalingMax, autoscale.MaxSize, autoscale.MaxSizeResource)
	if err != nil {
		return err
	}

	err = autoscale.ExpandAutoscalingDimension(elem.AutoscalingMin, autoscale.MinSize, autoscale.MinSizeResource)
	if err != nil {
		return err
	}

	// Ensure that if the Min and Max are empty, they're nil.
	if reflect.DeepEqual(elem.AutoscalingMin, new(models.TopologySize)) {
		elem.AutoscalingMin = nil
	}
	if reflect.DeepEqual(elem.AutoscalingMax, new(models.TopologySize)) {
		elem.AutoscalingMax = nil
	}

	if autoscale.PolicyOverrideJson.Value != "" {
		if err := json.Unmarshal([]byte(autoscale.PolicyOverrideJson.Value),
			&elem.AutoscalingPolicyOverrideJSON,
		); err != nil {
			return fmt.Errorf(
				"elasticsearch topology %s: unable to load policy_override_json: %w",
				topologyID, err,
			)
		}
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

func NewElasticsearchTopologyAutoscaling(topology *models.ElasticsearchClusterTopologyElement) (*ElasticsearchTopologyAutoscaling, error) {
	var a ElasticsearchTopologyAutoscaling

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
			return nil, fmt.Errorf(
				"elasticsearch topology %s: unable to persist policy_override_json: %w",
				topology.ID, err,
			)
		}
		a.PolicyOverrideJson.Value = string(b)
	}

	return &a, nil
}

// expandAutoscalingDimension centralises processing of %_size and %_size_resource attributes
// Due to limitations in the Terraform SDK, it's not possible to specify a Default on a Computed schema member
// to work around this limitation, this function will default the %_size_resource attribute to `memory`.
// Without this default, setting autoscaling limits on tiers which do not have those limits in the deployment
// template leads to an API error due to the empty resource field on the TopologySize model.
func (autoscale ElasticsearchTopologyAutoscaling) ExpandAutoscalingDimension(model *models.TopologySize, size, sizeResource types.String) error {
	if size.Value != "" {
		val, err := deploymentsize.ParseGb(size.Value)
		if err != nil {
			return err
		}
		model.Value = &val

		if model.Resource == nil {
			model.Resource = ec.String("memory")
		}
	}

	if sizeResource.Value != "" {
		model.Resource = &sizeResource.Value
	}

	return nil
}
