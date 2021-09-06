<p align=center><a href="https://nddocs.yndd.io"><img src=https://gitlab.com/rdodin/pics/-/wikis/uploads/9f2e581a8d207a21ff024a312679a239/containerlab_export_white_ink_3?sanitize=true/></a></p>

[![github release](https://img.shields.io/github/release/yndd/ndd-core.svg?style=flat-square&color=00c9ff&labelColor=bec8d2)](https://github.com/yndd/ndd-core/releases/)
[![Go Report](https://img.shields.io/badge/go%20report-A%2B-blue?style=flat-square&color=00c9ff&labelColor=bec8d2)](https://goreportcard.com/report/github.com/yndd/ndd-core)
[![Godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square&color=00c9ff&labelColor=bec8d2)](https://godoc.org/github.com/yndd/ndd-core)
[![Doc](https://img.shields.io/badge/Docs-ndddocs.yndd.io-blue?style=flat-square&color=00c9ff&labelColor=bec8d2)](https://ndddocs.yndd.io)
[![Docker Pulls](https://img.shields.io/docker/pulls/yndd/ndd-core-controller.svg?style=flat-square&color=00c9ff&labelColor=bec8d2)](https://img.shields.io/docker/pulls/yndd/ndd-core.svg) 
[![Twitter Follow](https://img.shields.io/twitter/follow/yndd.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=yndd&user_id=1434394355385651201)

## Overview
 
NDD is an opensource [Kubernetes] add-on that enables platform and application teams to consume network devices in a similar way as other resources are consumed in [Kubernetes]. 

NDD uses a modular approach, through providers, which allows multiple network device types to be supported. NDD allows the network providers to be generated from [YANG], which enables rapid enablement of multiple network device types. Through YANG we can provider automate input and dependency management between the various resource that are consumed within the device.

An NDD provider represents the device model through various [CRs] within the Kubernetes API in order to provide flexible management of the device resources.

NDD is build on the basis of the [kubebuilder] and [operator-pattern] within kubernetes.

Features:

* Device discovery and Provider registration
* Declaritive CRUD configuration of network devices through [CRs]
* Configuration Input Validation:
    - Declarative validation using an OpenAPI v3 schema derived from [YANG]
    - Runtime Dependency Management amongst the various resources comsumed within a device (parent dependency management and leaf reference dependency management amont resources)
* Automatic or Operator interacted configuration drift management
* Delete Policy, and Active etc  

## Releases

NDD is in alpha phase so dont use it in production

## Getting Started

Take a look at the [documentation] to get started.

## Get involved

ndd is a community driven project and we welcome contribution.

- Discord: [discord]
- Twitter: [@yndd]
- Email: [info@yndd.io]

For filling bugs, suggesting improvments, or requesting new feature, please open an [issue].

## Code of conduct

## Licensing

ndd-runtime is under the Apache 2.0 license.

[documentation]: https://ndddocs.yndd.io
[issue]: https://github.com/yndd/ndd-core/issues
[roadmap]: https//github.com/yndd/tbd
[discord]: https://discord.gg/prHcBMSq
[@yndd]: https://twitter.com/yndd
[info@yndd.io]: mailto:info@yndd.io

[Kubernetes]: https://kubernetes.io
[YANG]: https://en.wikipedia.org/wiki/YANG
[CRs]: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/
[kubebuilder]: https://kubebuilder.io
[operator-pattern]: https://kubernetes.io/docs/concepts/extend-kubernetes/operator/