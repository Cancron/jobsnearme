// Server-rendered fragment for the internal Scalar API reference — the session-cookie
// counterpart of ../+page.server.ts. Same reasoning throughout (dynamic import to dodge
// SvelteKit's build-time route analysis touching @scalar/server-side-rendering's
// worker_threads dependency chain; memoized per-process render), against the internal
// spec instead of the external one, so a request for either page never renders the
// other's content.
import spec from '$lib/docs/generated/api-reference.internal.openapi.json' with { type: 'json' };
import { scalarConfigFromContent } from '$lib/docs/scalarConfig';
import type { PageServerLoad } from './$types';

let cachedScalarHtml: Promise<string> | undefined;

export const load: PageServerLoad = async () => {
  cachedScalarHtml ??= (async () => {
    const { renderApiReferenceToString } = await import('@scalar/server-side-rendering');
    return renderApiReferenceToString(scalarConfigFromContent(spec));
  })();
  return { scalarHtml: await cachedScalarHtml };
};
