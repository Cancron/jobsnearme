import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

// A SOURCE-TEXT AUDIT, the same choice jobActionStrip.test.ts and aiInterviewBadge.test.ts
// make and for the same reason: `web/` has no component-test infrastructure (no Svelte
// plugin, no DOM — see vitest.config.ts), so a mounted test cannot exist here at all. This
// pins the presence rules from freehire#2755 rather than the render output: whether the
// direct-apply shortcut appears (has a destination / has none) and where (full action row /
// compact, which has none).
const ROW = readFileSync(join(import.meta.dirname, 'JobRow.svelte'), 'utf8');

// The compact card's control inventory is a separate `{#if compact} … {:else} … {/if}`
// branch (see the component's own comment on `compact`): splitting on that boundary is what
// lets the assertions below tell "renders in the action row" apart from "renders regardless
// of compact", which a plain `ROW.includes(...)` cannot.
const compactBranch = ROW.slice(ROW.indexOf('{#if compact}'), ROW.indexOf('{:else}'));
const actionRow = ROW.slice(ROW.indexOf('{:else}'), ROW.lastIndexOf('{/if}'));
// The button's own `{#if applyJob} … {/if}` block, isolated within the action row so the
// hover-reveal and link-hygiene assertions can't pass by matching Hide's copies instead.
const applyBlock = actionRow.slice(
  actionRow.indexOf('{#if applyJob}'),
  actionRow.indexOf('{/if}', actionRow.indexOf('{#if applyJob}')),
);

describe('JobRow direct-apply button', () => {
  it('is gated on a job that actually carries an outbound url', () => {
    expect(ROW).toContain("const applyJob = $derived('url' in job ? job : null);");
    expect(applyBlock).not.toBe('');
  });

  it('never renders on the compact card', () => {
    expect(compactBranch).not.toContain('applyJob');
  });

  it('is outline, never the brand-green primary', () => {
    expect(applyBlock).toContain('variant="outline"');
  });

  it('carries the same link hygiene as the job page apply link', () => {
    expect(applyBlock).toContain('target="_blank"');
    expect(applyBlock).toContain('rel="nofollow noopener noreferrer"');
  });

  it('fires the same apply-intent event the job page fires', () => {
    expect(ROW).toContain("track('job_apply', { slug: applyJob.public_slug, source: applyJob.source })");
  });

  it('is a sibling of the card link, inside the action row', () => {
    // The action row itself sits after the closing `</a>` of the card link.
    expect(ROW.indexOf('</a>')).toBeLessThan(ROW.indexOf('{#if applyJob}'));
  });

  it('hover-reveals like Hide, and stays visible on touch', () => {
    for (const cls of [
      'opacity-0',
      'focus-visible:opacity-100',
      'group-hover:opacity-100',
      'pointer-coarse:opacity-100',
    ]) {
      expect(applyBlock).toContain(cls);
    }
  });
});
