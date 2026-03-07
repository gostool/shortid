# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Breaking
- None.

### Deprecations
- `HTTPServer` / `NewHTTPServer` are marked as deprecated convenience APIs.

### Behavior Changes
- `ValidateConfig` added and used by `NewGenerator`.
- Config now rejects mixed `MachineID` + `MachineIDProvider` usage.

### Bug Fixes
- Added failure-injection and edge-case tests for generator and timestamp helpers.

### Performance
- Added benchmark regression guard workflow.
