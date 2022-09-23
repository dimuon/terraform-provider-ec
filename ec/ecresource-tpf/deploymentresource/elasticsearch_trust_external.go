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

type ElasticsearchTrustExternals []*ElasticsearchTrustExternal

func NewElasticsearchTrustExternals(in *models.ElasticsearchClusterTrustSettings) (ElasticsearchTrustExternals, error) {
	if in == nil || len(in.External) == 0 {
		return nil, nil
	}

	exts := make([]*ElasticsearchTrustExternal, 0, len(in.External))
	for _, model := range in.External {
		ext, err := NewElasticsearchTrustExternal(model)
		if err != nil {
			return nil, err
		}
		exts = append(exts, ext)
	}

	return exts, nil
}

func (externals ElasticsearchTrustExternals) Payload(model *models.ElasticsearchClusterSettings) *models.ElasticsearchClusterSettings {
	payloads := make([]*models.ExternalTrustRelationship, 0, len(externals))

	for _, external := range externals {
		id := external.RelationshipId.Value
		all := external.TrustAll.Value

		payloads = append(payloads, &models.ExternalTrustRelationship{
			TrustRelationshipID: &id,
			TrustAll:            &all,
			TrustAllowlist:      external.TrustAllowlist,
		})
	}

	if len(payloads) == 0 {
		return nil
	}

	if model == nil {
		model = &models.ElasticsearchClusterSettings{}
	}

	if model.Trust == nil {
		model.Trust = &models.ElasticsearchClusterTrustSettings{}
	}

	model.Trust.External = append(model.Trust.External, payloads...)

	return model
}

type ElasticsearchTrustExternal struct {
	RelationshipId types.String `tfsdk:"relationship_id"`
	TrustAll       types.Bool   `tfsdk:"trust_all"`
	TrustAllowlist []string     `tfsdk:"trust_allowlist"`
}

func NewElasticsearchTrustExternal(in *models.ExternalTrustRelationship) (*ElasticsearchTrustExternal, error) {
	var ext ElasticsearchTrustExternal
	if in.TrustRelationshipID != nil {
		ext.RelationshipId.Value = *in.TrustRelationshipID
	}
	if in.TrustAll != nil {
		ext.TrustAll.Value = *in.TrustAll
	}
	if in.TrustAllowlist != nil {
		ext.TrustAllowlist = make([]string, 0, len(in.TrustAllowlist))
		ext.TrustAllowlist = append(ext.TrustAllowlist, in.TrustAllowlist...)
	}
	return &ext, nil
}
