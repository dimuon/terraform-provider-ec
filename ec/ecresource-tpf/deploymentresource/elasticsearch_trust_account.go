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

type ElasticsearchTrustAccounts []ElasticsearchTrustAccount

func (accs *ElasticsearchTrustAccounts) fromModel(in *models.ElasticsearchClusterTrustSettings) error {
	if in == nil || len(in.Accounts) == 0 {
		*accs = nil
		return nil
	}

	*accs = make(ElasticsearchTrustAccounts, 0, len(in.Accounts))
	for _, model := range in.Accounts {
		var acc ElasticsearchTrustAccount
		if err := acc.fromModel(model); err != nil {
			return err
		}
		*accs = append(*accs, acc)
	}

	return nil
}

type ElasticsearchTrustAccount struct {
	AccountId      types.String `tfsdk:"account_id"`
	TrustAll       types.Bool   `tfsdk:"trust_all"`
	TrustAllowlist []string     `tfsdk:"trust_allowlist"`
}

func (acc *ElasticsearchTrustAccount) fromModel(in *models.AccountTrustRelationship) error {
	if in.AccountID != nil {
		acc.AccountId.Value = *in.AccountID
	}
	if in.TrustAll != nil {
		acc.TrustAll.Value = *in.TrustAll
	}
	if in.TrustAllowlist != nil {
		acc.TrustAllowlist = *&in.TrustAllowlist
	}
	return nil
}
