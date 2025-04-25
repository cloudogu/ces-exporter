# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## Changed
- Update makefiles to 9.8.0
- Implement system info endpoint [#6]
- implement maintenance mode endpoint [#18]
- Implement export mode endpoint for multinode [#14] and classic mode [#20]
- Implement configuration endpoint for classic ces [#16]

## Added
- New Make-Target (`make debian`) to build a debian package to install the exporter in a classic CES [#12]
  - This also contains several refactorings to make the application run without a kubernetes environment

## [v0.0.1] - 2025-03-14
- initial release