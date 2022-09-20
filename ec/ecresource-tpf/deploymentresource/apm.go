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
	"github.com/elastic/terraform-provider-ec/ec/internal/flatteners"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Apm struct {
	ElasticsearchClusterRefId types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String `tfsdk:"ref_id"`
	ResourceId                types.String `tfsdk:"resource_id"`
	Region                    types.String `tfsdk:"region"`
	HttpEndpoint              types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String `tfsdk:"https_endpoint"`
	Topology                  []Topology   `tfsdk:"topology"`
	Config                    ApmConfig    `tfsdk:"config"`
}

func NewApms(in []*models.ApmResourceInfo) ([]*Apm, error) {
	if len(in) == 0 {
		return nil, nil
	}

	apms := make([]*Apm, 0, len(in))
	for _, model := range in {
		if util.IsCurrentApmPlanEmpty(model) || isApmResourceStopped(model) {
			continue
		}
		apm, err := NewApm(model)
		if err != nil {
			return nil, err
		}
		apms = append(apms, apm)
	}
	return apms, nil
}

func NewApm(in *models.ApmResourceInfo) (*Apm, error) {
	var apm Apm

	if in.RefID != nil {
		apm.RefId.Value = *in.RefID
	}

	if in.Info.ID != nil {
		apm.ResourceId.Value = *in.Info.ID
	}

	if in.Region != nil {
		apm.Region.Value = *in.Region
	}

	plan := in.Info.PlanInfo.Current.Plan
	var err error

	apm.Topology, err = NewApmTopologies(plan.ClusterTopology)
	if err != nil {
		return nil, err
	}

	if in.ElasticsearchClusterRefID != nil {
		apm.ElasticsearchClusterRefId.Value = *in.ElasticsearchClusterRefID
	}

	apm.HttpEndpoint.Value, apm.HttpsEndpoint.Value = flatteners.FlattenEndpoints(in.Info.Metadata)

	cfg, err := NewApmConfig(plan.Apm)
	if err != nil {
		return nil, err
	}
	apm.Config = *cfg

	return &apm, nil
}
