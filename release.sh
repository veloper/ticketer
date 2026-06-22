#!/usr/bin/env bash
set -euo pipefail

# ──────────────────────────────────────────────────────────────────────────────
# Ticketer Release Script
# ──────────────────────────────────────────────────────────────────────────────
# 1. Commits any pending changes to main and pushes to GitHub
# 2. Cuts a release tag: vYYYY.MM.DD.<SSSSS> (e.g. v2026.06.21.45230)
#    where SSSSS is seconds since start of day, zero-padded to 5 digits
# 3. Builds the Docker image
# 4. Pushes to Docker Hub
# ──────────────────────────────────────────────────────────────────────────────

# Configuration
DOCKER_IMAGE="veloper/ticketer"

# ── Step 1: Commit and push ──────────────────────────────────────────────────

echo "==> Checking for pending changes..."

if ! git diff --quiet --cached || ! git diff --quiet; then
    echo "==> Staging all changes..."
    git add -A
    echo "==> Committing..."
    git commit -m "release prep"
fi

echo "==> Pushing to main..."
git push origin main

# ── Step 2: Create release tag ───────────────────────────────────────────────

DATE=$(date +%Y.%-m.%-d)                     # e.g. 2026.6.21
SECONDS=$(date +%H*3600+%M*60+%S | bc)      # seconds since midnight
TAG=$(printf "v%s.%05d" "$DATE" "$SECONDS")  # e.g. v2026.6.21.45230

echo "==> Creating tag: $TAG"
git tag "$TAG"
git push origin "$TAG"

# ── Step 3: Build Docker image ───────────────────────────────────────────────

echo "==> Building Docker image: $DOCKER_IMAGE:latest"
docker build -t "$DOCKER_IMAGE:latest" .

# Also tag with the release version
echo "==> Tagging: $DOCKER_IMAGE:$TAG"
docker tag "$DOCKER_IMAGE:latest" "$DOCKER_IMAGE:$TAG"

# ── Step 4: Push to Docker Hub ───────────────────────────────────────────────

echo "==> Pushing to Docker Hub..."
docker push "$DOCKER_IMAGE:latest"
docker push "$DOCKER_IMAGE:$TAG"

echo ""
echo "✅ Release $TAG complete!"
echo "   https://hub.docker.com/r/$DOCKER_IMAGE/tags"
