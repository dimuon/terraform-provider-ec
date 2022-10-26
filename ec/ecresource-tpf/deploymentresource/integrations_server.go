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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IntegrationsServerTF struct {
	ElasticsearchClusterRefId types.String `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String `tfsdk:"ref_id"`
	ResourceId                types.String `tfsdk:"resource_id"`
	Region                    types.String `tfsdk:"region"`
	HttpEndpoint              types.String `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String `tfsdk:"https_endpoint"`
	Topology                  types.List   `tfsdk:"topology"`
	Config                    types.List   `tfsdk:"config"`
}

type IntegrationsServer struct {
	ElasticsearchClusterRefId *string                   `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     *string                   `tfsdk:"ref_id"`
	ResourceId                *string                   `tfsdk:"resource_id"`
	Region                    *string                   `tfsdk:"region"`
	HttpEndpoint              *string                   `tfsdk:"http_endpoint"`
	HttpsEndpoint             *string                   `tfsdk:"https_endpoint"`
	Topology                  Topologies                `tfsdk:"topology"`
	Config                    IntegrationsServerConfigs `tfsdk:"config"`
}

type IntegrationsServers []IntegrationsServer

func readIntegrationsServers(in []*models.IntegrationsServerResourceInfo) (IntegrationsServers, error) {
	if len(in) == 0 {
		return nil, nil
	}

	for _, model := range in {
		if util.IsCurrentIntegrationsServerPlanEmpty(model) || isIntegrationsServerResourceStopped(model) {
			continue
		}

		srv, err := readIntegrationsServer(model)
		if err != nil {
			return nil, err
		}

		return IntegrationsServers{*srv}, nil
	}

	return nil, nil
}

func readIntegrationsServer(in *models.IntegrationsServerResourceInfo) (*IntegrationsServer, error) {
	var srv IntegrationsServer

	srv.RefId = in.RefID

	srv.ResourceId = in.Info.ID

	srv.Region = in.Region

	plan := in.Info.PlanInfo.Current.Plan

	var err error
	if srv.Topology, err = readIntegrationsServerTopologies(plan.ClusterTopology); err != nil {
		return nil, err
	}

	srv.ElasticsearchClusterRefId = in.ElasticsearchClusterRefID

	srv.HttpEndpoint, srv.HttpsEndpoint = converters.ExtractEndpoints(in.Info.Metadata)

	cfg, err := readIntegrationsServerConfig(plan.IntegrationsServer)
	if err != nil {
		return nil, err
	}
	srv.Config = cfg

	return &srv, nil
}

func (srv IntegrationsServerTF) Payload(ctx context.Context, payload models.IntegrationsServerPayload) (*models.IntegrationsServerPayload, diag.Diagnostics) {
	if !srv.ElasticsearchClusterRefId.IsNull() {
		payload.ElasticsearchClusterRefID = &srv.ElasticsearchClusterRefId.Value
	}

	if !srv.RefId.IsNull() {
		payload.RefID = &srv.RefId.Value
	}

	if srv.Region.Value != "" {
		payload.Region = &srv.Region.Value
	}

	if diags := payloadIntegrationsServerConfig(ctx, srv.Config, payload.Plan.IntegrationsServer); diags.HasError() {
		return nil, diags
	}

	var diags diag.Diagnostics
	payload.Plan.ClusterTopology, diags = integrationsServerTopologyPayload(ctx, payload.Plan.ClusterTopology, &srv.Topology)
	if diags.HasError() {
		return nil, diags
	}

	return &payload, nil
}

func integrationsServerPayload(ctx context.Context, list types.List, template *models.DeploymentTemplateInfoV2) (*models.IntegrationsServerPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	var srv *IntegrationsServerTF

	if diags = getFirst(ctx, list, &srv); diags.HasError() {
		return nil, diags
	}

	if srv == nil {
		return nil, nil
	}

	templatePayload := integrationsServerResource(template)

	if templatePayload == nil {
		diags.AddError("integrations_server payload error", "integrations_server specified but deployment template is not configured for it. Use a different template if you wish to add integrations_server")
		return nil, diags
	}

	payload, diags := srv.Payload(ctx, *templatePayload)

	if diags.HasError() {
		return nil, diags
	}

	return payload, nil
}

// IntegrationsServerResource returns the IntegrationsServerPayload from a deployment
// template or an empty version of the payload.
func integrationsServerResource(template *models.DeploymentTemplateInfoV2) *models.IntegrationsServerPayload {
	if template == nil || len(template.DeploymentTemplate.Resources.IntegrationsServer) == 0 {
		return nil
	}
	return template.DeploymentTemplate.Resources.IntegrationsServer[0]
}
