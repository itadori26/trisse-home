# Auto deploy

Every merge to `main` whose CI run is green updates the server (#211). Until a
release image exists on GHCR, the server builds trisse-home from source; this
automates exactly that, with a backup before and a rollback after.

## How it works

1. The workflow `.github/workflows/deploy.yml` starts when the **CI** run of a
   push to `main` succeeded. It never runs for pull requests and does nothing
   until the repository variable `DEPLOY_ENABLED` is `true`.
2. It connects to the server over SSH with a key that may run **one script and
   nothing else** (SSH forced command) and sends `deploy <commit>`.
3. `scripts/deploy/trisse-home-deploy` runs as root on the server:
   - checks the request (exactly `deploy` plus 40 hex characters, and the commit
     must be part of `origin/main`), takes a lock so only one deploy runs, and
     stops if the stack checkout has local changes;
   - builds the image from a separate worktree **while the old version keeps
     running**: a failing build changes nothing;
   - stops the service, copies the data volume to
     `/var/backups/trisse-home/<time>-<commit>` (the seven newest are kept),
     switches the stack checkout `/opt/stacks/trisse-home` to the commit and
     starts the new image;
   - waits until the container is `healthy` (up to two minutes). If it is not,
     it **restores the data from the backup, the previous image and the previous
     commit** and exits with an error, which turns the GitHub job red. Restoring
     the data matters because migrations may already have run.

The downtime is the time from stop to healthy, usually well under a minute.

## Security

- The key can only trigger a deploy of a commit that is already on `main`
  (protected, reviewed, merged by you). It cannot open a shell, forward ports or
  pass arguments to anything.
- The script and its `sudo` rule are owned by root. Changing the script in the
  repository does not change what runs on the server until it is installed again
  (see below).
- The private key lives only in the GitHub secret `DEPLOY_SSH_KEY`; the host key
  is pinned in `DEPLOY_KNOWN_HOSTS` (no trust on first use).

## One-time setup

On the server (as an administrator):

```sh
ssh-keygen -t ed25519 -N '' -C github-actions-trisse-home -f /tmp/trisse-deploy-key
sudo bash scripts/deploy/install.sh /tmp/trisse-deploy-key.pub
```

The installer creates the user `trisse-deploy`, installs the script and the
`sudo` rule and writes the key with its forced command. It prints the host key
line for `DEPLOY_KNOWN_HOSTS`.

In GitHub (Settings → Secrets and variables → Actions), with the GitHub CLI:

```sh
gh secret set DEPLOY_SSH_KEY < /tmp/trisse-deploy-key      # then delete the file
gh secret set DEPLOY_KNOWN_HOSTS                            # the line printed by the installer, HOST replaced
gh variable set DEPLOY_HOST --body your.server.example      # name or address GitHub can reach
gh variable set DEPLOY_ENABLED --body true                  # last: switches the workflow on
# only if different: DEPLOY_PORT (default 22), DEPLOY_USER (default trisse-deploy)
rm /tmp/trisse-deploy-key /tmp/trisse-deploy-key.pub
```

SSH must be reachable from the internet on that host and port. The `.env` of the
stack stays on the server and is never touched; a new environment variable that
a release needs must be added there by hand.

## Operating

- **Deploy by hand:** `ssh trisse-deploy@HOST "deploy <40 hex commit>"` with the
  key, or re-run the workflow in GitHub.
- **Updated script:** after changing `scripts/deploy/trisse-home-deploy`, run
  `sudo bash scripts/deploy/install.sh <key.pub>` on the server again.
- **Restore a backup** (service stopped): copy the directory back into the volume,
  `/var/lib/docker/volumes/trisse-home-data/_data`.
- **Pause:** set the variable `DEPLOY_ENABLED` to anything but `true`.
- **Tests:** `bash scripts/deploy/deploy_test.sh` (stubbed git and docker; covers
  success, nothing to do, build failure, rollback, rejected requests and the lock).
  CI runs it.
