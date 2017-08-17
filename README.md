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
[9]: https://godoc.org/fiunchinho/dmz-controller?status.svg "GoDoc badge"
[10]: https://godoc.org/fiunchinho/dmz-controller "GoDoc"
[11]: https://img.shields.io/docker/pulls/fiunchinho/dmz-controller.svg?maxAge=604800 "Docker Pulls"
[12]: https://hub.docker.com/r/fiunchinho/dmz-controller/ "DockerHub"
[13]: https://codeclimate.com/github/fiunchinho/dmz-controller/badges/gpa.svg "Code Climate badge"
[14]: https://codeclimate.com/github/fiunchinho/dmz-controller "Code Climate"
[15]: https://codeclimate.com/github/fiunchinho/dmz-controller/badges/issue_count.svg "Code Climate badge"
[16]: https://codeclimate.com/github/fiunchinho/dmz-controller "Code Climate Issues"

This is a kubernetes controller that watches Ingress objects that contain a specific annotation and adds whitelisted addresses to it.

## Motivation
We expose applications running on Kubernetes using Ingress rules. These applications can be either:
- Public: all the internet can access this application
- Private: only reachable from known sources like offices, VPN's and so on. 

Handling these lists of kwown sources is costly and error prone. This controllers tries to automate this process.

# How it works
Whenever an Ingress object is created containing the annotation `armesto.net/ingress: "office"`,
this controller will add the `ingress.kubernetes.io/whitelist-source-range` annotation to the Ingress object with some addresses.

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test-ingress
  namespace: default
  annotations:
    armesto.net/ingress: vpn
spec:
  backend:
    serviceName: testsvc
    servicePort: 80
```

Which ones? The whitelisted addresses come from a ConfigMap that contains a map for different sources. The addresses in the key specified in the `armesto.net/ingress` annotation will be whitelisted.
If we'd have the following `ConfigMap`, and our Ingress object annotated with `armesto.net/ingress: "office"`, the addresses 8.8.8.8/32 and 8.8.4.4/32 would be whitelisted.

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
The names of the keys in the ConfigMap are arbitrary, so you can write whatever data you like.

The controller is also watching the ConfigMap, so whenever a change is made to the ConfigMap (to add/remove addresses, for example), the controller will go over all the Ingress objects to see if a change needs to be done to the whitelist.

## Multiple providers
You can even choose multiple providers.
If we'd have the following `ConfigMap`, and our Ingress object annotated with `armesto.net/ingress: "office,vpn"`, the addresses 8.8.8.8/32, 8.8.4.4/32 and 123.123.123.123/28 would be whitelisted.
