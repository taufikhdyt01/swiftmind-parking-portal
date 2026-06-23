// Package fine holds the pure fine-calculation logic, decoupled from storage and
// transport so it can be unit-tested exhaustively and reused by any service.
//
// Formula: fine = base_amount(violation_type) × time_multiplier × repeat_multiplier
// Every value is configurable per rule version (see Ruleset). Money is integer
// IDR; multipliers are applied and the result is rounded half-up to the nearest
// rupiah.
package fine

import (
	"errors"
	"fmt"
	"math"
	"time"

	"parkwatch/pkg/domain"
)

// Ruleset is the full, versioned configuration that drives a fine calculation.
type Ruleset struct {
	BaseAmounts      map[string]int64 `json:"base_amounts"`
	TimeMultiplier   TimeMultiplier   `json:"time_multiplier"`
	RepeatMultiplier RepeatMultiplier `json:"repeat_multiplier"`
}

// TimeMultiplier configures the day/night multiplier based on the hour of the
// violation. The day window is [DayStartHour, NightStartHour); everything else
// (overnight) uses the night multiplier.
type TimeMultiplier struct {
	DayStartHour    int     `json:"day_start_hour"`
	NightStartHour  int     `json:"night_start_hour"`
	DayMultiplier   float64 `json:"day_multiplier"`
	NightMultiplier float64 `json:"night_multiplier"`
}

// RepeatMultiplier configures the multiplier based on the number of prior unpaid
// violations on the same plate. Tiers are evaluated by threshold; the highest
// tier whose MinPriorUnpaid is satisfied wins.
type RepeatMultiplier struct {
	Tiers []RepeatTier `json:"tiers"`
}

// RepeatTier maps a "prior unpaid count >= MinPriorUnpaid" threshold to a multiplier.
type RepeatTier struct {
	MinPriorUnpaid int     `json:"min_prior_unpaid"`
	Multiplier     float64 `json:"multiplier"`
}

// Input is everything needed to price a single violation.
type Input struct {
	ViolationType    string
	OccurredAt       time.Time
	PriorUnpaidCount int
}

// Breakdown is the itemized result of a calculation. These fields are persisted
// as the violation's immutable snapshot so the fine never changes when rules do.
type Breakdown struct {
	BaseAmount       int64   `json:"base_amount"`
	TimeMultiplier   float64 `json:"time_multiplier"`
	RepeatMultiplier float64 `json:"repeat_multiplier"`
	PriorUnpaidCount int     `json:"prior_unpaid_count"`
	FinalAmount      int64   `json:"final_amount"`
}

// Calculate prices a violation against this ruleset, returning the full breakdown.
func (r Ruleset) Calculate(in Input) (Breakdown, error) {
	base, ok := r.BaseAmounts[in.ViolationType]
	if !ok {
		return Breakdown{}, fmt.Errorf("fine: unknown violation type %q", in.ViolationType)
	}

	timeMult := r.TimeMultiplier.For(in.OccurredAt)
	repeatMult := r.RepeatMultiplier.For(in.PriorUnpaidCount)
	final := int64(math.Round(float64(base) * timeMult * repeatMult))

	return Breakdown{
		BaseAmount:       base,
		TimeMultiplier:   timeMult,
		RepeatMultiplier: repeatMult,
		PriorUnpaidCount: in.PriorUnpaidCount,
		FinalAmount:      final,
	}, nil
}

// For returns the day or night multiplier for the given timestamp's hour.
func (t TimeMultiplier) For(ts time.Time) float64 {
	h := ts.Hour()
	if h >= t.DayStartHour && h < t.NightStartHour {
		return t.DayMultiplier
	}
	return t.NightMultiplier
}

// For returns the multiplier for the given number of prior unpaid violations.
func (rm RepeatMultiplier) For(priorUnpaid int) float64 {
	mult := 1.0
	for _, tier := range rm.Tiers {
		if priorUnpaid >= tier.MinPriorUnpaid {
			mult = tier.Multiplier
		}
	}
	return mult
}

// ErrInvalidRuleset wraps every validation failure so callers can classify a
// bad ruleset with errors.Is rather than matching on the message text.
var ErrInvalidRuleset = errors.New("invalid ruleset")

// Validate checks that a ruleset is complete and sane before it is published.
// All failures wrap ErrInvalidRuleset.
func (r Ruleset) Validate() error {
	for _, vt := range domain.ViolationTypes() {
		amt, ok := r.BaseAmounts[vt.String()]
		if !ok {
			return fmt.Errorf("%w: missing base amount for %q", ErrInvalidRuleset, vt)
		}
		if amt <= 0 {
			return fmt.Errorf("%w: base amount for %q must be positive", ErrInvalidRuleset, vt)
		}
	}
	tm := r.TimeMultiplier
	if tm.DayMultiplier <= 0 || tm.NightMultiplier <= 0 {
		return fmt.Errorf("%w: time multipliers must be positive", ErrInvalidRuleset)
	}
	if !validHour(tm.DayStartHour) || !validHour(tm.NightStartHour) {
		return fmt.Errorf("%w: hour boundaries must be between 0 and 23", ErrInvalidRuleset)
	}
	if len(r.RepeatMultiplier.Tiers) == 0 {
		return fmt.Errorf("%w: at least one repeat tier is required", ErrInvalidRuleset)
	}
	for _, tier := range r.RepeatMultiplier.Tiers {
		if tier.Multiplier <= 0 {
			return fmt.Errorf("%w: repeat multipliers must be positive", ErrInvalidRuleset)
		}
		if tier.MinPriorUnpaid < 0 {
			return fmt.Errorf("%w: repeat tier thresholds must be non-negative", ErrInvalidRuleset)
		}
	}
	return nil
}

func validHour(h int) bool { return h >= 0 && h <= 23 }

// DefaultRuleset is the day-one ruleset specified by the assignment.
func DefaultRuleset() Ruleset {
	return Ruleset{
		BaseAmounts: map[string]int64{
			domain.ViolationExpiredMeter.String():    50_000,
			domain.ViolationNoParkingZone.String():   150_000,
			domain.ViolationBlockingHydrant.String(): 250_000,
			domain.ViolationDisabledSpot.String():    500_000,
		},
		TimeMultiplier: TimeMultiplier{
			DayStartHour:    6,
			NightStartHour:  22,
			DayMultiplier:   1.0,
			NightMultiplier: 1.5,
		},
		RepeatMultiplier: RepeatMultiplier{
			Tiers: []RepeatTier{
				{MinPriorUnpaid: 0, Multiplier: 1.0},
				{MinPriorUnpaid: 1, Multiplier: 1.5},
				{MinPriorUnpaid: 2, Multiplier: 2.0},
			},
		},
	}
}
