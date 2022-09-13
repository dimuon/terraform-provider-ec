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
	"strconv"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTopology struct {
	Id                      types.String                       `tfsdk:"id"`
	InstanceConfigurationId types.String                       `tfsdk:"instance_configuration_id"`
	Size                    types.String                       `tfsdk:"size"`
	SizeResource            types.String                       `tfsdk:"size_resource"`
	ZoneCount               types.Int64                        `tfsdk:"zone_count"`
	NodeTypeData            types.String                       `tfsdk:"node_type_data"`
	NodeTypeMaster          types.String                       `tfsdk:"node_type_master"`
	NodeTypeIngest          types.String                       `tfsdk:"node_type_ingest"`
	NodeTypeMl              types.String                       `tfsdk:"node_type_ml"`
	NodeRoles               types.Set                          `tfsdk:"node_roles"`
	Autoscaling             []ElasticsearchTopologyAutoscaling `tfsdk:"autoscaling"`
	Config                  []ElasticsearchTopologyConfig      `tfsdk:"config"`
}

func (est *ElasticsearchTopology) fromModel(topology *models.ElasticsearchClusterTopologyElement) error {
	est.Id.Value = topology.ID

	if topology.InstanceConfigurationID != "" {
		est.InstanceConfigurationId.Value = topology.InstanceConfigurationID
	}

	if topology.Size != nil {
		est.Size.Value = util.MemoryToState(*topology.Size.Value)
		est.SizeResource.Value = *topology.Size.Resource
	}

	est.ZoneCount.Value = int64(topology.ZoneCount)

	if nt := topology.NodeType; nt != nil {
		if nt.Data != nil {
			est.NodeTypeData.Value = strconv.FormatBool(*nt.Data)
		}

		if nt.Ingest != nil {
			est.NodeTypeIngest.Value = strconv.FormatBool(*nt.Ingest)
		}

		if nt.Master != nil {
			est.NodeTypeMaster.Value = strconv.FormatBool(*nt.Master)
		}

		if nt.Ml != nil {
			est.NodeTypeMl.Value = strconv.FormatBool(*nt.Ml)
		}
	}

	if len(topology.NodeRoles) > 0 {
		est.NodeRoles.ElemType = types.StringType
		est.NodeRoles.Elems = make([]attr.Value, 0, len(topology.NodeRoles))
		for _, role := range topology.NodeRoles {
			est.NodeRoles.Elems = append(est.NodeRoles.Elems, types.String{Value: role})
		}
	}

	// if err := est.Autoscaling.fromModel(topology); err != nil {
	// 	return err
	// }

	// Computed config object to avoid unsetting legacy topology config settings.
	// if err := est.Config.fromModel(topology.Elasticsearch); err != nil {
	// 	return err
	// }

	return nil
}

func isPotentiallySizedTopology(topology *models.ElasticsearchClusterTopologyElement, isAutoscaling bool) bool {
	currentlySized := topology.Size != nil && topology.Size.Value != nil && *topology.Size.Value > 0
	canBeSized := isAutoscaling && topology.AutoscalingMax != nil && topology.AutoscalingMax.Value != nil && *topology.AutoscalingMax.Value > 0

	return currentlySized || canBeSized
}
