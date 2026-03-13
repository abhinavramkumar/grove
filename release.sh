#!/bin/bash
# grove release helper
# Bumps version, creates tag, and pushes to trigger the release workflow.
#
# Usage:
#   ./release.sh patch    # 0.1.0 → 0.1.1
#   ./release.sh minor    # 0.1.0 → 0.2.0
#   ./release.sh major    # 0.1.0 → 1.0.0

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

BUMP_TYPE="${1:-patch}"

if [[ ! "$BUMP_TYPE" =~ ^(patch|minor|major)$ ]]; then
  echo -e "${RED}Usage: ./release.sh [patch|minor|major]${NC}"
  exit 1
fi

# Ensure we're on main and clean
BRANCH=$(git branch --show-current)
if [ "$BRANCH" != "main" ]; then
  echo -e "${RED}Error: must be on main branch (currently on ${BRANCH})${NC}"
  exit 1
fi

if [ -n "$(git status --porcelain)" ]; then
  echo -e "${RED}Error: working directory is not clean${NC}"
  git status --short
  exit 1
fi

# Read current version
CURRENT=$(cat VERSION | tr -d '[:space:]')
echo -e "${GREEN}grove release${NC}"
echo "===================="
echo "Current version: ${CURRENT}"
echo "Bump type: ${BUMP_TYPE}"

# Bump version
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"
case "$BUMP_TYPE" in
  major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
  minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
  patch) PATCH=$((PATCH + 1)) ;;
esac
NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
TAG="v${NEW_VERSION}"

echo "New version: ${NEW_VERSION}"
echo ""

# Check if tag already exists
if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo -e "${RED}Error: tag $TAG already exists${NC}"
  exit 1
fi

read -p "Release ${TAG}? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Aborted."
  exit 0
fi

# Write new version
echo "$NEW_VERSION" > VERSION

# Verify build
echo -e "${YELLOW}Building...${NC}"
go build -ldflags="-s -w -X main.version=${NEW_VERSION}" -o /dev/null ./cmd/grove

echo -e "${YELLOW}Running tests...${NC}"
go test ./...

# Commit, tag, push
git add VERSION
git commit -m "release: v${NEW_VERSION}"
git tag "$TAG"
git push origin main
git push origin "$TAG"

echo ""
echo -e "${GREEN}Released ${TAG}${NC}"
echo "GitHub Actions will build and publish binaries."
echo "Track progress: https://github.com/abhinavramkumar/grove/actions"
