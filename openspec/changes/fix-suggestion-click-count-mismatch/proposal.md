## Why

A title suggestion's job count is mined as an exact match on the posting's normalised
title (`internal/search/suggest/build.go`), but clicking the suggestion sends the
phrase as an unscoped, unquoted query against four searchable fields with default
typo tolerance (`title`, `company`, `description`, `location`). The two numbers
answer different questions, and the gap is large: "Founding Engineer" showed 708 in
the suggestion and 17,808 on the results page it led to. The count is not decoration
— it reads as a promise of what the click will show, and for every title suggestion
today it is not one.

## What Changes

- Clicking a title suggestion (or a composed suggestion whose last part is a title)
  sends its text as a quoted (every word required, no typo tolerance), title-field-
  restricted query instead of an unscoped, OR-of-tokens, typo-tolerant query against
  four fields — bringing the actual result count close to the count the suggestion
  displayed. This is not exact-phrase matching (this deployment's Meilisearch index
  cannot verify word adjacency — see `search-q-field-scoping`'s design notes); it
  requires every word to appear in the title, in any order, which is what closes
  most of the gap since the current query also matches on `company`, `description`,
  and `location`.
- `recordQuery` strips the quoting wrapper before feeding the query into demand
  tracking, so a suggestion-originated search still lands on the same normalised
  demand key as someone typing the same words unquoted — without this, every
  suggestion click would silently start a second, never-matching demand bucket for a
  phrase the dictionary already tracks.
- Skill, category, and company suggestions are unaffected: they already apply as
  exact facet filters against the same index the count was computed from, so their
  displayed count already matches what a click produces.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `search-suggestions`: applying a title suggestion narrows the resulting job search
  to a quoted, title-field-restricted query, rather than an unscoped multi-field
  query — so the suggestion's displayed count approximates what the click actually
  returns.

## Impact

- `web/src/lib/apiSuggestions.ts` (`applyParams`): a title part's `plan.q` becomes a
  quoted string, and the plan carries a title-only field restriction.
- `web/src/lib/browseTarget.ts` (`browseQuery`): passes the field restriction through
  to the `/jobs` URL as `q_fields=title` (an existing, already-shipped public param —
  see `search-q-field-scoping`).
- `internal/api/handler/search.go` (`recordQuery`): strips the quoting wrapper before
  calling `suggest.Title`, so demand tracking is unaffected by the new mechanism.
- No change to `cmd/build-suggestions`, the nightly dictionary build, or the
  `suggestions` Meilisearch index — the fix is entirely in how a title suggestion is
  applied, not in how its count is computed.
