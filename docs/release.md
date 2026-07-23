# Release

Agent-compose releases are Forgejo-canonical and single-stage. Every push to
`main` enters a no-cancel queue, validates the exact commit, bumps the minor
version, and publishes that commit before the next queued push begins.

The workflow:

* runs `ward exec test`
* creates the next semver tag with the shared AOS tag action
* cross-compiles macOS, Linux, and Windows binaries
* creates the Forgejo release and uploads binaries, checksums, and package files
* updates the Homebrew tap and Scoop bucket when their write tokens are present

The validation gate restores pre-commit environments by config hash before it
runs. A cold cache keeps the AOS-standard 30-minute window, and a successful
run saves the environments for the next queued release.

There is no staging environment, production environment, promote branch, or
draft-release tier. A manual dispatch can retry an explicit tag. Major releases
remain a deliberate dispatch choice rather than a commit-message inference.

## See also

* [../README.md](../README.md) - installation and product status.
* [FEATURES.md](FEATURES.md) - shipped capability inventory.
* [../.ward/ward.yaml](../.ward/ward.yaml) - local release build verbs.
