#!/usr/bin/env bash
# Tests scripts/deploy/trisse-home-deploy with stubbed git, docker and runuser (#211).
# The stubs model what matters: which image is "running", what the "database"
# contains (a new version migrates it) and which commit the checkout is on.
#   bash scripts/deploy/deploy_test.sh
set -uo pipefail

SCRIPT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/trisse-home-deploy"
ROOT=$(mktemp -d)
trap 'rm -rf "$ROOT"' EXIT
fails=0

PREV=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
MAIN=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
OLDER=cccccccccccccccccccccccccccccccccccccccc

make_stubs() {
	mkdir -p "$ROOT/bin"
	cat >"$ROOT/bin/runuser" <<'STUB'
#!/usr/bin/env bash
shift 2 # -u <user>
[[ $1 == -- ]] && shift
exec "$@"
STUB
	cat >"$ROOT/bin/git" <<'STUB'
#!/usr/bin/env bash
[[ $1 == -C ]] && shift 2
cmd=$1; shift
echo "git $cmd $*" >>"$T/calls.log"
case $cmd in
status) [[ -n ${DIRTY:-} ]] && echo " M file" ;;
fetch) ;;
rev-parse) case $1 in origin/main) echo "$MAIN" ;; HEAD) cat "$T/head" ;; esac ;;
cat-file) [[ -z ${UNKNOWN:-} ]] ;;
merge-base) [[ -z ${NOT_ANCESTOR:-} ]] ;;
worktree) if [[ $1 == remove ]]; then rm -rf "${*: -1}"; fi ;;
checkout) echo "${*: -1}" >"$T/head" ;;
esac
STUB
	cat >"$ROOT/bin/docker" <<'STUB'
#!/usr/bin/env bash
echo "docker $*" >>"$T/calls.log"
case $1 in
build)
	[[ -n ${FAIL_BUILD:-} ]] && exit 1
	echo candidate >"$T/img/candidate"
	;;
tag)
	src=${2#*:}; dst=${3#*:}
	[[ -f $T/img/$src ]] || exit 1
	cp "$T/img/$src" "$T/img/$dst"
	;;
rmi | image) ;;
volume) echo "$T/volume" ;;
inspect)
	run=$(cat "$T/running")
	if [[ $run == stopped ]]; then echo "exited "; exit 0; fi
	if [[ $run == candidate && -n ${NEW_UNHEALTHY:-} ]]; then echo "running starting"; else echo "running healthy"; fi
	;;
compose)
	shift
	while [[ $1 == -f ]]; do shift 2; done
	case $1 in
	stop) echo stopped >"$T/running" ;;
	up)
		img=$(cat "$T/img/local")
		echo "$img" >"$T/running"
		[[ $img == candidate ]] && echo migrated >"$T/volume/db.sqlite"
		;;
	esac
	;;
esac
exit 0
STUB
	chmod +x "$ROOT"/bin/*
}

# setup <test name>: a fresh world with the old version running.
setup() {
	T="$ROOT/$1"
	export T
	rm -rf "$T"
	mkdir -p "$T/stack" "$T/state" "$T/backups" "$T/volume" "$T/img"
	echo "$PREV" >"$T/head"
	echo old >"$T/img/local"
	echo old >"$T/running"
	echo old-data >"$T/volume/db.sqlite"
	unset DIRTY UNKNOWN NOT_ANCESTOR FAIL_BUILD NEW_UNHEALTHY SSH_ORIGINAL_COMMAND
	export MAIN
}

deploy() {
	STACK_DIR="$T/stack" STATE_DIR="$T/state" BACKUP_DIR="$T/backups" LOCK_FILE="$T/lock" \
		KEEP_BACKUPS=2 HEALTH_TIMEOUT=1 HEALTH_INTERVAL=1 PATH="$ROOT/bin:$PATH" \
		bash "$SCRIPT" >"$T/out.log" 2>&1
	code=$?
}

check() { # check <description> <condition...>
	local desc=$1
	shift
	if "$@"; then echo "ok   $desc"; else
		echo "FAIL $desc"
		fails=$((fails + 1))
		[[ -f $T/out.log ]] && sed 's/^/     | /' "$T/out.log"
	fi
}
eq() { [[ $(cat "$1") == "$2" ]]; }
called() { grep -q -- "$1" "$T/calls.log"; }
not_called() { ! grep -q -- "$1" "$T/calls.log"; }
before() { # before <pattern a> <pattern b>: first a comes before first b
	local a b
	a=$(grep -n -m1 -- "$1" "$T/calls.log" | cut -d: -f1)
	b=$(grep -n -m1 -- "$2" "$T/calls.log" | cut -d: -f1)
	[[ -n $a && -n $b && $a -lt $b ]]
}
exit_is() { [[ $code -eq $1 ]]; }

make_stubs

echo "== success"
setup success
for old in 20250101T000000Z-aaaaaaa 20250102T000000Z-bbbbbbb 20250103T000000Z-ccccccc; do mkdir -p "$T/backups/$old"; done
deploy
check "exit 0" exit_is 0
check "the image is built before the service is stopped" before "docker build" "compose.* stop"
check "the service is stopped before the new one starts" before "stop trisse-home" "up -d --no-build"
check "the checkout moved to the new commit" eq "$T/head" "$MAIN"
check "the new version is running" eq "$T/running" candidate
check "the state file records the commit" eq "$T/state/current" "$MAIN"
check "the backup holds the data from before the update" bash -c "cat \"$T\"/backups/*-bbbbbbb/db.sqlite | grep -qx old-data"
check "only the newest backups are kept" bash -c "[[ \$(ls \"$T/backups\" | wc -l) -eq 2 ]]"
check "the temporary build worktree is removed" bash -c "! ls \"$T/state\" | grep -q '^build'"

echo "== already deployed"
setup already
echo "$MAIN" >"$T/head"
deploy
check "exit 0" exit_is 0
check "nothing is built or restarted" not_called "docker build"

echo "== explicit commit"
setup explicit
export SSH_ORIGINAL_COMMAND="deploy $OLDER"
deploy
check "exit 0" exit_is 0
check "exactly that commit is deployed, not the head of main" eq "$T/head" "$OLDER"

echo "== build failure"
setup buildfail
FAIL_BUILD=1 deploy
check "exit 6" exit_is 6
check "the running service is not stopped" not_called "stop trisse-home"
check "the checkout is unchanged" eq "$T/head" "$PREV"
check "the data is untouched" bash -c "grep -qx old-data \"$T/volume/db.sqlite\""

echo "== new version does not become healthy"
setup rollback
NEW_UNHEALTHY=1 deploy
check "exit 8" exit_is 8
check "the checkout is switched back" eq "$T/head" "$PREV"
check "the old version is running again" eq "$T/running" old
check "the old image is restored" eq "$T/img/local" old
check "the data from before the migration is restored" bash -c "grep -qx old-data \"$T/volume/db.sqlite\""
check "the failure is reported" bash -c "grep -q 'rolled back' \"$T/out.log\""

echo "== rejected requests"
for req in "rm -rf /" "deploy abc" "deploy $MAIN extra" "deploy ${MAIN^^}" "$MAIN" "deploy \$(id)"; do
	setup bad
	SSH_ORIGINAL_COMMAND="$req" deploy
	check "refused (exit 2): $req" exit_is 2
	check "  and nothing was run" bash -c "[[ ! -s \"$T/calls.log\" ]]"
done
setup unknown
UNKNOWN=1 deploy
check "unknown commit is refused (exit 5)" exit_is 5
setup notancestor
NOT_ANCESTOR=1 deploy
check "a commit outside origin/main is refused (exit 5)" exit_is 5
check "  and nothing was built" not_called "docker build"
setup dirty
DIRTY=1 deploy
check "a dirty checkout is refused (exit 4)" exit_is 4

echo "== one deploy at a time"
setup locked
(
	exec 8>"$T/lock"
	flock -n 8
	sleep 3
) &
sleep 0.5
deploy
check "a second deploy is refused (exit 3)" exit_is 3
wait

if ((fails > 0)); then
	echo "$fails check(s) failed"
	exit 1
fi
echo "all checks passed"
