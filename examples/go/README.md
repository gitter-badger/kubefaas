# Hello World in Go on Kubefaas

`hello.go` contains a very simple kubefaas function that says "Hello, World!".

## Deploying this function on your cluster

```bash

# Create the Kubefaas Go environment and function, and wait for the
# function to build.  (Take a look at the YAML files in the specs
# directory for details about how the environment and function are
# specified.)

$ kubefaas spec apply --wait
1 environment created
1 package created
1 function created

# Now, run the function with the "kubefaas function test" command:

$ kubefaas function test --name hello-go
Hello, World!
```
