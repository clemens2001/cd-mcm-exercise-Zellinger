# Exercise 1: Git Basics -- PRs, Interactive Rebase & Unit Tests

**Course:** Continuous Delivery in Agile Software Development (Master)
**Points:** 24

## Learning Objectives

- Master the Pull Request workflow (branch, commit, review, merge)
- Use Git Interactive Rebase to clean up commit history
- Write meaningful unit tests in Go
- Apply Conventional Commits

## Prerequisites

- Go 1.22+ installed
- Git 2.30+
- GitHub account with access to this repository

## Project Overview

This is a **Product Catalog API** written in Go. It provides a REST API for managing products (CRUD operations). The code is structured as follows:

```
cmd/api/main.go              # Application entry point
internal/model/product.go    # Product data model + validation
internal/store/memory.go     # In-memory product store
internal/handler/handler.go  # HTTP request handlers
```

### Build & Test

```bash
go test ./...                    # Run all tests
go test -v ./internal/model/     # Run model tests with verbose output
go test -cover ./...             # Run tests with coverage
go build -o api-server ./cmd/api # Build the server binary
```

---

## Tasks

### Task 1: Repository Setup (2 Points)

1. Fork this repository into your own GitHub account. Name it `cd-mcm-exercise-[Nachname]`.
2. Clone your fork locally.
3. Verify the project builds and tests pass:
   ```bash
   go test ./...
   ```
4. **Deliverable:** Screenshot or link to your forked repository.

---

### Task 2: Unit Tests (6 Points)

The store package (`internal/store/memory_test.go`) contains incomplete tests marked with `TODO` comments. Your task:

1. **Complete `TestCreateAndGet`**: Create a product, retrieve it by ID, and verify all fields match.
2. **Add `TestUpdateProduct`**: Create a product, update it, verify the update was applied.
3. **Add `TestDeleteProduct`**: Create a product, delete it, verify it's gone (GetByID returns error).
4. **Add `TestGetByIDNotFound`**: Verify that GetByID with a non-existent ID returns `ErrNotFound`.

Requirements:
- Each test must have a clear name describing what it tests.
- Use table-driven tests for at least one test function.
- All tests must pass: `go test -v ./internal/store/`

**Deliverable:** Completed test file committed on a `feature/unit-tests` branch.

---

### Task 3: Feature Branch & Pull Request (8 Points)

1. Create a branch `feature/about-me` from `main`.
2. Create a file `about-me.md` in the repository root containing:
   - Your name and program of study
   - Your experience with Go and Git (1-2 sentences each)
   - What you expect to learn in this course
   - Why Continuous Delivery matters for agile teams (2-3 sentences)
3. Commit with a proper Conventional Commit message (e.g., `docs: add about-me with background info`).
4. Push the branch and open a **Pull Request** to `main`.
5. Request a review from a fellow student. Address at least one review comment.
6. Merge the PR after approval.

**Deliverable:** Link to the merged PR.

---

### Task 4: Interactive Rebase (8 Points)

A special branch `exercise/01-rebase-practice` contains a deliberately messy commit history. Your task is to clean it up using `git rebase -i`.

#### Setup

```bash
git checkout exercise/01-rebase-practice
git checkout -b my-rebase-practice    # Work on your own branch
git log --oneline                     # Review the commit history
```

You will see commits with these issues:
- **Typo in commit message:** "commmit G" has three m's
- **Commits that should be squashed:** Multiple small "fix" commits that belong together
- **Wrong commit order:** Some commits are in an illogical order
- **Unnecessary commit:** A debug commit that should be dropped

#### Your Tasks

Using `git rebase -i <base-commit>`:

1. **Reword** the commit with the typo ("commmit G" → "commit G")
2. **Squash** the three "fix readme" commits into a single commit
3. **Reorder** commits so that related changes are adjacent
4. **Drop** the "add debug output" commit

#### How to use Interactive Rebase

```bash
git rebase -i HEAD~10   # Rebase the last 10 commits

# In the editor, change the action keyword before each commit:
# pick   = keep the commit as-is
# reword = keep the commit but edit the message
# squash = meld into previous commit (combine messages)
# fixup  = meld into previous commit (discard this message)
# drop   = remove the commit entirely
# Reorder lines to reorder commits
```

**Deliverable:** Run `git log --oneline` after your rebase and include a screenshot showing the cleaned-up history. Push as `solution/01-rebase-[Nachname]`.

---

## Conventions

### Conventional Commits

Use these prefixes for all commit messages:

| Prefix | Usage |
|--------|-------|
| `feat:` | New feature |
| `fix:` | Bug fix |
| `docs:` | Documentation only |
| `test:` | Adding or updating tests |
| `refactor:` | Code refactoring (no behavior change) |
| `chore:` | Maintenance tasks |

### Branch Naming

- `feature/<short-description>`
- `fix/<issue-id>-<short-description>`
- `docs/<short-description>`

---

## Grading

| Task | Points |
|------|--------|
| Repository Setup | 2 |
| Unit Tests | 6 |
| Feature Branch & PR | 8 |
| Interactive Rebase | 8 |
| **Total** | **24** |
