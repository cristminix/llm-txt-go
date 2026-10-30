// Package search provides fuzzy search over documentation sections.
// The scoring algorithm approximates Fuse.js behavior with threshold-based filtering.
package search

import (
	"math"
	"sort"
	"strings"

	"github.com/llm-txt-go-go/internal/config"
	"github.com/llm-txt-go-go/internal/docloader"
)

// SearchResult holds a matched document section with its relevance score.
type SearchResult struct {
	Item  docloader.DocSection `json:"item"`
	Score float64              `json:"score,omitempty"`
}

// SearchEngine performs fuzzy search over documentation sections.
type SearchEngine struct {
	sections []docloader.DocSection
	config   config.SearchConfig
}

// NewSearchEngine creates a SearchEngine with the given sections and configuration.
func NewSearchEngine(sections []docloader.DocSection, cfg config.SearchConfig) *SearchEngine {
	return &SearchEngine{
		sections: sections,
		config:   cfg,
	}
}

// Search performs a fuzzy search and returns results up to limit, sorted by relevance.
func (se *SearchEngine) Search(query string, limit int) []SearchResult {
	queryLower := strings.ToLower(strings.TrimSpace(query))
	if len(queryLower) < se.config.MinMatchCharLength {
		return nil
	}

	var results []SearchResult

	for _, section := range se.sections {
		score := se.computeScore(queryLower, section)
		if score <= se.config.Threshold {
			results = append(results, SearchResult{
				Item:  section,
				Score: score,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score < results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// computeScore calculates a relevance score for a section against the query.
// Score ranges from 0 (perfect match) to 1 (no match). Lower is better.
func (se *SearchEngine) computeScore(queryLower string, section docloader.DocSection) float64 {
	fields := []string{section.Title, section.Content, section.Path}
	bestScore := 1.0

	for _, field := range fields {
		fieldLower := strings.ToLower(field)
		if fieldLower == "" {
			continue
		}

		score := fieldScore(queryLower, fieldLower)
		if score < bestScore {
			bestScore = score
		}
	}

	return bestScore
}

// fieldScore computes the match score for a single field.
// Uses a combination of substring position ratio and length ratio.
func fieldScore(query, field string) float64 {
	if query == field {
		return 0.0
	}

	// Exact substring match — score based on position in field
	if idx := strings.Index(field, query); idx != -1 {
		posRatio := float64(idx) / float64(len(field))
		lenRatio := float64(len(query)) / float64(len(field))
		return posRatio*0.3 + (1.0-lenRatio)*0.2
	}

	// Token-based partial matching
	queryTokens := strings.Fields(query)
	if len(queryTokens) == 0 {
		return 1.0
	}

	fieldTokens := strings.Fields(field)
	if len(fieldTokens) == 0 {
		return 1.0
	}

	totalScore := 0.0
	matched := 0

	for _, qt := range queryTokens {
		if len(qt) < 3 {
			continue
		}
		bestTokenScore := 1.0

		for _, ft := range fieldTokens {
			if strings.Contains(ft, qt) {
				posRatio := float64(strings.Index(ft, qt)) / float64(len(ft))
				lenRatio := float64(len(qt)) / float64(len(ft))
				score := posRatio*0.3 + (1.0-lenRatio)*0.2
				if score < bestTokenScore {
					bestTokenScore = score
				}
			}
		}

		if bestTokenScore < 1.0 {
			totalScore += bestTokenScore
			matched++
		}
	}

	if matched == 0 {
		return 1.0
	}

	// Average token score + penalty for unmatched tokens
	avgScore := totalScore / float64(matched)
	coveragePenalty := float64(len(queryTokens)-matched) / float64(len(queryTokens)) * 0.4

	return math.Min(avgScore+coveragePenalty, 1.0)
}
