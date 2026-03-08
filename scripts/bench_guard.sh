#!/usr/bin/env bash
set -euo pipefail

# Conservative upper bounds (ns/op) to catch obvious regressions.
BENCH_NAMES=(
  "BenchmarkEncodeBase62"
  "BenchmarkDecodeBase62"
  "BenchmarkToTimestampShort"
  "BenchmarkFromTimestampShort"
  "BenchmarkToTimestampDynamic"
  "BenchmarkToTimestampCompact"
  "BenchmarkFromTimestampCompact"
)

BENCH_LIMITS=(
  "120"
  "80"
  "300"
  "80"
  "20"
  "350"
  "150"
)

output_file="$(mktemp)"
trap 'rm -f "$output_file"' EXIT

# Run only root-package benchmarks tracked in LIMITS.
go test -run '^$' -bench 'Benchmark(EncodeBase62|DecodeBase62|ToTimestampShort|FromTimestampShort|ToTimestampDynamic|ToTimestampCompact|FromTimestampCompact)$' ./... >"$output_file"

status=0
for i in "${!BENCH_NAMES[@]}"; do
  name="${BENCH_NAMES[$i]}"
  limit="${BENCH_LIMITS[$i]}"
  line="$(grep -E "^${name}-" "$output_file" || true)"
  if [[ -z "$line" ]]; then
    echo "[bench-guard] missing benchmark output for ${name}"
    status=1
    continue
  fi

  ns="$(awk '{for(i=1;i<=NF;i++){if($i ~ /ns\/op$/){print $(i-1); exit}}}' <<<"$line")"
  if [[ -z "$ns" ]]; then
    echo "[bench-guard] cannot parse ns/op for ${name}: $line"
    status=1
    continue
  fi

  if ! awk -v val="$ns" -v max="$limit" 'BEGIN { exit !(val <= max) }'; then
    echo "[bench-guard] regression: ${name} ${ns}ns/op > ${limit}ns/op"
    status=1
  else
    echo "[bench-guard] ok: ${name} ${ns}ns/op <= ${limit}ns/op"
  fi
done

exit "$status"
