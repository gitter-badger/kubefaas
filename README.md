# Kubefaas: Serverless Functions for Kubernetes

Kubefaas is a Function-as-a-Service framework for Kubernetes based on [Fission](https://github.com/fission/fission).

Kubefaas abstracts away the complexity of Kubernetes which allows 
you to focus on your daily routines. 

Kubefaas currently supports most of the popular programming languages 
like NodeJS, Python, Ruby, Go, PHP, and any Linux executable. 
Want a custom runtime environment? No problem, you can also create a custom runtime 
environment and use it for your function. 

# Kubernetes is the right place for Serverless

We're built on Kubernetes because we think any non-trivial app will
use a combination of serverless functions and more conventional
microservices, and Kubernetes is a great framework to bring these
together seamlessly.

Building on Kubernetes also means that anything you do for operations
on your Kubernetes cluster &mdash; such as monitoring or log
aggregation &mdash; also helps with ops on your Kubefaas deployment.

# Getting started and documentation

## Documentations

You can learn more about Kubefaas and get started from [documentation](https://kubefaas.com/docs).
* Concepts: https://kubefaas.com/docs/concepts/
* Installation: https://kubefaas.com/docs/installation/
* Troubleshooting: https://kubefaas.com/docs/trouble-shooting/

## Usage

```bash
  # Add the stock NodeJS env to your Kubefaas deployment
  $ kubefaas env create --name nodejs --image kubefaas/node-env

  # A javascript one-liner that prints "hello world"
  $ curl https://raw.githubusercontent.com/srcmesh/kubefaas/master/examples/nodejs/hello.js > hello.js

  # Upload your function code to kubefaas
  $ kubefaas function create --name hello --env nodejs --code hello.js

  # Map GET /hello to your new function
  $ kubefaas route create --method GET --url /hello --function hello

  # Run the function.  This takes about 100msec the first time.
  $ kubefaas function test --name hello
  Hello, world!
```

# Official Releases

Official releases of Kubefaas can be found on [the releases page](https://github.com/srcmesh/kubefaas/releases). 
Please note that it is strongly recommended that you use official releases of Kubefaas, as unreleased versions from 
the master branch are subject to changes and incompatibilities that will not be supported in the official releases. 
Builds from the master branch can have functionality changed and even removed at any time without compatibility support 
and without prior notice.

# Licensing

Kubefaas is an open core project maintained by [Srcmesh](https://srcmesh.com/) and released under the Apache 2.0 license.
