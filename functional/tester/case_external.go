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

package tester

import (
	"fmt"
	"os/exec"

	"github.com/scaledata/etcd/functional/sdrpcpb"
)

type caseExternal struct {
	Case

	desc      string
	sdrpcpbCase sdrpcpb.Case

	scriptPath string
}

func (c *caseExternal) Inject(clus *Cluster) error {
	return exec.Command(c.scriptPath, "enable", fmt.Sprintf("%d", clus.rd)).Run()
}

func (c *caseExternal) Recover(clus *Cluster) error {
	return exec.Command(c.scriptPath, "disable", fmt.Sprintf("%d", clus.rd)).Run()
}

func (c *caseExternal) Desc() string {
	return c.desc
}

func (c *caseExternal) TestCase() sdrpcpb.Case {
	return c.sdrpcpbCase
}

func new_Case_EXTERNAL(scriptPath string) Case {
	return &caseExternal{
		desc:       fmt.Sprintf("external fault injector (script: %q)", scriptPath),
		sdrpcpbCase:  sdrpcpb.Case_EXTERNAL,
		scriptPath: scriptPath,
	}
}
