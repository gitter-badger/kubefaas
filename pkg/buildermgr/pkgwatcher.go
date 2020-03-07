/*
Copyright 2017 The Fission Authors.

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

package buildermgr

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	apiv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	k8sCache "k8s.io/client-go/tools/cache"

	fv1 "github.com/srcmesh/kubefaas/pkg/apis/core/v1"
	"github.com/srcmesh/kubefaas/pkg/cache"
	"github.com/srcmesh/kubefaas/pkg/crd"
	"github.com/srcmesh/kubefaas/pkg/utils"
)

type (
	packageWatcher struct {
		logger           *zap.Logger
		fissionClient    *crd.FissionClient
		k8sClient        *kubernetes.Clientset
		podStore         k8sCache.Store
		pkgStore         k8sCache.Store
		builderNamespace string
		storageSvcUrl    string
	}
)

func makePackageWatcher(logger *zap.Logger, fissionClient *crd.FissionClient, k8sClientSet *kubernetes.Clientset,
	builderNamespace string, storageSvcUrl string) *packageWatcher {
	lw := k8sCache.NewListWatchFromClient(k8sClientSet.CoreV1().RESTClient(), "pods", metav1.NamespaceAll, fields.Everything())
	store, controller := k8sCache.NewInformer(lw, &apiv1.Pod{}, 30*time.Second, k8sCache.ResourceEventHandlerFuncs{})
	go controller.Run(make(chan struct{}))

	pkgw := &packageWatcher{
		logger:           logger.Named("package_watcher"),
		fissionClient:    fissionClient,
		k8sClient:        k8sClientSet,
		podStore:         store,
		builderNamespace: builderNamespace,
		storageSvcUrl:    storageSvcUrl,
	}
	return pkgw
}

// build helps to update package status, checks environment builder pod status and
// dispatches buildPackage to build source package into deployment package.
// Following is the steps build function takes to complete the whole process.
// 1. Check package status
// 2. Update package status to running state
// 3. Check environment builder pod status
// 4. Call buildPackage to build package
// 5. Update package resource in package ref of functions that share the same package
// 6. Update package status to succeed state
// *. Update package status to failed state,if any one of steps above failed/time out
func (pkgw *packageWatcher) build(buildCache *cache.Cache, srcpkg *fv1.Package) {
	// Ignore duplicate build requests
	key := fmt.Sprintf("%v-%v", srcpkg.ObjectMeta.Name, srcpkg.ObjectMeta.ResourceVersion)
	_, err := buildCache.Set(key, srcpkg)
	if err != nil {
		return
	}
	defer buildCache.Delete(key)

	pkgw.logger.Info("start processing package", zap.String("package_name", srcpkg.ObjectMeta.Name), zap.String("resource_version", srcpkg.ObjectMeta.ResourceVersion))
	defer pkgw.logger.Info("end processing package", zap.String("package_name", srcpkg.ObjectMeta.Name), zap.String("resource_version", srcpkg.ObjectMeta.ResourceVersion))

	if srcpkg.Spec.Source.IsEmpty() {
		var buildStatus fv1.BuildStatus
		var buildLogs string

		// if source archive is empty while the deploy archive isn't, mark the pkg as status "None".
		// Mark as "Failed" if both archives are empty.

		if srcpkg.Spec.Deployment.IsEmpty() {
			buildStatus = fv1.BuildStatusFailed
			buildLogs = "Both source & deploy archives are empty"
		} else {
			buildStatus = fv1.BuildStatusNone
		}

		_, err := updatePackageStatus(pkgw.logger, pkgw.fissionClient, srcpkg, buildStatus, buildLogs)
		if err != nil {
			pkgw.logger.Error("error updating package state",
				zap.Any("status", buildStatus),
				zap.Any("buildlogs", buildLogs),
				zap.Error(err),
			)
		}
		return
	}

	pkg, err := updatePackageStatus(pkgw.logger, pkgw.fissionClient, srcpkg, fv1.BuildStatusRunning, "")
	if err != nil {
		pkgw.logger.Error("error setting package pending state", zap.Error(err))
		return
	}

	env, err := pkgw.fissionClient.CoreV1().Environments(pkg.Spec.Environment.Namespace).Get(pkg.Spec.Environment.Name, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		e := "environment does not exist"
		pkgw.logger.Error(e, zap.String("environment", pkg.Spec.Environment.Name))
		updatePackageStatus(pkgw.logger, pkgw.fissionClient, pkg,
			fv1.BuildStatusFailed, fmt.Sprintf("%s: %q", e, pkg.Spec.Environment.Name))
		return
	}

	// Do health check for environment builder pod
	for i := 0; i < 15; i++ {
		// Informer store is not able to use label to find the pod,
		// iterate all available environment builders.
		items := pkgw.podStore.List()
		if err != nil {
			pkgw.logger.Error("error retrieving pod information for environment", zap.Error(err), zap.String("environment", env.ObjectMeta.Name))
			return
		}

		if len(items) == 0 {
			pkgw.logger.Info("builder pod does not exist for environment, will retry again later", zap.String("environment", pkg.Spec.Environment.Name))
			time.Sleep(time.Duration(i*1) * time.Second)
			continue
		}

		for _, item := range items {
			pod := item.(*apiv1.Pod)

			// In order to support backward compatibility, for all builder images created in default env,
			// the pods will be created in fission-builder namespace
			builderNs := pkgw.builderNamespace
			if env.ObjectMeta.Namespace != metav1.NamespaceDefault {
				builderNs = env.ObjectMeta.Namespace
			}

			// Filter non-matching pods
			if pod.ObjectMeta.Labels[LABEL_ENV_NAME] != env.ObjectMeta.Name ||
				pod.ObjectMeta.Labels[LABEL_ENV_NAMESPACE] != builderNs ||
				pod.ObjectMeta.Labels[LABEL_ENV_RESOURCEVERSION] != env.ObjectMeta.ResourceVersion {
				continue
			}

			// Pod may become "Running" state but still failed at health check, so use
			// pod.Status.ContainerStatuses instead of pod.Status.Phase to check pod readiness states.
			podIsReady := true

			for _, cStatus := range pod.Status.ContainerStatuses {
				podIsReady = podIsReady && cStatus.Ready
			}

			if !podIsReady {
				pkgw.logger.Info("builder pod is not ready for environment, will retry again later", zap.String("environment", pkg.Spec.Environment.Name))
				time.Sleep(time.Duration(i*1) * time.Second)
				break
			}

			// Add the package getter rolebinding to builder sa
			// we continue here if role binding was not setup succeesffully. this is because without this, the fetcher wont be able to fetch the source pkg into the container and
			// the build will fail eventually
			err := utils.SetupRoleBinding(pkgw.logger, pkgw.k8sClient, fv1.PackageGetterRB, pkg.ObjectMeta.Namespace, fv1.PackageGetterCR, fv1.ClusterRole, fv1.FissionBuilderSA, builderNs)
			if err != nil {
				pkgw.logger.Error("error setting up role binding for package",
					zap.Error(err),
					zap.String("role_binding", fv1.PackageGetterRB),
					zap.String("package_name", pkg.ObjectMeta.Name),
					zap.String("package_namespace", pkg.ObjectMeta.Namespace))
				continue
			} else {
				pkgw.logger.Info("setup rolebinding for sa package",
					zap.String("sa", fmt.Sprintf("%s.%s", fv1.FissionBuilderSA, builderNs)),
					zap.String("package", fmt.Sprintf("%s.%s", pkg.ObjectMeta.Name, pkg.ObjectMeta.Namespace)))
			}

			ctx := context.Background()
			uploadResp, buildLogs, err := buildPackage(ctx, pkgw.logger, pkgw.fissionClient, builderNs, pkgw.storageSvcUrl, pkg)
			if err != nil {
				pkgw.logger.Error("error building package", zap.Error(err), zap.String("package_name", pkg.ObjectMeta.Name))
				updatePackageStatus(pkgw.logger, pkgw.fissionClient, pkg, fv1.BuildStatusFailed, buildLogs)
				return
			}

			pkgw.logger.Info("starting package info update", zap.String("package_name", pkg.ObjectMeta.Name))

			fnList, err := pkgw.fissionClient.CoreV1().
				Functions(metav1.NamespaceAll).List(metav1.ListOptions{})
			if err != nil {
				e := "error getting function list"
				pkgw.logger.Error(e, zap.Error(err))
				buildLogs += fmt.Sprintf("%s: %v\n", e, err)
				updatePackageStatus(pkgw.logger, pkgw.fissionClient, pkg, fv1.BuildStatusFailed, buildLogs)
			}

			// A package may be used by multiple functions. Update
			// functions with old package resource version
			for _, fn := range fnList.Items {
				if fn.Spec.Package.PackageRef.Name == pkg.ObjectMeta.Name &&
					fn.Spec.Package.PackageRef.Namespace == pkg.ObjectMeta.Namespace &&
					fn.Spec.Package.PackageRef.ResourceVersion != pkg.ObjectMeta.ResourceVersion {
					fn.Spec.Package.PackageRef.ResourceVersion = pkg.ObjectMeta.ResourceVersion
					// update CRD
					_, err = pkgw.fissionClient.CoreV1().Functions(fn.ObjectMeta.Namespace).Update(&fn)
					if err != nil {
						e := "error updating function package resource version"
						pkgw.logger.Error(e, zap.Error(err))
						buildLogs += fmt.Sprintf("%s: %v\n", e, err)
						updatePackageStatus(pkgw.logger, pkgw.fissionClient, pkg, fv1.BuildStatusFailed, buildLogs)
						return
					}
				}
			}

			_, err = updatePackage(pkgw.logger, pkgw.fissionClient, pkg,
				fv1.BuildStatusSucceeded, buildLogs, uploadResp)
			if err != nil {
				pkgw.logger.Error("error updating package info", zap.Error(err), zap.String("package_name", pkg.ObjectMeta.Name))
				updatePackageStatus(pkgw.logger, pkgw.fissionClient, pkg, fv1.BuildStatusFailed, buildLogs)
				return
			}

			pkgw.logger.Info("completed package build request", zap.String("package_name", pkg.ObjectMeta.Name))
			return
		}
	}
	// build timeout
	updatePackageStatus(pkgw.logger, pkgw.fissionClient, pkg,
		fv1.BuildStatusFailed, "Build timeout due to environment builder not ready")

	pkgw.logger.Error("max retries exceeded in building source package, timeout due to environment builder not ready",
		zap.String("package", fmt.Sprintf("%s.%s", pkg.ObjectMeta.Name, pkg.ObjectMeta.Namespace)))
}

func (pkgw *packageWatcher) watchPackages() {
	buildCache := cache.MakeCache(0, 0)
	lw := k8sCache.NewListWatchFromClient(pkgw.fissionClient.CoreV1().RESTClient(), "packages", apiv1.NamespaceAll, fields.Everything())

	pkgStore, controller := k8sCache.NewInformer(lw, &fv1.Package{}, 60*time.Minute, k8sCache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pkg := obj.(*fv1.Package)
			var err error

			if len(pkg.Status.BuildStatus) == 0 {
				pkg.Status.BuildStatus = fv1.BuildStatusPending
				pkg, err = pkgw.fissionClient.CoreV1().Packages(pkg.Namespace).UpdateStatus(pkg)
				if err != nil {
					pkgw.logger.Error("error filling package status", zap.Error(err))
					return
				}
				// once we update the package status, an update event
				// will arrive and handle by UpdateFunc later. So we
				// don't need to build the package at this moment.
				return
			}

			if pkg.Status.BuildStatus == fv1.BuildStatusPending {
				go pkgw.build(buildCache, pkg)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPkg := oldObj.(*fv1.Package)
			pkg := newObj.(*fv1.Package)
			var err error

			if oldPkg.Generation == pkg.Generation {
				// If both old and new package's spec and status are the same
				// means this is an status update event and so we only need to
				// handle package in "pending" state.
				if pkg.Status.BuildStatus == fv1.BuildStatusPending {
					go pkgw.build(buildCache, pkg)
				}
			} else if oldPkg.Generation != pkg.Generation {
				// The buildermgr updates the deploy archive of spec when the package is "running" state.
				// Hence, we ignore this event.
				// TODO: handle the case that user updates the package
				//  when the build is still in progress.
				if pkg.Status.BuildStatus == fv1.BuildStatusRunning {
					return
				}

				pkg.Status.BuildStatus = fv1.BuildStatusPending
				pkg, err = pkgw.fissionClient.CoreV1().Packages(pkg.Namespace).UpdateStatus(pkg)
				if err != nil {
					pkgw.logger.Error("error filling package status", zap.Error(err))
				}
			}
		},
	})

	pkgw.pkgStore = pkgStore
	controller.Run(make(chan struct{}))
}
