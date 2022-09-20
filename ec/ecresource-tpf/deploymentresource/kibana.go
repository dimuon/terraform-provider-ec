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

type Kibana struct {
	ElasticsearchClusterRefId types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String `tfsdk:"ref_id"`
	ResourceId                types.String `tfsdk:"resource_id"`
	Region                    types.String `tfsdk:"region"`
	HttpEndpoint              types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String `tfsdk:"https_endpoint"`
	Topology                  []*Topology  `tfsdk:"topology"`
	Config                    KibanaConfig `tfsdk:"config"`
}

func NewKibanas(in []*models.KibanaResourceInfo) ([]*Kibana, error) {
	if len(in) == 0 {
		return nil, nil
	}

	kibanas := make([]*Kibana, 0, len(in))
	for _, model := range in {
		if util.IsCurrentKibanaPlanEmpty(model) || isKibanaResourceStopped(model) {
			continue
		}

		kibana, err := NewKibana(model)
		if err != nil {
			return nil, err
		}
		kibanas = append(kibanas, kibana)
	}
	return kibanas, nil
}

func NewKibana(in *models.KibanaResourceInfo) (*Kibana, error) {
	var kibana Kibana

	if in.RefID != nil {
		kibana.RefId.Value = *in.RefID
	}

	if in.Info.ClusterID != nil {
		kibana.ResourceId.Value = *in.Info.ClusterID
	}

	if in.Region != nil {
		kibana.Region.Value = *in.Region
	}

	plan := in.Info.PlanInfo.Current.Plan
	var err error

	if kibana.Topology, err = NewKibanaTopologies(plan.ClusterTopology); err != nil {
		return nil, err
	}

	if in.ElasticsearchClusterRefID != nil {
		kibana.ElasticsearchClusterRefId = types.String{Value: *in.ElasticsearchClusterRefID}
	}

	kibana.HttpEndpoint.Value, kibana.HttpsEndpoint.Value = flatteners.FlattenEndpoints(in.Info.Metadata)

	cfg, err := NewKibanaConfig(plan.Kibana)
	if err != nil {
		return nil, err
	}
	kibana.Config = *cfg

	return &kibana, nil
}
