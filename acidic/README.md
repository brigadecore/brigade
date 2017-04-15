# acidic: Az Continuous Integration and Delivery Interface Containers

The `acidic` package contains all the IC docker definitions that are part of the
Acid core.

- `acid-go`: Builder for Go-specific projects
- `acid-ubuntu`: Basic Ubuntu image plus `make` and `git`

These can be build from the top-level `make docker-build` target.

## What is `hook.sh` for?

The `hook.sh` script is the top-level command
