

build:
  #!/usr/bin/env bash
  set -euo pipefail

  pnpm run build

  go build -o app ./...

