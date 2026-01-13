# Docker Best Practices (Project)

## Base Images and Pinning
- Pin Go and Alpine versions via build args to keep builds reproducible.
- Use multi-stage builds to keep runtime image small.

## Build Caching
- Use BuildKit cache mounts for `go mod download` and `go build` to speed up CI and local builds.

## Build Flags
- Use `-trimpath` and `-ldflags="-s -w"` to remove build paths and reduce binary size.
- Keep `CGO_ENABLED=0` for static binaries unless a dependency requires CGO.

## File Copying and Ownership
- Copy only required artifacts into the runtime image.
- Use `--chown` on copied files to avoid extra `chown` layers.

## Users and Permissions
- Run as a non-root user in the runtime image.
- Create only the directories you need and set ownership explicitly.

## Runtime Image Hygiene
- Install only runtime dependencies (e.g. `ca-certificates`).
- Keep `.dockerignore` up to date to avoid bloated build contexts.
