#!/usr/bin/env bash
# Bootstrap a minimal SSH test server inside a base Linux container.
#
# Usage: test-server-bootstrap.sh <user> <pubkey-path>
#
# Creates the user, installs openssh-server, enables passwordless (NOPASSWD)
# sudo, injects the given SSH public key, and starts sshd in the foreground.
# This mirrors the contract the agent expects from lscr.io/linuxserver/openssh-server
# (USER_NAME + SUDO_ACCESS=true + PUBLIC_KEY) for distros that don't ship it.
set -euo pipefail

USER_NAME="${1:?usage: test-server-bootstrap.sh <user> <pubkey-path>}"
PUBKEY_PATH="${2:?usage: test-server-bootstrap.sh <user> <pubkey-path>}"

echo "==> Detecting package manager"
if command -v apt-get >/dev/null 2>&1; then
  PKG=apt
elif command -v dnf >/dev/null 2>&1; then
  PKG=dnf
elif command -v yum >/dev/null 2>&1; then
  PKG=yum
elif command -v apk >/dev/null 2>&1; then
  PKG=apk
else
  echo "No supported package manager found" >&2
  exit 1
fi

echo "==> Installing openssh-server and sudo ($PKG)"
case "$PKG" in
  apt)
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -y
    apt-get install -y openssh-server sudo
    mkdir -p /run/sshd
    ;;
  dnf|yum)
    "$PKG" install -y openssh-server sudo
    mkdir -p /run/sshd
    ;;
  apk)
    apk add --no-cache openssh sudo
    mkdir -p /run/sshd
    ;;
esac

echo "==> Creating user $USER_NAME"
if ! id -u "$USER_NAME" >/dev/null 2>&1; then
  useradd -m -s /bin/bash "$USER_NAME"
fi

echo "==> Configuring passwordless sudo"
echo "$USER_NAME ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/$USER_NAME"
chmod 440 "/etc/sudoers.d/$USER_NAME"

echo "==> Installing SSH public key"
USER_HOME=$(getent passwd "$USER_NAME" | cut -d: -f6)
mkdir -p "$USER_HOME/.ssh"
chmod 700 "$USER_HOME/.ssh"
cat "$PUBKEY_PATH" >> "$USER_HOME/.ssh/authorized_keys"
chmod 600 "$USER_HOME/.ssh/authorized_keys"
chown -R "$USER_NAME:$USER_NAME" "$USER_HOME/.ssh"

echo "==> Ensuring SSH host keys exist"
if [ ! -f /etc/ssh/ssh_host_ed25519_key ]; then
  ssh-keygen -A >/dev/null 2>&1 || true
fi

echo "==> Starting sshd"
exec /usr/sbin/sshd -D -e
