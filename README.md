# Kogito Serveless Workflow using Knative Functions Example: E-Commerce App

This example aims to port [an example of AWS Step functions](https://github.com/aws-samples/aws-step-functions-long-lived-transactions) into the Knative/Serverless Workflow platform. 

## Prereqs

1. [Ko installed](https://github.com/google/ko)
2. [Java SDK installed](https://adoptopenjdk.net/)
3. [Maven installed](https://maven.apache.org/install.html)
4. [Quarkus CLI](https://quarkus.io/guides/cli-tooling)
5. [Knative quickstart plugin](https://knative.dev/docs/getting-started/)
6. [Kind](https://kind.sigs.k8s.io/docs/user/quick-start)
7. PostgreSQL

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
```

Below is the name of the kind cluster created by kn quickstart

```shell
export KIND_CLUSTER_NAME=knative 
```

## Installing postgreSQL

The commerce application stores orders and inventories in postgreSQL.

Install kubegres:

```shell
kubectl apply -f https://raw.githubusercontent.com/reactive-tech/kubegres/v1.15/kubegres.yaml
```

```shell
kubectl wait deployment -n kubegres-system kubegres-controller-manager --for condition=Available=True --timeout=90s
```

```shell
kubectl apply -f config/infra/postgres
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
kn workflow create --name e-commerce-ksw
```

This command will create a Maven Quarkus project in the `e-commerce-ksw` directory with all required Kogito dependencies.


### Copy the e-commerce workflow

```shell
rm -rf e-commerce-ksw/src/main/resources/* && cp config/sw/* e-commerce-ksw/src/main/resources
```

## Building your project's image and Deploying to Knative

Navigate to your project's directory. For this example:

```shell
cd e-commerce-ksw
```

You can use the Serverless Workflow plug-in for the Knative CLI to build your image with the following command:

```shell
kn workflow build --image dev.local/e-commerce-ksw:1.0
```

Load the produced container image into Kind:

```shell
kind load docker-image dev.local/e-commerce-ksw:1.0 --name=knative
```

Then deploy the workflow as a Knative application:

```shell
kn service create -f target/kubernetes/knative.yml
```

You should see something like "service.serving.knative.dev/e-commerce-ksw created" in the terminal


Check the service is ready:

```shell
kn service list
```

Expected output:
```
NAME           URL                                              LATEST               AGE   CONDITIONS   READY   REASON
e-commerce-ksw   http://e-commerce-ksw.default.127.0.0.1.sslip.io   e-commerce-ksw-00001   12s   3 OK / 3     True  
```

## Interacting with your application

To interact with the application, you can call the service via command line

```shell
curl -v -X POST -H 'Content-Type:application/json' -H 'Accept:application/json' -d '{"workflowdata" : {"order_id":"8dee2","order_info":{"order_date":"2022-01-01T02:30:50Z","customer_id":"id001","order_status":"fillIn","items": [{"item_id":"itemID456","qty":1,"description":"Pencil","unit_price":2.5},{"item_id":"itemID789","qty":1,"description":"Paper","unit_price":4}],"payment":{"merchant_id":"merchantID1234","payment_amount":6.5,"transaction_id":"54c512","transaction_date":"2022-01-01T02:30:50Z","order_id":"8dee2","payment_type":"creditcard"},"inventory":{"transaction_id":"54c512","transaction_date":"2022-01-01T02:30:50Z","order_id":"8dee2","items":["Pencil","Paper"],"transaction_type":"online"}}}}'  http://e-commerce-ksw.default.127.0.0.1.sslip.io/commerce
```

## Resources

- [Quarkus Container Images Guide](https://quarkus.io/guides/container-image)
- [Getting Started With Quarkus](https://quarkus.io/guides/getting-started)
- [CNCF Serverless Workflow](https://serverlessworkflow.io/)
- [Kogito Serverless Workflow](https://github.com/kiegroup/kogito-runtimes/tree/main/kogito-serverless-workflow)
