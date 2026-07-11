#!/usr/bin/env just --justfile

[group('release')]
create-prod-release:
    nix build --impure .#default

# Check release readiness before tagging.
[group('release')]
check-release:
    #!/usr/bin/env bash
    set -euo pipefail

    @just @check-release-version
    # Check if builds
    nix build .#default --no-link

# Check if the version in plugin/manifest.json and flake.nix is greater than the latest release on GitHub.
[group('release')]
check-release-version:
    #!/usr/bin/env bash
    set -euo pipefail

    for command in nix jq perl git curl; do
        if ! command -v "$command" >/dev/null 2>&1; then
            echo "Missing required command: $command" >&2
            exit 1
        fi
    done

    # Check if versions are ok
    manifest_version=$(jq -r '.version' plugin/manifest.json)
    flake_version=$(perl -ne 'print "$1\n" if /version = "([^"]+)";/' flake.nix | head -n 1)

    if [[ -z "$flake_version" ]]; then
        echo "Could not read version from flake.nix" >&2
        exit 1
    fi

    # All version files must agree before a release tag is created.
    if [[ "$manifest_version" != "$flake_version" ]]; then
        echo "Version mismatch:" >&2
        echo "  plugin/manifest.json: $manifest_version" >&2
        echo "  flake.nix: $flake_version" >&2
        exit 1
    fi

    # Check if local version is higher than latest release on github
    remote_url=$(git remote get-url origin)
    repo=$(perl -ne 'print "$1\n" if m{github\.com[:/]([^/]+/[^/.]+)(?:\.git)?$}' <<<"$remote_url")
    if [[ -z "$repo" ]]; then
        echo "Could not determine GitHub repo from origin: $remote_url" >&2
        exit 1
    fi

    latest_tag=$(curl -fsSL "https://api.github.com/repos/$repo/releases/latest" | jq -r '.tag_name')
    latest_version=${latest_tag#v}

    semver_gt() {
        local current=$1 latest=$2
        IFS=. read -r current_major current_minor current_patch <<<"$current"
        IFS=. read -r latest_major latest_minor latest_patch <<<"$latest"

        # Ignore prerelease/build metadata for numeric comparison.
        current_patch=${current_patch%%[-+]*}
        latest_patch=${latest_patch%%[-+]*}

        for part in current_major current_minor current_patch latest_major latest_minor latest_patch; do
            if [[ ! ${!part} =~ ^[0-9]+$ ]]; then
                echo "Unsupported semver value: current=$current latest=$latest" >&2
                return 2
            fi
        done

        if (( current_major != latest_major )); then
            (( current_major > latest_major ))
            return
        fi
        if (( current_minor != latest_minor )); then
            (( current_minor > latest_minor ))
            return
        fi
        (( current_patch > latest_patch ))
    }

    if ! semver_gt "$manifest_version" "$latest_version"; then
        echo "Version must be greater than latest GitHub release:" >&2
        echo "  current: $manifest_version" >&2
        echo "  latest:  $latest_tag" >&2
        exit 1
    fi

    echo "Ready: $manifest_version > $latest_tag"
