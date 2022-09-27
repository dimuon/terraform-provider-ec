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

type ElasticsearchTrustExternals types.Set

func (externals ElasticsearchTrustExternals) Read(ctx context.Context, in *models.ElasticsearchClusterTrustSettings) diag.Diagnostics {
	if in == nil || len(in.External) == 0 {
		return nil
	}

	exts := make([]ElasticsearchTrustExternal, 0, len(in.External))
	for _, model := range in.External {
		var external ElasticsearchTrustExternal
		if diags := external.Read(ctx, model); diags.HasError() {
			return nil
		}
		exts = append(exts, external)
	}

	return tfsdk.ValueFrom(ctx, exts, elasticsearchTrustExternal().FrameworkType(), externals)
}

func (externals ElasticsearchTrustExternals) Payload(ctx context.Context, model *models.ElasticsearchClusterSettings) (*models.ElasticsearchClusterSettings, diag.Diagnostics) {
	payloads := make([]*models.ExternalTrustRelationship, 0, len(externals.Elems))

	for _, elem := range externals.Elems {
		var external ElasticsearchTrustExternal
		if diags := tfsdk.ValueAs(ctx, elem, &external); diags.HasError() {
			return nil, diags
		}
		id := external.RelationshipId.Value
		all := external.TrustAll.Value

		payload := &models.ExternalTrustRelationship{
			TrustRelationshipID: &id,
			TrustAll:            &all,
		}
		if diags := tfsdk.ValueAs(ctx, external.TrustAllowlist, payload.TrustAllowlist); diags.HasError() {
			return nil, diags
		}
		payloads = append(payloads, payload)
	}

	if len(payloads) == 0 {
		return nil, nil
	}

	if model == nil {
		model = &models.ElasticsearchClusterSettings{}
	}

	if model.Trust == nil {
		model.Trust = &models.ElasticsearchClusterTrustSettings{}
	}

	model.Trust.External = append(model.Trust.External, payloads...)

	return model, nil
}

type ElasticsearchTrustExternal struct {
	RelationshipId types.String `tfsdk:"relationship_id"`
	TrustAll       types.Bool   `tfsdk:"trust_all"`
	TrustAllowlist types.Set    `tfsdk:"trust_allowlist"`
}

func (ext ElasticsearchTrustExternal) Read(ctx context.Context, in *models.ExternalTrustRelationship) diag.Diagnostics {
	if in.TrustRelationshipID != nil {
		ext.RelationshipId = types.String{Value: *in.TrustRelationshipID}
	}
	if in.TrustAll != nil {
		ext.TrustAll = types.Bool{Value: *in.TrustAll}
	}
	if in.TrustAllowlist != nil {
		if diags := tfsdk.ValueFrom(ctx, in.TrustAllowlist, types.ListType{ElemType: types.StringType}, ext.TrustAllowlist); diags.HasError() {
			return diags
		}
	}
	return nil
}
