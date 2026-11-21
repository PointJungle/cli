# Contributing

Thank you for helping improve codeforge's CLI and SDKs.

## Issues

Bugs, questions and ideas are welcome in [GitHub issues](https://github.com/pointjungle/codeforge-sdk/issues). For a bug, include the command, `codeforge version`, your system, and the output with `--json` if it helps. Never paste tokens or the content of `tokens.json`.

For a security issue, do not open a public issue: write to the maintainers through the contact given on [codeforge.pointjungle.com](https://codeforge.pointjungle.com).

## Pull requests

This repository is published from PointJungle's internal repository: each commit here is a sync of changes made there. Pull requests are welcome all the same. A maintainer reviews yours here, applies it internally with you as co-author, and it comes back in the next sync; your pull request is then closed with a link to that commit.

Before you open one:

- open an issue first for anything larger than a fix, so we can agree on the approach;
- follow [docs/cli/development/](docs/cli/development/README.md) — checks, layout and the rules for new commands;
- write commit messages as [Conventional Commits](https://www.conventionalcommits.org) (`feat(cli): …`, `fix(cli): …`): release notes are generated from them.

By contributing, you agree that your contribution is licensed under the [MIT License](LICENSE).
