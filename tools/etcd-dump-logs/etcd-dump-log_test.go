// Copyright 2018 The etcd Authors
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

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scaledata/etcd/auth/sdauthpb"
	"github.com/scaledata/etcd/etcdserver/sdetcdserverpb"
	"github.com/scaledata/etcd/pkg/fileutil"
	"github.com/scaledata/etcd/pkg/pbutil"
	"github.com/scaledata/etcd/raft/sdraftpb"
	"github.com/scaledata/etcd/wal"
	"go.uber.org/zap"
)

func TestEtcdDumpLogEntryType(t *testing.T) {
	// directory where the command is
	binDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	dumpLogsBinary := path.Join(binDir + "/etcd-dump-logs")
	if !fileutil.Exist(dumpLogsBinary) {
		t.Skipf("%q does not exist", dumpLogsBinary)
	}

	decoder_correctoutputformat := filepath.Join(binDir, "/testdecoder/decoder_correctoutputformat.sh")
	decoder_wrongoutputformat := filepath.Join(binDir, "/testdecoder/decoder_wrongoutputformat.sh")

	p, err := ioutil.TempDir(os.TempDir(), "etcddumplogstest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(p)

	memberdir := filepath.Join(p, "member")
	err = os.Mkdir(memberdir, 0744)
	if err != nil {
		t.Fatal(err)
	}
	waldir := walDir(p)
	snapdir := snapDir(p)

	w, err := wal.Create(zap.NewExample(), waldir, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(snapdir, 0744)
	if err != nil {
		t.Fatal(err)
	}

	ents := make([]sdraftpb.Entry, 0)

	// append entries into wal log
	appendConfigChangeEnts(&ents)
	appendNormalRequestEnts(&ents)
	appendNormalIRREnts(&ents)
	appendUnknownNormalEnts(&ents)

	// force commit newly appended entries
	err = w.Save(sdraftpb.HardState{}, ents)
	if err != nil {
		t.Fatal(err)
	}
	w.Close()

	argtests := []struct {
		name         string
		args         []string
		fileExpected string
	}{
		{"no entry-type", []string{p}, "expectedoutput/listAll.output"},
		{"confchange entry-type", []string{"-entry-type", "ConfigChange", p}, "expectedoutput/listConfigChange.output"},
		{"normal entry-type", []string{"-entry-type", "Normal", p}, "expectedoutput/listNormal.output"},
		{"request entry-type", []string{"-entry-type", "Request", p}, "expectedoutput/listRequest.output"},
		{"internalRaftRequest entry-type", []string{"-entry-type", "InternalRaftRequest", p}, "expectedoutput/listInternalRaftRequest.output"},
		{"range entry-type", []string{"-entry-type", "IRRRange", p}, "expectedoutput/listIRRRange.output"},
		{"put entry-type", []string{"-entry-type", "IRRPut", p}, "expectedoutput/listIRRPut.output"},
		{"del entry-type", []string{"-entry-type", "IRRDeleteRange", p}, "expectedoutput/listIRRDeleteRange.output"},
		{"txn entry-type", []string{"-entry-type", "IRRTxn", p}, "expectedoutput/listIRRTxn.output"},
		{"compaction entry-type", []string{"-entry-type", "IRRCompaction", p}, "expectedoutput/listIRRCompaction.output"},
		{"lease grant entry-type", []string{"-entry-type", "IRRLeaseGrant", p}, "expectedoutput/listIRRLeaseGrant.output"},
		{"lease revoke entry-type", []string{"-entry-type", "IRRLeaseRevoke", p}, "expectedoutput/listIRRLeaseRevoke.output"},
		{"confchange and txn entry-type", []string{"-entry-type", "ConfigChange,IRRCompaction", p}, "expectedoutput/listConfigChangeIRRCompaction.output"},
		{"decoder_correctoutputformat", []string{"-stream-decoder", decoder_correctoutputformat, p}, "expectedoutput/decoder_correctoutputformat.output"},
		{"decoder_wrongoutputformat", []string{"-stream-decoder", decoder_wrongoutputformat, p}, "expectedoutput/decoder_wrongoutputformat.output"},
	}

	for _, argtest := range argtests {
		t.Run(argtest.name, func(t *testing.T) {
			cmd := exec.Command(dumpLogsBinary, argtest.args...)
			actual, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatal(err)
			}
			expected, err := ioutil.ReadFile(path.Join(binDir, argtest.fileExpected))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(actual, expected) {
				t.Errorf(`Got input of length %d, wanted input of length %d
==== BEGIN RECEIVED FILE ====
%s
==== END RECEIVED FILE ====
==== BEGIN EXPECTED FILE ====
%s
==== END EXPECTED FILE ====
`, len(actual), len(expected), actual, expected)
			}
		})
	}

}

func appendConfigChangeEnts(ents *[]sdraftpb.Entry) {
	configChangeData := []sdraftpb.ConfChange{
		{ID: 1, Type: sdraftpb.ConfChangeAddNode, NodeID: 2, Context: []byte("")},
		{ID: 2, Type: sdraftpb.ConfChangeRemoveNode, NodeID: 2, Context: []byte("")},
		{ID: 3, Type: sdraftpb.ConfChangeUpdateNode, NodeID: 2, Context: []byte("")},
		{ID: 4, Type: sdraftpb.ConfChangeAddLearnerNode, NodeID: 3, Context: []byte("")},
	}
	configChangeEntries := []sdraftpb.Entry{
		{Term: 1, Index: 1, Type: sdraftpb.EntryConfChange, Data: pbutil.MustMarshal(&configChangeData[0])},
		{Term: 2, Index: 2, Type: sdraftpb.EntryConfChange, Data: pbutil.MustMarshal(&configChangeData[1])},
		{Term: 2, Index: 3, Type: sdraftpb.EntryConfChange, Data: pbutil.MustMarshal(&configChangeData[2])},
		{Term: 2, Index: 4, Type: sdraftpb.EntryConfChange, Data: pbutil.MustMarshal(&configChangeData[3])},
	}
	*ents = append(*ents, configChangeEntries...)
}

func appendNormalRequestEnts(ents *[]sdraftpb.Entry) {
	a := true
	b := false

	requests := []sdetcdserverpb.Request{
		{ID: 0, Method: "", Path: "/path0", Val: "{\"hey\":\"ho\",\"hi\":[\"yo\"]}", Dir: true, PrevValue: "", PrevIndex: 0, PrevExist: &b, Expiration: 9, Wait: false, Since: 1, Recursive: false, Sorted: false, Quorum: false, Time: 1, Stream: false, Refresh: &b},
		{ID: 1, Method: "QGET", Path: "/path1", Val: "{\"0\":\"1\",\"2\":[\"3\"]}", Dir: false, PrevValue: "", PrevIndex: 0, PrevExist: &b, Expiration: 9, Wait: false, Since: 1, Recursive: false, Sorted: false, Quorum: false, Time: 1, Stream: false, Refresh: &b},
		{ID: 2, Method: "SYNC", Path: "/path2", Val: "{\"0\":\"1\",\"2\":[\"3\"]}", Dir: false, PrevValue: "", PrevIndex: 0, PrevExist: &b, Expiration: 2, Wait: false, Since: 1, Recursive: false, Sorted: false, Quorum: false, Time: 1, Stream: false, Refresh: &b},
		{ID: 3, Method: "DELETE", Path: "/path3", Val: "{\"hey\":\"ho\",\"hi\":[\"yo\"]}", Dir: false, PrevValue: "", PrevIndex: 0, PrevExist: &a, Expiration: 2, Wait: false, Since: 1, Recursive: false, Sorted: false, Quorum: false, Time: 1, Stream: false, Refresh: &b},
		{ID: 4, Method: "RANDOM", Path: "/path4/superlong" + strings.Repeat("/path", 30), Val: "{\"hey\":\"ho\",\"hi\":[\"yo\"]}", Dir: false, PrevValue: "", PrevIndex: 0, PrevExist: &b, Expiration: 2, Wait: false, Since: 1, Recursive: false, Sorted: false, Quorum: false, Time: 1, Stream: false, Refresh: &b},
	}

	for i, request := range requests {
		var currentry sdraftpb.Entry
		currentry.Term = 3
		currentry.Index = uint64(i + 5)
		currentry.Type = sdraftpb.EntryNormal
		currentry.Data = pbutil.MustMarshal(&request)
		*ents = append(*ents, currentry)
	}
}

func appendNormalIRREnts(ents *[]sdraftpb.Entry) {
	irrrange := &sdetcdserverpb.RangeRequest{Key: []byte("1"), RangeEnd: []byte("hi"), Limit: 6, Revision: 1, SortOrder: 1, SortTarget: 0, Serializable: false, KeysOnly: false, CountOnly: false, MinModRevision: 0, MaxModRevision: 20000, MinCreateRevision: 0, MaxCreateRevision: 20000}

	irrput := &sdetcdserverpb.PutRequest{Key: []byte("foo1"), Value: []byte("bar1"), Lease: 1, PrevKv: false, IgnoreValue: false, IgnoreLease: true}

	irrdeleterange := &sdetcdserverpb.DeleteRangeRequest{Key: []byte("0"), RangeEnd: []byte("9"), PrevKv: true}

	delInRangeReq := &sdetcdserverpb.RequestOp{Request: &sdetcdserverpb.RequestOp_RequestDeleteRange{
		RequestDeleteRange: &sdetcdserverpb.DeleteRangeRequest{
			Key: []byte("a"), RangeEnd: []byte("b"),
		},
	},
	}

	irrtxn := &sdetcdserverpb.TxnRequest{Success: []*sdetcdserverpb.RequestOp{delInRangeReq}, Failure: []*sdetcdserverpb.RequestOp{delInRangeReq}}

	irrcompaction := &sdetcdserverpb.CompactionRequest{Revision: 0, Physical: true}

	irrleasegrant := &sdetcdserverpb.LeaseGrantRequest{TTL: 1, ID: 1}

	irrleaserevoke := &sdetcdserverpb.LeaseRevokeRequest{ID: 2}

	irralarm := &sdetcdserverpb.AlarmRequest{Action: 3, MemberID: 4, Alarm: 5}

	irrauthenable := &sdetcdserverpb.AuthEnableRequest{}

	irrauthdisable := &sdetcdserverpb.AuthDisableRequest{}

	irrauthenticate := &sdetcdserverpb.InternalAuthenticateRequest{Name: "myname", Password: "password", SimpleToken: "token"}

	irrauthuseradd := &sdetcdserverpb.AuthUserAddRequest{Name: "name1", Password: "pass1"}

	irrauthuserdelete := &sdetcdserverpb.AuthUserDeleteRequest{Name: "name1"}

	irrauthuserget := &sdetcdserverpb.AuthUserGetRequest{Name: "name1"}

	irrauthuserchangepassword := &sdetcdserverpb.AuthUserChangePasswordRequest{Name: "name1", Password: "pass2"}

	irrauthusergrantrole := &sdetcdserverpb.AuthUserGrantRoleRequest{User: "user1", Role: "role1"}

	irrauthuserrevokerole := &sdetcdserverpb.AuthUserRevokeRoleRequest{Name: "user2", Role: "role2"}

	irrauthuserlist := &sdetcdserverpb.AuthUserListRequest{}

	irrauthrolelist := &sdetcdserverpb.AuthRoleListRequest{}

	irrauthroleadd := &sdetcdserverpb.AuthRoleAddRequest{Name: "role2"}

	irrauthroledelete := &sdetcdserverpb.AuthRoleDeleteRequest{Role: "role1"}

	irrauthroleget := &sdetcdserverpb.AuthRoleGetRequest{Role: "role3"}

	perm := &sdauthpb.Permission{
		PermType: sdauthpb.WRITE,
		Key:      []byte("Keys"),
		RangeEnd: []byte("RangeEnd"),
	}

	irrauthrolegrantpermission := &sdetcdserverpb.AuthRoleGrantPermissionRequest{Name: "role3", Perm: perm}

	irrauthrolerevokepermission := &sdetcdserverpb.AuthRoleRevokePermissionRequest{Role: "role3", Key: []byte("key"), RangeEnd: []byte("rangeend")}

	irrs := []sdetcdserverpb.InternalRaftRequest{
		{ID: 5, Range: irrrange},
		{ID: 6, Put: irrput},
		{ID: 7, DeleteRange: irrdeleterange},
		{ID: 8, Txn: irrtxn},
		{ID: 9, Compaction: irrcompaction},
		{ID: 10, LeaseGrant: irrleasegrant},
		{ID: 11, LeaseRevoke: irrleaserevoke},
		{ID: 12, Alarm: irralarm},
		{ID: 13, AuthEnable: irrauthenable},
		{ID: 14, AuthDisable: irrauthdisable},
		{ID: 15, Authenticate: irrauthenticate},
		{ID: 16, AuthUserAdd: irrauthuseradd},
		{ID: 17, AuthUserDelete: irrauthuserdelete},
		{ID: 18, AuthUserGet: irrauthuserget},
		{ID: 19, AuthUserChangePassword: irrauthuserchangepassword},
		{ID: 20, AuthUserGrantRole: irrauthusergrantrole},
		{ID: 21, AuthUserRevokeRole: irrauthuserrevokerole},
		{ID: 22, AuthUserList: irrauthuserlist},
		{ID: 23, AuthRoleList: irrauthrolelist},
		{ID: 24, AuthRoleAdd: irrauthroleadd},
		{ID: 25, AuthRoleDelete: irrauthroledelete},
		{ID: 26, AuthRoleGet: irrauthroleget},
		{ID: 27, AuthRoleGrantPermission: irrauthrolegrantpermission},
		{ID: 28, AuthRoleRevokePermission: irrauthrolerevokepermission},
	}

	for i, irr := range irrs {
		var currentry sdraftpb.Entry
		currentry.Term = uint64(i + 4)
		currentry.Index = uint64(i + 10)
		currentry.Type = sdraftpb.EntryNormal
		currentry.Data = pbutil.MustMarshal(&irr)
		*ents = append(*ents, currentry)
	}
}

func appendUnknownNormalEnts(ents *[]sdraftpb.Entry) {
	var currentry sdraftpb.Entry
	currentry.Term = 27
	currentry.Index = 34
	currentry.Type = sdraftpb.EntryNormal
	currentry.Data = []byte("?")
	*ents = append(*ents, currentry)
}
