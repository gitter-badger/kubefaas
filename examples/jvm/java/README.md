# Hello World in JVM/Java on Kubefaas

The `io.fission.HelloWorld.java` class is a very simple kubefaas function that implements `io.fission.Function` and says "Hello, World!" .

## Building and deploying using Kubefaas

Kubefaas's builder can be used to create the binary artifact from source code. Create an environment with builder image and then create a package. 

```
$ zip -r java-src-pkg.zip *
$ kubefaas env create --name java --image kubefaas/jvm-env --version 2 --keeparchive --builder kubefaas/jvm-builder
$ kubefaas package create --sourcearchive java-src-pkg.zip --env java
java-src-pkg-zip-tvd0
$ kubefaas package info --name java-src-pkg-zip-tvd0
Name:        java-src-pkg-zip-tvd0
Environment: java
Status:      succeeded
Build Logs:
[INFO] Scanning for projects...
[INFO] 
[INFO] -----------------------< io.fission:hello-world >-----------------------
[INFO] Building hello-world 1.0-SNAPSHOT
[INFO] --------------------------------[ jar ]---------------------------------
```

Once package's status is `succeeded` then that package can be used to create and execute a function.

```
$ kubefaas fn create --name hello --pkg java-src-pkg-zip-tvd0 --env java --entrypoint io.fission.HelloWorld
$ kubefaas fn test --name hello
Hello World!
```

## Building locally and deploying with Kubefaas

You can build the jar file in one of the two ways below based on your setup:

- You can use docker without the need to install JDK and Maven to build the jar file from source code:

```bash
$ ./build.sh

```
- If you have JDK and Maven installed, you can directly build the JAR file using command:

```
$ mvn clean package
```

Both of above steps will generate a target subdirectory which has the archive `target/hello-world-1.0-SNAPSHOT-jar-with-dependencies.jar` which will be used for creating function.

- The archive created above will be used as a deploy package when creating the function.

```
$ kubefaas env create --name jvm --image kubefaas/jvm-env --version 2 --keeparchive=true
$ kubefaas fn create --name hello --deploy target/hello-world-1.0-SNAPSHOT-jar-with-dependencies.jar --env jvm --entrypoint io.fission.HelloWorld
$ kubefaas route create --function hello --url /hellop --method GET
$ kubefaas fn test --name hello
Hello World!
```
