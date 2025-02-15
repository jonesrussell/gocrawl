# TODO List for Codebase Improvements

## Lint Errors to Address

### 1. Error Return Value Not Checked (errcheck)
- [ ] Check and handle error return values in the following files:
  - `internal/storage/mock_storage.go`
  - `internal/storage/scroll.go`
  - `cmd/crawl.go`

### 2. Unused Parameters (revive)
- [ ] Remove or rename unused parameters in the following files:
  - `internal/app/app.go`
  - `cmd/crawl.go`
  - `internal/crawler/crawler_test.go`

### 3. Magic Numbers (mnd)
- [ ] Replace magic numbers with named constants in the following files:
  - `internal/storage/storage.go`
  - `cmd/crawl.go`

### 4. Empty Blocks (revive)
- [ ] Remove or comment on empty blocks in the following files:
  - `internal/storage/config.go`

### 5. Shadowing Variables (govet)
- [ ] Rename variables to avoid shadowing in the following files:
  - `internal/storage/article.go`
  - `internal/storage/storage.go`

### 6. Global Variables (gochecknoglobals)
- [ ] Refactor global variables in the following files:
  - `cmd/root.go`
  - `internal/logger/module.go`

### 7. Use of `fmt.Println` and `fmt.Printf` (forbidigo)
- [ ] Replace `fmt.Println` and `fmt.Printf` with logger statements in the following files:
  - `internal/app/app.go`
  - `cmd/root.go`

### 8. Package Naming (testpackage)
- [ ] Rename test files to follow the convention of `*_test.go`:
  - Change `package logger` to `package logger_test` in relevant test files.

### 9. Indentation Errors (revive)
- [ ] Fix indentation issues in the following files:
  - `cmd/root.go`

### 10. Superfluous Else (revive)
- [ ] Remove unnecessary `else` statements in the following files:
  - Various files as indicated by the linter.

## General Code Quality Improvements
- [ ] Review and refactor code for readability and maintainability.
- [ ] Ensure consistent logging practices across the codebase.
- [ ] Document any significant changes made during the refactoring process.

## Notes
- After addressing the above tasks, run the linter again to ensure all issues are resolved.
- Consider setting up a CI/CD pipeline to automatically run lint checks on pull requests.
