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

type EnterpriseSearchTF struct {
	ElasticsearchClusterRefId types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String `tfsdk:"ref_id"`
	ResourceId                types.String `tfsdk:"resource_id"`
	Region                    types.String `tfsdk:"region"`
	HttpEndpoint              types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String `tfsdk:"https_endpoint"`
	Topology                  types.List   `tfsdk:"topology"`
	Config                    types.Object `tfsdk:"config"`
}

type EnterpriseSearch struct {
	ElasticsearchClusterRefId *string                    `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     *string                    `tfsdk:"ref_id"`
	ResourceId                *string                    `tfsdk:"resource_id"`
	Region                    *string                    `tfsdk:"region"`
	HttpEndpoint              *string                    `tfsdk:"http_endpoint"`
	HttpsEndpoint             *string                    `tfsdk:"https_endpoint"`
	Topology                  EnterpriseSearchTopologies `tfsdk:"topology"`
	Config                    *EnterpriseSearchConfig    `tfsdk:"config"`
}

func readEnterpriseSearch(in *models.EnterpriseSearchResourceInfo) (*EnterpriseSearch, error) {
	var ess EnterpriseSearch

	ess.RefId = in.RefID

	ess.ResourceId = in.Info.ID

	ess.Region = in.Region

	plan := in.Info.PlanInfo.Current.Plan
	var err error
	if ess.Topology, err = readEnterpriseSearchTopologies(plan.ClusterTopology); err != nil {
		return nil, err
	}

	ess.ElasticsearchClusterRefId = in.ElasticsearchClusterRefID

	ess.HttpEndpoint, ess.HttpsEndpoint = converters.ExtractEndpoints(in.Info.Metadata)

	cfg, err := readEnterpriseSearchConfig(plan.EnterpriseSearch)
	if err != nil {
		return nil, err
	}
	ess.Config = cfg

	return &ess, nil
}

func (es *EnterpriseSearchTF) Payload(ctx context.Context, payload models.EnterpriseSearchPayload) (*models.EnterpriseSearchPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !es.ElasticsearchClusterRefId.IsNull() {
		payload.ElasticsearchClusterRefID = &es.ElasticsearchClusterRefId.Value
	}

	if !es.RefId.IsNull() {
		payload.RefID = &es.RefId.Value
	}

	if es.Region.Value != "" {
		payload.Region = &es.Region.Value
	}

	if !es.Config.IsNull() {
		var config EnterpriseSearchConfigTF

		ds := tfsdk.ValueAs(ctx, es.Config, &config)
		diags = append(diags, ds...)

		if !ds.HasError() {
			diags.Append(config.Payload(ctx, payload.Plan.EnterpriseSearch)...)
		}
	}

	topology, err := enterpriseSearchTopologiesPayload(ctx, es.Topology, payload.Plan.ClusterTopology)
	if err != nil {
		return nil, err
	}
	payload.Plan.ClusterTopology = topology

	return &payload, diags
}

func readEnterpriseSearches(in []*models.EnterpriseSearchResourceInfo) (*EnterpriseSearch, error) {
	for _, model := range in {
		if util.IsCurrentEssPlanEmpty(model) || isEssResourceStopped(model) {
			continue
		}

		ess, err := readEnterpriseSearch(model)
		if err != nil {
			return nil, err
		}

		return ess, nil
	}

	return nil, nil
}

func enterpriseSearchesPayload(ctx context.Context, esObj types.Object, template *models.DeploymentTemplateInfoV2) (*models.EnterpriseSearchPayload, diag.Diagnostics) {
	if esObj.IsNull() {
		return nil, nil
	}

	templatePayload := essResource(template)

	var diags diag.Diagnostics

	if templatePayload == nil {
		diags.AddError(
			"enterprise_search payload error",
			"enterprise_search specified but deployment template is not configured for it. Use a different template if you wish to add enterprise_search",
		)
		return nil, diags
	}

	var es EnterpriseSearchTF

	if diags = tfsdk.ValueAs(ctx, esObj, &es); diags.HasError() {
		return nil, diags
	}

	payload, diags := es.Payload(ctx, *templatePayload)

	if diags.HasError() {
		return nil, diags
	}

	return payload, nil
}

// essResource returns the EnterpriseSearchPayload from a deployment
// template or an empty version of the payload.
func essResource(template *models.DeploymentTemplateInfoV2) *models.EnterpriseSearchPayload {
	if template == nil || len(template.DeploymentTemplate.Resources.EnterpriseSearch) == 0 {
		return nil
	}
	return template.DeploymentTemplate.Resources.EnterpriseSearch[0]
}
