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
	"github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchSnapshotSourceTF struct {
	SourceElasticsearchClusterId types.String `tfsdk:"source_elasticsearch_cluster_id"`
	SnapshotName                 types.String `tfsdk:"snapshot_name"`
}

type ElasticsearchSnapshotSource struct {
	SourceElasticsearchClusterId string `tfsdk:"source_elasticsearch_cluster_id"`
	SnapshotName                 string `tfsdk:"snapshot_name"`
}

type ElasticsearchSnapshotSources []ElasticsearchSnapshotSource

func elasticsearchSnapshotSourcePayload(ctx context.Context, list types.List, payload *models.ElasticsearchClusterPlan) diag.Diagnostics {
	var snapshot *ElasticsearchSnapshotSourceTF

	if diags := utils.GetFirst(ctx, list, &snapshot); diags.HasError() {
		return diags
	}

	if snapshot == nil {
		return nil
	}

	if payload.Transient == nil {
		payload.Transient = &models.TransientElasticsearchPlanConfiguration{
			RestoreSnapshot: &models.RestoreSnapshotConfiguration{},
		}
	}

	if !snapshot.SourceElasticsearchClusterId.IsNull() {
		payload.Transient.RestoreSnapshot.SourceClusterID = snapshot.SourceElasticsearchClusterId.Value
	}

	if !snapshot.SnapshotName.IsNull() {
		payload.Transient.RestoreSnapshot.SnapshotName = &snapshot.SnapshotName.Value
	}

	return nil
}
