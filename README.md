# Kogito Serveless Workflow using Knative Functions Example: E-Commerce App

This example aims to port [an example of AWS Step functions](https://github.com/aws-samples/aws-step-functions-long-lived-transactions) into the Knative/Serverless Workflow platform. 

## Prereqs

1. [Ko installed](https://github.com/google/ko)
2. [Java SDK installed](https://adoptopenjdk.net/)
3. [Maven installed](https://maven.apache.org/install.html)
4. [Quarkus CLI](https://quarkus.io/guides/cli-tooling)
5. [Knative quickstart plugin](https://knative.dev/docs/getting-started/)
6. [Kind](https://kind.sigs.k8s.io/docs/user/quick-start)

To edit your workflows:

1. Visual Studio Code with [Red Hat Java Plugin](https://marketplace.visualstudio.com/items?itemName=redhat.java) installed
2. [Serverless Workflow Editor](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-extension-serverless-workflow-editor)

## Creating a Knative cluster

Skip this step if you have already access to a Kubernetes with Knative Serving.

Create a Kind cluster and install Knative Serving by using this command:

```shell
kn quickstart kind --install-serving
```

ko supports loading images directly to Kind local registry: 

```shell
export KO_DOCKER_REPO=kind.local
export KIND_CLUSTER_NAME=knative # name of the kind cluster created by kn quickstart
```

### Kubernetes integration

In this example we use a YAML file to run the image for each function in the workflow. Below is the YAAML file for the first function order-new:

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: order-new
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/min-scale: "1"
    spec:
      containers:
      - image: ko://e-commerce-app/order-new
```

### Apply the e-commerce YAML configs

Use `ko apply` to apply the YAML config for all functions:

```shell
ko apply -f config
```

## Building and Deploying the Workflow 

### Creating the project

To create the project skeleton, run:

```shell
quarkus create app \
  -x=kogito-quarkus-serverless-workflow \
  -x=quarkus-container-image-jib \
  -x=quarkus-resteasy-jackson \
  -x=quarkus-smallrye-openapi \
  -x=kubernetes \
 org.acme:e-commerce-ksw
```

The `org.acme:e-commerce-ksw` is the group id, artifact id, and version of your project.

This command will create a Maven Quarkus project in the `e-commerce-ksw` directory with all required Kogito dependencies.

### Creating your first Workflow

Go to the directory `src/main/resources` and create a file named `e-commerce.sw.yaml`. 
You can play around and type the workflow definition by hand using the editor intellisense feature or copy and paste from the snnipet below:

```yaml
---
id: commerce
version: '1.0'
name: Hello Person
start: NewOrder
functions:
- name: orderNew
  type: custom
  operation: "rest:post:/"
- name: payment
  type: custom
  operation: "rest:post:/"
- name: inventoryReserve
  type: custom
  operation: "rest:post:/"
states:
- name: NewOrder
  type: operation
  actions:
  - functionRef:
      refName: payment
  transition: ProcessPayment
- name: ProcessPayment
  type: operation
  actions:
  - functionRef:
      refName: orderNew
  transition: InvReserve
- name: InvReserve
  type: operation
  actions:
  - functionRef:
      refName: inventoryReserve
  end: true
```

## Building your project's image and Deploying to Knative

You can use the Quarkus CLI to build your image with the following command:

```shell
quarkus build \
  -Dquarkus.container-image.build=true \
  -Dquarkus.kubernetes.deployment-target=knative \
  -Dquarkus.container-image.registry=kind.local \
  -Dquarkus.container-image.tag=latest
```

Load the produced container image into Kind:

```shell
kind load docker-image kind.local/ruizcm/e-commerce-ksw:latest --name=knative
```

Then deploy the workflow as a Knative application:

```shell
# Install the app!
kn service create -f target/kubernetes/knative.yml

# You should see something like "service.serving.knative.dev/e-commerce-ksw created" in the terminal
```

Check the service is ready:

```shell
kn service list

NAME           URL                                              LATEST               AGE   CONDITIONS   READY   REASON
e-commerce-ksw   http://e-commerce-ksw.default.127.0.0.1.sslip.io   e-commerce-ksw-00001   12s   3 OK / 3     True  
```

## Interacting with your application

To interact with the application, you can call the service via command line

```shell
curl -v -H 'Content-Type:application/json' \
  -H 'Accept:application/json'\
  -H "ce-specversion: 1.0" \ 
  -H "ce-type: dunno" \ 
  -H "ce-id: 1" \ 
  -H "ce-source: local" \
  -d '{"workflowdata" : {"name": "John"}}' \
  http://e-commerce-ksw.default.127.0.0.1.sslip.io/commerce
```

## Resources

- [Quarkus Container Images Guide](https://quarkus.io/guides/container-image)
- [Getting Started With Quarkus](https://quarkus.io/guides/getting-started)
- [CNCF Serverless Workflow](https://serverlessworkflow.io/)
- [Kogito Serverless Workflow](https://github.com/kiegroup/kogito-runtimes/tree/main/kogito-serverless-workflow)
