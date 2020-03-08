# Kubefaas: Ruby Environment

This is the Ruby environment for Kubefaas.

It's a Docker image containing a Ruby 2.6.1 runtime. The image uses
Rack with WEBrick to host the internal web server.

The environment works via convention where you create a Ruby method
called `handler` with a single optional argument, a `Kubefaas::Context`
object.

The `Kubefaas::Context` object gives access to the Rack env, a
request object, and a logger. Please see `kubefaas/context.rb` for the
public api.

The `Kubefaas::Request` object is a subclass of `Rack::Request` and
provides access to parameters and headers. See `kubefaas/request.rb`
for the public api.

Looking for ready-to-run examples? See the [Ruby examples directory](../../examples/ruby).

## Customizing this image

To add package dependencies, edit Gemfile to add what you
need, and rebuild this image (instructions below).

## Rebuilding and pushing the image

You'll need access to a Docker registry to push the image: you can
sign up for Docker hub at hub.docker.com, or use registries from
gcr.io, quay.io, etc.  Let's assume you're using a docker hub account
called USER.  Build and push the image to the the registry:

```
   docker build -t USER/ruby-env . && docker push USER/ruby-env
```

## Using the image in kubefaas

You can add this customized image to kubefaas with "kubefaas env
create":

```
   kubefaas env create --name ruby --image USER/ruby-env
```

Or, if you already have an environment, you can update its image:

```
   kubefaas env update --name ruby --image USER/ruby-env
```

After this, kubefaas functions that have the env parameter set to the
same environment name as this command will use this environment.
