# kubernetes-operator-assignment

## <a id="description"></a> Description
This application revolves around the operator **CustomDeployment** which creates:
- a regular **Deployment** that creates pods running the **nginx** image
- a **Service**
- a **ClusterIssuer** using **cert-manager**

The kubernetes cluster used was created using **kind**.

## Prerequisites
- Install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/):
    ```sh
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    curl -LO "https://dl.k8s.io/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl.sha256"
    echo "$(cat kubectl.sha256)  kubectl" | sha256sum --check
    sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
    ```

- Install [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) for the local cluster:
    ```sh
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.12.0/kind-linux-amd64
    chmod +x ./kind
    mv ./kind usr/local/bin/kind
    ```
- Install [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder/releases/tag/v3.4.1) for initializing, creating and running your CRDs:
    ```text
    Download the latest release and put it in your path.
    ```

- Install [cert-manager](https://cert-manager.io/docs/installation/) for the certificate issuer:
    ```sh
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
    ```

- Install [nginx ingress controller](https://kubernetes.github.io/ingress-nginx/deploy/):
    ```sh
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.2.0/deploy/static/provider/cloud/deploy.yaml
    ```

- Install [MetalLB](https://metallb.universe.tf/installation/) for asigning an IP address to the service:
    - Preparation:
        - Edit **kube-proxy config**:
            ```sh
            kubectl edit configmap -n kube-system kube-proxy
            ```
        - Set the following fields:
            ```yaml
            apiVersion: kubeproxy.config.k8s.io/v1alpha1
            kind: KubeProxyConfiguration
            mode: "ipvs"
            ipvs:
            strictARP: true
            ```
    - Install:
        ```sh
        kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/namespace.yaml
        kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/metallb.yaml
        ```
    - Install the config map for MetalLB:
        ```sh
        make apply-metallb-configmap 
        ```
## How to use
- Create a `yaml` file that contains has the following structure:
    ```yaml
    apiVersion: crds.k8s.op.asgn/v1
    kind: CustomDeployment
    metadata:
    name: <custom-deployment-name>
    spec:
    host: <host-name>
    port: <port>
    replicas: <number-of-replicas>
    image:
        name: <image-name>
        tag: <image-tag>
    ```

    where:
    - `<custom-deployment-name>`: the name of the resource
    - `<host-name>`: the host where the application is accessible
    - `<port>`: port value (integer between `30000-32767`). Default value: `30000`
    - `<number-of-replicas>`: the number of pods running the image. Default value: `1`
    - `<image-name>`: the name of the image (ex: `nginx`)
    - `<image-tag>`: the tag/version of the image (ex: `latest`)

- Create the Custom Deployment, along with all the other resources mentioned in the [description](#description) section:
    ```sh
    kubectl apply -f <your-file>.yaml
    ```

### Running sample on the cluster
1. Install the CRDs into the cluster:

```sh
make install
```

2.  Run the **CustomDeployment** controller:

```sh
make run
```

3. Install Instances of Custom Resource:

```sh
make apply-sample
```

### Delete CRD resource
```sh
make delete-sample
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

