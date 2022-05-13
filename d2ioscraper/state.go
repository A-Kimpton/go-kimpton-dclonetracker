package d2ioscraper

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Region int64

// Enum Printer ()
func (r Region) String() string {
    return [...]string{"", "Americas", "Europe", "Asia"}[r]
}

const (
	MAX_CLONE_LEVELS = 6 // This may change in future updates

	// Consts for region
	AMERICAS Region = 1
	EUROPE   Region = 2
	ASIA     Region = 3
)

var(
	globalState []D2CloneState
	lock sync.RWMutex // Lock for global state
)

// Struct to store state of the
type D2CloneState struct {
	Region      Region // Americas, Europe, Asia
	IsLadder    bool
	IsHardcore  bool
	LastUpdated time.Time // According to diablo2.io
	Progress    uint64    // Current progress level
}

func (s D2CloneState) String() string {
	return fmt.Sprintf("Region: %s, Ladder: %t, Hardcore: %t, Progress: %d/%d, Last Updated: %s", s.Region, s.IsLadder, s.IsHardcore, s.Progress, MAX_CLONE_LEVELS, s.LastUpdated)
}

func (s D2CloneState) PrettyString() string {
	ladder := "Non-Ladder"
	if s.IsLadder {
		ladder = "Ladder"
	}
	hardcore := "Softcore"
	if s.IsHardcore {
		hardcore = "Hardcore"
	}
	progress := fmt.Sprintf("%d/%d", s.Progress, MAX_CLONE_LEVELS)
	return fmt.Sprintf("[%s][%s][%s] has progress: %s", s.Region, ladder, hardcore, progress)
}

// Checks if they are the same server
func (s D2CloneState) IsSameServer(s2 D2CloneState) bool {
	if s.Region != s2.Region {
		return false
	}
	if s.IsLadder != s2.IsLadder {
		return false
	}
	if s.IsHardcore != s2.IsHardcore {
		return false
	}

	return true
}

// Checks if progress has updated - returns error if theyre not the same server
func (s D2CloneState) HasUpdated(s2 D2CloneState) (bool, error) {
	if !s.IsSameServer(s2) {
		return false, fmt.Errorf("servers are not the same s1(region-%s, ld-%t, hc-%t) != s2(region-%s, ld-%t, hc-%t)", s.Region, s.IsLadder, s.IsHardcore, s2.Region, s2.IsLadder, s2.IsHardcore)
	}
	return s.Progress != s2.Progress, nil
}

func Update(state []D2CloneState) {
	lock.Lock()
    defer lock.Unlock()
	globalState = state
}

// Gets the progress of a set server - threadsafe
func GetProgress(region Region, isLadder, isHardcore bool) (D2CloneState, error) {
	return GetProgressFromState(GetGlobalState(), region, isLadder, isHardcore)
}

// Helper func
func GetProgressFromState(state []D2CloneState, region Region, isLadder, isHardcore bool) (D2CloneState, error) {
	// Search
	for _, s := range(state) {
		if s.Region == region && s.IsLadder == isLadder && s.IsHardcore == isHardcore {
			// Found it
			return s, nil
		}
	}
	// Could not find required state
	return D2CloneState{}, errors.New(fmt.Sprintf("could not find d2clone state matching description - Region: %d, Ladder: %t, Hardcore: %t", region, isLadder, isHardcore))
}

// Make a copy of the state and return it (So that its threadsafe)
func GetGlobalState() []D2CloneState {
	lock.RLock()
    defer lock.RUnlock()
	toReturn := make([]D2CloneState, len(globalState))

	// If global state has any values, copy it
	if globalState != nil {
		copy(toReturn, globalState)
	}
	return toReturn
}