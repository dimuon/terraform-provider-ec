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
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deploymentsize"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTopologyAutoscalingTF struct {
	MaxSizeResource    types.String `tfsdk:"max_size_resource"`
	MaxSize            types.String `tfsdk:"max_size"`
	MinSizeResource    types.String `tfsdk:"min_size_resource"`
	MinSize            types.String `tfsdk:"min_size"`
	PolicyOverrideJson types.String `tfsdk:"policy_override_json"`
}

type ElasticsearchTopologyAutoscaling struct {
	MaxSizeResource    *string `tfsdk:"max_size_resource"`
	MaxSize            *string `tfsdk:"max_size"`
	MinSizeResource    *string `tfsdk:"min_size_resource"`
	MinSize            *string `tfsdk:"min_size"`
	PolicyOverrideJson *string `tfsdk:"policy_override_json"`
}

type ElasticsearchTopologyAutoscalings []ElasticsearchTopologyAutoscaling

func ReadElasticsearchTopologyAutoscalings(in *models.ElasticsearchClusterTopologyElement) (ElasticsearchTopologyAutoscalings, error) {
	autoscaling, err := ReadElasticsearchTopologyAutoscaling(in)
	if err != nil {
		return nil, err
	}

	if autoscaling.isEmpty() {
		return nil, nil
	}

	return ElasticsearchTopologyAutoscalings{*autoscaling}, nil
}

func ElasticsearchTopologyAutoscalingsPayload(ctx context.Context, autos types.List, topologyID string, payload *models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	if len(autos.Elems) == 0 {
		return nil
	}

	return ElasticsearchTopologyAutoscalingPayload(ctx, autos.Elems[0], topologyID, payload)
}

func ElasticsearchTopologyAutoscalingPayload(ctx context.Context, autoObj attr.Value, topologyID string, payload *models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	var diag diag.Diagnostics

	if autoObj.IsNull() || autoObj.IsUnknown() {
		return nil
	}

	// it should be only one element if any
	var autoscale ElasticsearchTopologyAutoscalingTF

	tfsdk.ValueAs(ctx, autoObj, &autoscale)

	if payload.AutoscalingMax == nil {
		payload.AutoscalingMax = new(models.TopologySize)
	}

	if payload.AutoscalingMin == nil {
		payload.AutoscalingMin = new(models.TopologySize)
	}

	err := autoscale.ExpandAutoscalingDimension(payload.AutoscalingMax, autoscale.MaxSize, autoscale.MaxSizeResource)
	if err != nil {
		diag.AddError("fail to parse autoscale max size", err.Error())
		return diag
	}

	err = autoscale.ExpandAutoscalingDimension(payload.AutoscalingMin, autoscale.MinSize, autoscale.MinSizeResource)
	if err != nil {
		diag.AddError("fail to parse autoscale min size", err.Error())
		return diag
	}

	// Ensure that if the Min and Max are empty, they're nil.
	if reflect.DeepEqual(payload.AutoscalingMin, new(models.TopologySize)) {
		payload.AutoscalingMin = nil
	}
	if reflect.DeepEqual(payload.AutoscalingMax, new(models.TopologySize)) {
		payload.AutoscalingMax = nil
	}

	if autoscale.PolicyOverrideJson.Value != "" {
		if err := json.Unmarshal([]byte(autoscale.PolicyOverrideJson.Value),
			&payload.AutoscalingPolicyOverrideJSON,
		); err != nil {
			diag.AddError(fmt.Sprintf("elasticsearch topology %s: unable to load policy_override_json", topologyID), err.Error())
			return diag
		}
	}

	return diag
}

func ReadElasticsearchTopologyAutoscaling(topology *models.ElasticsearchClusterTopologyElement) (*ElasticsearchTopologyAutoscaling, error) {
	var a ElasticsearchTopologyAutoscaling

	if ascale := topology.AutoscalingMax; ascale != nil {
		a.MaxSizeResource = ascale.Resource
		a.MaxSize = ec.String(util.MemoryToState(*ascale.Value))
	}

	if ascale := topology.AutoscalingMin; ascale != nil {
		a.MinSizeResource = ascale.Resource
		a.MinSize = ec.String(util.MemoryToState(*ascale.Value))
	}

	if topology.AutoscalingPolicyOverrideJSON != nil {
		b, err := json.Marshal(topology.AutoscalingPolicyOverrideJSON)
		if err != nil {
			return nil, fmt.Errorf("elasticsearch topology %s: unable to persist policy_override_json - %w", topology.ID, err)
		}
		a.PolicyOverrideJson = ec.String(string(b))
	}

	return &a, nil
}

func (a ElasticsearchTopologyAutoscaling) isEmpty() bool {
	return reflect.ValueOf(a).IsZero()
}

// expandAutoscalingDimension centralises processing of %_size and %_size_resource attributes
// Due to limitations in the Terraform SDK, it's not possible to specify a Default on a Computed schema member
// to work around this limitation, this function will default the %_size_resource attribute to `memory`.
// Without this default, setting autoscaling limits on tiers which do not have those limits in the deployment
// template leads to an API error due to the empty resource field on the TopologySize model.
func (autoscale ElasticsearchTopologyAutoscalingTF) ExpandAutoscalingDimension(model *models.TopologySize, size, sizeResource types.String) error {
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
