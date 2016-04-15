# Deis Router v2

[![Build Status](https://travis-ci.org/deis/router.svg?branch=master)](https://travis-ci.org/deis/router)
[![Go Report Card](http://goreportcard.com/badge/deis/router)](http://goreportcard.com/report/deis/router)
[![Docker Repository on Quay](https://quay.io/repository/deis/router/status "Docker Repository on Quay")](https://quay.io/repository/deis/router)

Deis (pronounced DAY-iss) Workflow is an open source Platform as a Service (PaaS) that adds a developer-friendly layer to any [Kubernetes](http://kubernetes.io) cluster, making it easy to deploy and manage applications on your own servers.

## Beta Status

This Deis component is currently in beta status, and we welcome your input! If you have feedback, please submit an [issue][issues]. If you'd like to participate in development, please read the "Development" section below and submit a [pull request][prs].

# About

The Deis router handles ingress and routing of HTTP/S traffic bound for the Deis API and for your own applications. This component is 100% Kubernetes native and, while it's intended for use inside of the Deis [PaaS](https://en.wikipedia.org/wiki/Platform_as_a_service), it's flexible enough to be used standalone inside any Kubernetes cluster.

# Development

The Deis project welcomes contributions from all developers. The high level process for development matches many other open source projects. See below for an outline.

* Fork this repository
* Make your changes
* Submit a [pull request][prs] (PR) to this repository with your changes, and unit tests whenever possible.
	* If your PR fixes any [issues][issues], make sure you write Fixes #1234 in your PR description (where #1234 is the number of the issue you're closing)
* The Deis core contributors will review your code. After each of them sign off on your code, they'll label your PR with `LGTM1` and `LGTM2` (respectively). Once that happens, they'll merge it.

## Installation

This section documents simple procedures for installing the Deis Router for evaluation or use.  Those wishing to contribute to Deis Router development might consider the more developer-oriented instructions in the [Hacking Router](#hacking) section.

Deis Router can be installed with or without the rest of the Deis platform.  In either case, begin with a healthy Kubernetes cluster.  Kubernetes getting started documentation is available [here](http://kubernetes.io/gettingstarted/).

Next, install the [helm](http://helm.sh) package manager, then use the commands below to initialize that tool and load the [deis/charts](https://github.com/deis/charts) repository.

```
$ helm update
$ helm repo add deis https://github.com/deis/charts
```

To install the router:

```
$ helm fetch deis/<chart>
$ helm generate <chart>
$ helm install <chart>
```
Where `<chart>` is selected from the options below:

| Chart | Description |
|-------|-------------|
| deis | Install the latest router release along with the rest of the latest Deis platform release. |
| deis-dev | Install the edge router (from master) with the rest of the edge Deis platform. |
| router | Install the latest router release with its minimal set of dependencies. |
| router-dev | Install the edge router (from master) with its minimal set of dependencies. |


For next steps, skip ahead to the [How it Works](#how-it-works) and [Configuration Guide](#configuration) sections.

## <a name="hacking"></a>Hacking Router

The only dependencies for hacking on / contributing to this component are:

* `git`
* `make`
* `docker`
* `kubectl`, properly configured to manipulate a healthy Kubernetes cluster that you presumably use for development
* Your favorite text editor

Although the router is written in Go, you do _not_ need Go or any other development tools installed.  Any parts of the developer workflow requiring tools not listed above are delegated to a containerized Go development environment.

### Registry

If your Kubernetes cluster is running locally on one or more virtual machines, it's advisable to also run your own local Docker registry.  This provides a place where router images built from source-- possibly containing your own experimental hacks-- can be pushed relatively quickly and can be accessed readily by your Kubernetes cluster.

Fortunately, this is very easy to set up as long as you have Docker already functioning properly:

```
$ make dev-registry
```

This will produce output containing further instructions such as:

```
59ba57a3628fe04016634760e039a3202036d5db984f6de96ea8876a7ba8a945

To use a local registry for Deis development:
    export DEIS_REGISTRY=192.168.99.102:5000/
```

Following those instructions will make your local registry usable by the various `make` targets mentioned in following sections.

If you do not want to run your own local registry or if the Kubernetes cluster you will be deploying to is remote, then you can easily make use of a public registry such as [hub.docker.com](http://hub.docker.com), provided you have an account.  To do so:

```
$ export DEIS_REGISTRY=registry.hub.docker.com/
$ export IMAGE_PREFIX=your-username
```

### If I can `make` it there, I'll `make` it anywhere...

The entire developer workflow for anyone hacking on the router is implemented as a set of `make` targets.  They are simple and easy to use, and collectively provide a workflow that should feel familiar to anyone who has hacked on Deis v1.x in the past.

#### Setup:

To "bootstrap" the development environment:

```
$ make bootstrap
```

In router's case, this step carries out some extensive dependency management using glide within the containerized development environment.  Because the router leverages the Kubernetes API, which in turn has in excess of one hundred dependencies, this step can take quite some time.  __Be patient, and allow up to 20 minutes.  You generally only ever have to do this once.__


#### To build:

```
$ make build
```

Make sure to have defined the variable `DEIS_REGISTRY` previous to this step, as your image tags will be prefixed according to this.

Built images will be tagged with the sha of the latest git commit.  __This means that for a new image to have its own unique tag, experimental changes should be committed _before_ building.  Do this in a branch.  Commits can be squashed later when you are done hacking.__

#### To deploy:

```
$ make deploy
```

The deploy target will implicitly build first, then push the built image (which has its own unique tags) to your development registry (i.e. that specified by `DEIS_REGISTRY`).  A Kubernetes manifest is prepared, referencing the uniquely tagged image, and that manifest is submitted to your Kubernetes cluster.  If a router component is already running in your Kubernetes cluster, it will be deleted and replaced with your build.

To see that the router is running, you can look for its pod(s):

```
$ kubectl get pods --namespace=deis
```

## Trying it Out

To deploy some sample routable applications:

```
$ make examples
```

This will deploy Nginx and Apache to your Kubernetes cluster as if they were user applications.

To test, first modify your `/etc/hosts` such that the following four hostnames are resolvable to the IP of the Kubernetes node that is hosting the router:

* nginx.example.com
* apache.example.com
* httpd.example.com
* unknown.example.com

By requesting the following three URLs from your browser, you should find that one is routed to a pod running Nginx, while the other two are routed to a pod running Apache:

* http://nginx.example.com
* http://apache.example.com
* http://httpd.example.com

Requesting http://unknown.example.com should result in a 404 from the router since no route exists for that domain name.

## <a name="how-it-works"></a>How it Works

The router is implemented as a simple Go program that manages Caddyserver and Caddy configuration.  It regularly queries the Kubernetes API for services labeled with `router.deis.io/routable: "true"`.  Such services are compared to known services resident in memory.  If there are differences, new Caddy configuration is generated and Caddy is reloaded.

When generating configuration, the program reads all annotations of each service prefixed with `router.deis.io`.  These annotations describe all the configuration options that allow the program to dynamically construct Caddy configuration, including virtual hosts for all the domain names associated with each routable application.

Similarly, the router watches the annotations on its _own_ replication controller to dynamically construct global Caddy configuration.

## <a name="configuration"></a>Configuration Guide

### Environment variables

Router configuration is driven almost entirely by annotations on the router's replication controller and the services of all routable applications-- those labeled with `router.deis.io/routable: "true"`.

One exception to this, however, is that in order for the router to discover its own annotations, the router must be configured via environment variable with some awareness of its own namespace.  (It cannot query the API for information about itself without knowing this.)

The `POD_NAMESPACE` environment variable is required by the router and it should be configured to match the Kubernetes namespace that the router is deployed into.  If no value is provided, the router will assume a value of `default`.

For example, consider the following Kubernetes manifest.  Given a manifest containing the following metadata:

```
apiVersion: v1
kind: ReplicationController
metadata:
  name: deis-router
  namespace: deis
# ...
```

The corresponding template must inject a `POD_NAMESPACE=deis` environment variable into router containers.  The most elegant way to achieve this is by means of the Kubernetes "downward API," as in this snippet from the same manifest:

```
# ...
spec:
  # ...
  template:
    # ...
    spec:
      containers:
      - name: deis-router
        # ...
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
# ...
```

Altering the value of the `POD_NAMESPACE` environment variable requires the router to be restarted for changes to take effect.

### Annotations

All remaining options are configured through annotations.  Any of the following three Kubernetes resources can be configured:

| Resource | Notes |
|----------|-------|
| <ul><li>deis-router replication controller</li><li>deis-builder service (if in use)</li></ul> | All of these configuration options are specific to _this_ implementation of the router (as indicated by the inclusion of the token `nginx` in the annotations' names).  Customized and alternative router implementations are possible.  Such routers are under no obligation to honor these annotations, as many or all of these may not be applicable in such scenarios.  Customized and alternative implementations _should_ document their own configuration options. |
| <ul><li>routable application services</li></ul> | These are services labeled with `router.deis.io/routable: "true"`.  In the context of the broader Deis PaaS, these annotations are _written_ by the Deis workflow component (the API).  These annotations, therefore, represent the contract or _interface_ between that component and the router.  As such, any customized or alternative router implementations that wishes to remain compatible with deis-workflow must honor (or ignore) these annotations, but may _not_ alter their names or redefine their meanings. |

The table below details the configuration options that are available for each of the above.

_Note that Kubernetes annotation maps are all of Go type `map[string]string`.  As such, all configuration values must also be strings.  To avoid Kubernetes attempting to populate the `map[string]string` with non-string values, all numeric and boolean configuration values should be enclosed in double quotes to help avoid confusion._


| Component | Resource Type | Annotation | Default Value | Description |
|-----------|---------------|------------|---------------|-------------|
| <a name="platform-domain"></a>deis-router | RC | router.deis.io/caddy.platformDomain | N/A | This defines the router's platform domain.  Any domains added to a routable application _not_ containing the `.` character will be assumed to be subdomains of this platform domain.  Thus, for example, a platform domain of `example.com` coupled with a routable app counting `foo` among its domains will result in router configuration that routes traffic for `foo.example.com` to that application. |
| routable application | service | router.deis.io/domains | N/A | Comma-delimited list of domains for which traffic should be routed to the application.  These may be fully qualified (e.g. `foo.example.com`) or, if not containing any `.` character, will be considered subdomains of the router's domain, if that is defined. |

#### Annotations by example

##### router replication controller:

```
apiVersion: v1
kind: ReplicationController
metadata:
  name: deis-router
  namespace: deis
  # ...
  annotations:
    router.deis.io/caddy.platformDomain: example.com
# ...
```

##### builder service:

```
apiVersion: v1
kind: Service
metadata:
  name: deis-builder
  namespace: deis
  # ...
# ...
```

##### routable service:

```
apiVersion: v1
kind: Service
metadata:
  name: foo
  labels:
  	router.deis.io/routable: "true"
  namespace: examples
  # ...
  annotations:
    router.deis.io/domains: foo,bar,www.foobar.com
# ...
```

### <a name="ssl"></a>SSL

SSL is on by default. Certs will be registered through LetsEncrypt due to the direct integration built in to Caddy.

### Front-facing load balancer

Depending on what distribution of Kubernetes you use and where you host it, installation of the router _may_ automatically include an external (to Kubernetes) load balancer or similar mechanism for routing inbound traffic from beyond the cluster into the cluster to the router(s).  For example, [kube-aws](https://coreos.com/kubernetes/docs/latest/kubernetes-on-aws.html) and [Google Container Engine](https://cloud.google.com/container-engine/) both do this.  On some other platforms-- Vagrant or bare metal, for instance-- this must either be accomplished manually or does not apply at all.

#### Idle connection timeouts

If a load balancer such as the one described above does exist (whether created automatically or manually) _and_ if you intend on handling any long-running requests, the load balancer (or similar) _may_ require some manual configuration to increase the idle connection timeout.  Typically, this is most applicable to AWS and Elastic Load Balancers, but may apply in other cases as well.  It does _not_ apply to Google Container Engine, as the idle connection timeout cannot be configured there, but also works fine as-is.

If, for instance, router were installed on kube-aws, in conjunction with the rest of the Deis platform, this timeout should be increased to a recommended value of 1200 seconds.  This will ensure the load balancer does not hang up on the client during long-running operations like an application deployment.  Directions for this can be found [here](http://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/config-idle-timeout.html).

#### Manually configuring a load balancer

If using a Kubernetes distribution or underlying infrastructure that does not support the automated provisioning of a front-facing load balancer, operators will wish to manually configure a load balancer (or use other tricks) to route inbound traffic from beyond the cluster _into_ the cluster to the platform's own router(s).  There are many ways to accomplish this.  The remainder of this section discusses three general options for accomplishing this.

##### Option 1

This manually replicates the configuration that would be achieved automatically with some distributions on some infrastructure providers, as discussed above.

First, determine the "node ports" for the `deis-router` service:

```
$ kubectl describe service deis-router --namespace=deis
```

This will yield output similar to the following:

```
...
Port:			http	80/TCP
NodePort:		http	32477/TCP
Endpoints:		10.2.80.11:80
Port:			https	443/TCP
NodePort:		https	32389/TCP
Endpoints:		10.2.80.11:443
Port:			builder	2222/TCP
NodePort:		builder	30729/TCP
Endpoints:		10.2.80.11:2222
Port:			healthz	9090/TCP
NodePort:		healthz	31061/TCP
Endpoints:		10.2.80.11:9090
...
```

The node ports shown above are high-numbered ports that are allocated on _every_ Kubernetes worker node for use by the router service.  The kube-proxy component on _every_ Kubernetes node will listen on these ports and proxy traffic through to the corresponding port within an "endpoint--" that is, a pod running the Deis router.

If manually creating a load balancer, configure the load balancer to have _all_ Kubernetes worker nodes in the back-end pool, and listen on ports 80, 443, and 2222 (port 9090 can be ignored).  Each of these listeners should proxy inbound traffic to the corresponding node ports on the worker nodes.  Ports 80 and 443 may use either HTTP/S or TCP as protocols.  Port 2222 must use TCP.

With this configuration, the path a request takes from the end-user to an application pod is as follows:

```
user agent (browser) --> front-facing load balancer --> kube-proxy on _any_ Kubernetes worker node --> _any_ Deis router pod --> kube-proxy on that same node --> _any_ application pod
```

##### Option 2

Option 2 differs only slightly from option 1, but is more efficient.  As such, even operators who had a front-facing load balancer automatically provisioned on their infrastructure by Kubernetes might consider manually reconfiguring that load balancer as follows.

Deis router pods will listen on _host_ ports 80, 443, 2222, and 9090 wherever they run.  (They will not run on any worker nodes where all of these four ports are not available.)  Taking advantage of this, an operator may completely dismiss the node ports discussed in option 1.  The load balancer can be configured to have _all_ Kubernetes worker nodes in the back-end pool, and listen on ports 80, 443, and 2222.  Each of these listeners should proxy inbound traffic to the _same_ ports on the worker nodes.  Ports 80 and 443 may use either HTTP/S or TCP as protocols.  Port 2222 must use TCP.

Additionally, a health check _must_ be configured using the HTTP protocol, port 9090, and the `/healthz` endpoint.  With such a health check in place, _only_ nodes that are actually hosting a router pods will pass and be included in the load balancer's pool of active back end instances.

With this configuration, the path a request takes from the end-user to an application pod is as follows:

```
user agent (browser) --> front-facing load balancer --> a Deis router pod --> kube-proxy on that same node --> _any_ application pod
```

##### Option 3

Option 3 is similar to option 2, but does not actually utilize a load balancer at all.  Instead, a DNS A record may be created that lists the public IP addresses of _all_ Kubernetes worker nodes.  This will leverage DNS round-robining to direct requests to all nodes.  To guarantee _all_ nodes can adequately route incoming traffic, the Deis router component should be scaled out by increasing the number of replicas specified in the replication controller to match the number of worker nodes.  Anti-affinity should ensure exactly one router pod runs per worker node.

__This configuration is not suitable for production.__ The primary use case for this configuration is demonstrating or evaluating Deis on bare metal Kubernetes clusters without incurring the effort to configure an _actual_ front-facing load balancer.

## Production Considerations

### Customizing the charts

The Helm charts available for installing router (either with or without the rest of Deis) are intended to get users up and running as quickly as possible.  As such, the charts do not strictly require any editing prior to installation in order to successfully bootstrap a cluster.  However, there are some useful customizations that should be applied for use in production environments:

* __Specify a [platform domain](#platform-domain).__  Without a platform domain specified, any routable service specifying one or more non-fully-qualified domain names (not containing any `.` character) among its `router.deis.io/domains` will be matched using a regular expression of the form `^{{ $domain }}\.(?<domain>.+)$` where `{{ $domain }}` resolves to the non-fully-qualified domain name.  By way of example, the idiosyncrasy that this exposes is that traffic bound for the `foo` subdomain of _any_ domain would be routed to an application that lists the non-fully-qualified domain name `foo` among its `router.deis.io/domains`.  While this behavior is not innately wrong, it may not be desirable.  To circumvent this, specify a [platform domain](#platform-domain).  This will cause routable services specifying one or more non-fully-qualified domain names to be matched, explicitly, as subdomains of the platform domain.  Apart from remediating this minor idiosyncrasy, this is required in order to properly utilize a wildcard SSL certificate and may also result in a very modest performance improvement.

* __Do you need to scale the router?__ For greater availability, it's desirable to run more than one instance of the router.  _How many_ can only be informed by stress/performance testing the applications in your cluster.  To increase the number of router instances from the default of one, increase the number of replicas specified by the `deis-router` replication controller.  Do not specify a number of replicas greater than the number of worker nodes in your Kubernetes cluster.

## License

Copyright 2013, 2014, 2015, 2016 Engine Yard, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

[issues]: https://github.com/deis/router/issues
[prs]: https://github.com/deis/router/pulls
