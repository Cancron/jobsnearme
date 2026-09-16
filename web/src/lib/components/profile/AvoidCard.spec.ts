import { fireEvent, render, screen } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { UserProfile } from '$lib/types';
import AvoidCard from './AvoidCard.svelte';

const baseProfile: UserProfile = {
  specializations: ['backend'],
  skills: ['go', 'python'],
  seniorities: [],
  excluded_skills: ['java'],
  excluded_sources: ['greenhouse'],
  excluded_companies: ['acme'],
  location_preferences: null,
  derived_location: null,
  cv: null,
  created_at: null,
  updated_at: null,
};

const { avoidSkill, unavoidSkill, avoidSource, unavoidSource, avoidCompany, unavoidCompany } = vi.hoisted(() => ({
  avoidSkill: vi.fn(),
  unavoidSkill: vi.fn(),
  avoidSource: vi.fn(),
  unavoidSource: vi.fn(),
  avoidCompany: vi.fn(),
  unavoidCompany: vi.fn(),
}));

vi.mock('$lib/profile.svelte', () => ({
  profileStore: {
    get profile() {
      return baseProfile;
    },
    avoidSkill,
    unavoidSkill,
    avoidSource,
    unavoidSource,
    avoidCompany,
    unavoidCompany,
  },
}));

// Both dictionary fetches are real network in production; stubbed here so the pickers'
// mount effects resolve instantly with no candidates to search.
vi.mock('$lib/skillDictionary', () => ({
  loadSkillDistribution: vi.fn().mockResolvedValue([]),
}));
vi.mock('$lib/sourceDictionary', () => ({
  loadSourceDistribution: vi.fn().mockResolvedValue([]),
}));
vi.mock('$lib/facets', () => ({
  companySearch: vi.fn().mockResolvedValue([]),
}));

beforeEach(() => {
  avoidSkill.mockReset().mockResolvedValue(baseProfile);
  unavoidSkill.mockReset().mockResolvedValue(baseProfile);
  avoidSource.mockReset().mockResolvedValue(baseProfile);
  unavoidSource.mockReset().mockResolvedValue(baseProfile);
  avoidCompany.mockReset().mockResolvedValue(baseProfile);
  unavoidCompany.mockReset().mockResolvedValue(baseProfile);
});

describe('AvoidCard', () => {
  it('notifies onProfileChanged after un-avoiding a skill succeeds', async () => {
    const onProfileChanged = vi.fn();
    render(AvoidCard, { props: { onProfileChanged } });

    await fireEvent.click(screen.getByTitle('java'));

    expect(unavoidSkill).toHaveBeenCalledWith('java');
    expect(onProfileChanged).toHaveBeenCalledTimes(1);
  });

  it('notifies onProfileChanged after un-avoiding a source succeeds', async () => {
    const onProfileChanged = vi.fn();
    render(AvoidCard, { props: { onProfileChanged } });

    await fireEvent.click(screen.getByTitle('greenhouse'));

    expect(unavoidSource).toHaveBeenCalledWith('greenhouse');
    expect(onProfileChanged).toHaveBeenCalledTimes(1);
  });

  it('notifies onProfileChanged after un-avoiding a company succeeds', async () => {
    const onProfileChanged = vi.fn();
    render(AvoidCard, { props: { onProfileChanged } });

    await fireEvent.click(screen.getByTitle('acme'));

    expect(unavoidCompany).toHaveBeenCalledWith('acme');
    expect(onProfileChanged).toHaveBeenCalledTimes(1);
  });

  it('does not notify onProfileChanged when the save fails', async () => {
    unavoidSkill.mockReset().mockRejectedValue(new Error('network error'));
    const onProfileChanged = vi.fn();
    render(AvoidCard, { props: { onProfileChanged } });

    await fireEvent.click(screen.getByTitle('java'));

    expect(onProfileChanged).not.toHaveBeenCalled();
  });

  it('still un-avoids the skill when no onProfileChanged prop is given', async () => {
    render(AvoidCard, { props: {} });

    await fireEvent.click(screen.getByTitle('java'));

    expect(unavoidSkill).toHaveBeenCalledWith('java');
  });
});
