# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [v1.1.0] - 2026-03-07

### Breaking
- None.

### Deprecations
- `HTTPServer` / `NewHTTPServer` remain deprecated convenience APIs.

### Behavior Changes
- Added lease-based machine-id abstraction: `MachineIDLeaseProvider` and `MachineIDLease`.
- `Generator` now supports lease-driven machine-id lifecycle (`acquire -> renew -> lost`).
- Added convenience constructors `New(machineID, businessType)` and `MustNew(...)`.
- Added Redis lease provider with token-CAS renew/release semantics.
- Redis lease acquisition now uses single Lua script (cursor + slot acquisition in one call).
- Redis lease slot count is configurable via `RedisMachineIDLeaseOptions.Slots` (default 64).

### Bug Fixes
- Strengthened generator initialization and lease-loss error handling.
- Added lease token CAS safety tests.

### Performance
- Added Redis lease performance tests for single-instance and two-instance scenarios.
- Documented measured throughput and architecture scaling path in README.

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
