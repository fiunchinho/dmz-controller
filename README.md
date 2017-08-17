# DMZ Controller
[![Build Status][1]][2] [![codecov.io][3]][4] [![Go Report][5]][6] [![GoDoc][7]][8] [![Docker Pulls][9]][10] [![Code CLimate][11]][12] [![Code CLimate Issues][13]][14]

[1]: https://travis-ci.org/fiunchinho/dmz-controller.svg?branch=master "Build Status badge"
[2]: https://travis-ci.org/fiunchinho/dmz-controller "Travis-CI Build Status"
[3]: https://codecov.io/github/v2ray/v2ray-core/coverage.svg?branch=master "Coverage badge"
[4]: https://codecov.io/github/v2ray/v2ray-core?branch=master "Codecov Status"
[5]: https://goreportcard.com/badge/github.com/fiunchinho/dmz-controller "Go Report badge"
[6]: https://goreportcard.com/report/github.com/fiunchinho/dmz-controller "Go Report"
[7]: https://godoc.org/v2ray.com/core?status.svg "GoDoc badge"
[8]: https://godoc.org/v2ray.com/core "GoDoc"
[9]: https://img.shields.io/docker/pulls/fiunchinho/dmz-controller.svg?maxAge=604800 "Docker Pulls"
[10]: https://hub.docker.com/u/fiunchinho/dmz-controller "DockerHub"
[11]: https://codeclimate.com/github/fiunchinho/dmz-controller/badges/gpa.svg "Code Climate badge"
[12]: https://codeclimate.com/github/fiunchinho/dmz-controller "Code Climate"
[13]: https://codeclimate.com/github/fiunchinho/dmz-controller/badges/issue_count.svg "Code Climate badge"
[14]: https://codeclimate.com/github/fiunchinho/dmz-controller "Code Climate Issues"

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
