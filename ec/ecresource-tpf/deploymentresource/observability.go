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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Observability struct {
	DeploymentId types.String `tfsdk:"deployment_id"`
	RefId        types.String `tfsdk:"ref_id"`
	Logs         types.Bool   `tfsdk:"logs"`
	Metrics      types.Bool   `tfsdk:"metrics"`
}

func NewObservability(in *models.DeploymentSettings) ([]*Observability, error) {
	if in == nil || in.Observability == nil {
		return nil, nil
	}

	var obs Observability

	// We are only accepting a single deployment ID and refID for both logs and metrics.
	// If either of them is not nil the deployment ID and refID will be filled.
	if in.Observability.Metrics != nil {
		if in.Observability.Metrics.Destination.DeploymentID != nil {
			obs.DeploymentId.Value = *in.Observability.Metrics.Destination.DeploymentID
		}

		obs.RefId.Value = in.Observability.Metrics.Destination.RefID
		obs.Metrics.Value = true
	}

	if in.Observability.Logging != nil {
		if in.Observability.Logging.Destination.DeploymentID != nil {
			obs.DeploymentId.Value = *in.Observability.Logging.Destination.DeploymentID
		}
		obs.RefId.Value = in.Observability.Logging.Destination.RefID
		obs.Logs.Value = true
	}

	if obs == (Observability{}) {
		return nil, nil
	}

	return []*Observability{&obs}, nil
}
