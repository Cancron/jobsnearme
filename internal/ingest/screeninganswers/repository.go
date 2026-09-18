package screeninganswers

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/strelov1/freehire/internal/platform/db"
	"github.com/strelov1/freehire/internal/platform/pgconv"
)

// Compile-time proof that QueriesRepository satisfies Repository.
var _ Repository = (*QueriesRepository)(nil)

// QueriesRepository adapts *db.Queries + a pool to the Repository. It maps the no-row
// condition on Get to ErrNotFound; UpdateLocked needs the pool directly, alongside
// *db.Queries, to open the transaction the locked read-merge-write runs inside.
type QueriesRepository struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

// NewQueriesRepository constructs a QueriesRepository.
func NewQueriesRepository(q *db.Queries, pool *pgxpool.Pool) *QueriesRepository {
	return &QueriesRepository{q: q, pool: pool}
}

// Get returns the user's screening answers, mapping no row to ErrNotFound.
func (r *QueriesRepository) Get(ctx context.Context, userID int64) (Answers, error) {
	row, err := r.q.GetScreeningAnswers(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Answers{}, ErrNotFound
	}
	if err != nil {
		return Answers{}, err
	}
	return answersFromRow(row), nil
}

// UpdateLocked runs the whole read-merge-write in one transaction: GetScreeningAnswersForUpdate
// takes a row lock on the caller's existing record (nothing to lock when there is no row
// yet, which reads as a fully-unstated Answers{}), merge combines it with the caller's
// update, and the merged result is written back with the same UpsertScreeningAnswers
// Store.Update used to call directly — all before the commit that releases the lock. A
// second concurrent call for the same userID blocks on the SELECT ... FOR UPDATE until this
// transaction commits, so it merges onto this write's result instead of the same stale row.
func (r *QueriesRepository) UpdateLocked(ctx context.Context, userID int64, merge func(existing Answers) Answers) (Answers, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Answers{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := r.q.WithTx(tx)
	existingRow, err := qtx.GetScreeningAnswersForUpdate(ctx, userID)
	var existing Answers
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		existing = Answers{}
	case err != nil:
		return Answers{}, err
	default:
		existing = answersFromRow(existingRow)
	}

	merged := merge(existing)
	row, err := qtx.UpsertScreeningAnswers(ctx, db.UpsertScreeningAnswersParams{
		UserID:                userID,
		AuthorizedCountries:   merged.AuthorizedCountries,
		VisaSponsorshipNeeded: pgconv.Bool(merged.VisaSponsorshipNeeded),
		DesiredSalaryAmount:   pgconv.Int4(merged.DesiredSalaryAmount),
		DesiredSalaryCurrency: pgconv.Text(derefString(merged.DesiredSalaryCurrency)),
		DesiredSalaryPeriod:   pgconv.Text(derefString(merged.DesiredSalaryPeriod)),
		NoticePeriodDays:      pgconv.Int4(merged.NoticePeriodDays),
		WillingToRelocate:     pgconv.Bool(merged.WillingToRelocate),
		Age18OrOlder:          pgconv.Bool(merged.Age18OrOlder),
	})
	if err != nil {
		return Answers{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Answers{}, err
	}
	return answersFromRow(row), nil
}

// answersFromRow maps the generated db row to the package domain type.
func answersFromRow(row db.ScreeningAnswer) Answers {
	return Answers{
		AuthorizedCountries:   row.AuthorizedCountries,
		VisaSponsorshipNeeded: pgconv.BoolPtr(row.VisaSponsorshipNeeded),
		DesiredSalaryAmount:   pgconv.IntPtr(row.DesiredSalaryAmount),
		DesiredSalaryCurrency: pgconv.TextPtr(row.DesiredSalaryCurrency),
		DesiredSalaryPeriod:   pgconv.TextPtr(row.DesiredSalaryPeriod),
		NoticePeriodDays:      pgconv.IntPtr(row.NoticePeriodDays),
		WillingToRelocate:     pgconv.BoolPtr(row.WillingToRelocate),
		Age18OrOlder:          pgconv.BoolPtr(row.Age18OrOlder),
	}
}

// derefString returns the empty string for a nil pointer, its content otherwise —
// pgconv.Text already treats "" as NULL, so this is the one-line bridge from Answers'
// *string fields to that convention.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
