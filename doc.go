// Package shortid provides a stable, production-ready ID generation SDK.
//
// Core responsibilities:
//   - Distributed unique ID generation (Sonyflake-like layout)
//   - Base encoding/decoding helpers
//   - Timestamp compression helpers
//   - Pluggable machine-id and sequence providers
//
// Non-goals:
//   - Transport/runtime concerns (HTTP/gRPC servers) are not a core evolution
//     target of this package. Build transport layers in application services.
package shortid
