// Copyright 2017 The etcd Lockors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpcproxy

import (
	"context"

	"github.com/scaledata/etcd/clientv3"
	"github.com/scaledata/etcd/etcdserver/api/v3lock/sdv3lockpb"
)

type lockProxy struct {
	client *clientv3.Client
}

func NewLockProxy(client *clientv3.Client) sdv3lockpb.LockServer {
	return &lockProxy{client: client}
}

func (lp *lockProxy) Lock(ctx context.Context, req *sdv3lockpb.LockRequest) (*sdv3lockpb.LockResponse, error) {
	return sdv3lockpb.NewLockClient(lp.client.ActiveConnection()).Lock(ctx, req)
}

func (lp *lockProxy) Unlock(ctx context.Context, req *sdv3lockpb.UnlockRequest) (*sdv3lockpb.UnlockResponse, error) {
	return sdv3lockpb.NewLockClient(lp.client.ActiveConnection()).Unlock(ctx, req)
}
