/*
Copyright 2018 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	"fmt"
	"path/filepath"

	"github.com/srcmesh/kubefaas/pkg/controller/client"
	"github.com/srcmesh/kubefaas/pkg/cli/util"
)

type FissionVersion struct {
	client client.Interface
}

func NewFissionVersion(client client.Interface) Resource {
	return FissionVersion{client: client}
}

func (res FissionVersion) Dump(dumpDir string) {
	ver := util.GetVersion(res.client)
	file := filepath.Clean(fmt.Sprintf("%v/%v", dumpDir, "kubefaas-version.txt"))
	writeToFile(file, ver)
}