package evaluation

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	frontierTier         = "frontier"
	ossTier              = "oss"
	roleDimension        = "role-understanding"
	personalityDimension = "personality-expression"
	defaultCellMaximum   = 8
)

type scorecardCell struct {
	total   int
	maximum int
	passed  bool
}

type scorecardModel struct {
	name  string
	cells map[string]*scorecardCell
}

type scorecardResult struct {
	role     string
	frontier []*scorecardModel
	oss      []*scorecardModel
}

type scorecardRow struct {
	role     string
	frontier *scorecardModel
	oss      *scorecardModel
}

// MarkdownScorecard validates scored result files and renders one compact page.
func MarkdownScorecard(resultsDir, seat string) ([]byte, error) {
	return markdownScorecard(resultsDir, seat, false)
}

// MarkdownHistoricalScorecard renders preserved records without rebinding
// their immutable pack digests to the current roster.
func MarkdownHistoricalScorecard(resultsDir, seat string) ([]byte, error) {
	return markdownScorecard(resultsDir, seat, true)
}

func markdownScorecard(resultsDir, seat string, historical bool) ([]byte, error) {
	if strings.TrimSpace(resultsDir) == "" {
		return nil, fmt.Errorf("scorecard results directory is required")
	}
	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		return nil, fmt.Errorf("read scorecard results: %w", err)
	}

	var results []scorecardResult
	seenRoles := make(map[string]bool)
	dates := make(map[string]bool)
	totalPoints := 0
	maximumPoints := 0
	passedCases := 0
	caseCount := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(resultsDir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read scorecard result %q: %w", entry.Name(), err)
		}
		result, err := DecodeResult(raw)
		if err != nil {
			return nil, fmt.Errorf("decode scorecard result %q: %w", entry.Name(), err)
		}
		if seat != "" && result.Seat != seat {
			continue
		}
		if seenRoles[result.Role] {
			return nil, fmt.Errorf(
				"scorecard repeats role %q for seat %q",
				result.Role,
				result.Seat,
			)
		}
		var packCases []Case
		if historical {
			packCases, err = historicalScorecardCases(result)
			if err != nil {
				return nil, fmt.Errorf("validate historical scorecard result %q: %w", entry.Name(), err)
			}
		} else {
			pack, buildErr := Build(result.Role, result.Seat)
			if buildErr != nil {
				return nil, fmt.Errorf("build scorecard pack for %q: %w", entry.Name(), buildErr)
			}
			if err := ValidateResult(result, pack); err != nil {
				return nil, fmt.Errorf("validate scorecard result %q: %w", entry.Name(), err)
			}
			packCases = pack.Cases
		}

		models := map[string]map[string]*scorecardModel{
			frontierTier: {},
			ossTier:      {},
		}
		cases := make(map[string]Case, len(packCases))
		for _, evalCase := range packCases {
			cases[evalCase.ID] = evalCase
		}
		for _, scored := range result.Cases {
			evalCase := cases[scored.ID]
			tierModels, ok := models[evalCase.ModelTier]
			if !ok {
				return nil, fmt.Errorf(
					"scorecard result %q has unsupported model tier %q",
					entry.Name(),
					evalCase.ModelTier,
				)
			}
			model := tierModels[scored.Model]
			if model == nil {
				model = &scorecardModel{
					name:  scored.Model,
					cells: make(map[string]*scorecardCell),
				}
				tierModels[scored.Model] = model
			}
			cell := model.cells[evalCase.Dimension]
			if cell == nil {
				cell = &scorecardCell{passed: true}
				model.cells[evalCase.Dimension] = cell
			}
			cell.total += scored.Total
			cell.maximum += len(evalCase.Rubric) * 2
			cell.passed = cell.passed && scored.Passed

			totalPoints += scored.Total
			maximumPoints += len(evalCase.Rubric) * 2
			caseCount++
			if scored.Passed {
				passedCases++
			}
		}

		card := scorecardResult{
			role:     result.Role,
			frontier: sortedScorecardModels(models[frontierTier]),
			oss:      sortedScorecardModels(models[ossTier]),
		}
		if err := validateScorecardResult(card); err != nil {
			return nil, fmt.Errorf("scorecard result %q: %w", entry.Name(), err)
		}
		results = append(results, card)
		seenRoles[result.Role] = true
		dates[result.Provenance.EvaluatedAt] = true
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("scorecard has no results for seat %q", seat)
	}

	var rows []scorecardRow
	frontierModels := make(map[string]bool)
	ossModels := make(map[string]bool)
	for _, result := range results {
		for _, frontier := range result.frontier {
			frontierModels[frontier.name] = true
			for _, oss := range result.oss {
				ossModels[oss.name] = true
				rows = append(rows, scorecardRow{
					role:     result.role,
					frontier: frontier,
					oss:      oss,
				})
			}
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].role != rows[j].role {
			return rows[i].role < rows[j].role
		}
		if rows[i].frontier.name != rows[j].frontier.name {
			return rows[i].frontier.name < rows[j].frontier.name
		}
		return rows[i].oss.name < rows[j].oss.name
	})

	dateList := sortedKeys(dates)
	uniformModels := len(frontierModels) == 1 && len(ossModels) == 1
	var out bytes.Buffer
	out.WriteString("<!-- generated by agent-compose scorecard. Do not edit. -->\n\n")
	out.WriteString("# Evaluation scorecard\n\n")
	fmt.Fprintf(
		&out,
		"`%s · seat %s",
		strings.Join(dateList, ", "),
		seat,
	)
	if uniformModels {
		fmt.Fprintf(
			&out,
			" · F %s · O %s",
			sortedKeys(frontierModels)[0],
			sortedKeys(ossModels)[0],
		)
	}
	fmt.Fprintf(
		&out,
		" · %d/%d pass · %d/%d points`\n\n",
		passedCases,
		caseCount,
		totalPoints,
		maximumPoints,
	)

	if uniformModels {
		out.WriteString("| role | FR | FP | OR | OP | Σ |\n")
		out.WriteString("|---|---:|---:|---:|---:|---:|\n")
	} else {
		out.WriteString("| role | F model | O model | FR | FP | OR | OP | Σ |\n")
		out.WriteString("|---|---|---|---:|---:|---:|---:|---:|\n")
	}
	for _, row := range rows {
		fr := row.frontier.cells[roleDimension]
		fp := row.frontier.cells[personalityDimension]
		or := row.oss.cells[roleDimension]
		op := row.oss.cells[personalityDimension]
		total := fr.total + fp.total + or.total + op.total
		maximum := fr.maximum + fp.maximum + or.maximum + op.maximum
		if uniformModels {
			fmt.Fprintf(
				&out,
				"| %s | %s | %s | %s | %s | %d/%d |\n",
				markdownCell(row.role),
				formatScorecardCell(fr),
				formatScorecardCell(fp),
				formatScorecardCell(or),
				formatScorecardCell(op),
				total,
				maximum,
			)
			continue
		}
		fmt.Fprintf(
			&out,
			"| %s | %s | %s | %s | %s | %s | %s | %d/%d |\n",
			markdownCell(row.role),
			markdownCell(row.frontier.name),
			markdownCell(row.oss.name),
			formatScorecardCell(fr),
			formatScorecardCell(fp),
			formatScorecardCell(or),
			formatScorecardCell(op),
			total,
			maximum,
		)
	}
	out.WriteString("\n`F` frontier · `O` OSS · `R` role · `P` personality · `✓` pass · `×` fail. Cells are points")
	if allCellsUseDefaultTotal(rows) {
		fmt.Fprintf(&out, " out of %d", defaultCellMaximum)
	}
	out.WriteString(".\n")
	return out.Bytes(), nil
}

func historicalScorecardCases(result *ScoredResult) ([]Case, error) {
	if result == nil || result.Format != ResultFormatV2 ||
		result.Role == "" || result.Seat == "" ||
		result.Provenance.EvaluatedAt == "" ||
		result.Provenance.SourceRevision == "" ||
		result.Provenance.PackDigest == "" {
		return nil, fmt.Errorf("historical result identity or provenance is incomplete")
	}
	cases := make([]Case, 0, len(result.Cases))
	for _, scored := range result.Cases {
		tier, dimension, ok := strings.Cut(scored.ID, "-")
		if !ok || (tier != frontierTier && tier != ossTier) ||
			(dimension != roleDimension && dimension != personalityDimension) {
			return nil, fmt.Errorf("case %q does not encode a supported tier and dimension", scored.ID)
		}
		if scored.Model == "" || strings.TrimSpace(scored.RawResponse) == "" ||
			len(scored.Scores) == 0 {
			return nil, fmt.Errorf("case %q has incomplete evidence", scored.ID)
		}
		total := 0
		rubric := make([]Criterion, 0, len(scored.Scores))
		for _, score := range scored.Scores {
			if score.Criterion == "" || score.Evidence == "" || score.Score < 0 || score.Score > 2 {
				return nil, fmt.Errorf("case %q has invalid criterion evidence", scored.ID)
			}
			total += score.Score
			rubric = append(rubric, Criterion{ID: score.Criterion})
		}
		if total != scored.Total {
			return nil, fmt.Errorf("case %q total is %d, want %d", scored.ID, scored.Total, total)
		}
		cases = append(cases, Case{
			ID: scored.ID, ModelTier: tier, Dimension: dimension, Rubric: rubric,
		})
	}
	return cases, nil
}

func sortedScorecardModels(models map[string]*scorecardModel) []*scorecardModel {
	names := make([]string, 0, len(models))
	for name := range models {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]*scorecardModel, 0, len(names))
	for _, name := range names {
		out = append(out, models[name])
	}
	return out
}

func validateScorecardResult(result scorecardResult) error {
	if len(result.frontier) == 0 || len(result.oss) == 0 {
		return fmt.Errorf("frontier and OSS models are required")
	}
	for _, models := range [][]*scorecardModel{result.frontier, result.oss} {
		for _, model := range models {
			for _, dimension := range []string{roleDimension, personalityDimension} {
				if model.cells[dimension] == nil {
					return fmt.Errorf(
						"model %q has no %s score",
						model.name,
						dimension,
					)
				}
			}
		}
	}
	return nil
}

func formatScorecardCell(cell *scorecardCell) string {
	mark := "×"
	if cell.passed {
		mark = "✓"
	}
	if cell.maximum == defaultCellMaximum {
		return fmt.Sprintf("%d%s", cell.total, mark)
	}
	return fmt.Sprintf("%d/%d%s", cell.total, cell.maximum, mark)
}

func allCellsUseDefaultTotal(rows []scorecardRow) bool {
	for _, row := range rows {
		for _, cell := range []*scorecardCell{
			row.frontier.cells[roleDimension],
			row.frontier.cells[personalityDimension],
			row.oss.cells[roleDimension],
			row.oss.cells[personalityDimension],
		} {
			if cell.maximum != defaultCellMaximum {
				return false
			}
		}
	}
	return true
}

func markdownCell(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, "|", `\|`)
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
