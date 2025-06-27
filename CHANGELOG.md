# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Fixed
- [#43] Exclude ces-exporter from multinode system info 

## [v1.0.0] - 2025-06-23
- 🎉🎉 First release 🎉🎉

### Fixed
- Do not error when deactivating already deactivated maintenance mode [#38]
### Changed
- Interpret volumeIncreaseFactor as a factor instead of percentage. [#34]

## [v0.0.4] - 2025-06-12
### Changed
- BackupSchedules are only returned when the backup is active [#32]

### Fixed
- Rename BackupSchedule name to be a valid lowercase RFC 1123 subdomain [#32]

## [v0.0.3] - 2025-05-23
### Fixed
- Do not add binary to debian-package
- Rename systemctl start-script
- Return correct dogu-volume-path in export-api

## [v0.0.2] - 2025-05-21
### Changed
- Update makefiles to 9.8.0
- Implement system info endpoint [#6]
- implement maintenance mode endpoint [#18]
- Implement export mode endpoint for multinode [#14] and classic mode [#20]
- Implement configuration endpoint for classic ces [#16]

### Added
- New Make-Target (`make debian`) to build a debian package to install the exporter in a classic CES [#12]
  - This also contains several refactorings to make the application run without a kubernetes environment

### Fixed
- Use full dogu-name in the system-info api-endpoint of the multinode-exporter 
- Registration with registrator 
- Error getting backup schedules when backup-dogu was not installed

## [v0.0.1] - 2025-03-14
- initial release