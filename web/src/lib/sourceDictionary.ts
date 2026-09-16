// The sources-to-avoid typeahead's candidate universe — the same `source` facet
// distribution the job-search filter panel uses, so a value picked here always matches a
// real `jobs.source`. Mirrors skillDictionary.ts's shape for the analogous source picker.

import { api } from '$lib/api';
import { dynamicOptions, type FacetOption } from '$lib/facets';

/** Fetch the live source distribution and shape it into sorted typeahead options.
 *  Best-effort: any failure (network, decode) resolves to an empty list rather than
 *  throwing, so a caller can render "nothing to suggest yet" instead of an error. */
export async function loadSourceDistribution(): Promise<FacetOption[]> {
  try {
    const counts = await api.facetCounts(new URLSearchParams(), { facets: ['source'] });
    return dynamicOptions('source', counts.facets?.source ?? {}, []);
  } catch {
    return [];
  }
}
