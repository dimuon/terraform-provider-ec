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
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deploymentsize"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTopologyAutoscalings types.List

func (autoscalings ElasticsearchTopologyAutoscalings) Read(ctx context.Context, in *models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	var autoscaling ElasticsearchTopologyAutoscaling

	diag := autoscaling.Read(in)
	if diag.HasError() {
		return diag
	}

	if autoscaling == (ElasticsearchTopologyAutoscaling{}) {
		return nil
	}

	return tfsdk.ValueFrom(ctx, []*ElasticsearchTopologyAutoscaling{&autoscaling}, elasticsearchTopologyAutoscalingAttribute().Type, &autoscalings)
}

func (autos ElasticsearchTopologyAutoscalings) Payload(ctx context.Context, topologyID string, elem *models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	var diag diag.Diagnostics

	if len(autos.Elems) == 0 {
		return nil
	}

	// it should be only one element if any
	var autoscale ElasticsearchTopologyAutoscaling
	tfsdk.ValueAs(ctx, autos.Elems[0], &autoscale)

	if elem.AutoscalingMax == nil {
		elem.AutoscalingMax = new(models.TopologySize)
	}

	if elem.AutoscalingMin == nil {
		elem.AutoscalingMin = new(models.TopologySize)
	}

	err := autoscale.ExpandAutoscalingDimension(elem.AutoscalingMax, autoscale.MaxSize, autoscale.MaxSizeResource)
	if err != nil {
		diag.AddError("fail to parse autoscale max size", err.Error())
		return diag
	}

	err = autoscale.ExpandAutoscalingDimension(elem.AutoscalingMin, autoscale.MinSize, autoscale.MinSizeResource)
	if err != nil {
		diag.AddError("fail to parse autoscale min size", err.Error())
		return diag
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
			diag.AddError(fmt.Sprintf("elasticsearch topology %s: unable to load policy_override_json", topologyID), err.Error())
			return diag
		}
	}

	return diag
}

type ElasticsearchTopologyAutoscaling struct {
	MaxSizeResource    types.String `tfsdk:"max_size_resource"`
	MaxSize            types.String `tfsdk:"max_size"`
	MinSizeResource    types.String `tfsdk:"min_size_resource"`
	MinSize            types.String `tfsdk:"min_size"`
	PolicyOverrideJson types.String `tfsdk:"policy_override_json"`
}

func (a ElasticsearchTopologyAutoscaling) Read(topology *models.ElasticsearchClusterTopologyElement) diag.Diagnostics {
	if ascale := topology.AutoscalingMax; ascale != nil {
		a.MaxSizeResource = types.String{Value: *ascale.Resource}
		a.MaxSize = types.String{Value: util.MemoryToState(*ascale.Value)}
	}

	if ascale := topology.AutoscalingMin; ascale != nil {
		a.MinSizeResource = types.String{Value: *ascale.Resource}
		a.MinSize = types.String{Value: util.MemoryToState(*ascale.Value)}
	}

	if topology.AutoscalingPolicyOverrideJSON != nil {
		b, err := json.Marshal(topology.AutoscalingPolicyOverrideJSON)
		if err != nil {
			var diag diag.Diagnostics
			diag.AddError(fmt.Sprintf("elasticsearch topology %s: unable to persist policy_override_json", topology.ID), err.Error())
			return diag
		}
		a.PolicyOverrideJson = types.String{Value: string(b)}
	}

	return nil
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
