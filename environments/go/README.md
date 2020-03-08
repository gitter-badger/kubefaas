# Kubefaas: Go Environment

This is the Go environment for Kubefaas.

It's a Docker image containing a Go runtime, along with a dynamic loader.

Looking for ready-to-run examples? See the [Go examples directory](../../examples/go).

## Build this image

```
docker build -t USER/go-runtime . && docker push USER/go-runtime
```

Note that if you build the runtime, you must also build the go-builder
image, to ensure that it's at the same version of go:

```
cd builder && docker build -t USER/go-builder . && docker push USER/go-builder
```

## Using the image in kubefaas

You can add this customized image to kubefaas with "kubefaas env
create":

```
kubefaas env create --name go --image USER/go-runtime --builder USER/go-builder --version 2
```

Or, if you already have an environment, you can update its image:

```
kubefaas env update --name go --image USER/go-runtime --builder USER/go-builder
```

After this, kubefaas functions that have the env parameter set to the
same environment name as this command will use this environment.
