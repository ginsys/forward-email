---
name: Bug report
about: Create a bug report to help improve the Forward Email CLI
title: "bug: "
labels: [bug, needs-triage]
assignees: []
---

## Description
A clear description of the problem.

## Steps To Reproduce
Provide exact commands and inputs.
```bash
# example
forward-email domain list --profile=default --output=json
```

## Expected Behavior
What you expected to happen.

## Actual Behavior
What actually happened, including errors and output.

## Environment
- CLI version: `forward-email --version`
- OS/Arch: `uname -a`
- Profile/Config details (if relevant): path `~/.config/forwardemail/config.yaml`

## Diagnostics
Include any helpful logs or debug outputs.
```bash
# optional diagnostics
forward-email debug auth [profile]
forward-email debug keys [profile]
```

## Additional Context
Screenshots, links, or any other context.
