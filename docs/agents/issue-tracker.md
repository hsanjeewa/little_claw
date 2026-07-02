# Issue tracker

This repo uses **GitHub Issues** as its issue tracker.

- Repository: `hsanjeewa/little_claw`
- Remote: `https://github.com/hsanjeewa/little_claw.git`
- Issues: `https://github.com/hsanjeewa/little_claw/issues`

## Creating issues

Use the `gh` CLI:

```bash
gh issue create --title "..." --body "..."
```

## Pull requests as a request surface

Yes. External pull requests are treated as a triage surface. `/triage` will include external PRs in the same queue as issues and apply the same labels/states. In-flight PRs from collaborators are left alone.
