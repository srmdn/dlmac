# AGENTS.md - dlmac

## Project identity

`dlmac` is a macOS-only command-line wrapper for `yt-dlp` and `ffmpeg`.
It downloads permitted online media, extracts audio, and keeps output in
`./downloads/`.

The current product is a small CLI. A local web interface may be added later,
but the CLI stays the core engine unless the maintainer changes that direction.

## Read before acting

Before changing behavior, structure, dependencies, or documentation, read:

1. `README.md`
2. `SPEC.md`
3. `docs/roadmap.md`
4. `docs/workflows.md`

Private notes, handoffs, local agent state, and environment-specific material
belong in `.local/` or `.sisyphus/`. These directories must stay ignored and
must not be staged.

## Non-negotiable rules

- Use RTK for every shell command.
- Keep changes small, verifiable, and dependency-light.
- Do not add dependencies without explicit maintainer approval.
- Preserve macOS compatibility.
- Keep `README.md` aligned with actual behavior.
- Validate inputs and quote shell variables safely.
- Do not add features outside `SPEC.md` unless the maintainer approves a scope
  update first.
- Do not commit downloaded media, generated output, local handoffs, secrets, or
  machine-specific state.

## Safety and legal boundaries

This tool is for local use with content the user owns, has permission to
download, or that is legally available.

Do not add support for:

- DRM bypass.
- Paywall bypass.
- Private or restricted content bypass.
- Login, cookie, browser-session, or credential workflows.
- Platform rate-limit bypass.
- Features that encourage copyright infringement.

Pass public URLs to `yt-dlp`; do not build access-circumvention flows.

## Architecture boundaries

- Keep the current CLI implementation in `dlmac`.
- Keep installation logic in `install.sh`.
- Keep public project documentation in root files and `docs/`.
- Keep private drafts, handoffs, and machine output in `.local/`.
- Keep local agent continuation state in `.sisyphus/`.
- Keep downloaded media and conversion output in `downloads/`.

For the planned transcript feature, prefer public YouTube captions exposed by
`yt-dlp`. Do not add speech-to-text, AI summarization, accounts, queues, or a
web server until the maintainer approves that scope.

## Runtime and verification

Run the relevant checks before delivery:

```bash
rtk bash -n dlmac install.sh
rtk shellcheck dlmac install.sh
rtk ./dlmac --help
```

For command behavior changes, also run the relevant manual command from
`SPEC.md` or `docs/workflows.md`.

Review the full diff before committing or reporting completion:

```bash
rtk git diff
rtk git status --short --ignored
```

## Maintainer shorthand

When the maintainer says "lanjutkan", "oke lanjutkan", "gas", or a similar
short continuation prompt, continue the next safe step in `docs/workflows.md`.

Treat the prompt as permission to proceed within the active workflow when the
next step is already implied by the current state. This can include local
verification, staging, committing, pushing the current branch, creating or
updating a PR, and reporting the result.

Pause and ask for explicit confirmation before destructive or high-risk actions:

- Force push.
- Reset, rebase, or history rewrite.
- Delete unmerged branches.
- Delete another maintainer's branch.
- Publish a release or deployment.
- Add a new dependency.
- Expand scope beyond `SPEC.md`.

## Git and delivery

- Keep `main` stable and releasable.
- Use one short-lived `codex/<scope>` branch for non-trivial changes.
- Direct commits to `main` are allowed only for tiny verified mechanical fixes.
- Keep commits small, traceable, and single-scope.
- Push, PR creation, merge, release tags, and deployment require explicit
  maintainer authorization.
- PRs must include a summary and verification commands.
- Delete short-lived branches after their PRs are merged, unless the branch is
  intentionally kept for an active experiment, staged rollout, or long-running
  work line.
- Do not delete another maintainer's branch without explicit approval.
- Prune stale local branches only after confirming their work is merged or no
  longer needed.

## Recovery prompts

If the work starts to over-expand, use:

> Stop. Return to `SPEC.md`. Remove anything not approved there.

If new dependencies appear without approval, use:

> Stop. Allowed dependencies must be approved first. Remove the new dependency.
