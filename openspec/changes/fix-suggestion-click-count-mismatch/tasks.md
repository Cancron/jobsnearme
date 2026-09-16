## 1. Frontend: apply a title suggestion as a phrase, scoped to the title field

- [x] 1.1 In `web/src/lib/apiSuggestions.test.ts`, add failing tests for `applyParams`:
      a bare title part produces a quoted `plan.q` and a title-only field
      restriction; an embedded `"` in the title text is stripped before quoting; a
      composed suggestion (title + company) still applies the company facet
      alongside the quoted, field-restricted title.
- [x] 1.2 Extend the `ApplyPlan` shape in `apiSuggestions.ts` to carry the field
      restriction (e.g. `qFields?: string[]`), and update the title branch of
      `applyParams` to strip embedded `"` characters and wrap the result in quotes.
- [x] 1.3 In `web/src/lib/browseTarget.test.ts`, add a failing test asserting
      `browseQuery` threads the plan's field restriction through as `q_fields` on the
      `/jobs` target.
- [x] 1.4 Update `browseTarget.ts`'s `browseQuery` to set `q_fields` from the plan
      when present.
- [x] 1.5 Run `pnpm --filter web test` (or the project's equivalent) and confirm all
      four new/updated tests pass.

## 2. Backend: keep demand tracking unaffected by the new quoting

- [x] 2.1 In `internal/api/handler/search_test.go`, add a failing test for
      `recordQuery`: a raw query wrapped in a single matching pair of `"` is recorded
      under the same normalised key as the same words without quotes; a query with
      only one quote, or quotes not at both ends, is left untouched before
      normalisation (it is not the suggestion-click shape and should not be
      guessed at).
- [x] 2.2 Implement the quote-stripping in `recordQuery` (`internal/api/handler/
      search.go`), before the existing `suggest.Title(raw)` call — strip one matching
      leading and trailing `"` only, not every quote character in the string.
- [x] 2.3 Run `go test ./internal/api/handler/...` and confirm the new test passes
      alongside the existing suite.

## 3. Verification

- [x] 3.1 `go build ./... && go vet ./...` and `go test ./...` clean (one pre-existing,
      unrelated failure: `cmd/billing-sync`'s `TestTheStoreProviderAloneKeepsTheWorkerRunning`
      fails in this environment because an ambient local Postgres on :5432, not this
      change, answers when the test clears `DATABASE_URL` — not touched by this diff).
- [x] 3.2 `gofmt -l .` prints nothing for touched Go files.
- [ ] 3.3 (Deferred — see note) Manually verify against a live stack: type "Founding
      Engineer" (or another known-mismatched phrase) in the search box, click the
      title suggestion, and confirm the resulting `/jobs` URL carries a quoted,
      title-scoped query and that the result count is close to the suggestion's
      displayed count rather than an order of magnitude larger. Skipped in this
      session — it needs a stack with real, indexed catalogue data (`make up` +
      seeding + a `cmd/build-suggestions` run), which is disproportionate
      infrastructure to stand up for a fix already covered end-to-end by unit tests.
      Do this once against staging or after deploy, not by standing up a local stack
      just for it.
- [x] 3.4 Confirm a suggestion click's search still appears under its plain
      (unquoted) form in `search_queries` rather than under a quoted variant —
      proven deterministically by `TestDemandKey_StripsMatchingQuotePairBeforeNormalising`
      (2.1); no live-DB spot-check needed.
