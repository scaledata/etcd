// Copyright 2017 The etcd Authors
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

package adapter

import (
	"context"

	"github.com/scaledata/etcd/etcdserver/api/v3election/sdv3electionpb"

	"google.golang.org/grpc"
)

type es2ec struct{ es sdv3electionpb.ElectionServer }

func ElectionServerToElectionClient(es sdv3electionpb.ElectionServer) sdv3electionpb.ElectionClient {
	return &es2ec{es}
}

func (s *es2ec) Campaign(ctx context.Context, r *sdv3electionpb.CampaignRequest, opts ...grpc.CallOption) (*sdv3electionpb.CampaignResponse, error) {
	return s.es.Campaign(ctx, r)
}

func (s *es2ec) Proclaim(ctx context.Context, r *sdv3electionpb.ProclaimRequest, opts ...grpc.CallOption) (*sdv3electionpb.ProclaimResponse, error) {
	return s.es.Proclaim(ctx, r)
}

func (s *es2ec) Leader(ctx context.Context, r *sdv3electionpb.LeaderRequest, opts ...grpc.CallOption) (*sdv3electionpb.LeaderResponse, error) {
	return s.es.Leader(ctx, r)
}

func (s *es2ec) Resign(ctx context.Context, r *sdv3electionpb.ResignRequest, opts ...grpc.CallOption) (*sdv3electionpb.ResignResponse, error) {
	return s.es.Resign(ctx, r)
}

func (s *es2ec) Observe(ctx context.Context, in *sdv3electionpb.LeaderRequest, opts ...grpc.CallOption) (sdv3electionpb.Election_ObserveClient, error) {
	cs := newPipeStream(ctx, func(ss chanServerStream) error {
		return s.es.Observe(in, &es2ecServerStream{ss})
	})
	return &es2ecClientStream{cs}, nil
}

// es2ecClientStream implements Election_ObserveClient
type es2ecClientStream struct{ chanClientStream }

// es2ecServerStream implements Election_ObserveServer
type es2ecServerStream struct{ chanServerStream }

func (s *es2ecClientStream) Send(rr *sdv3electionpb.LeaderRequest) error {
	return s.SendMsg(rr)
}
func (s *es2ecClientStream) Recv() (*sdv3electionpb.LeaderResponse, error) {
	var v interface{}
	if err := s.RecvMsg(&v); err != nil {
		return nil, err
	}
	return v.(*sdv3electionpb.LeaderResponse), nil
}

func (s *es2ecServerStream) Send(rr *sdv3electionpb.LeaderResponse) error {
	return s.SendMsg(rr)
}
func (s *es2ecServerStream) Recv() (*sdv3electionpb.LeaderRequest, error) {
	var v interface{}
	if err := s.RecvMsg(&v); err != nil {
		return nil, err
	}
	return v.(*sdv3electionpb.LeaderRequest), nil
}
