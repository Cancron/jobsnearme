package experience

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/strelov1/freehire/internal/candidate/cv"
)

func TestSeedHistoryFromBankSplitsJobsAndProjects(t *testing.T) {
	jobID := uuid.New()
	projectID := uuid.New()
	employments := []Employment{
		{ID: jobID, Kind: KindJob, Company: "RingCentral", Role: "SWE", Location: "Remote"},
		{ID: projectID, Kind: KindProject, Company: "telagon.io", Link: "https://telagon.io"},
	}
	atoms := []Atom{
		{EmploymentID: &jobID, Claim: "Cut latency", Provenance: ProvenanceCVImport},
		{EmploymentID: &projectID, Claim: "1.4M+ channels", Provenance: ProvenanceStatedInChat},
		{EmploymentID: &projectID, Claim: "model guess", Provenance: ProvenanceAgentInferred},
	}

	got := seedHistoryFromBank(employments, atoms)

	if !got.HasJobEmployments || !got.HasProjectEmployments {
		t.Fatalf("flags = jobs:%v projects:%v, want both true", got.HasJobEmployments, got.HasProjectEmployments)
	}
	if len(got.Experience) != 1 || got.Experience[0].Company != "RingCentral" {
		t.Errorf("experience = %+v, want the job only", got.Experience)
	}
	if len(got.Experience[0].Highlights) != 1 || got.Experience[0].Highlights[0] != "Cut latency" {
		t.Errorf("job highlights = %+v", got.Experience[0].Highlights)
	}
	if len(got.Projects) != 1 {
		t.Fatalf("projects = %+v, want one banked project", got.Projects)
	}
	p := got.Projects[0]
	if p.Name != "telagon.io" || p.Link != "https://telagon.io" {
		t.Errorf("project identity = %+v, want name and link", p)
	}
	if len(p.Highlights) != 1 || p.Highlights[0] != "1.4M+ channels" {
		t.Errorf("project highlights = %+v, want publishable only", p.Highlights)
	}
}

func TestSeedHistoryFromBankJobsOnly(t *testing.T) {
	jobID := uuid.New()
	got := seedHistoryFromBank(
		[]Employment{{ID: jobID, Kind: KindJob, Company: "Acme", Role: "Dev"}},
		nil,
	)
	if !got.HasJobEmployments || got.HasProjectEmployments {
		t.Errorf("flags = %+v, want jobs without projects", got)
	}
	if len(got.Projects) != 0 {
		t.Errorf("projects = %+v, want none", got.Projects)
	}
}

func TestSeedHistoryFromBankEmpty(t *testing.T) {
	got := seedHistoryFromBank(nil, nil)
	if got.HasJobEmployments || got.HasProjectEmployments {
		t.Errorf("empty bank flags = %+v", got)
	}
	if len(got.Experience) != 0 || len(got.Projects) != 0 {
		t.Errorf("empty bank rows = %+v", got)
	}
}

// The regression this PR shipped with: a bank holding ONLY a project-kind row must not be
// read as "has job history" — HasJobEmployments must stay false so the seeder falls back
// to the structure's own Experience instead of blanking it. See cv_seed.go.
func TestSeedHistoryFromBankProjectsOnlyLeavesJobEmploymentsFalse(t *testing.T) {
	id := uuid.New()
	got := seedHistoryFromBank(
		[]Employment{{ID: id, Kind: KindProject, Company: "opensched", Link: "https://opensched.dev"}},
		[]Atom{{EmploymentID: &id, Claim: "shipped a feature", Provenance: ProvenanceStatedInChat}},
	)
	if got.HasJobEmployments {
		t.Errorf("flags = %+v, want HasJobEmployments false for a projects-only bank", got)
	}
	if !got.HasProjectEmployments {
		t.Errorf("flags = %+v, want HasProjectEmployments true", got)
	}
	if len(got.Experience) != 0 {
		t.Errorf("experience = %+v, want none — the project's highlight belongs to Projects, not Experience", got.Experience)
	}
}

// Confirmed evidence with no place still lands in Experience (as a blank-header entry),
// even when the bank holds no employment rows at all — HasJobEmployments must reflect that.
func TestSeedHistoryFromBankPlacelessOnlySetsHasJobEmployments(t *testing.T) {
	got := seedHistoryFromBank(nil, []Atom{
		{Claim: "shipped a feature nobody wrote down", Provenance: ProvenanceStatedInChat},
	})
	if !got.HasJobEmployments {
		t.Errorf("flags = %+v, want HasJobEmployments true — placeless evidence is still Experience content", got)
	}
	if len(got.Experience) != 1 || len(got.Experience[0].Highlights) != 1 {
		t.Errorf("experience = %+v, want one placeless entry", got.Experience)
	}
}

func TestSeedHistoryProjectWithoutPublishableAtomsStillListed(t *testing.T) {
	id := uuid.New()
	got := seedHistoryFromBank(
		[]Employment{{ID: id, Kind: KindProject, Company: "opensched", Link: "https://opensched.dev"}},
		[]Atom{{EmploymentID: &id, Claim: "inferred", Provenance: ProvenanceAgentInferred}},
	)
	if len(got.Projects) != 1 || got.Projects[0].Link != "https://opensched.dev" {
		t.Errorf("projects = %+v, want identity even without publishable highlights", got.Projects)
	}
	if len(got.Projects[0].Highlights) != 0 {
		t.Errorf("highlights = %+v, want none from agent_inferred", got.Projects[0].Highlights)
	}
}

// The bug this PR fixed: a long-tenured role banking more achievements over the years
// than cv.MaxBullets meant "reset base CV from résumé" refused forever, since the seed
// it built was already over the ceiling cvedit's write gate enforces. Capping here, and
// keeping the MOST RECENT claims (atoms arrive oldest-created first, see
// publishableHighlights), is what closes that dead end.
func TestSeedHistoryFromBankCapsAnEmploymentBucketAtMaxBullets(t *testing.T) {
	jobID := uuid.New()
	const extra = 5
	total := cv.MaxBullets + extra
	atoms := make([]Atom, total)
	for i := range atoms {
		// Oldest first, matching ListExperienceAtoms' ORDER BY created_at — claim i is
		// older than claim i+1.
		atoms[i] = Atom{EmploymentID: &jobID, Claim: fmt.Sprintf("claim %d", i), Provenance: ProvenanceCVImport}
	}

	got := seedHistoryFromBank([]Employment{{ID: jobID, Kind: KindJob, Company: "Acme"}}, atoms)

	if len(got.Experience) != 1 {
		t.Fatalf("experience = %+v, want one row", got.Experience)
	}
	highlights := got.Experience[0].Highlights
	if len(highlights) != cv.MaxBullets {
		t.Fatalf("highlights = %d, want capped at %d", len(highlights), cv.MaxBullets)
	}
	if highlights[0] != fmt.Sprintf("claim %d", extra) {
		t.Errorf("first surviving highlight = %q, want the oldest DROPPED — kept ones should start at claim %d", highlights[0], extra)
	}
	if last := highlights[len(highlights)-1]; last != fmt.Sprintf("claim %d", total-1) {
		t.Errorf("last surviving highlight = %q, want the most recently added claim", last)
	}
}

// The placeless bucket is not special-cased in the bucketing loop, so it needs the same
// cap — this is the exact shape of the bug found via freehire.me/api/v1/me/cvs/base/reseed
// (a member with a company-less employment history, reported as "experience[9]").
func TestSeedHistoryFromBankCapsThePlacelessBucketAtMaxBullets(t *testing.T) {
	const extra = 3
	total := cv.MaxBullets + extra
	atoms := make([]Atom, total)
	for i := range atoms {
		atoms[i] = Atom{Claim: fmt.Sprintf("placeless claim %d", i), Provenance: ProvenanceStatedInChat}
	}

	got := seedHistoryFromBank(nil, atoms)

	if len(got.Experience) != 1 {
		t.Fatalf("experience = %+v, want the one placeless entry", got.Experience)
	}
	highlights := got.Experience[0].Highlights
	if len(highlights) != cv.MaxBullets {
		t.Fatalf("placeless highlights = %d, want capped at %d", len(highlights), cv.MaxBullets)
	}
	if last := highlights[len(highlights)-1]; last != fmt.Sprintf("placeless claim %d", total-1) {
		t.Errorf("last surviving placeless highlight = %q, want the most recently added claim", last)
	}
}

// WorkHistory still flattens projects into experience-shaped rows for fit analysis.
func TestWorkHistoryStillFlattensProjects(t *testing.T) {
	jobID := uuid.New()
	projectID := uuid.New()
	flat := experienceFromBank(
		[]Employment{
			{ID: jobID, Kind: KindJob, Company: "Acme", Role: "Dev"},
			{ID: projectID, Kind: KindProject, Company: "opensched", Link: "https://opensched.dev"},
		},
		nil,
	)
	if len(flat) != 2 {
		t.Fatalf("WorkHistory-shaped rows = %d, want job and project flattened", len(flat))
	}
	var companies []string
	for _, e := range flat {
		companies = append(companies, e.Company)
	}
	if companies[0] != "Acme" || companies[1] != "opensched" {
		t.Errorf("companies = %v, want Acme then opensched", companies)
	}
}
