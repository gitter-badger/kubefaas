# Kubefaas

Kubefaas is a framework for serverless functions on Kubernetes.

## Prerequisites

- Kubernetes 1.9 or later

## Helm charts

The following table lists two helm charts for Kubefaas.

| Parameter      | Description                                                                            |
| ---------------| ---------------------------------------------------------------------------------------|
| `kubefaas-core` | FaaS essentials, and triggers for HTTP, Timers and Kubernetes Watches                 |
| `kubefaas-all`  | Log aggregation with fluent-bit and InfluxDB; NATS for message queue triggers;        |

## Installing the chart

To install the chart with the release name `my-release`,

```bash
$ helm install --name my-release kubefaas-all
```

## Uninstalling the chart

To uninstall/delete chart,

```bash
$ helm delete my-release
```

## Configuration

The following table lists the configurable parameters of the Kubefaas chart and their default values.

Parameter | Description | Default
--------- | ----------- | -------
`serviceType` | Type of Controller service to use. For minikube, set this to NodePort, elsewhere use LoadBalancer or ClusterIP. | `ClusterIP`
`routerServiceType` | Type of Router service to use. For minikube, set this to NodePort, elsewhere use LoadBalancer or ClusterIP. | `LoadBalancer`
`repository` | Image base repository | `index.docker.io`
`image` | Bundle image repository | `srcmesh/kubefaas-bundle`
`imageTag` | Bundle image tag | `1.8.0`
`pullPolicy` | Image pull policy | `IfNotPresent`
`fetcher.image` | Fetcher repository | `kubefaas/fetcher`
`fetcher.imageTag` | Fetcher image tag | `1.8.0`
`controllerPort` | Controller service port | `31313`
`routerPort` | Router service port | ` 31314`
`functionNamespace` | Namespace in which to run functions (this is different from the release namespace) | `kubefaas-function`
`builderNamespace` | Namespace in which to run builders (this is different from the release namespace) | `kubefaas-builder`
`enableIstio` | Enable istio integration | `false`
`persistence.enabled` | If true, persist data to a persistent volume | `true`
`persistence.existingClaim` | Provide an existing PersistentVolumeClaim instead of creating a new one | `nil`
`persistence.storageClass` | PersistentVolumeClaim storage class | `nil`
`persistence.accessMode` | PersistentVolumeClaim access mode | `ReadWriteOnce`
`persistence.size` | PersistentVolumeClaim size | `8Gi`
`analytics` | Analytics let us count how many people installed kubefaas. Set to false to disable analytics | `true`
`analyticsNonHelmInstall` | Internally used for generating an analytics job for non-helm installs | `false`
`pruneInterval` | The frequency of archive pruner (in minutes) | `60`
`preUpgradeChecksImage` | Pre-install/pre-upgrade checks live in this image | `kubefaas/pre-upgrade-checks`
`debugEnv` | If there are any pod specialization errors when a function is triggered and this flag is set to true, the error summary is returned as part of http response | `true`
`prometheus.enabled` | Set to true if prometheus needs to be deployed along with kubefaas | `true` in `kubefaas-all`, `false` in `kubefaas-core`
`prometheus.serviceEndpoint` | If prometheus.enabled is false, please assign the prometheus service URL that is accessible by components. | `nil`
`canaryDeployment.enabled` | Set to true if you need canary deployment feature | `true` in `kubefaas-all`, `false` in `kubefaas-core`
`extraCoreComponentPodConfig` | Extend the container specs for the core kubefaas pods. Can be used to add things like affinty/tolerations/nodeSelectors/etc. | None
`executor.adoptExistingResources` | If true, executor will try to adopt existing resources created by the old executor instance. | `false`
`router.deployAsDaemonSet` | Deploy router as DaemonSet instead of Deployment | `false`
`router.svcAddressMaxRetries` | Max retries times for router to retry on a certain service URL returns from cache/executor | `5`
`router.svcAddressUpdateTimeout` | The length of update lock expiry time for router to get a service URL returns from executor | `30`
`router.svcAnnotations` | Annotations for router service | None
`router.useEncodedPath` | For router to match encoded path. If true, "/foo%2Fbar" will match the path "/{var}"; Otherwise, it will match the path "/foo/bar". | `false`
`router.traceSamplingRate` | Uniformly sample traces with the given probabilistic sampling rate | `0.5`
`router.roundTrip.disableKeepAlive` | Disable transport keep-alive for fast switching function version | `true`
`router.roundTrip.keepAliveTime` | The keep-alive period for an active network connection to function pod | `30s`
`router.roundTrip.timeout` | HTTP transport request timeout | `50ms`
`router.roundTrip.timeoutExponent` | The length of request timeout will multiply with timeoutExponent after each retry | `2` 
`router.roundTrip.maxRetries` | Max retries times of a failed request | `10`

### Extra configuration for `kubefaas-all`

Parameter | Description | Default
--------- | ----------- | -------
`createNamespace` | If true, create `kubefaas-function` and `kubefaas-builder` namespaces | ` true`
`logger.influxdbAdmin` | Log database admin username | `admin`
`logger.fluentdImageRepository` | Logger fluentbit image repository | `index.docker.io`
`logger.fluentdImage` | Logger fluentbit image | `fluent/fluent-bit`
`logger.fluentdImageTag` | Logger fluentbit image tag | `1.0.4`
`nats.enabled` | Nats streaming enabled | `true`
`nats.authToken` | Nats streaming auth token | `defaultKubefaasAuthToken`
`nats.clusterID` | Nats streaming clusterID | `kubefaasMQTrigger`
`natsStreamingPort` | Nats streaming service port | `31316`
`azureStorageQueue.enabled` | Azure storage account name | `false`
`azureStorageQueue.key` | Azure storage account name | `""`
`azureStorageQueue.accountName` | Azure storage access key | `""`
`kafka.enabled` | Kafka trigger enabled | `false`
`kafka.brokers` | Kafka brokers uri | `broker.kafka:9092`
`kafka.version` | Kafka broker version | `nil`
`heapster` | Enable Heapster (only enable this in clusters where heapster does not exist already) | `false`

Please note that deploying of Azure Storage Queue or Kafka is not done by Kubefaas chart and you will have to explicitly deploy them.

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```bash
$ helm install --name my-release --set image=custom/bundle,imageTag=v1 kubefaas-all
```

If you're using minikube, set serviceType and routerServiceType to NodePort:

```bash
$ helm install --name my-release --set serviceType=NodePort,routerServiceType=NodePort kubefaas-all
```

You can also set parameters with a yaml file (see [values.yaml](kubefaas-all/values.yaml) for
what it should look like):

```bash
$ helm install --name my-release -f values.yaml kubefaas-all
```
