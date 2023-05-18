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

// IdentityServer holds the name of NooBaa's COSI driver
type IdentityServer struct {
	driver string
}

// DriverGetInfo returns the info of the Noobaa's COSI driver
func (id *IdentityServer) DriverGetInfo(ctx context.Context,
	req *cosi.DriverGetInfoRequest) (*cosi.DriverGetInfoResponse, error) {
	log := util.Logger()
	if id.driver == "" {
		log.Errorf("Driver name cannot be empty")
		return nil, status.Error(codes.InvalidArgument, "Driver name is empty")
	}

	return &cosi.DriverGetInfoResponse{
		Name: id.driver,
	}, nil
}
