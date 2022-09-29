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

type ElasticsearchSnapshotSource struct {
	SourceElasticsearchClusterId types.String `tfsdk:"source_elasticsearch_cluster_id"`
	SnapshotName                 types.String `tfsdk:"snapshot_name"`
}

type ElasticsearchSnapshotSources types.List

func (snapshots ElasticsearchSnapshotSources) Payload(ctx context.Context) (*models.TransientElasticsearchPlanConfiguration, diag.Diagnostics) {
	if len(snapshots.Elems) == 0 {
		return nil, nil
	}

	payload := models.TransientElasticsearchPlanConfiguration{
		RestoreSnapshot: &models.RestoreSnapshotConfiguration{},
	}

	for _, elem := range snapshots.Elems {
		var snapshot ElasticsearchSnapshotSource
		if diags := tfsdk.ValueAs(ctx, elem, &snapshot); diags.HasError() {
			return nil, diags
		}

		if !snapshot.SourceElasticsearchClusterId.IsNull() {
			payload.RestoreSnapshot.SourceClusterID = snapshot.SourceElasticsearchClusterId.Value
		}

		if !snapshot.SnapshotName.IsNull() {
			payload.RestoreSnapshot.SnapshotName = &snapshot.SnapshotName.Value
		}
	}

	return &payload, nil
}