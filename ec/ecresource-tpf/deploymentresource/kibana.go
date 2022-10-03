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

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type KibanaTF struct {
	ElasticsearchClusterRefId types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String `tfsdk:"ref_id"`
	ResourceId                types.String `tfsdk:"resource_id"`
	Region                    types.String `tfsdk:"region"`
	HttpEndpoint              types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String `tfsdk:"https_endpoint"`
	Topology                  types.List   `tfsdk:"topology"`
	Config                    types.List   `tfsdk:"config"`
}

type Kibana struct {
	ElasticsearchClusterRefId *string       `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     *string       `tfsdk:"ref_id"`
	ResourceId                *string       `tfsdk:"resource_id"`
	Region                    *string       `tfsdk:"region"`
	HttpEndpoint              *string       `tfsdk:"http_endpoint"`
	HttpsEndpoint             *string       `tfsdk:"https_endpoint"`
	Topology                  Topologies    `tfsdk:"topology"`
	Config                    KibanaConfigs `tfsdk:"config"`
}

type Kibanas []Kibana

func readKibana(in *models.KibanaResourceInfo) (*Kibana, error) {
	var kibana Kibana

	kibana.RefId = in.RefID

	kibana.ResourceId = in.Info.ClusterID

	kibana.Region = in.Region

	plan := in.Info.PlanInfo.Current.Plan
	var err error

	if kibana.Topology, err = readKibanaTopologies(plan.ClusterTopology); err != nil {
		return nil, err
	}

	kibana.ElasticsearchClusterRefId = in.ElasticsearchClusterRefID

	kibana.HttpEndpoint, kibana.HttpsEndpoint = converters.ExtractEndpoints(in.Info.Metadata)

	cfg, err := readKibanaConfigs(plan.Kibana)
	if err != nil {
		return nil, err
	}
	kibana.Config = cfg

	return &kibana, nil
}

func (kibana KibanaTF) Payload(ctx context.Context, payload models.KibanaPayload) (*models.KibanaPayload, diag.Diagnostics) {
	if !kibana.ElasticsearchClusterRefId.IsNull() {
		payload.ElasticsearchClusterRefID = &kibana.ElasticsearchClusterRefId.Value
	}

	if !kibana.RefId.IsNull() {
		payload.RefID = &kibana.RefId.Value
	}

	if kibana.Region.Value != "" {
		payload.Region = &kibana.Region.Value
	}

	if diags := kibanaConfigPayload(ctx, payload.Plan.Kibana, &kibana.Config); diags.HasError() {
		return nil, diags
	}

	topology, diags := kibanaTopologyPayload(ctx, payload.Plan.ClusterTopology, &kibana.Topology)
	if diags.HasError() {
		return nil, diags
	}
	payload.Plan.ClusterTopology = topology

	return &payload, nil
}

func readKibanas(in []*models.KibanaResourceInfo) (Kibanas, error) {
	if len(in) == 0 {
		return nil, nil
	}

	kibanas := make(Kibanas, 0, len(in))
	for _, model := range in {
		if util.IsCurrentKibanaPlanEmpty(model) || isKibanaResourceStopped(model) {
			continue
		}

		kibana, err := readKibana(model)
		if err != nil {
			return nil, err
		}
		kibanas = append(kibanas, *kibana)
	}
	return kibanas, nil
}

func kibanaPayload(ctx context.Context, template *models.DeploymentTemplateInfoV2, kibanas types.List) ([]*models.KibanaPayload, diag.Diagnostics) {
	if len(kibanas.Elems) == 0 {
		return nil, nil
	}

	templatePlayload := kibanaResource(template)

	var diags diag.Diagnostics

	if templatePlayload == nil {
		diags.AddError("Kibana payload error", "kibana specified but deployment template is not configured for it. Use a different template if you wish to add kibana")
	}

	payloads := make([]*models.KibanaPayload, 0, len(kibanas.Elems))

	for _, elem := range kibanas.Elems {
		var kibana KibanaTF
		if tfsdk.ValueAs(ctx, elem, &kibana); diags.HasError() {
			return nil, diags
		}

		payload, diags := kibana.Payload(ctx, *templatePlayload)
		if diags.HasError() {
			return nil, diags
		}
		payloads = append(payloads, payload)
	}

	return payloads, nil
}

// kibanaResource returns the KibanaPayload from a deployment
// template or an empty version of the payload.
func kibanaResource(res *models.DeploymentTemplateInfoV2) *models.KibanaPayload {
	if len(res.DeploymentTemplate.Resources.Kibana) == 0 {
		return nil
	}
	return res.DeploymentTemplate.Resources.Kibana[0]
}
