## ADDED Requirements

### Requirement: Applying a title suggestion narrows to a quoted, title-scoped query

A title suggestion's displayed job count is mined as an exact match on the
posting's normalised title. Applying the suggestion SHALL therefore search with
its text quoted (every word required, in any order, with typo tolerance off)
and restricted to the job title field, rather than as an unscoped multi-field
query — so the count a visitor saw approximates the results the click
produces, instead of a materially larger set drawn from matches anywhere
across title, company, description, and location. This is a closer
approximation, not exact parity: quoting in this deployment requires every
word to be present rather than verifying they are adjacent (word-adjacency
verification needs a different Meilisearch proximity configuration and a full
reindex, out of scope here — see the `search-q-field-scoping` capability).

This requirement covers title suggestions only. A skill, category, or company
suggestion already applies as an exact facet filter against the same index its
count was drawn from, so its displayed count already matches what a click
produces and neither its application nor its count computation changes.

#### Scenario: A title suggestion's count approximates its click-through results

- **WHEN** a visitor applies a title suggestion
- **THEN** the resulting job search is a quoted query restricted to the title
  field, and the number of results it returns is close to the job count the
  suggestion displayed, not a materially larger set drawn from unrelated
  matches in company, description, or location

#### Scenario: A composed suggestion's title part is quoted and title-scoped

- **WHEN** a visitor applies a suggestion carrying both a title part and a
  facet part (for example, a role at a named company)
- **THEN** the title part is applied as a quoted query restricted to the title
  field, alongside the facet filter for the other part

#### Scenario: A suggestion-originated search still counts as ordinary demand

- **WHEN** a title suggestion's quoted, field-restricted search is recorded
  for the suggestion dictionary's demand tracking
- **THEN** it is recorded under the same normalised phrase a visitor typing the
  same words unquoted would produce, not as a distinct phrase carrying the
  search's own quoting or field-restriction notation

#### Scenario: A non-title suggestion's count is unaffected

- **WHEN** a visitor applies a skill, category, or company suggestion
- **THEN** it is applied as an exact facet filter, exactly as before this
  change, and its displayed count is unaffected
