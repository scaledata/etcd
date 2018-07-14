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
	"github.com/scaledata/etcd/etcdserver/api/v3election/sdv3electionpb"
)

type electionProxy struct {
	client *clientv3.Client
}

func NewElectionProxy(client *clientv3.Client) sdv3electionpb.ElectionServer {
	return &electionProxy{client: client}
}

func (ep *electionProxy) Campaign(ctx context.Context, req *sdv3electionpb.CampaignRequest) (*sdv3electionpb.CampaignResponse, error) {
	return sdv3electionpb.NewElectionClient(ep.client.ActiveConnection()).Campaign(ctx, req)
}

func (ep *electionProxy) Proclaim(ctx context.Context, req *sdv3electionpb.ProclaimRequest) (*sdv3electionpb.ProclaimResponse, error) {
	return sdv3electionpb.NewElectionClient(ep.client.ActiveConnection()).Proclaim(ctx, req)
}

func (ep *electionProxy) Leader(ctx context.Context, req *sdv3electionpb.LeaderRequest) (*sdv3electionpb.LeaderResponse, error) {
	return sdv3electionpb.NewElectionClient(ep.client.ActiveConnection()).Leader(ctx, req)
}

func (ep *electionProxy) Observe(req *sdv3electionpb.LeaderRequest, s sdv3electionpb.Election_ObserveServer) error {
	conn := ep.client.ActiveConnection()
	ctx, cancel := context.WithCancel(s.Context())
	defer cancel()
	sc, err := sdv3electionpb.NewElectionClient(conn).Observe(ctx, req)
	if err != nil {
		return err
	}
	for {
		rr, err := sc.Recv()
		if err != nil {
			return err
		}
		if err = s.Send(rr); err != nil {
			return err
		}
	}
}

func (ep *electionProxy) Resign(ctx context.Context, req *sdv3electionpb.ResignRequest) (*sdv3electionpb.ResignResponse, error) {
	return sdv3electionpb.NewElectionClient(ep.client.ActiveConnection()).Resign(ctx, req)
}
