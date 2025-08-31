#!/usr/bin/env sh
set -eu

VERSION_FILE="VERSION"
CHANGELOG="CHANGELOG.md"
DRY_RUN="${DRY_RUN:-0}"

die() { echo "error: $*" >&2; exit 1; }

read_version() {
  [ -f "$VERSION_FILE" ] || die "VERSION file not found"
  ver=$(tr -d '\n' < "$VERSION_FILE")
  echo "$ver"
}

write_version_real() {
  v="$1"
  printf "%s\n" "$v" > "$VERSION_FILE"
}

is_prerelease() {
  echo "$1" | grep -Eq '^[^-]+-(alpha|beta|rc)\.[0-9]+$'
}

base_part() {
  echo "${1%%-*}"
}

bump_patch() {
  v="$1"; base=$(base_part "$v")
  base_no_v=${base#v}
  major=${base_no_v%%.*}
  rest=${base_no_v#*.}
  minor=${rest%%.*}
  patch=${rest#*.}
  patch=$((patch + 1))
  echo "v${major}.${minor}.${patch}"
}

bump_minor() {
  v="$1"; base=$(base_part "$v")
  base_no_v=${base#v}
  major=${base_no_v%%.*}
  rest=${base_no_v#*.}
  minor=${rest%%.*}
  minor=$((minor + 1))
  echo "v${major}.${minor}.0"
}

bump_major() {
  v="$1"; base=$(base_part "$v")
  base_no_v=${base#v}
  major=${base_no_v%%.*}
  major=$((major + 1))
  echo "v${major}.0.0"
}

start_prerelease() {
  base="$1"; type="$2" # alpha|beta|rc
  # start prerelease series at .0 (e.g., v1.2.3-alpha.0)
  echo "${base}-${type}.0"
}

bump_prerelease() {
  v="$1"; type="$2"
  # use -- to prevent pattern beginning with '-' from being treated as options
  echo "$v" | grep -Eq -- "-${type}\\.[0-9]+$" || die "version $v is not a ${type} prerelease"
  n=$(echo "$v" | sed -E "s/.*-${type}\.([0-9]+)/\1/")
  n=$((n + 1))
  echo "$(echo "$v" | sed -E "s/(-${type}\.)[0-9]+/\1${n}/")"
}

need_changelog_header() {
  v="$1"
  [ -f "$CHANGELOG" ] || { echo 0; return; }
  if grep -q "^## $v" "$CHANGELOG" 2>/dev/null; then echo 0; else echo 1; fi
}

add_changelog_header_real() {
  v="$1"
  [ -f "$CHANGELOG" ] || return 0
  date=$(date +%Y-%m-%d)
  if ! grep -q "^## $v" "$CHANGELOG" 2>/dev/null; then
    tmp=$(mktemp)
    {
      echo "## $v - $date"
      echo
      echo "- Summary: (fill in)"
      echo
      cat "$CHANGELOG"
    } > "$tmp"
    mv "$tmp" "$CHANGELOG"
  fi
}

commit_if_any() {
  changed="$1"; msg="$2"
  if [ "$DRY_RUN" = "1" ]; then
    if [ "$changed" = "1" ]; then
      echo "commit: $msg"
    else
      echo "commit: $msg (no changes; skip)"
    fi
    return 0
  fi
  [ "$changed" = "1" ] || return 0
  git add "$VERSION_FILE" "$CHANGELOG" 2>/dev/null || true
  git commit -m "$msg"
}

plan_write_version_if_needed() {
  from="$1"; to="$2"
  if [ "$from" = "$to" ]; then return 1; fi
  if [ "$DRY_RUN" = "1" ]; then
    echo "write VERSION: $from -> $to"
  else
    write_version_real "$to"
  fi
  return 0
}

plan_add_changelog_header() {
  v="$1"
  need=$(need_changelog_header "$v")
  if [ "$need" = "1" ]; then
    if [ "$DRY_RUN" = "1" ]; then
      echo "prepend CHANGELOG section: $v"
    else
      add_changelog_header_real "$v"
    fi
    return 0
  fi
  return 1
}

plan_tag() {
  tag="$1"
  if [ "$DRY_RUN" = "1" ]; then
    echo "tag: $tag"
  else
    git tag -a "$tag" -m "$tag"
  fi
}

case "${1:-}" in
  release)
    kind="${2:-}" # alpha|beta|stable
    [ -n "$kind" ] || die "usage: semver.sh release {alpha|beta|stable|bump} [...]"
    if [ "$kind" = "bump" ]; then
      scope="${3:-}"
      [ -n "$scope" ] || die "usage: semver.sh release bump {minor|major}"
      curr=$(read_version)
      base=$(base_part "$curr")
      case "$scope" in
        minor) next_base=$(bump_minor "$base") ;;
        major) next_base=$(bump_major "$base") ;;
        *) die "invalid bump scope: $scope (use minor|major)" ;;
      esac
      next_dev=$(start_prerelease "$next_base" "alpha")
      if [ "$DRY_RUN" = "1" ]; then
        echo "prepared: $next_dev (no tag)"
        echo "Plan (dry-run): release bump $scope"
        echo "current VERSION: $curr"
      fi
      if plan_write_version_if_needed "$curr" "$next_dev"; then changed=1; else changed=0; fi
      if plan_add_changelog_header "$next_dev"; then :; fi
      commit_if_any "$changed" "chore: start next cycle: $next_dev"
      exit 0
    fi
    # normal release flow
    curr=$(read_version)
    base=$(base_part "$curr")
    case "$kind" in
      alpha|beta|rc)
        # Compute release version to display plan summary first
        if is_prerelease "$curr" && echo "$curr" | grep -q "$kind"; then
          release_v="$curr"
        else
          release_v=$(start_prerelease "$base" "$kind")
        fi
        if [ "$DRY_RUN" = "1" ]; then
          echo "prepared: $release_v"
          echo "Plan (dry-run): release $kind"
          echo "current VERSION: $curr"
        fi
        # Ensure VERSION reflects release_v if starting prerelease
        if [ "$release_v" != "$curr" ]; then
          if plan_write_version_if_needed "$curr" "$release_v"; then changed1=1; else changed1=0; fi
        fi
        if plan_add_changelog_header "$release_v"; then changed2=1; else changed2=0; fi
        if [ ${changed1:-0} -eq 1 ] || [ ${changed2:-0} -eq 1 ]; then changed=1; else changed=0; fi
        commit_if_any "$changed" "release: $release_v"
        plan_tag "$release_v"

        next_dev=$(bump_prerelease "$release_v" "$kind")
        if plan_write_version_if_needed "$release_v" "$next_dev"; then changed3=1; else changed3=0; fi
        if plan_add_changelog_header "$next_dev"; then :; fi
        commit_if_any "$changed3" "chore: bump version: $release_v -> $next_dev"
        echo "Prepared release $release_v. Review and push tag:"
        echo "  git push origin $release_v"
        ;;
      stable)
        # Formal release: tag base version (strip prerelease) and update changelog for it
        release_v="$base"
        if [ "$DRY_RUN" = "1" ]; then
          echo "prepared: $release_v"
          echo "Plan (dry-run): release stable"
          echo "current VERSION: $curr"
        fi
        if plan_write_version_if_needed "$curr" "$release_v"; then changed1=1; else changed1=0; fi
        if plan_add_changelog_header "$release_v"; then changed2=1; else changed2=0; fi
        if [ ${changed1:-0} -eq 1 ] || [ ${changed2:-0} -eq 1 ]; then changed=1; else changed=0; fi
        commit_if_any "$changed" "release: $release_v"
        plan_tag "$release_v"

        # Next development cycle: next patch alpha.0
        next_base=$(bump_patch "$release_v")
        next_dev=$(start_prerelease "$next_base" "alpha")
        if plan_write_version_if_needed "$release_v" "$next_dev"; then changed3=1; else changed3=0; fi
        if plan_add_changelog_header "$next_dev"; then :; fi
        commit_if_any "$changed3" "chore: start next cycle: $next_dev"
        echo "Prepared release $release_v. Review and push tag:"
        echo "  git push origin $release_v"
        ;;
      *) die "invalid release kind: $kind" ;;
    esac
    ;;

  *)
    die "usage: semver.sh release ..."
    ;;
esac
