# Message Queue Trigger Demonstration - NATS Streaming 

## Create spec

```bash
$ kubefaas spec init
$ kubefaas env create --name go --image kubefaas/go-env-1.12 --builder kubefaas/go-builder-1.12 --period 5 --spec
$ kubefaas pkg create --name publisher --src mqtrigger/* --spec
$ kubefaas fn create --name publisher --env go --pkg publisher --entrypoint "Handler" --spec
$ kubefaas fn create --name hello --env go --src https://raw.githubusercontent.com/srcmesh/kubefaas/master/examples/go/hello.go --entrypoint "Handler" --spec
$ kubefaas mqt create --name mqtrigger --function hello --mqtype nats-streaming --topic foobar --spec
```

## Apply CRDs

```bash
$ kubefaas spec apply

# wait for package build status become succeeded
$ kubefaas pkg list
NAME                                       BUILD_STATUS ENV       LASTUPDATEDAT
hello-98476132-84ff-4e74-8b0f-2d1005871d1c succeeded    go        19 Dec 19 17:31 UTC
publisher                                  succeeded    go        19 Dec 19 17:19 UTC

# you can rebuild the package if it shows failed
$ kubefaas pkg rebuild --name <pkg-name>

$ kubefaas fn test --name publisher
Publish Success

$ kubectl -n kubefaas-function get pod -l functionName=hello
NAME                                         READY   STATUS        RESTARTS   AGE
poolmgr-go-default-610954-55664ccc68-b258c   2/2     Running       0          18m

# you should be able to see the function prints message
$ kubectl -n kubefaas-function logs -f -c go poolmgr-go-default-610954-55664ccc68-b258c
{"level":"info","ts":1576775701.7085218,"caller":"go/server.go:209","msg":"listening on 8888 ..."}
{"level":"info","ts":1576776720.3545933,"logger":"specialize_v2_handler","caller":"go/server.go:171","msg":"specializing ..."}
{"level":"info","ts":1576776720.3546736,"logger":"specialize_v2_handler","caller":"go/server.go:62","msg":"loading plugin","location":"/userfunc/15382797-f381-48af-9189-561f45f9285c/hello-98476132-84ff-4e74-8b0f-2d1005871d1c-7693uh-pwsz5u"}
{"level":"info","ts":1576776720.3640525,"logger":"specialize_v2_handler","caller":"go/server.go:180","msg":"done"}
2019/12/19 17:32:00 Hello, world!
```
