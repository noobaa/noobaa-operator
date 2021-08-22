/*
Copyright 2021 The Ceph-COSI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cosi

import (
	"context"

	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cosi "sigs.k8s.io/container-object-storage-interface-spec"
)

// IdentityServer holds the name of NooBaa's COSI provisioner
type IdentityServer struct {
	provisioner string
}

// ProvisionerGetInfo returns the info of the Noobaa's COSI provisioner
func (id *IdentityServer) ProvisionerGetInfo(ctx context.Context,
	req *cosi.ProvisionerGetInfoRequest) (*cosi.ProvisionerGetInfoResponse, error) {
	log := util.Logger()
	if id.provisioner == "" {
		log.Errorf("Provisioner name cannot be empty")
		return nil, status.Error(codes.InvalidArgument, "Provisioner name is empty")
	}

	return &cosi.ProvisionerGetInfoResponse{
		Name: id.provisioner,
	}, nil
}
