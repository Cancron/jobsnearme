## Context

See proposal.md for the symptom and root cause. Two mechanisms already exist and this
design reuses both rather than adding anything new:

- Meilisearch quoting: wrapping part of a `q` string in double quotes disables typo
  tolerance for those words and requires every one of them to be present. **It is not
  a contiguous-phrase match in this deployment**: the `jobs` index runs
  `ProximityPrecision: byAttribute` (`internal/search/search/client.go`, `#1637`),
  which gives Meilisearch only attribute-level, not word-level, distance data, so it
  cannot verify word adjacency. A quoted `q="founding engineer"` therefore matches a
  document containing both words anywhere in a searched field, in any order — this is
  documented and was measured, not assumed, in the sibling (not yet archived) OpenSpec
  change `search-q-field-scoping`, whose design.md also records that true contiguous
  matching would need `ProximityPrecision: byWord` plus a full catalogue reindex, with
  a measured indexing-cost regression (~1.7x slower bulk load, ~3.3x slower warm-index
  incremental push at 60k-document scale) — out of scope for that change and for this
  one.
- Field-scoped search: `q_fields` (`internal/search/search/query_params.go`,
  `QFieldsFromValues`, shipped by `search-q-field-scoping`) is already a public
  `/jobs`/`/jobs/search` param that restricts `q` to a named subset of
  `SearchableAttributes` via Meilisearch's `AttributesToSearchOn`, with no reindex.

Skill, category, and company suggestions are out of scope for the fix itself (see
proposal) because clicking them already sets an exact facet filter
(`skills=`/`category=`/`company_slug=`) against the same index their count was drawn
from — the mismatch is structural to title suggestions only, since `title` has no
facet and falls back to a free-text query (`apiSuggestions.ts`'s `facetFor` map).

## Goals / Non-Goals

**Goals:**
- Make a title suggestion's click-through result count closely approximate its
  displayed count, using only existing search mechanisms.
- Leave `cmd/build-suggestions` and the nightly dictionary build untouched — the count
  computation is not the thing being changed.
- Leave demand tracking (`search_queries`, `recordQuery`) counting a suggestion-driven
  search under the same key an equivalent typed search would use.

**Non-Goals:**
- Exact-phrase (word-adjacency) matching, or byte-exact parity between the suggestion
  count and the click-through total. Quoting a title-scoped query requires every word
  to be present, not that they are adjacent or in order — a title like "Engineer,
  Founding Team Lead" would match a quoted, title-scoped "founding engineer" query
  even though its normalised form was never counted in the "founding engineer"
  dictionary bucket. Closing the ~25x gap to a small residual is the bar; true
  adjacency matching is `search-q-field-scoping`'s explicitly deferred follow-up
  (`ProximityPrecision: byWord` + full reindex), not something to bundle into a fix
  that touches only how a suggestion is applied.
- Any change to skill/category/company suggestion behavior.
- Any change to `/jobs`'s general free-text search semantics for a query a visitor
  types directly (unscoped, un-quoted, multi-field) — only a suggestion-originated
  search changes shape.

## Decisions

**Decision: fix how a title suggestion is applied, not how its count is computed.**

Three directions were weighed:

1. **Chosen — narrow the click to match the count** (quote + `q_fields=title`).
   Cheap, reuses existing mechanisms, touches only the suggestion-application path.
2. **Rejected — widen the count to approximate the click.** Would require running a
   live (or periodically cached) full-text query per dictionary entry at build time —
   79,727 documents as of the last rebuild — turning a ~2-minute nightly job mining
   SQL aggregates into one issuing tens of thousands of search queries, for a number
   that would still drift between rebuilds exactly as it does today. Rejected on cost
   for a benefit direction 1 already delivers more cheaply.
3. **Rejected — UI-only, stop implying a promise.** Relabelling or dropping the count
   fixes the honesty problem but throws away real information the count carries
   ("this role exists in volume") that pointed 1 preserves. Reached for only if 1 had
   turned out infeasible.

**Decision: quote and scope in the frontend, strip the quoting in the backend's demand
recorder only.**

`apiSuggestions.ts` already owns turning a suggestion's parts into `plan.q` /
`plan.facets` (`applyParams`), and `browseTarget.ts` already turns a plan into `/jobs`
URL params — both already the single place this logic lives, so the title branch of
`applyParams` gains the quoting and a field-restriction, and `browseTarget.ts` threads
the restriction through as `q_fields=title`. No new API surface.

The one place this leaks is `recordQuery` (`internal/api/handler/search.go`), which
records `c.Query("q")` — the same raw string Meilisearch received — for demand
tracking. Left alone, a suggestion click for "Founding Engineer" would record the
literal string `"founding engineer"` (quotes included, since `suggest.Title` does not
treat `"` as a separator), permanently splitting demand tracking for that phrase into
a quoted bucket that never matches the dictionary's own key and never accumulates
enough count to surface. `recordQuery` therefore strips one matching pair of leading
and trailing `"` before calling `suggest.Title`, rather than teaching `Title` about
quotes: `Title` is also applied to raw mined posting titles (`cmd/build-suggestions`),
and a title that genuinely starts and ends with a quote character is a different case
that should not be silently unwrapped there.

**Decision: strip an embedded `"` in the suggestion text before wrapping it.**

A suggestion's title text is a mined, normalised catalogue title
(`suggest.build.displayTitle`), so an embedded double quote is unlikely but not
provably impossible. Wrapping `He said "wow" Engineer` naively would produce
`"He said "wow" Engineer"`, which Meilisearch would parse as more than one quoted
segment rather than one. The frontend strips any `"` from the text before wrapping it,
rather than escaping it — the same choice `check-alert-rules.py`'s comment makes
elsewhere in this codebase for a different reason (rewrite past the character, don't
escape it), and simpler than teaching the one call site Meilisearch's escaping rules
for a case that has not been observed to occur.

## Risks / Trade-offs

- **[Risk] A quoted, title-scoped query still isn't the dictionary's exact-match
  count** (see Non-Goals — no word-adjacency guarantee) → Mitigation: this closes the
  gap from ~25x to a small residual, not a guarantee of equality; no scenario in the
  spec claims exact parity, and the residual gap is explicitly named as
  `search-q-field-scoping`'s deferred follow-up, not a defect of this change.
- **[Risk] `q_fields=title` on a composed suggestion (title + company) narrows the
  title portion correctly but the URL now carries both `q_fields=title` and a
  `company_slug` facet filter — worth confirming these compose without surprising
  interaction** → Mitigation: `q_fields` only restricts which fields the free-text `q`
  matches against; it has no interaction with facet filters, which are a separate
  Meilisearch `filter` expression. Existing behavior, not new to this change.
- **[Risk] Forgetting the `recordQuery` fix silently pollutes demand tracking** →
  Mitigation: called out explicitly as its own spec scenario
  ("A suggestion-originated search still counts as ordinary demand"), so it has a
  dedicated test rather than riding along as an implementation footnote.

## Migration Plan

No data migration. No index rebuild. No deploy ordering concern — both the frontend
and backend halves of the change deploy together as an ordinary release, and neither
half is meaningful without the other (the frontend change alone would search
correctly but pollute demand tracking on every suggestion click; the backend
`recordQuery` fix alone has nothing to strip until the frontend sends quoted queries).
Rollback is a plain revert, since no persisted data shape changed.
