# Workflows

## Local verification

Run these checks before reporting a code or documentation change complete:

```bash
rtk bash -n dlmac install.sh
rtk shellcheck dlmac install.sh
rtk ./dlmac --help
rtk ./dlmac --version
```

For behavior changes, also run the closest manual command from `SPEC.md`.

## Feature workflow

Use this workflow for non-trivial changes:

1. Read `AGENTS.md`, `SPEC.md`, `README.md`, and relevant docs.
2. Create or switch to a short-lived branch named `codex/<scope>`.
3. Update `SPEC.md` first when behavior changes.
4. Implement one feature at a time.
5. Update `README.md` and docs to match behavior.
6. Run local verification.
7. Review the full diff.
8. Commit with one clear scope.
9. Push and open a PR only when the maintainer asks for it.

## Continuation shorthand

When the maintainer says "lanjutkan", "oke lanjutkan", "gas", or a similar
short continuation prompt, continue the next safe step from the current state.

Use this order:

1. Finish any in-progress file edit.
2. Run local verification.
3. Review the diff and status.
4. Stage files that belong to the active scope.
5. Commit with a single-scope message.
6. Push the current branch.
7. Create or update the PR.
8. Report the result and the next recommended step.

Do not use shorthand permission for destructive or high-risk actions:

- Force push.
- Reset, rebase, or history rewrite.
- Delete unmerged branches.
- Delete another maintainer's branch.
- Publish a release or deployment.
- Add a new dependency.
- Expand scope beyond `SPEC.md`.

## Git branch policy

Keep `main` stable and releasable.

Use short-lived branches for normal work:

```text
codex/transcript-command
codex/repo-structure
codex/local-web-mvp
```

Delete a branch after its PR is merged when:

- The scope is complete.
- The branch was temporary.
- The merged work is available on `main`.
- No one needs the branch for follow-up work.

Keep a branch when:

- The PR is still active.
- The feature is long-running.
- The branch is an approved experiment.
- The branch is a staged rollout line.

Do not delete another maintainer's branch without explicit approval.

## Commit policy

Keep commits small and single-scope.

Good examples:

```text
Docs: define repo workflow
Add transcript command
Fix transcript language validation
```

Do not commit:

- `.local/`
- `.sisyphus/`
- `downloads/`
- `.DS_Store`
- Secrets
- Local machine state
- Generated media

## PR policy

Each PR must include:

```md
## Summary

- ...

## Verification

- [ ] `rtk bash -n dlmac install.sh`
- [ ] `rtk shellcheck dlmac install.sh`
- [ ] `rtk ./dlmac --help`
```

Merge only after verification passes and the maintainer approves the PR.

Use squash merge for short-lived feature branches unless the maintainer asks
to preserve individual commits.

## Release workflow

Use this flow for user-facing releases:

1. Update the version in `dlmac`.
2. Update `README.md` and `SPEC.md`.
3. Run verification.
4. Commit the release update.
5. Create a tag only when the maintainer explicitly asks.
6. Push the tag only when the maintainer explicitly asks.
