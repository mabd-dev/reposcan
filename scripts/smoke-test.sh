#!/usr/bin/env bash
#
# End-to-end smoke test for reposcan.
#
# Builds the binary, creates throwaway git/jj fixture repos in a temp dir, runs
# reposcan against them and asserts on the JSON output. Nothing outside the temp
# dir is touched: HOME is redirected so the real ~/.config/reposcan is untouched,
# and --no-telemetry is passed on every invocation.
#
# Usage:
#   ./scripts/smoke-test.sh                 # build from source, run all checks
#   REPOSCAN_BIN=/path/to/reposcan ./scripts/smoke-test.sh   # test a release binary
#   KEEP_FIXTURES=1 ./scripts/smoke-test.sh # leave the temp dir behind for poking
#
# Exit code 0 = all checks passed. 1 = at least one check failed.

set -uo pipefail

PASS=0
FAIL=0
SKIP=0

green() { printf '\033[32m%s\033[0m' "$1"; }
red()   { printf '\033[31m%s\033[0m' "$1"; }
yellow(){ printf '\033[33m%s\033[0m' "$1"; }

ok()   { PASS=$((PASS+1)); printf '  %s  %s\n' "$(green ' ok ')" "$1"; }
bad()  { FAIL=$((FAIL+1)); printf '  %s  %s\n' "$(red 'FAIL')" "$1"; [ $# -gt 1 ] && printf '        %s\n' "$2"; }
skip() { SKIP=$((SKIP+1)); printf '  %s  %s\n' "$(yellow 'skip')" "$1"; }
section() { printf '\n%s\n' "$1"; }

# assert_eq <label> <want> <got>
assert_eq() {
  if [ "$2" = "$3" ]; then ok "$1"; else bad "$1" "want '$2', got '$3'"; fi
}

# known_bug <label> <condition-cmd...> -- reports KNOWN if the bug still
# reproduces, FIXED if it no longer does. Never fails the run.
known_bug() {
  local label="$1"; shift
  if "$@" >/dev/null 2>&1; then
    printf '  %s %s\n' "$(yellow 'KNOWN')" "$label"
  else
    printf '  %s %s\n' "$(green 'FIXED')" "$label — bug no longer reproduces, drop it from docs/known-issues.md"
  fi
}

command -v jq >/dev/null || { echo "jq is required (brew install jq / apt install jq)"; exit 1; }
command -v git >/dev/null || { echo "git is required"; exit 1; }

HAS_JJ=0
command -v jj >/dev/null && HAS_JJ=1

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP="$(mktemp -d)"
FIX="$TMP/fix"

# Keep using the real Go caches before HOME is redirected, otherwise every run
# re-downloads the module cache into the sandbox HOME.
if command -v go >/dev/null; then
  export GOMODCACHE="$(go env GOMODCACHE)" GOCACHE="$(go env GOCACHE)"
fi

# Sandbox HOME: reposcan reads/writes ~/.config/reposcan, and we do not want the
# tester's real config, logs or colorscheme touched.
export HOME="$TMP/home"
mkdir -p "$HOME/.config/reposcan"

cleanup() {
  if [ "${KEEP_FIXTURES:-0}" = "1" ]; then
    echo; echo "fixtures kept at $TMP"
  else
    chmod -R u+rwX "$TMP" 2>/dev/null
    rm -rf "$TMP" 2>/dev/null
  fi
}
trap cleanup EXIT

# ---------------------------------------------------------------- build

section "build"
if [ -n "${REPOSCAN_BIN:-}" ]; then
  BIN="$REPOSCAN_BIN"
  [ -x "$BIN" ] && ok "using binary $BIN" || { bad "binary $BIN not executable"; exit 1; }
else
  BIN="$TMP/reposcan"
  if (cd "$REPO_ROOT" && go build -o "$BIN" . 2>"$TMP/build.err"); then
    ok "go build"
  else
    bad "go build" "$(cat "$TMP/build.err")"; exit 1
  fi
fi

if (cd "$REPO_ROOT" && go test ./... >"$TMP/test.out" 2>&1); then
  ok "go test ./..."
else
  bad "go test ./..." "$(grep -E '^(FAIL|---)' "$TMP/test.out" | head -5)"
fi

# reposcan wrapper: always isolated + no telemetry
rs() { "$BIN" --no-telemetry "$@" 2>/dev/null; }

# reset_config wipes the sandbox config.toml and lets reposcan regenerate the
# defaults. Needed between config tests: an invalid config.toml aborts the
# process before CLI flags are parsed, so a leftover bad config breaks
# everything that follows.
reset_config() { rm -f "$HOME/.config/reposcan/config.toml"; "$BIN" --no-telemetry -r "$FIX" -o none -f all >/dev/null 2>&1; }

# repos <filter> -> newline-separated repo names
repos() { rs -r "$FIX" -o json -f "$1" | jq -r '.repoStates[].repo' | sort | tr '\n' ' ' | sed 's/ $//'; }
count() { rs -r "$FIX" -o json -f "$1" | jq '.repoStates | length'; }

# ---------------------------------------------------------------- fixtures

section "fixtures"
G="git -c user.email=qa@example.com -c user.name=QA -c init.defaultBranch=main -c commit.gpgsign=false -c protocol.file.allow=always"

mkdir -p "$FIX/remotes"
seed() { # $1=name -> repo with a bare origin and one pushed commit on main
  $G init --bare -q "$FIX/remotes/$1.git"
  $G init -q "$FIX/$1"
  echo hello > "$FIX/$1/README.md"
  $G -C "$FIX/$1" add -A
  $G -C "$FIX/$1" commit -qm init
  $G -C "$FIX/$1" remote add origin "$FIX/remotes/$1.git"
  $G -C "$FIX/$1" push -q -u origin main
}

seed clean                                                   # synced, nothing to do

seed uncommitted                                             # 1 modified + 1 untracked
echo changed >> "$FIX/uncommitted/README.md"
echo new > "$FIX/uncommitted/untracked.txt"

seed ahead                                                   # 1 commit not pushed
echo local > "$FIX/ahead/local.txt"
$G -C "$FIX/ahead" add -A
$G -C "$FIX/ahead" commit -qm "unpushed commit"

seed behind                                                  # 1 commit on origin, already fetched
$G clone -q "$FIX/remotes/behind.git" "$TMP/pusher"
echo remote > "$TMP/pusher/remote.txt"
$G -C "$TMP/pusher" add -A
$G -C "$TMP/pusher" commit -qm "remote commit"
$G -C "$TMP/pusher" push -q origin main
rm -rf "$TMP/pusher"
$G -C "$FIX/behind" fetch -q origin

seed stashed                                                 # clean tree, 1 stash entry
echo wip >> "$FIX/stashed/README.md"
$G -C "$FIX/stashed" stash push -q -m wip

$G init -q "$FIX/noremote"                                   # no remote configured
echo x > "$FIX/noremote/a.txt"
$G -C "$FIX/noremote" add -A
$G -C "$FIX/noremote" commit -qm init

seed wt-main                                                 # linked worktree -> .git file
$G -C "$FIX/wt-main" worktree add -q -b feature "$FIX/wt-linked"
echo "wt change" >> "$FIX/wt-linked/README.md"

mkdir -p "$FIX/nested/node_modules"                          # must be skipped by default dirIgnore
$G init -q "$FIX/nested/node_modules/ignored"
echo z > "$FIX/nested/node_modules/ignored/z.txt"
$G -C "$FIX/nested/node_modules/ignored" add -A
$G -C "$FIX/nested/node_modules/ignored" commit -qm init

mkdir -p "$FIX/plain"                                        # not a repo at all
echo nothing > "$FIX/plain/file.txt"

EXPECT_ALL="ahead behind clean noremote stashed uncommitted wt-main wt-main"
EXPECT_DIRTY="ahead behind uncommitted wt-main"
EXPECT_UNCOMMITTED="uncommitted wt-main"
N_ALL=8; N_DIRTY=4; N_UNCOMMITTED=2

if [ "$HAS_JJ" = "1" ]; then
  jj git init "$FIX/jjrepo" >/dev/null 2>&1
  echo "jj file" > "$FIX/jjrepo/file.txt"
  EXPECT_ALL="ahead behind clean jjrepo noremote stashed uncommitted wt-main wt-main"
  EXPECT_DIRTY="ahead behind jjrepo uncommitted wt-main"
  EXPECT_UNCOMMITTED="jjrepo uncommitted wt-main"
  N_ALL=9; N_DIRTY=5; N_UNCOMMITTED=3
  ok "fixtures created (git + jj)"
else
  skip "jj fixtures (jj binary not installed)"
  ok "fixtures created (git only)"
fi

# ---------------------------------------------------------------- discovery

section "discovery"
assert_eq "filter=all finds every repo"            "$EXPECT_ALL" "$(repos all)"
assert_eq "totalScannedRepos matches"              "$N_ALL"      "$(rs -r "$FIX" -o json -f all | jq '.totalScannedRepos')"
assert_eq "node_modules repo ignored by default"   "0"           "$(rs -r "$FIX" -o json -f all | jq '[.repoStates[] | select(.repo=="ignored")] | length')"
assert_eq "linked worktree discovered (.git file)" "feature"     "$(rs -r "$FIX" -o json -f all | jq -r '.repoStates[] | select(.path|endswith("/wt-linked")) | .branch')"
assert_eq "no duplicates when roots overlap"       "$N_ALL"      "$(rs -r "$FIX" -r "$FIX/ahead" -o json -f all | jq '.repoStates | map(.path) | unique | length')"
assert_eq "multiple roots scan both"               "ahead behind" "$(rs -r "$FIX/ahead" -r "$FIX/behind" -o json -f all | jq -r '.repoStates[].repo' | sort | tr '\n' ' ' | sed 's/ $//')"
assert_eq "custom --dirIgnore excludes matches"    "0"           "$(rs -r "$FIX" -o json -f all -d '**/wt-*/**' | jq '[.repoStates[] | select(.repo=="wt-main")] | length')"

# ---------------------------------------------------------------- filters

section "filters"
assert_eq "filter=dirty"       "$EXPECT_DIRTY"       "$(repos dirty)"
assert_eq "filter=uncommitted" "$EXPECT_UNCOMMITTED" "$(repos uncommitted)"
assert_eq "filter=unpushed"    "ahead"               "$(repos unpushed)"
assert_eq "filter=unpulled"    "behind"              "$(repos unpulled)"
assert_eq "filter=stash"       "stashed"             "$(repos stash)"

# ---------------------------------------------------------------- repo state detail

section "repo state"
assert_eq "uncommitted lists modified + untracked" "2" \
  "$(rs -r "$FIX" -o json -f all | jq '[.repoStates[] | select(.repo=="uncommitted")][0].uncommitedFiles | length')"
assert_eq "ahead count is 1"    "1" "$(rs -r "$FIX" -o json -f all | jq '[.repoStates[] | select(.repo=="ahead")][0].remoteStatus[0].ahead')"
assert_eq "behind count is 1"   "1" "$(rs -r "$FIX" -o json -f all | jq '[.repoStates[] | select(.repo=="behind")][0].remoteStatus[0].behind')"
assert_eq "outgoing commit subject captured" "unpushed commit" \
  "$(rs -r "$FIX" -o json -f all | jq -r '[.repoStates[] | select(.repo=="ahead")][0].remoteStatus[0].outgoingCommits[0]' | cut -d' ' -f2-)"
assert_eq "stash entry captured" "1" \
  "$(rs -r "$FIX" -o json -f all | jq '[.repoStates[] | select(.repo=="stashed")][0].stashes | length')"
assert_eq "git repos tagged vcsType=git" "git" \
  "$(rs -r "$FIX" -o json -f all | jq -r '[.repoStates[] | select(.repo=="clean")][0].vcsType')"
if [ "$HAS_JJ" = "1" ]; then
  assert_eq "jj repo tagged vcsType=jj" "jj" \
    "$(rs -r "$FIX" -o json -f all | jq -r '[.repoStates[] | select(.repo=="jjrepo")][0].vcsType')"
  assert_eq "jj working-copy change detected" "1" \
    "$(rs -r "$FIX" -o json -f all | jq '[.repoStates[] | select(.repo=="jjrepo")][0].uncommitedFiles | length')"
else
  skip "jj state checks"
fi

# ---------------------------------------------------------------- output formats

section "output"
out="$(rs -r "$FIX" -o none -f all)"
assert_eq "--output none prints nothing" "0" "${#out}"
rs -r "$FIX" -o json -f all | jq -e . >/dev/null 2>&1 && ok "--output json is valid JSON" || bad "--output json is valid JSON"

OUTDIR="$TMP/reports/nested/deep"
rs -r "$FIX" -o none -f all --json-output-path "$OUTDIR" >/dev/null
reportfile="$(find "$OUTDIR" -name '*.json' 2>/dev/null | head -1)"
if [ -n "$reportfile" ]; then
  ok "--json-output-path creates missing directories"
  assert_eq "report file has the same repos" "$N_ALL" "$(jq '.repoStates | length' "$reportfile")"
else
  bad "--json-output-path creates missing directories" "no .json found under $OUTDIR"
fi

rm -rf "$HOME/.config/reposcan/logs"
rs -r "$FIX" -o none -f all --debug >/dev/null
[ -n "$(ls "$HOME/.config/reposcan/logs"/*.log 2>/dev/null)" ] \
  && ok "--debug writes a log file" || bad "--debug writes a log file"

# ---------------------------------------------------------------- config file

section "config"
CFG="$HOME/.config/reposcan/config.toml"
rm -f "$CFG"
rs -r "$FIX" -o none -f all >/dev/null
[ -f "$CFG" ] && ok "missing config.toml is created with defaults" || bad "missing config.toml is created with defaults"

cat > "$CFG" <<EOF
version = 1
roots = ["$FIX"]
only = "stash"
countStashAsDirty = false
maxWorkers = 8
no-telemetry = true
dirignore = ["**/node_modules/**"]
[output]
type = "json"
EOF
assert_eq "config roots + only used when no flags" "stashed" \
  "$("$BIN" 2>/dev/null | jq -r '.repoStates[].repo')"
assert_eq "--filter flag overrides config only"    "ahead" \
  "$("$BIN" -f unpushed 2>/dev/null | jq -r '.repoStates[].repo')"

sed -i.bak 's/countStashAsDirty = false/countStashAsDirty = true/' "$CFG"
assert_eq "countStashAsDirty=true adds stash-only repos to dirty" "$((N_DIRTY+1))" \
  "$("$BIN" -f dirty 2>/dev/null | jq '.repoStates | length')"
sed -i.bak 's/countStashAsDirty = true/countStashAsDirty = false/' "$CFG"

assert_eq '$VAR expansion in config roots' "$N_ALL" \
  "$(FIXROOT="$FIX" sh -c "printf 'version=1\nroots=[\"\$FIXROOT\"]\nonly=\"all\"\nno-telemetry=true\ndirignore=[\"**/node_modules/**\"]\n[output]\ntype=\"json\"\n' > '$CFG'"; \
     FIXROOT="$FIX" "$BIN" 2>/dev/null | jq '.repoStates | length')"

reset_config

# ---------------------------------------------------------------- errors

section "error handling"
"$BIN" --no-telemetry -r "$FIX" -o json -f bogus  >/dev/null 2>&1; assert_eq "invalid --filter exits 1"      "1" "$?"
"$BIN" --no-telemetry -r "$FIX" -o bogus          >/dev/null 2>&1; assert_eq "invalid --output exits 1"      "1" "$?"
"$BIN" --no-telemetry -r "$TMP/nope" -o json      >/dev/null 2>&1; assert_eq "nonexistent root exits 1"      "1" "$?"
"$BIN" --no-telemetry --bogus-flag                >/dev/null 2>&1; assert_eq "unknown flag exits 1"          "1" "$?"
"$BIN" --no-telemetry bogus-subcommand            >/dev/null 2>&1; assert_eq "unknown subcommand exits 1"    "1" "$?"

mkdir -p "$TMP/perm/locked"; chmod 000 "$TMP/perm/locked"
assert_eq "unreadable dir is a warning, not a crash" "1" \
  "$(rs -r "$TMP/perm" -o json -f all | jq '[.warnings[] | select(contains("permission denied"))] | length')"
chmod 755 "$TMP/perm/locked"

mkdir -p "$TMP/loop/a"; ln -sfn "$TMP/loop" "$TMP/loop/a/back"
rs -r "$TMP/loop" -o json -f all >/dev/null 2>&1
assert_eq "symlink cycle does not hang or crash" "0" "$?"

mkdir -p "$TMP/emptyroot"
assert_eq "root with no repos yields empty report" "0" \
  "$(rs -r "$TMP/emptyroot" -o json -f all | jq '.repoStates | length')"

# ---------------------------------------------------------------- subcommands

section "subcommands"
"$BIN" version 2>/dev/null | grep -q '^reposcan v' && ok "version prints reposcan vX.Y.Z" || bad "version prints reposcan vX.Y.Z"
"$BIN" --help  2>/dev/null | grep -q 'Available Commands' && ok "--help renders" || bad "--help renders"
"$BIN" update --help 2>/dev/null | grep -q '\--alias' && ok "update --help shows --alias" || bad "update --help shows --alias"

# ---------------------------------------------------------------- known issues
# Bugs found during release testing that are not fixed yet, written up in
# docs/known-issues.md. These never fail the run: KNOWN = still reproduces,
# FIXED = gone, so drop the entry there and delete the check here.

section "known issues (informational)"

field() { # field <repo-name> <jq-path-after-the-selector>
  rs -r "$FIX" -o json -f all | jq -r "[.repoStates[] | select(.repo==\"$1\")][0]$2"
}

bug_tilde_root() {
  printf 'version = 1\nroots = ["~"]\nonly = "all"\nno-telemetry = true\n[output]\ntype = "json"\n' > "$CFG"
  ! "$BIN" >/dev/null 2>&1
}
bug_partial_config() {
  printf 'roots = ["%s"]\n' "$FIX" > "$CFG"
  ! "$BIN" -o json -f all >/dev/null 2>&1
}
bug_malformed_config() {
  printf 'not = valid toml [[[\n' > "$CFG"
  ! "$BIN" -r "$FIX" -o json -f all >/dev/null 2>&1
}
bug_no_upstream_minus_one() { [ "$(field noremote .remoteStatus[0].ahead)" = "-1" ]; }
bug_worktree_repo_name() {
  local name
  name=$(rs -r "$FIX" -o json -f all | jq -r --arg p /wt-linked '[.repoStates[] | select(.path|endswith($p))][0].repo')
  [ "$name" = "wt-main" ]
}
bug_jj_relative_paths()     { field jjrepo .uncommitedFiles[0] | grep -q '\.\./\.\./'; }
bug_jj_null_stashes()       { [ "$(field jjrepo .stashes)" = "null" ]; }

reset_config
known_bug "tilde (~) in config roots is not expanded"                        bug_tilde_root
known_bug "partial config.toml loses defaults instead of merging them"       bug_partial_config
known_bug "malformed config.toml is swallowed, shown as empty-field errors"  bug_malformed_config

reset_config
known_bug "repo with no upstream reports ahead/behind as -1"                 bug_no_upstream_minus_one
known_bug "linked worktree row is named after the main repo, not its dir"    bug_worktree_repo_name
if [ "$HAS_JJ" = "1" ]; then
  known_bug "jj uncommitted paths rendered relative to cwd (../../.. chain)" bug_jj_relative_paths
  known_bug "jj repos emit stashes:null while git emits stashes:[]"          bug_jj_null_stashes
fi

reset_config

# ---------------------------------------------------------------- summary

section "summary"
printf '  %d passed, %d failed, %d skipped\n\n' "$PASS" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
