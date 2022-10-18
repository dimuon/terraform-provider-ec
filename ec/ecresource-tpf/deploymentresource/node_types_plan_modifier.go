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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Use `self` as value of `observability`'s `deployment_id` attribute
func UseNodeTypesDefault() tfsdk.AttributePlanModifier {
	return nodeTypesDefault{}
}

type nodeTypesDefault struct{}

func (r nodeTypesDefault) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	if !resp.AttributePlan.IsUnknown() {
		return
	}

	// if the config is the unknown value, use the unknown value otherwise, interpolation gets messed up
	if req.AttributeConfig.IsUnknown() {
		return
	}

	deploymentVersionChanged, diags := isAttributeChanged(ctx, path.Root("version"), req)

	resp.Diagnostics = append(resp.Diagnostics, diags...)

	if diags.HasError() {
		return
	}

	if deploymentVersionChanged {
		return
	}

	topologyElementChanged, diags := isElasticTopologyChanged(ctx, req.AttributePath.ParentPath().ParentPath(), req)

	resp.Diagnostics = append(resp.Diagnostics, diags...)

	if diags.HasError() {
		return
	}

	if topologyElementChanged {
		return
	}

	if !req.AttributeState.IsUnknown() && !req.AttributeState.IsNull() {
		resp.AttributePlan = req.AttributeState
		return
	}

	nodeRolesPath := req.AttributePath.ParentPath().AtName("node_roles")

	var nodeRoles types.Set

	diags = req.State.GetAttribute(ctx, nodeRolesPath, &nodeRoles)

	resp.Diagnostics = append(resp.Diagnostics, diags...)

	if diags.HasError() {
		return
	}

	if nodeRoles.IsUnknown() || nodeRoles.IsNull() {
		return
	}

	resp.AttributePlan = req.AttributeState
}

// Description returns a human-readable description of the plan modifier.
func (r nodeTypesDefault) Description(ctx context.Context) string {
	return "Calculate node type value based on current state and `node_roles`'s value."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r nodeTypesDefault) MarkdownDescription(ctx context.Context) string {
	return "Calculate node type value based on current state and `node_roles`'s value."
}

func isAttributeChanged[V terraformValue](ctx context.Context, p path.Path, req tfsdk.ModifyAttributePlanRequest) (bool, diag.Diagnostics) {
	var planValue V

	if diags := req.Plan.GetAttribute(ctx, p, &planValue); diags.HasError() {
		return false, diags
	}

	var stateValue V

	if diags := req.State.GetAttribute(ctx, p, &stateValue); diags.HasError() {
		return false, diags
	}

	return !planValue.Equal(stateValue), nil
}

type terraformValue interface {
	types.String
	attr.Value
}

// we need to compare tologoy lists with disregard to computed values whose values are unknown in plan
func isElasticTopologyChanged(ctx context.Context, p path.Path, req tfsdk.ModifyAttributePlanRequest) (bool, diag.Diagnostics) {
	var planValue types.List

	if diags := req.Plan.GetAttribute(ctx, p, &planValue); diags.HasError() {
		return false, diags
	}

	var planTopology []ElasticsearchTopologyTF

	if diags := planValue.ElementsAs(ctx, &planTopology, false); diags.HasError() {
		return false, diags
	}

	var stateValue types.List

	if diags := req.State.GetAttribute(ctx, p, &stateValue); diags.HasError() {
		return false, diags
	}

	var stateTopology []ElasticsearchTopologyTF

	if diags := stateValue.ElementsAs(ctx, &stateTopology, false); diags.HasError() {
		return false, diags
	}

	return elasticTopologiesChanged(ctx, planTopology, stateTopology)
}

func elasticTopologiesChanged(ctx context.Context, plan, state []ElasticsearchTopologyTF) (bool, diag.Diagnostics) {
	if len(plan) != len(state) {
		return true, nil
	}

	planSet := elasticsearchTopologiesToSet(plan)
	stateSet := elasticsearchTopologiesToSet(state)

	for id, pl := range planSet {
		st, exist := stateSet[id]

		if !exist {
			return true, nil
		}

		typ := elasticsearchTopologyAttribute().FrameworkType().(types.ListType).ElemType

		var planObj types.Object
		if diags := tfsdk.ValueFrom(ctx, pl, typ, &planObj); diags.HasError() {
			return false, diags
		}

		var stateObj types.Object
		if diags := tfsdk.ValueFrom(ctx, st, typ, &stateObj); diags.HasError() {
			return false, diags
		}

		for attrKey, attrVal := range planObj.Attrs {
			if attrVal.IsUnknown() {
				continue
			}

			if value, exist := stateObj.Attrs[attrKey]; !exist || !value.Equal(attrVal) {
				return true, nil
			}
		}
	}

	return false, nil
}

func elasticsearchTopologiesToSet(topologies []ElasticsearchTopologyTF) map[string]ElasticsearchTopologyTF {
	set := make(map[string]ElasticsearchTopologyTF, len(topologies))

	for _, topology := range topologies {
		set[topology.Id.Value] = topology
	}

	return set
}

/*
func EqualLists(l, other types.List) bool {
	if l.Unknown != other.Unknown {
		return false
	}
	if l.Null != other.Null {
		return false
	}
	if l.ElemType == nil && other.ElemType != nil {
		return false
	}
	if l.ElemType != nil && !l.ElemType.Equal(other.ElemType) {
		return false
	}
	if len(l.Elems) != len(other.Elems) {
		return false
	}
	for pos, lElem := range l.Elems {
		if lElem.IsUnknown() {
			continue
		}
		oElem := other.Elems[pos]
		if !lElem.Equal(oElem) {
			return false
		}
	}
	return true
}
*/
