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

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ObservabilityTF struct {
	DeploymentId types.String `tfsdk:"deployment_id"`
	RefId        types.String `tfsdk:"ref_id"`
	Logs         types.Bool   `tfsdk:"logs"`
	Metrics      types.Bool   `tfsdk:"metrics"`
}

type Observability struct {
	DeploymentId *string `tfsdk:"deployment_id"`
	RefId        *string `tfsdk:"ref_id"`
	Logs         bool    `tfsdk:"logs"`
	Metrics      bool    `tfsdk:"metrics"`
}

type Observabilities []Observability

func ReadObservabilities(in *models.DeploymentSettings) (Observabilities, error) {
	obs, err := ReadObservability(in)

	if err != nil {
		return nil, err
	}

	if obs == nil {
		return nil, nil
	}

	return Observabilities{*obs}, nil
}

func ReadObservability(in *models.DeploymentSettings) (*Observability, error) {
	if in == nil || in.Observability == nil {
		return nil, nil
	}

	var obs Observability

	// We are only accepting a single deployment ID and refID for both logs and metrics.
	// If either of them is not nil the deployment ID and refID will be filled.
	if in.Observability.Metrics != nil {
		if in.Observability.Metrics.Destination.DeploymentID != nil {
			obs.DeploymentId = in.Observability.Metrics.Destination.DeploymentID
		}

		obs.RefId = &in.Observability.Metrics.Destination.RefID
		obs.Metrics = true
	}

	if in.Observability.Logging != nil {
		if in.Observability.Logging.Destination.DeploymentID != nil {
			obs.DeploymentId = in.Observability.Logging.Destination.DeploymentID
		}
		obs.RefId = &in.Observability.Logging.Destination.RefID
		obs.Logs = true
	}

	if obs == (Observability{}) {
		return nil, nil
	}

	return &obs, nil
}

func ObservabilityPayload(ctx context.Context, list types.List, client *api.API) (*models.DeploymentObservabilitySettings, diag.Diagnostics) {
	var observability *ObservabilityTF

	if diags := utils.GetFirst(ctx, list, &observability); diags.HasError() {
		return nil, nil
	}

	if observability == nil {
		return nil, nil
	}

	var payload models.DeploymentObservabilitySettings

	if observability.DeploymentId.Value == "" {
		return nil, nil
	}

	refID := observability.RefId.Value

	if observability.DeploymentId.Value != "self" && refID == "" {
		// Since ms-77, the refID is optional.
		// To not break ECE users with older versions, we still pre-calculate the refID here
		params := deploymentapi.PopulateRefIDParams{
			Kind:         util.Elasticsearch,
			API:          client,
			DeploymentID: observability.DeploymentId.Value,
			RefID:        ec.String(""),
		}

		if err := deploymentapi.PopulateRefID(params); err != nil {
			var diags diag.Diagnostics
			diags.AddError("observability ref_id auto discovery", err.Error())
			return nil, diags
		}

		refID = *params.RefID
	}

	if observability.Logs.Value {
		payload.Logging = &models.DeploymentLoggingSettings{
			Destination: &models.ObservabilityAbsoluteDeployment{
				DeploymentID: ec.String(observability.DeploymentId.Value),
				RefID:        refID,
			},
		}
	}

	if observability.Metrics.Value {
		payload.Metrics = &models.DeploymentMetricsSettings{
			Destination: &models.ObservabilityAbsoluteDeployment{
				DeploymentID: ec.String(observability.DeploymentId.Value),
				RefID:        refID,
			},
		}
	}

	return &payload, nil
}