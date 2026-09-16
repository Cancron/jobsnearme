<script lang="ts">
  // The Avoid view: three independent exclusion lists — skills, sources, companies —
  // each autosaving straight to profileStore, mirroring how SkillsCard autosaves the
  // wanted-skills list. Unlike skills, sources and companies have no "wanted" counterpart
  // list to reconcile against, so toggling one is a plain add/remove (see
  // profileExclusions.ts) rather than the mutual-exclusion dance withAvoidedSkill does.
  import type { FacetOption } from '$lib/facets';
  import { companySearch } from '$lib/facets';
  import { loadSkillDistribution } from '$lib/skillDictionary';
  import { loadSourceDistribution } from '$lib/sourceDictionary';
  import { profileStore } from '$lib/profile.svelte';
  import RemoteSearchSelect from '../facets/RemoteSearchSelect.svelte';

  let {
    onProfileChanged,
  }: {
    /** Fired after an exclusion-list change saves — the profile-derived saved-search
     *  alert needs to stay in step with it, the same way SkillsCard's autosave does. */
    onProfileChanged?: () => void;
  } = $props();

  const excludedSkills = $derived(profileStore.profile?.excluded_skills ?? []);
  const excludedSources = $derived(profileStore.profile?.excluded_sources ?? []);
  const excludedCompanies = $derived(profileStore.profile?.excluded_companies ?? []);

  // One write in flight at a time, mirroring SkillsCard: a second toggle started before
  // the first settles would race against a profile row it hasn't seen yet.
  let pending = $state(false);
  let failed = $state<string | null>(null);

  // Builds one list's toggle handler: avoid a value not yet excluded, unavoid one already
  // there. `list` is read lazily (a closure, not a captured array) so it always sees the
  // current profile, not the value at the time the handler was built.
  function makeToggle(
    list: () => string[],
    avoid: (v: string) => Promise<unknown>,
    unavoid: (v: string) => Promise<unknown>,
  ) {
    return async (value: string) => {
      if (pending) return;
      pending = true;
      failed = null;
      try {
        await (list().includes(value) ? unavoid(value) : avoid(value));
        onProfileChanged?.();
      } catch {
        failed = value;
      } finally {
        pending = false;
      }
    };
  }

  const toggleExcludedSkill = makeToggle(
    () => excludedSkills,
    (v) => profileStore.avoidSkill(v),
    (v) => profileStore.unavoidSkill(v),
  );
  const toggleExcludedSource = makeToggle(
    () => excludedSources,
    (v) => profileStore.avoidSource(v),
    (v) => profileStore.unavoidSource(v),
  );
  const toggleExcludedCompany = makeToggle(
    () => excludedCompanies,
    (v) => profileStore.avoidCompany(v),
    (v) => profileStore.unavoidCompany(v),
  );

  let skillDist = $state.raw<FacetOption[]>([]);
  let skillDistReady = $state(false);
  $effect(() => {
    void loadSkillDistribution().then((dist) => {
      skillDist = dist;
      skillDistReady = true;
    });
  });

  let sourceDist = $state.raw<FacetOption[]>([]);
  let sourceDistReady = $state(false);
  $effect(() => {
    void loadSourceDistribution().then((dist) => {
      sourceDist = dist;
      sourceDistReady = true;
    });
  });

  // A locally-held distribution search: an empty query lists the popular first page (8),
  // a real query filters by label and returns up to 50 matches.
  function searchDist(dist: FacetOption[], query: string): Promise<FacetOption[]> {
    const q = query.trim().toLowerCase();
    const matches = q ? dist.filter((o) => o.label.toLowerCase().includes(q)) : dist;
    return Promise.resolve(matches.slice(0, q ? 50 : 8));
  }

  const searchSkills = (query: string) => searchDist(skillDist, query);
  const searchSources = (query: string) => searchDist(sourceDist, query);
</script>

<div class="flex flex-col gap-6 {pending ? 'pointer-events-none opacity-60' : ''}">
  <div class="flex flex-col gap-2">
    <div class="flex items-baseline justify-between">
      <span class="text-sm font-medium">Skills to avoid</span>
      <span class="text-xs tabular-nums text-muted-foreground">{excludedSkills.length}</span>
    </div>
    <RemoteSearchSelect
      search={searchSkills}
      include={[]}
      exclude={excludedSkills}
      placeholder="Search skills to exclude"
      onToggle={toggleExcludedSkill}
      fallbackLabel={(v) => v}
      clearOnSelect
      ready={skillDistReady}
      techIcons
    />
  </div>

  <div class="flex flex-col gap-2">
    <div class="flex items-baseline justify-between">
      <span class="text-sm font-medium">Sources to avoid</span>
      <span class="text-xs tabular-nums text-muted-foreground">{excludedSources.length}</span>
    </div>
    <RemoteSearchSelect
      search={searchSources}
      include={[]}
      exclude={excludedSources}
      placeholder="Search sources to exclude"
      onToggle={toggleExcludedSource}
      fallbackLabel={(v) => v}
      clearOnSelect
      ready={sourceDistReady}
    />
  </div>

  <div class="flex flex-col gap-2">
    <div class="flex items-baseline justify-between">
      <span class="text-sm font-medium">Companies to avoid</span>
      <span class="text-xs tabular-nums text-muted-foreground">{excludedCompanies.length}</span>
    </div>
    <RemoteSearchSelect
      search={companySearch}
      include={[]}
      exclude={excludedCompanies}
      placeholder="Search companies to exclude"
      onToggle={toggleExcludedCompany}
      fallbackLabel={(v) => v}
      clearOnSelect
    />
  </div>

  {#if failed}
    <p class="text-sm text-destructive">Could not update {failed} in your profile. Try again.</p>
  {/if}
</div>
