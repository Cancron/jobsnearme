import { fireEvent, render, screen } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { UserProfile } from '$lib/types';
import SkillsCard from './SkillsCard.svelte';

const baseProfile: UserProfile = {
  specializations: ['backend'],
  skills: ['go', 'python'],
  seniorities: [],
  excluded_skills: [],
  excluded_sources: [],
  excluded_companies: [],
  location_preferences: null,
  derived_location: null,
  cv: null,
  created_at: null,
  updated_at: null,
};

const { addSkill, removeSkill } = vi.hoisted(() => ({
  addSkill: vi.fn(),
  removeSkill: vi.fn(),
}));

vi.mock('$lib/profile.svelte', () => ({
  profileStore: {
    get profile() {
      return baseProfile;
    },
    addSkill,
    removeSkill,
  },
}));

// The dictionary fetch is real network (api.facetCounts) in production; stubbed here
// so the picker's mount effect resolves instantly with no candidates to search.
vi.mock('$lib/skillDictionary', () => ({
  loadSkillDistribution: vi.fn().mockResolvedValue([]),
}));

beforeEach(() => {
  addSkill.mockReset().mockResolvedValue(baseProfile);
  removeSkill.mockReset().mockResolvedValue(baseProfile);
});

describe('SkillsCard', () => {
  it('notifies onProfileChanged after removing a skill succeeds', async () => {
    const onProfileChanged = vi.fn();
    render(SkillsCard, { props: { onProfileChanged } });

    // `go` is already a selected skill, so RemoteSearchSelect renders it as a chip
    // (title=labelOf(value)) regardless of the debounced search/dictionary.
    await fireEvent.click(screen.getByTitle('go'));

    expect(removeSkill).toHaveBeenCalledWith('go');
    expect(onProfileChanged).toHaveBeenCalledTimes(1);
  });

  it('does not notify onProfileChanged when the save fails', async () => {
    removeSkill.mockReset().mockRejectedValue(new Error('network error'));
    const onProfileChanged = vi.fn();
    render(SkillsCard, { props: { onProfileChanged } });

    await fireEvent.click(screen.getByTitle('go'));

    expect(onProfileChanged).not.toHaveBeenCalled();
  });

  it('still removes the skill when no onProfileChanged prop is given', async () => {
    // Names the break a careless `if (onProfileChanged) { await store.removeSkill(...);
    // onProfileChanged() }` would introduce: the core write must not be gated on the
    // optional callback being present.
    render(SkillsCard, { props: {} });

    await fireEvent.click(screen.getByTitle('go'));

    expect(removeSkill).toHaveBeenCalledWith('go');
  });
});
