Hi there!
Thanks for contributing to Swarm!

Before opening a Pull Request or reviewing a Pull Request, please get yourself acquainted with the following guidelines:

1. Open PRs preemptively (i.e. whenever you start working on it), especially in the case of large tasks
2. Create PRs with a healthy amount of description. A PR without a description will not be approved
3. PR names should adhere to the following standard: "(affected package(s)): description".When many top-level packages are affected, use "all" instead of individual names. Eg: "storage: fix panic on hasherstore put", "network, storage, p2p: integrate capabilities api", "all: first swap version"
4. Mark the PR with a `ready for review` label once it is. If a second(or more) round of reviews is required, remove the `ready for review` label and add a `ready for another review` label
5. Never(!) force-push a PR once reviews have started! This has the unfortunate side-effect of resolving/outdating previous comments over the PR. Rebasing while the PR is under review is to be done only if requested to do so by reviewers
6. It is up to the reviewer to resolve a comment
7. A comment does not account as a blocker to merging a PR unless explicitly mentioned in the comment: `BLOCKER: .....`
8. When reviewing - please make the comment on a line that would constitute a change when addressed - this will resolve the comments automatically once addressed
9. Once you address a review comment with a code change - no need to comment on it with "done"/"ok"/:+1: - this generates redundant email notifications (to those who have them enabled) and once you push git will automatically resolve/outdate the comment and the reviewer could then mark the comment as resolved without any further taxation of reading even more comments on the PR
10. We are all here to ship working software with a continuous attention to technical excellence and good design (https://agilemanifesto.org/principles.html); not to make each other's lives miserable. Thus, we ask you to take PR reviews as a process of reflection rather than a fencing duel(!)
11. Please reference related issues that should be closed with the most favorite keyword of your choice - https://github.com/blog/1506-closing-issues-via-pull-requests
12. Always use the `Squash` method of merging the PR to the branch
13. Always tidy up the commit message when using squash - having a very long commit message does not help anyone. Summerize the work done in bullet points and only then merge. Mention intermediate commits only if would provide value while bisecting or looking for regressions

