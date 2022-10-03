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

type ApmTF struct {
	ElasticsearchClusterRefId types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String `tfsdk:"ref_id"`
	ResourceId                types.String `tfsdk:"resource_id"`
	Region                    types.String `tfsdk:"region"`
	HttpEndpoint              types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String `tfsdk:"https_endpoint"`
	Topology                  types.List   `tfsdk:"topology"`
	Config                    types.List   `tfsdk:"config"`
}

type Apm struct {
	ElasticsearchClusterRefId *string    `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     *string    `tfsdk:"ref_id"`
	ResourceId                *string    `tfsdk:"resource_id"`
	Region                    *string    `tfsdk:"region"`
	HttpEndpoint              *string    `tfsdk:"http_endpoint"`
	HttpsEndpoint             *string    `tfsdk:"https_endpoint"`
	Topology                  Topologies `tfsdk:"topology"`
	Config                    ApmConfigs `tfsdk:"config"`
}

func ReadApm(in *models.ApmResourceInfo) (*Apm, error) {
	var apm Apm

	apm.RefId = in.RefID

	apm.ResourceId = in.Info.ID

	apm.Region = in.Region

	plan := in.Info.PlanInfo.Current.Plan

	topologies, err := readApmTopologies(plan.ClusterTopology)
	if err != nil {
		return nil, err
	}

	apm.Topology = topologies

	apm.ElasticsearchClusterRefId = in.ElasticsearchClusterRefID

	apm.HttpEndpoint, apm.HttpsEndpoint = converters.ExtractEndpoints(in.Info.Metadata)

	configs, err := readApmConfigs(plan.Apm)
	if err != nil {
		return nil, err
	}

	apm.Config = configs

	return &apm, nil
}

func (apm ApmTF) Payload(ctx context.Context, payload models.ApmPayload) (*models.ApmPayload, diag.Diagnostics) {
	if !apm.ElasticsearchClusterRefId.IsNull() {
		payload.ElasticsearchClusterRefID = &apm.ElasticsearchClusterRefId.Value
	}

	if !apm.RefId.IsNull() {
		payload.RefID = &apm.RefId.Value
	}

	if apm.Region.Value != "" {
		payload.Region = &apm.Region.Value
	}

	if diags := ApmConfigsTF(apm.Config).Payload(ctx, payload.Plan.Apm); diags.HasError() {
		return nil, diags
	}

	topology, diags := ApmTopologiesTF(apm.Topology).Payload(ctx, payload.Plan.ClusterTopology)
	if diags.HasError() {
		return nil, diags
	}
	payload.Plan.ClusterTopology = topology

	return &payload, nil
}

type Apms []Apm

func readApms(in []*models.ApmResourceInfo) (Apms, error) {
	apms := make([]Apm, 0, len(in))

	for _, model := range in {
		if util.IsCurrentApmPlanEmpty(model) || isApmResourceStopped(model) {
			continue
		}

		apm, err := ReadApm(model)
		if err != nil {
			return nil, err
		}

		apms = append(apms, *apm)
	}

	return apms, nil
}

func apmsPayload(ctx context.Context, template *models.DeploymentTemplateInfoV2, apms types.List) ([]*models.ApmPayload, diag.Diagnostics) {
	if len(apms.Elems) == 0 {
		return nil, nil
	}

	templatePayload := apmResource(template)

	var diags diag.Diagnostics
	if templatePayload == nil {
		diags.AddError("Apm reading error", "apm specified but deployment template is not configured for it. Use a different template if you wish to add apm")
		return nil, diags
	}

	payloads := make([]*models.ApmPayload, 0, len(apms.Elems))
	for _, elem := range apms.Elems {
		var apm ApmTF
		if diags := tfsdk.ValueAs(ctx, elem, &apm); diags.HasError() {
			return nil, diags
		}
		payload, diags := apm.Payload(ctx, *templatePayload)
		if diags.HasError() {
			return nil, diags
		}
		payloads = append(payloads, payload)
	}

	return payloads, nil
}

// apmResource returns the ApmPayload from a deployment
// template or an empty version of the payload.
func apmResource(template *models.DeploymentTemplateInfoV2) *models.ApmPayload {
	if template == nil || len(template.DeploymentTemplate.Resources.Apm) == 0 {
		return nil
	}
	return template.DeploymentTemplate.Resources.Apm[0]
}
