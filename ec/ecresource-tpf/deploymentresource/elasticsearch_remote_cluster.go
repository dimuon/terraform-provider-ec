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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchRemoteClusters types.Set

func (clusters ElasticsearchRemoteClusters) Read(ctx context.Context, in []*models.RemoteResourceRef) diag.Diagnostics {
	if len(in) == 0 {
		return nil
	}

	rems := make([]ElasticsearchRemoteCluster, 0, len(in))

	for _, model := range in {
		var cluster ElasticsearchRemoteCluster
		diags := cluster.Read(ctx, model)
		if diags.HasError() {
			return diags
		}
		rems = append(rems, cluster)
	}

	return tfsdk.ValueFrom(ctx, rems, elasticsearchRemoteCluster().Type(), clusters)
}

func (clusters ElasticsearchRemoteClusters) Payload(ctx context.Context) (*models.RemoteResources, diag.Diagnostics) {
	payloads := models.RemoteResources{Resources: []*models.RemoteResourceRef{}}

	for _, elem := range clusters.Elems {
		var cluster ElasticsearchRemoteCluster
		diags := tfsdk.ValueAs(ctx, elem, &cluster)

		if diags.HasError() {
			return nil, diags
		}
		var payload models.RemoteResourceRef

		if !cluster.DeploymentId.IsNull() {
			payload.DeploymentID = &cluster.DeploymentId.Value
		}

		if !cluster.RefId.IsNull() {
			payload.ElasticsearchRefID = &cluster.RefId.Value
		}

		if !cluster.Alias.IsNull() {
			payload.Alias = &cluster.Alias.Value
		}

		if !cluster.SkipUnavailable.IsNull() {
			payload.SkipUnavailable = &cluster.SkipUnavailable.Value
		}

		payloads.Resources = append(payloads.Resources, &payload)
	}

	return &payloads, nil
}

type ElasticsearchRemoteCluster struct {
	DeploymentId    types.String `tfsdk:"deployment_id"`
	Alias           types.String `tfsdk:"alias"`
	RefId           types.String `tfsdk:"ref_id"`
	SkipUnavailable types.Bool   `tfsdk:"skip_unavailable"`
}

func (cluster *ElasticsearchRemoteCluster) Read(_ context.Context, in *models.RemoteResourceRef) diag.Diagnostics {
	if in.DeploymentID != nil && *in.DeploymentID != "" {
		cluster.DeploymentId = types.String{Value: *in.DeploymentID}
	}

	if in.ElasticsearchRefID != nil && *in.ElasticsearchRefID != "" {
		cluster.RefId = types.String{Value: *in.ElasticsearchRefID}
	}

	if in.Alias != nil && *in.Alias != "" {
		cluster.Alias = types.String{Value: *in.Alias}
	}

	if in.SkipUnavailable != nil {
		cluster.SkipUnavailable = types.Bool{Value: *in.SkipUnavailable}
	}

	return nil
}
