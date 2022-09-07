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
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/api"
)

// CreateResource will createResource a new deployment from the specified settings.
func Create(ctx context.Context, client *api.API, cfg, plan *DeploymentData) (resp *DeploymentData, errors []error) {
	// reqID := deploymentapi.RequestID("")

	// req, err := createResourceToModel(client, cfg, plan)
	// if err != nil {
	// 	return nil, []error{err}
	// }

	// res, err := deploymentapi.Create(deploymentapi.CreateParams{
	// 	API:       client,
	// 	RequestID: reqID,
	// 	Request:   req,
	// 	Overrides: &deploymentapi.PayloadOverrides{
	// 		Name:    plan.Name.Value,
	// 		Version: plan.Version.Value,
	// 		Region:  plan.Region.Value,
	// 	},
	// })

	// if err != nil {
	// 	merr := multierror.NewPrefixed("failed creating deployment", err)
	// 	return nil, []error{merr.Append(newCreationError(reqID))}
	// }

	// if err := WaitForPlanCompletion(client, *res.ID); err != nil {
	// 	merr := multierror.NewPrefixed("failed tracking create progress", err)
	// 	return nil, []error{merr.Append(newCreationError(reqID))}
	// }

	// d.SetId(*res.ID)

	// if err := handleRemoteClusters(d, client); err != nil {
	// 	errors = append(errors, err)
	// }

	/* 	if errs := readResource(ctx, d, meta); errs != nil {
	   		errors = append(errors, errs...)
	   	}

	   	if err := parseCredentials(d, res.Resources); err != nil {
	   		errors = append(errors, err)
	   	}

	   	return errors
	*/
	return nil, nil
}

func newCreationError(reqID string) error {
	return fmt.Errorf(
		`set "request_id" to "%s" to recreate the deployment resources`, reqID,
	)
}
