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

	"github.com/elastic/cloud-sdk-go/pkg/models"
	topologyv1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/topology/v1"
	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/utils"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	ElasticsearchClusterRefId *string               `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     *string               `tfsdk:"ref_id"`
	ResourceId                *string               `tfsdk:"resource_id"`
	Region                    *string               `tfsdk:"region"`
	HttpEndpoint              *string               `tfsdk:"http_endpoint"`
	HttpsEndpoint             *string               `tfsdk:"https_endpoint"`
	Topology                  topologyv1.Topologies `tfsdk:"topology"`
	Config                    KibanaConfigs         `tfsdk:"config"`
}

type Kibanas []Kibana

func ReadKibana(in *models.KibanaResourceInfo) (*Kibana, error) {
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

	cfg, err := ReadKibanaConfig(plan.Kibana)
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

	if diags := KibanaConfigPayload(ctx, kibana.Config, payload.Plan.Kibana); diags.HasError() {
		return nil, diags
	}

	topology, diags := kibanaTopologyPayload(ctx, payload.Plan.ClusterTopology, &kibana.Topology)
	if diags.HasError() {
		return nil, diags
	}
	payload.Plan.ClusterTopology = topology

	return &payload, nil
}

func ReadKibanas(in []*models.KibanaResourceInfo) (Kibanas, error) {
	if len(in) == 0 {
		return nil, nil
	}

	for _, model := range in {
		if util.IsCurrentKibanaPlanEmpty(model) || utils.IsKibanaResourceStopped(model) {
			continue
		}

		kibana, err := ReadKibana(model)
		if err != nil {
			return nil, err
		}

		return Kibanas{*kibana}, nil
	}

	return nil, nil
}

func KibanaPayload(ctx context.Context, list types.List, template *models.DeploymentTemplateInfoV2) (*models.KibanaPayload, diag.Diagnostics) {
	var kibanaTF *KibanaTF

	var diags diag.Diagnostics

	if diags = utils.GetFirst(ctx, list, &kibanaTF); diags.HasError() {
		return nil, diags
	}

	if kibanaTF == nil {
		return nil, nil
	}

	templatePlayload := kibanaResource(template)

	if templatePlayload == nil {
		diags.AddError("kibana payload error", "kibana specified but deployment template is not configured for it. Use a different template if you wish to add kibana")
		return nil, diags
	}

	payload, diags := kibanaTF.Payload(ctx, *templatePlayload)
	if diags.HasError() {
		return nil, diags
	}

	return payload, nil
}

// kibanaResource returns the KibanaPayload from a deployment
// template or an empty version of the payload.
func kibanaResource(res *models.DeploymentTemplateInfoV2) *models.KibanaPayload {
	if res == nil || len(res.DeploymentTemplate.Resources.Kibana) == 0 {
		return nil
	}
	return res.DeploymentTemplate.Resources.Kibana[0]
}
