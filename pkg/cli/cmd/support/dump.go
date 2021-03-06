/*
Copyright 2019 The Fission Authors.

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

package support

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/srcmesh/kubefaas/pkg/cli/cliwrapper/cli"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd/support/resources"
	flagkey "github.com/srcmesh/kubefaas/pkg/cli/flag/key"
	"github.com/srcmesh/kubefaas/pkg/cli/util"
	"github.com/srcmesh/kubefaas/pkg/utils"
)

const (
	DUMP_ARCHIVE_PREFIX = "kubefaas-dump"
	DEFAULT_OUTPUT_DIR  = "kubefaas-dump"
)

type DumpSubCommand struct {
	cmd.CommandActioner
}

func Dump(input cli.Input) error {
	return (&DumpSubCommand{}).do(input)
}

func (opts *DumpSubCommand) do(input cli.Input) error {
	fmt.Println("Start dumping process...")

	nozip := input.Bool(flagkey.SupportNoZip)
	outputDir := input.String(flagkey.SupportOutput)

	// check whether the dump directory exists.
	_, err := os.Stat(outputDir)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(outputDir, 0755)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(errors.Wrap(err, "Error checking dump directory status"))
	}

	outputDir, err = filepath.Abs(outputDir)
	if err != nil {
		panic(errors.Wrap(err, "Error creating dump directory for dumping files"))
	}

	_, k8sClient, err := util.GetKubernetesClient()
	if err != nil {
		return err
	}

	ress := map[string]resources.Resource{
		// kubernetes info
		"kubernetes-version": resources.NewKubernetesVersion(k8sClient),
		"kubernetes-nodes":   resources.NewKubernetesObjectDumper(k8sClient, resources.KubernetesNode, ""),

		// kubefaas info
		"version": resources.NewFissionVersion(opts.Client()),

		// kubefaas component logs & spec
		"components-svc-spec": resources.NewKubernetesObjectDumper(k8sClient, resources.KubernetesService,
			"svc in (buildermgr, controller, executor, influxdb, kubewatcher, logger, mqtrigger, nats-streaming, router, storagesvc, timer)"),
		"components-deployment-spec": resources.NewKubernetesObjectDumper(k8sClient, resources.KubernetesDeployment,
			"svc in (buildermgr, controller, executor, influxdb, kubewatcher, logger, mqtrigger, nats-streaming, router, storagesvc, timer)"),
		"components-daemonset-spec": resources.NewKubernetesObjectDumper(k8sClient, resources.KubernetesDaemonSet,
			"svc in (buildermgr, controller, executor, influxdb, kubewatcher, logger, mqtrigger, nats-streaming, router, storagesvc, timer)"),
		"components-pod-spec": resources.NewKubernetesObjectDumper(k8sClient, resources.KubernetesPod,
			"svc in (buildermgr, controller, executor, influxdb, kubewatcher, logger, mqtrigger, nats-streaming, router, storagesvc, timer)"),
		"components-pod-log": resources.NewKubernetesPodLogDumper(k8sClient,
			"svc in (buildermgr, controller, executor, influxdb, kubewatcher, logger, mqtrigger, nats-streaming, router, storagesvc, timer)"),

		// kubefaas builder logs & spec
		"builder-svc-spec":        resources.NewKubernetesObjectDumper(k8sClient, resources.KubernetesService, "owner=buildermgr"),
		"builder-deployment-spec": resources.NewKubernetesObjectDumper(k8sClient, resources.KubernetesDeployment, "owner=buildermgr"),
		"builder-pod-spec":        resources.NewKubernetesObjectDumper(k8sClient, resources.KubernetesPod, "owner=buildermgr"),
		"builder-pod-log":         resources.NewKubernetesPodLogDumper(k8sClient, "owner=buildermgr"),

		// kubefaas function logs & spec
		"function-svc-spec":        resources.NewKubernetesObjectDumper(k8sClient, resources.KubernetesService, "executorType=newdeploy"),
		"function-deployment-spec": resources.NewKubernetesObjectDumper(k8sClient, resources.KubernetesDeployment, "executorType in (poolmgr, newdeploy)"),
		"function-pod-spec":        resources.NewKubernetesObjectDumper(k8sClient, resources.KubernetesPod, "executorType in (poolmgr, newdeploy)"),
		"function-pod-log":         resources.NewKubernetesPodLogDumper(k8sClient, "executorType in (poolmgr, newdeploy)"),

		// CRD resources
		"crd-packages":     resources.NewCrdDumper(opts.Client(), resources.CrdPackage),
		"crd-environments": resources.NewCrdDumper(opts.Client(), resources.CrdEnvironment),
		"crd-functions":    resources.NewCrdDumper(opts.Client(), resources.CrdFunction),
		"crd-httptriggers": resources.NewCrdDumper(opts.Client(), resources.CrdHttpTrigger),
		"crd-kubewatchers": resources.NewCrdDumper(opts.Client(), resources.CrdKubeWatcher),
		"crd-mqtriggers":   resources.NewCrdDumper(opts.Client(), resources.CrdMessageQueueTrigger),
		"crd-timetriggers": resources.NewCrdDumper(opts.Client(), resources.CrdTimeTrigger),
	}

	dumpName := fmt.Sprintf("%v_%v", DUMP_ARCHIVE_PREFIX, time.Now().Unix())
	dumpDir := filepath.Join(outputDir, dumpName)

	wg := &sync.WaitGroup{}

	tempDir, err := utils.GetTempDir()
	if err != nil {
		fmt.Printf("Error creating temporary directory: %v\n", err.Error())
		return err
	}

	for key, res := range ress {
		dir := fmt.Sprintf("%v/%v/", tempDir, key)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				panic(err)
			}
		}
		wg.Add(1)
		go func(res resources.Resource, dir string) {
			defer wg.Done()
			res.Dump(dir)
		}(res, dir)
	}

	wg.Wait()

	if !nozip {
		defer os.RemoveAll(tempDir)
		path := filepath.Join(outputDir, fmt.Sprintf("%v.zip", dumpName))
		_, err := utils.MakeZipArchive(path, tempDir)
		if err != nil {
			fmt.Printf("Error creating archive for dump files: %v", err)
			return err
		}
		fmt.Printf("The archive dump file is %v\n", path)
	} else {
		err = os.Rename(tempDir, dumpDir)
		if err != nil {
			fmt.Printf("Error creating dump directory: %v\n", err.Error())
			return err
		}
		fmt.Printf("The dump files are placed at %v\n", dumpDir)
	}

	return nil
}
