#!/usr/bin/env bash
# One-time (and repeatable) server setup for the auto deploy (#211). Run as
# root from a checkout of the repository:
#
#   sudo bash scripts/deploy/install.sh /path/to/key.pub
#
# It creates the user "trisse-deploy", installs the deploy script root-owned,
# allows that user to run exactly that script with sudo and installs the public
# key so that it can run nothing else (SSH forced command). Run it again after
# changing scripts/deploy/trisse-home-deploy: what runs as root is the installed
# copy, not whatever the repository contains.
set -euo pipefail

[[ $(id -u) -eq 0 ]] || { echo "run as root (sudo)" >&2; exit 1; }
pub=${1:?usage: install.sh <public key file>}
[[ -r $pub ]] || { echo "cannot read $pub" >&2; exit 1; }
user=${DEPLOY_USER:-trisse-deploy}
here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
target=/usr/local/bin/trisse-home-deploy

key=$(head -n1 "$pub")
[[ $key =~ ^(ssh-ed25519|ecdsa-sha2-nistp[0-9]+|ssh-rsa)\ [A-Za-z0-9+/=]+ ]] || { echo "$pub does not look like a public key" >&2; exit 1; }

id "$user" >/dev/null 2>&1 || useradd --system --create-home --shell /bin/bash --comment "trisse-home deploy" "$user"
home=$(getent passwd "$user" | cut -d: -f6)

install -o root -g root -m 755 "$here/trisse-home-deploy" "$target"
install -d -o root -g root -m 700 /var/backups/trisse-home /var/lib/trisse-home-deploy

sudoers=$(mktemp)
cat >"$sudoers" <<SUDOERS
# trisse-home auto deploy (#211): this user may run exactly one script.
Defaults:$user env_keep += "SSH_ORIGINAL_COMMAND"
$user ALL=(root) NOPASSWD: $target
SUDOERS
visudo -cf "$sudoers" >/dev/null
install -o root -g root -m 440 "$sudoers" /etc/sudoers.d/trisse-deploy
rm -f "$sudoers"

install -d -o "$user" -g "$user" -m 700 "$home/.ssh"
printf 'command="sudo %s",no-pty,no-port-forwarding,no-agent-forwarding,no-X11-forwarding,no-user-rc %s\n' "$target" "$key" >"$home/.ssh/authorized_keys"
chown "$user:$user" "$home/.ssh/authorized_keys"
chmod 600 "$home/.ssh/authorized_keys"

echo "installed: $target, /etc/sudoers.d/trisse-deploy, $home/.ssh/authorized_keys"
echo "host key for DEPLOY_KNOWN_HOSTS (replace HOST, use [HOST]:PORT if the port is not 22):"
awk -v h=HOST '{print h, $1, $2}' /etc/ssh/ssh_host_ed25519_key.pub
