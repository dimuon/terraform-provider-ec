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

package v2

import (
	"context"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/kibana/v1"
	topologyv1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/topology/v1"
	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/utils"
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
	Topology                  types.Object `tfsdk:"topology"`
	Config                    types.Object `tfsdk:"config"`
}

type Kibana struct {
	ElasticsearchClusterRefId *string              `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     *string              `tfsdk:"ref_id"`
	ResourceId                *string              `tfsdk:"resource_id"`
	Region                    *string              `tfsdk:"region"`
	HttpEndpoint              *string              `tfsdk:"http_endpoint"`
	HttpsEndpoint             *string              `tfsdk:"https_endpoint"`
	Topology                  *topologyv1.Topology `tfsdk:"topology"`
	Config                    *v1.KibanaConfig     `tfsdk:"config"`
}

func ReadKibana(in *models.KibanaResourceInfo) (*Kibana, error) {
	if util.IsCurrentKibanaPlanEmpty(in) || utils.IsKibanaResourceStopped(in) {
		return nil, nil
	}

	var kibana Kibana

	kibana.RefId = in.RefID

	kibana.ResourceId = in.Info.ClusterID

	kibana.Region = in.Region

	plan := in.Info.PlanInfo.Current.Plan
	var err error

	topologies, err := v1.ReadKibanaTopologies(plan.ClusterTopology)
	if err != nil {
		return nil, err
	}

	if len(topologies) > 0 {
		kibana.Topology = &topologies[0]
	}

	kibana.ElasticsearchClusterRefId = in.ElasticsearchClusterRefID

	kibana.HttpEndpoint, kibana.HttpsEndpoint = converters.ExtractEndpoints(in.Info.Metadata)

	configs, err := v1.ReadKibanaConfig(plan.Kibana)
	if err != nil {
		return nil, err
	}

	if len(configs) > 0 {
		kibana.Config = &configs[0]
	}

	return &kibana, nil
}

func (kibana KibanaTF) Payload(ctx context.Context, payload models.KibanaPayload) (*models.KibanaPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !kibana.ElasticsearchClusterRefId.IsNull() {
		payload.ElasticsearchClusterRefID = &kibana.ElasticsearchClusterRefId.Value
	}

	if !kibana.RefId.IsNull() {
		payload.RefID = &kibana.RefId.Value
	}

	if kibana.Region.Value != "" {
		payload.Region = &kibana.Region.Value
	}

	if !kibana.Config.IsNull() && !kibana.Config.IsUnknown() {
		var config v1.KibanaConfigTF

		ds := tfsdk.ValueAs(ctx, kibana.Config, &config)

		diags.Append(ds...)

		if !ds.HasError() {
			diags.Append(config.Payload(payload.Plan.Kibana)...)
		}
	}

	topologyPayload, ds := v1.KibanaTopologyPayload(ctx, v1.DefaultKibanaTopology(payload.Plan.ClusterTopology), 0, kibana.Topology)

	diags.Append(ds...)

	if !ds.HasError() && topologyPayload != nil {
		payload.Plan.ClusterTopology = []*models.KibanaClusterTopologyElement{topologyPayload}
	}

	return &payload, diags
}

func KibanaPayload(ctx context.Context, kibanaObj types.Object, template *models.DeploymentTemplateInfoV2) (*models.KibanaPayload, diag.Diagnostics) {
	var kibanaTF *KibanaTF

	var diags diag.Diagnostics

	if diags = tfsdk.ValueAs(ctx, kibanaObj, &kibanaTF); diags.HasError() {
		return nil, diags
	}

	if kibanaTF == nil {
		return nil, nil
	}

	templatePlayload := v1.KibanaResource(template)

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
