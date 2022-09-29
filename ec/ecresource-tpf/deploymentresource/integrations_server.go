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
	"errors"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IntegrationsServer struct {
	ElasticsearchClusterRefId types.String                 `tfsdk:"elasticsearch_cluster_ref_id"`
	RefId                     types.String                 `tfsdk:"ref_id"`
	ResourceId                types.String                 `tfsdk:"resource_id"`
	Region                    types.String                 `tfsdk:"region"`
	HttpEndpoint              types.String                 `tfsdk:"http_endpoint"`
	HttpsEndpoint             types.String                 `tfsdk:"https_endpoint"`
	Topology                  IntegrationsServerTopologies `tfsdk:"topology"`
	Config                    IntegrationsServerConfigs    `tfsdk:"config"`
}

func NewIntegrationsServers(in []*models.IntegrationsServerResourceInfo) ([]*IntegrationsServer, error) {
	if len(in) == 0 {
		return nil, nil
	}

	srvs := make([]*IntegrationsServer, 0, len(in))
	for _, model := range in {
		if util.IsCurrentIntegrationsServerPlanEmpty(model) || isIntegrationsServerResourceStopped(model) {
			continue
		}

		srv, err := NewIntegrationsServer(model)
		if err != nil {
			return nil, err
		}
		srvs = append(srvs, srv)
	}

	return srvs, nil
}

func NewIntegrationsServer(in *models.IntegrationsServerResourceInfo) (*IntegrationsServer, error) {
	var srv IntegrationsServer

	if in.RefID != nil {
		srv.RefId = types.String{Value: *in.RefID}
	}

	if in.Info.ID != nil {
		srv.ResourceId = types.String{Value: *in.Info.ID}
	}

	if in.Region != nil {
		srv.Region = types.String{Value: *in.Region}
	}

	plan := in.Info.PlanInfo.Current.Plan
	var err error
	if srv.Topology, err = NewIntegrationsServerTopologies(plan.ClusterTopology); err != nil {
		return nil, err
	}

	if in.ElasticsearchClusterRefID != nil {
		srv.ElasticsearchClusterRefId = types.String{Value: *in.ElasticsearchClusterRefID}
	}

	srv.HttpEndpoint, srv.HttpsEndpoint = converters.ExtractEndpointsTF(in.Info.Metadata)

	cfg, err := NewIntegrationsServerConfig(plan.IntegrationsServer)
	if err != nil {
		return nil, err
	}
	srv.Config = cfg

	return &srv, nil
}

func (srv IntegrationsServer) Payload(payload models.IntegrationsServerPayload) (*models.IntegrationsServerPayload, error) {
	if !srv.ElasticsearchClusterRefId.IsNull() {
		payload.ElasticsearchClusterRefID = &srv.ElasticsearchClusterRefId.Value
	}

	if !srv.RefId.IsNull() {
		payload.RefID = &srv.RefId.Value
	}

	if srv.Region.Value != "" {
		payload.Region = &srv.Region.Value
	}

	if err := srv.Config.Payload(payload.Plan.IntegrationsServer); err != nil {
		return nil, err
	}

	var err error
	payload.Plan.ClusterTopology, err = srv.Topology.Payload(payload.Plan.ClusterTopology)
	if err != nil {
		return nil, err
	}

	return &payload, nil
}

type IntegrationsServers []*IntegrationsServer

func (srvs IntegrationsServers) Payload(template *models.DeploymentTemplateInfoV2) ([]*models.IntegrationsServerPayload, error) {
	if len(srvs) == 0 {
		return nil, nil
	}

	templatePayload := integrationsServerResource(template)

	if templatePayload == nil {
		return nil, errors.New("IntegrationsServer specified but deployment template is not configured for it. Use a different template if you wish to add IntegrationsServer")
	}

	payloads := make([]*models.IntegrationsServerPayload, 0, len(srvs))
	for _, srv := range srvs {
		payload, err := srv.Payload(*templatePayload)
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, payload)
	}

	return payloads, nil
}

// IntegrationsServerResource returns the IntegrationsServerPayload from a deployment
// template or an empty version of the payload.
func integrationsServerResource(template *models.DeploymentTemplateInfoV2) *models.IntegrationsServerPayload {
	if len(template.DeploymentTemplate.Resources.IntegrationsServer) == 0 {
		return nil
	}
	return template.DeploymentTemplate.Resources.IntegrationsServer[0]
}
