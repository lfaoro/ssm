#!/usr/bin/env bash
set -e

ROOT=$(git rev-parse --show-toplevel)
TAG=$(git describe --tags --abbrev=0)

echo "=== cloning AUR repo ==="
rm -rf /tmp/ssm-aur
git clone ssh://aur@aur.archlinux.org/ssm-bin.git /tmp/ssm-aur

echo "=== copying PKGBUILD + .SRCINFO ==="
cp "$ROOT/build/aur/ssm-bin.pkgbuild" /tmp/ssm-aur/PKGBUILD
cp "$ROOT/build/aur/ssm-bin.srcinfo"   /tmp/ssm-aur/.SRCINFO

echo "=== committing ==="
cd /tmp/ssm-aur
git add PKGBUILD .SRCINFO
git commit -m "$TAG"

echo "=== pushing ==="
git push origin master

echo "=== cleanup ==="
cd "$ROOT"
rm -rf /tmp/ssm-aur
echo "AUR pushed ($TAG)"
