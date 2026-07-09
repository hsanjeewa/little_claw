#!/usr/bin/env bash
# Bootstrap an SSH test server that REQUIRES a sudo password (no NOPASSWD),
# used to verify the agent feeds the per-host sudo password from the vault.
#
# Usage: test-server-sudopw-bootstrap.sh <user> <pubkey-path> <sudo-password>
set -euo pipefail

USER_NAME="${1:?usage: <user> <pubkey-path> <sudo-password>}"
PUBKEY_PATH="${2:?usage: <user> <pubkey-path> <sudo-password>}"
SUDO_PASS="${3:?usage: <user> <pubkey-path> <sudo-password>}"

export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y openssh-server sudo
mkdir -p /run/sshd

id -u "$USER_NAME" >/dev/null 2>&1 || useradd -m -s /bin/bash "$USER_NAME"

# Password REQUIRED sudo (explicitly NOT NOPASSWD) + set the user's password.
echo "$USER_NAME ALL=(ALL) ALL" > "/etc/sudoers.d/$USER_NAME"
chmod 440 "/etc/sudoers.d/$USER_NAME"
echo "$USER_NAME:$SUDO_PASS" | chpasswd

USER_HOME=$(getent passwd "$USER_NAME" | cut -d: -f6)
mkdir -p "$USER_HOME/.ssh"
chmod 700 "$USER_HOME/.ssh"
cat "$PUBKEY_PATH" >> "$USER_HOME/.ssh/authorized_keys"
chmod 600 "$USER_HOME/.ssh/authorized_keys"
chown -R "$USER_NAME:$USER_NAME" "$USER_HOME/.ssh"

ssh-keygen -A >/dev/null 2>&1 || true

exec /usr/sbin/sshd -D -e
