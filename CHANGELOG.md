# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.6.0] - 2022-07-05

### Changed

- Now uses HSM v2 API.

## [1.5.3] - 2021-08-09

### Changed

- Added GitHub configuration files and fixed snyk warning.

## [1.5.2] - 2021-07-26

### Changed

- Github migration phase 3.

## [1.5.1] - 2021-07-22

### Changed

Add support for building within the CSM Jenkins.

## [1.5.0] - 2021-06-28

### Security

- CASMHMS-4898 - Updated base container images for security updates.

## [1.4.3] - 2021-04-15

### Changed

- Fixed HTTP response leaks.

## [1.4.2] - 2021-04-06

### Changed

- Updated Dockerfiles to pull base images from Artifactory instead of DTR.

## [1.4.1] - 2021-02-04

### Changed

- Added User-Agent headers to outbound HTTP requests.

## [1.4.0] - 2021-02-03

### Changed

- Update Copyright/License and re-vendor go packages

## [1.3.0] - 2021-01-14

### Changed

- Updated license file.

## [1.2.0] - 2020-12-08

### Changed

- CASMINST-568: Strip colons from MAC addresses when building up the Patch URL

## [1.1.1] - 2020-10-20

### Security

- CASMHMS-4105 - Updated base Golang Alpine image to resolve libcrypto vulnerability.
- CASMHMS-4092 - Vendor library code.

## [1.1.0] - 2020-06-11

### Changed

- CASMHMS-3467 - Removed all traces of logic to deal with SMNetManager and converted over to logic targeted at HSM.

## [1.0.1] - 2020-04-29

### Changed

- CASMHMS-2972 - Updated hms-dns-dhcp to use trusted baseOS.

## [1.0.0] - 2019-09-17

### Added

- This is the initial release of the `hms-dns-dhcp` repo. It contains common code for service that wants to interact with the smnetmanager DNS and DHCP services.

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

