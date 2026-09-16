import { fireEvent, render, screen } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { CompanyListItem, ExperienceEmploymentWithAtoms } from '$lib/types';
import ExperienceEmploymentCard from './ExperienceEmploymentCard.svelte';

// The inline edit form gains a company-catalogue autocomplete for job-kind employments
// only — project-kind entries name a project, not a company — see openspec change
// experience-company-picker-present-checkbox.

const { listCompanies } = vi.hoisted(() => ({ listCompanies: vi.fn() }));

vi.mock('$lib/api', () => ({ api: { listCompanies } }));

const ringCentral: CompanyListItem = {
  slug: 'ringcentral',
  name: 'RingCentral',
  job_count: 57,
  feedback_count: 0,
  feedback_rating_avg: null,
};

const jobEmployment: ExperienceEmploymentWithAtoms = {
  id: 'e1',
  kind: 'job',
  company: 'Acme',
  role: 'Engineer',
  atoms: [],
};

const projectEmployment: ExperienceEmploymentWithAtoms = {
  id: 'e2',
  kind: 'project',
  name: 'Side Project',
  atoms: [],
};

function baseProps(employment: ExperienceEmploymentWithAtoms) {
  return {
    employment,
    selectedIds: [],
    busy: false,
    turnActive: false,
    onToggleSelect: vi.fn(),
    onConfirmAtom: vi.fn(),
    onSaveAtomEdit: vi.fn(),
    onSavePromote: vi.fn(),
    onRemoveAtom: vi.fn(),
    onSaveEmployment: vi.fn(),
    onRemoveEmployment: vi.fn(),
  };
}

beforeEach(() => {
  listCompanies.mockReset().mockResolvedValue({ items: [ringCentral], hasMore: false });
});

describe('ExperienceEmploymentCard edit form', () => {
  it('offers catalogue suggestions for a job-kind Company field', async () => {
    render(ExperienceEmploymentCard, { props: baseProps(jobEmployment) });

    await fireEvent.click(screen.getByRole('button', { name: 'Edit' }));
    await fireEvent.input(screen.getByLabelText('Company'), { target: { value: 'ring' } });

    expect(await screen.findByRole('option', { name: /RingCentral/ })).toBeTruthy();
  });

  it('keeps a plain text input with no suggestions for a project-kind name', async () => {
    render(ExperienceEmploymentCard, { props: baseProps(projectEmployment) });

    await fireEvent.click(screen.getByRole('button', { name: 'Edit' }));
    await fireEvent.input(screen.getByLabelText('Project name'), {
      target: { value: 'ring' },
    });

    expect(listCompanies).not.toHaveBeenCalled();
    expect(screen.queryByRole('listbox')).toBeNull();
  });
});
