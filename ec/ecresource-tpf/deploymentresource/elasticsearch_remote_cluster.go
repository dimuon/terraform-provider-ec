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

type ElasticsearchRemoteClusters []ElasticsearchRemoteCluster

func (rems *ElasticsearchRemoteClusters) fromModel(in []*models.RemoteResourceRef) error {
	if len(in) == 0 {
		*rems = nil
		return nil
	}

	*rems = make(ElasticsearchRemoteClusters, 0, len(in))
	for _, model := range in {
		var remote ElasticsearchRemoteCluster
		remote.fromModel(model)
		*rems = append(*rems, remote)
	}

	return nil
}

type ElasticsearchRemoteCluster struct {
	DeploymentId    types.String `tfsdk:"deployment_id"`
	Alias           types.String `tfsdk:"alias"`
	RefId           types.String `tfsdk:"ref_id"`
	SkipUnavailable types.Bool   `tfsdk:"skip_unavailable"`
}

func (rem *ElasticsearchRemoteCluster) fromModel(in *models.RemoteResourceRef) error {
	if in.DeploymentID != nil && *in.DeploymentID != "" {
		rem.DeploymentId.Value = *in.DeploymentID
	}

	if in.ElasticsearchRefID != nil && *in.ElasticsearchRefID != "" {
		rem.RefId.Value = *in.ElasticsearchRefID
	}

	if in.Alias != nil && *in.Alias != "" {
		rem.Alias.Value = *in.Alias
	}

	if in.SkipUnavailable != nil {
		rem.SkipUnavailable.Value = *in.SkipUnavailable
	}

	return nil
}
