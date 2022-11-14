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

package acc

import (
	//"fmt"
	//"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDeployment_UpgradeFrom0_4_1(t *testing.T) {
	resName := "ec_deployment.defaults"
	randomName := prefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	oldConfigFile := "testdata/deployment_state_upgrade_v0.tf"
	newConfigFile := "testdata/deployment_state_upgrade_v2.tf"
	oldConfig := fixtureAccDeploymentResourceBasicDefaults(t, oldConfigFile, randomName, getRegion(), defaultTemplate)
	newConfig := fixtureAccDeploymentResourceBasicDefaults(t, newConfigFile, randomName, getRegion(), defaultTemplate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"ec": {
						VersionConstraint: "0.4.1",
						Source:            "elastic/ec",
					},
				},
				Config: oldConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "elasticsearch.#", "1"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.size_resource", "memory"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_data", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ingest", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_master", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.node_type_ml", ""),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.id", "hot_content"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.0.topology.0.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.0.topology.0.zone_count", "2"),
					resource.TestCheckResourceAttr(resName, "kibana.#", "0"),
					resource.TestCheckResourceAttr(resName, "apm.#", "0"),
					resource.TestCheckResourceAttr(resName, "enterprise_search.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: testAccProviderFactory,
				Config:                   newConfig,
				PlanOnly:                 true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.instance_configuration_id"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size", "1g"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.size_resource", "memory"),
					resource.TestCheckResourceAttrSet(resName, "elasticsearch.hot.node_roles.#"),
					resource.TestCheckResourceAttr(resName, "elasticsearch.hot.zone_count", "3"),
					resource.TestCheckNoResourceAttr(resName, "kibana"),
					resource.TestCheckNoResourceAttr(resName, "apm"),
					resource.TestCheckNoResourceAttr(resName, "enterprise_search"),
				),
			},
		},
	})
}

/*
func fixtureAccDeploymentResourceBasicDefaults(t *testing.T, fileName, name, region, depTpl string) string {
	t.Helper()

	deploymentTpl := setDefaultTemplate(region, depTpl)
	b, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(string(b), region, name, region, deploymentTpl)
}
*/
