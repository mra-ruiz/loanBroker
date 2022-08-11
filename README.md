# Kogito Serveless Workflow using Knative Functions Example: E-Commerce App

This example aims to port [an example of AWS Step functions](https://github.com/aws-samples/aws-step-functions-long-lived-transactions) into the Knative/Serverless Workflow platform. 

## Prereqs

1. [Ko installed](https://github.com/google/ko)
2. [Java SDK installed](https://adoptopenjdk.net/)
3. [Maven installed](https://maven.apache.org/install.html)
4. [Quarkus CLI](https://quarkus.io/guides/cli-tooling)
5. [Knative quickstart plugin](https://knative.dev/docs/getting-started/)

To edit your workflows:

1. Visual Studio Code with [Red Hat Java Plugin](https://marketplace.visualstudio.com/items?itemName=redhat.java) installed
2. [Serverless Workflow Editor](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-extension-serverless-workflow-editor)

## Building images for your functions with ko

ko depends on an environment variable, KO_DOCKER_REPO, to identify where it should push images that it builds. Typically this will be a remote registry, e.g.:

```shell
export KO_DOCKER_REPO=docker.io/<your-docker-id>
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

### Apply the resolved YAML config

Use `ko apply` to apply the resolved YAML config for each function.

```shell
ko apply -f config/<resolved-yaml-config>.yaml
```

## Creating the project

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

## Creating your first Workflow

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
  -Dquarkus.container-image.group=<your-group> \
    -Dquarkus.container-image.tag=latest
```

Assuming you have [installed Knative locally](https://knative.dev/docs/getting-started/), on your kind cluster, run:

```shell
# Install the app!
kubectl apply -f target/kubernetes/knative.yml

# You should see something like "service.serving.knative.dev/e-commerce-ksw created" in the terminal
```

Wait a couple of seconds and run `kn service list`:

```shell
kn service list

NAME           URL                                              LATEST               AGE   CONDITIONS   READY   REASON
e-commerce-ksw   http://e-commerce-ksw.default.127.0.0.1.sslip.io   e-commerce-ksw-00001   12s   3 OK / 3     True  
```

## Interacting with your application

To interact with the application, you can call the service via command line

```shell
curl -v -X POST -H 'Content-Type:application/json' -H 'Accept:application/json' -H "ce-specversion: 1.0" -H "ce-type: dunno" -H "ce-id: 1" -H "ce-source: local" -d '{"workflowdata" : {"name": "John"}}'  http://e-commerce-app-ksw.default.127.0.0.1.sslip.io/commerce
```

## Resources

- [Quarkus Container Images Guide](https://quarkus.io/guides/container-image)
- [Getting Started With Quarkus](https://quarkus.io/guides/getting-started)
- [CNCF Serverless Workflow](https://serverlessworkflow.io/)
- [Kogito Serverless Workflow](https://github.com/kiegroup/kogito-runtimes/tree/main/kogito-serverless-workflow)