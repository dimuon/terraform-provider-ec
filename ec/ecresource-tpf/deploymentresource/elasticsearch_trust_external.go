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
)

type ElasticsearchTrustExternals []ElasticsearchTrustExternal

func (exts *ElasticsearchTrustExternals) fromModel(in *models.ElasticsearchClusterTrustSettings) error {
	if in == nil || len(in.External) == 0 {
		*exts = nil
		return nil
	}

	*exts = make(ElasticsearchTrustExternals, 0, len(in.External))
	for _, model := range in.External {
		var ext ElasticsearchTrustExternal
		if err := ext.fromModel(model); err != nil {
			return err
		}
		*exts = append(*exts, ext)
	}

	return nil

}

type ElasticsearchTrustExternal struct {
	RelationshipId string   `tfsdk:"relationship_id"`
	TrustAll       bool     `tfsdk:"trust_all"`
	TrustAllowlist []string `tfsdk:"trust_allowlist"`
}

func (ext *ElasticsearchTrustExternal) fromModel(in *models.ExternalTrustRelationship) error {
	if in.TrustRelationshipID != nil {
		ext.RelationshipId = *in.TrustRelationshipID
	}
	if in.TrustAll != nil {
		ext.TrustAll = *in.TrustAll
	}
	if in.TrustAllowlist != nil {
		ext.TrustAllowlist = make([]string, 0, len(in.TrustAllowlist))
		ext.TrustAllowlist = append(ext.TrustAllowlist, in.TrustAllowlist...)
	}
	return nil
}
