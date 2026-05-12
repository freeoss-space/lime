package cooldown_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freeoss-space/lime/internal/cooldown"
)

// --- Parse ---

func TestParse_ZeroVariants(t *testing.T) {
	tests := []string{"0", "0d", "0h", "0w", ""}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			d, err := cooldown.Parse(s)
			require.NoError(t, err)
			assert.True(t, d.IsZero())
		})
	}
}

func TestParse_Hours(t *testing.T) {
	d, err := cooldown.Parse("24h")
	require.NoError(t, err)
	assert.False(t, d.IsZero())
	assert.Equal(t, 24*time.Hour, d.D())
}

func TestParse_Days(t *testing.T) {
	d, err := cooldown.Parse("14d")
	require.NoError(t, err)
	assert.Equal(t, 14*24*time.Hour, d.D())
}

func TestParse_Weeks(t *testing.T) {
	d, err := cooldown.Parse("2w")
	require.NoError(t, err)
	assert.Equal(t, 14*24*time.Hour, d.D())
}

func TestParse_LargeValues(t *testing.T) {
	d, err := cooldown.Parse("365d")
	require.NoError(t, err)
	assert.Equal(t, 365*24*time.Hour, d.D())
}

func TestParse_InvalidFormats(t *testing.T) {
	bad := []string{"14", "2weeks", "1month", "-1d", "d", "1.5d", "14D", "14H"}
	for _, s := range bad {
		t.Run(s, func(t *testing.T) {
			_, err := cooldown.Parse(s)
			assert.Error(t, err)
		})
	}
}

func TestParse_ErrorContainsInput(t *testing.T) {
	_, err := cooldown.Parse("badvalue")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "badvalue")
}

// --- IsZero ---

func TestZero_IsZero(t *testing.T) {
	assert.True(t, cooldown.Zero.IsZero())
}

func TestNonZero_IsNotZero(t *testing.T) {
	d, _ := cooldown.Parse("7d")
	assert.False(t, d.IsZero())
}

// --- Blocks ---

var now = time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)

func TestBlocks_ZeroCooldownNeverBlocks(t *testing.T) {
	releaseYesterday := now.Add(-24 * time.Hour)
	assert.False(t, cooldown.Zero.Blocks(releaseYesterday, now))
}

func TestBlocks_ReleaseTooNew(t *testing.T) {
	d, _ := cooldown.Parse("14d")
	releasedTwoDaysAgo := now.Add(-2 * 24 * time.Hour)
	assert.True(t, d.Blocks(releasedTwoDaysAgo, now), "2-day-old release should be blocked by 14d cooldown")
}

func TestBlocks_ReleaseOldEnough(t *testing.T) {
	d, _ := cooldown.Parse("14d")
	releasedTwoWeeksAgo := now.Add(-15 * 24 * time.Hour)
	assert.False(t, d.Blocks(releasedTwoWeeksAgo, now), "15-day-old release passes 14d cooldown")
}

func TestBlocks_ExactlyAtBoundary(t *testing.T) {
	d, _ := cooldown.Parse("14d")
	releasedExactly14DaysAgo := now.Add(-14 * 24 * time.Hour)
	// Exactly at the boundary: now.Sub(t) == d → not blocked (>= cooldown required)
	assert.False(t, d.Blocks(releasedExactly14DaysAgo, now))
}

func TestBlocks_HourCooldown(t *testing.T) {
	d, _ := cooldown.Parse("48h")
	releasedOneHourAgo := now.Add(-time.Hour)
	assert.True(t, d.Blocks(releasedOneHourAgo, now))

	released3DaysAgo := now.Add(-3 * 24 * time.Hour)
	assert.False(t, d.Blocks(released3DaysAgo, now))
}

func TestBlocks_WeekCooldown(t *testing.T) {
	d, _ := cooldown.Parse("2w")
	releasedOneWeekAgo := now.Add(-7 * 24 * time.Hour)
	assert.True(t, d.Blocks(releasedOneWeekAgo, now))

	released3WeeksAgo := now.Add(-21 * 24 * time.Hour)
	assert.False(t, d.Blocks(released3WeeksAgo, now))
}

// --- String ---

func TestString_Zero(t *testing.T) {
	assert.Equal(t, "0d", cooldown.Zero.String())
}

func TestString_Days(t *testing.T) {
	// 9d is not a whole number of weeks, so stays as days.
	d, _ := cooldown.Parse("9d")
	assert.Equal(t, "9d", d.String())
}

func TestString_Weeks(t *testing.T) {
	d, _ := cooldown.Parse("2w")
	assert.Equal(t, "2w", d.String())
}

func TestString_Hours(t *testing.T) {
	d, _ := cooldown.Parse("36h")
	assert.Equal(t, "36h", d.String())
}

func TestString_DaysThatAreExactWeeks(t *testing.T) {
	d, _ := cooldown.Parse("7d")
	// 7d == 1w, prefer weeks
	assert.Equal(t, "1w", d.String())
}
