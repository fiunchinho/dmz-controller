# DMZ Controller
[![License][1]][2] [![Build Status][3]][4] [![codecov.io][5]][6] [![Go Report][7]][8] [![GoDoc][9]][10] [![Docker Pulls][11]][12] [![Code CLimate][13]][14] [![Code CLimate Issues][15]][16]

[1]: https://img.shields.io/badge/license-MIT-blue.svg "MIT License"
[2]: https://github.com/fiunchinho/dmz-controller/blob/master/LICENSE.md "License"
[3]: https://travis-ci.org/fiunchinho/dmz-controller.svg?branch=master "Build Status badge"
[4]: https://travis-ci.org/fiunchinho/dmz-controller "Travis-CI Build Status"
[5]: https://codecov.io/github/fiunchinho/dmz-controller/coverage.svg?branch=master "Coverage badge"
[6]: https://codecov.io/github/fiunchinho/dmz-controller?branch=master "Codecov Status"
[7]: https://goreportcard.com/badge/github.com/fiunchinho/dmz-controller "Go Report badge"
[8]: https://goreportcard.com/report/github.com/fiunchinho/dmz-controller "Go Report"
[9]: https://godoc.org/github.com/fiunchinho/dmz-controller?status.svg "GoDoc badge"
[10]: https://godoc.org/github.com/fiunchinho/dmz-controller "GoDoc"
[11]: https://img.shields.io/docker/pulls/fiunchinho/dmz-controller.svg?maxAge=604800 "Docker Pulls"
[12]: https://hub.docker.com/r/fiunchinho/dmz-controller/ "DockerHub"
[13]: https://codeclimate.com/github/fiunchinho/dmz-controller/badges/gpa.svg "Code Climate badge"
[14]: https://codeclimate.com/github/fiunchinho/dmz-controller "Code Climate"
[15]: https://codeclimate.com/github/fiunchinho/dmz-controller/badges/issue_count.svg "Code Climate badge"
[16]: https://codeclimate.com/github/fiunchinho/dmz-controller "Code Climate Issues"

This is a kubernetes controller that watches Ingress objects that contain a specific annotation and adds whitelisted addresses to it.

## Motivation
We use Ingress rules to expose applications that need to be accessed by other applications running outside our Kubernetes cluster.
Even though they are exposed to outside the cluster, sometimes we don't want them to be exposed to the whole internet.
We want to be able to whitelist the known sources that are allowed to access those applications, like offices or VPN's IPs.
So basically there are two type of applications
- Public: everybody can access this application
- Private: only traffic from known sources (like offices, VPN's and so on) is allowed 

Manually handling these lists of kwown sources is costly and error prone. This controllers tries to automate this process.

# How it works
Let's say we want to create an `Ingress` object to expose our application to the outside.
We could manually add IP's to the [ingress.kubernetes.io/whitelist-source-range annotation](https://github.com/kubernetes/ingress/blob/master/controllers/nginx/configuration.md#whitelist-source-range) to allow traffic from those IP's.
Instead, we'll add the `armesto.net/ingress` annotation, so the dmz-controller will take care of adding the whitelisted IP's. For example:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: my-application-ingress
  namespace: default
  annotations:
    armesto.net/ingress: office
spec:
  backend:
    serviceName: my-application-service
    servicePort: 80
```

Once we create the Ingress object, this controller will add the `ingress.kubernetes.io/whitelist-source-range` annotation with some IP addresses.
These addresses come from the dmz-controller `ConfigMap` that contains a map where we can store different IP sources, like

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dmz-controller
  namespace: default
data:
  office: 8.8.8.8/32,8.8.4.4/32
  vpn: 123.123.123.123/28
```

The IP's named with the key specified in the `armesto.net/ingress` annotation will be whitelisted.
Using the previous `Ingress` and `ConfigMap`, the IP's `8.8.8.8/32` and `8.8.4.4/32` would be whitelisted, because they are the office IP's.

The names of the keys in the `ConfigMap` are arbitrary: you can choose the names you like.

The controller is also watching the `ConfigMap`, so whenever a change is made (to add/remove addresses, for example), the controller will go over all the `Ingress` objects to see if a change needs to be done to its whitelist.

## Hybrid providers
The controller will respect whitelisted sources that were added to the Ingress object manually.
It only manages the [CIDRs](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing) that come from the ConfigMap, leaving the rest untouched.

It accomplishes it by adding an internal annotation `dmz-controller` to keep track of the addresses managed by the controller that came from the `ConfigMap`.

## Multiple providers
You can even choose multiple providers.

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: my-application-ingress
  namespace: default
  annotations:
    armesto.net/ingress: office,vpn
spec:
  backend:
    serviceName: my-application-service
    servicePort: 80
```

In this case, the addresses `8.8.8.8/32`, `8.8.4.4/32` and `123.123.123.123/28` would be added to the `Ingress` whitelist.

## Address Format
Addresses added to the `ConfigMap` need to be valid IP's or [CIDRs](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing).
If you store an IP, it will be transformed to a CIDR. For example, if you add the `8.8.8.8` IP, the controller will use it as if you had added the `8.8.8.8/32` CIDR. 
