// Copyright 2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package patternmatcher

import (
	"path/filepath"
)

// PatternMatcher is a stateful interface that tracks if all its patterns have been matched.
type PatternMatcher interface {
	// MatchesPattern is like filepath.Match, but returns true if any enclosed patterns match.
	MatchesPattern(name string) bool
	// StillMatching returns true unless all required patterns are matched.
	StillMatching() bool
	// Unmatched returns non-empty if MatchesPattern hasn't matched all patterns, yet.
	Unmatched() []string
}

type patternMatcher struct {
	patterns map[string]bool
	fastRead bool
}

// New returns a possibly no-op PatternMatcher based on the inputs
func New(patterns []string, fastRead bool) PatternMatcher {
	pm := &patternMatcher{patterns: map[string]bool{}, fastRead: fastRead}
	for _, pattern := range patterns {
		pm.patterns[pattern] = false
	}
	return pm
}

func (pm *patternMatcher) MatchesPattern(name string) bool {
	if len(pm.patterns) == 0 {
		return true
	}
	for pattern := range pm.patterns {
		if ok, _ := filepath.Match(pattern, name); ok {
			pm.patterns[pattern] = true
			return true
		}
	}
	return false
}

func (pm *patternMatcher) StillMatching() bool {
	return !pm.fastRead || len(pm.patterns) == 0 || len(pm.Unmatched()) > 0
}

func (pm *patternMatcher) Unmatched() []string {
	unmatched := make([]string, 0, len(pm.patterns))
	for pattern, matched := range pm.patterns {
		if !matched {
			unmatched = append(unmatched, pattern)
		}
	}
	return unmatched
}
