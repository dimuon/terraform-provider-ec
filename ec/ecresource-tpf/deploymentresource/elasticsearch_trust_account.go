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

type ElasticsearchTrustAccountTF struct {
	AccountId      types.String `tfsdk:"account_id"`
	TrustAll       types.Bool   `tfsdk:"trust_all"`
	TrustAllowlist types.List   `tfsdk:"trust_allowlist"`
}

type ElasticsearchTrustAccountsTF types.Set

type ElasticsearchTrustAccount struct {
	AccountId      *string  `tfsdk:"account_id"`
	TrustAll       *bool    `tfsdk:"trust_all"`
	TrustAllowlist []string `tfsdk:"trust_allowlist"`
}

type ElasticsearchTrustAccounts []ElasticsearchTrustAccount

func readElasticsearchTrustAccounts(in *models.ElasticsearchClusterTrustSettings) (ElasticsearchTrustAccounts, error) {
	if in == nil {
		return nil, nil
	}

	accounts := make(ElasticsearchTrustAccounts, 0, len(in.Accounts))

	for _, model := range in.Accounts {
		account, err := readElasticsearchTrustAccount(model)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, *account)
	}

	return accounts, nil
}

func (accounts ElasticsearchTrustAccountsTF) Payload(ctx context.Context, model *models.ElasticsearchClusterSettings) (*models.ElasticsearchClusterSettings, diag.Diagnostics) {
	payloads := make([]*models.AccountTrustRelationship, 0, len(accounts.Elems))

	for _, elem := range accounts.Elems {
		var account ElasticsearchTrustAccountTF
		if diags := tfsdk.ValueAs(ctx, elem, &account); diags.HasError() {
			return nil, diags
		}
		id := account.AccountId.Value
		all := account.TrustAll.Value

		payload := &models.AccountTrustRelationship{
			AccountID: &id,
			TrustAll:  &all,
		}
		payloads = append(payloads, payload)
		if diags := tfsdk.ValueAs(ctx, account.TrustAllowlist, payload.TrustAllowlist); diags.HasError() {
			return nil, diags
		}
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

	model.Trust.Accounts = append(model.Trust.Accounts, payloads...)

	return model, nil
}

func readElasticsearchTrustAccount(in *models.AccountTrustRelationship) (*ElasticsearchTrustAccount, error) {
	var acc ElasticsearchTrustAccount

	if in.AccountID != nil {
		acc.AccountId = in.AccountID
	}

	if in.TrustAll != nil {
		acc.TrustAll = in.TrustAll
	}

	acc.TrustAllowlist = in.TrustAllowlist

	return &acc, nil
}
