# Gardener Extension for S3 Compatible Storage

[![GitHub License](https://img.shields.io/github/license/metal-stack/os-metal-extension)](https://github.com/metal-stack/os-metal-extension/blob/master/LICENSE)
[![Build](https://github.com/metal-stack/os-metal-extension/actions/workflows/build.yaml/badge.svg)](https://github.com/metal-stack/os-metal-extension/actions/workflows/build.yaml)

This controller operates on the [`OperatingSystemConfig`](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md#cloud-config-user-data-for-bootstrapping-machines) resource in the `extensions.gardener.cloud/v1alpha1` API group. It manages those objects that are requesting:

- [Ubuntu](https://www.ubuntu.com/) configuration (`.spec.type=ubuntu`)
- [Debian](https://www.debian.org/) configuration (`.spec.type=debian`)

In comparison to the official OSC extensions from Gardener, this extension transforms the `OperatingSystemConfig` resources into [Ignition](https://www.flatcar.org/docs/latest/provisioning/ignition/) userdata. This userdata can be applied during machine provisioning as done by the metal-stack project.

This extension was made for working with operating system images built in the [metal-images](https://github.com/metal-stack/metal-images) repository.

## Example

An example `ControllerRegistration` resource that can be used to register this controller to Gardener can be found [here](example/controller-registration.yaml).

## Development

Development currently needs to happen against a real environment because there are many dependencies to external APIs for reconciliation. It is planned to allow development in the [mini-lab](https://github.com/metal-stack/mini-lab) soon.

## Feedback and Support

Feedback and contributions are always welcome! Please report bugs or suggestions as [GitHub issues](https://github.com/metal-stack/os-metal-extension/issues) or reach out to our [community](https://metal-stack.io/community).
