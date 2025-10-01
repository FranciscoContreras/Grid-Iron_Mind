package styleagent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// StyleViolation represents a style guideline violation
type StyleViolation struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"` // "error", "warning", "info"
	Line        int    `json:"line,omitempty"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion"`
	RuleID      string `json:"rule_id"`
}

// StyleAnalysisResult contains the full analysis
type StyleAnalysisResult struct {
	Valid       bool              `json:"valid"`
	Violations  []StyleViolation  `json:"violations"`
	Score       float64           `json:"score"` // 0-100
	Summary     string            `json:"summary"`
}

// StyleAgent manages style consistency
type StyleAgent struct {
	rules []StyleRule
}

// StyleRule defines a rule for checking
type StyleRule struct {
	ID          string
	Name        string
	Description string
	Check       func(content string) []StyleViolation
}

// NewStyleAgent creates a new style agent
func NewStyleAgent() *StyleAgent {
	agent := &StyleAgent{}
	agent.initializeRules()
	return agent
}

// initializeRules sets up all style checking rules
func (a *StyleAgent) initializeRules() {
	a.rules = []StyleRule{
		// Color Rules
		{
			ID:          "no-blue-colors",
			Name:        "No Blue in Generic UI",
			Description: "Blue colors should only appear in team-specific contexts",
			Check:       a.checkBlueColors,
		},
		{
			ID:          "no-colored-shadows",
			Name:        "Monochrome Shadows",
			Description: "Shadows should use black rgba, not colored",
			Check:       a.checkColoredShadows,
		},
		{
			ID:          "monochrome-focus",
			Name:        "Monochrome Focus States",
			Description: "Focus states should use black glow, not blue",
			Check:       a.checkFocusStates,
		},

		// Typography Rules
		{
			ID:          "font-family-outfit",
			Name:        "Outfit Font for Body",
			Description: "Use Outfit font for all text except buttons",
			Check:       a.checkFontFamily,
		},
		{
			ID:          "pixelify-buttons-only",
			Name:        "Pixelify Sans for Buttons Only",
			Description: "Pixelify Sans should only be used on buttons",
			Check:       a.checkPixelifySans,
		},

		// Component Rules
		{
			ID:          "backdrop-filter-required",
			Name:        "Backdrop Filter on Translucent Elements",
			Description: "Elements with opacity < 1 should have backdrop-filter: blur(10px)",
			Check:       a.checkBackdropFilter,
		},
		{
			ID:          "button-structure",
			Name:        "Button 3-Layer Structure",
			Description: "Buttons must have btn-skeu > btn-cap > btn-text structure",
			Check:       a.checkButtonStructure,
		},
		{
			ID:          "shadow-opacity",
			Name:        "Low Opacity Shadows",
			Description: "Shadow opacity should be 0.04-0.12 for subtle effects",
			Check:       a.checkShadowOpacity,
		},

		// Border Radius Rules
		{
			ID:          "border-radius-scale",
			Name:        "Consistent Border Radius",
			Description: "Use standard border-radius values: 12px, 14px, 16px, 24px",
			Check:       a.checkBorderRadius,
		},
	}
}

// AnalyzeHTML analyzes HTML content for style violations
func (a *StyleAgent) AnalyzeHTML(ctx context.Context, content string) (*StyleAnalysisResult, error) {
	var allViolations []StyleViolation

	// Run all rules
	for _, rule := range a.rules {
		violations := rule.Check(content)
		allViolations = append(allViolations, violations...)
	}

	// Calculate score
	score := a.calculateScore(allViolations)

	result := &StyleAnalysisResult{
		Valid:      len(allViolations) == 0,
		Violations: allViolations,
		Score:      score,
		Summary:    a.generateSummary(allViolations, score),
	}

	return result, nil
}

// AnalyzeCSS analyzes CSS content for style violations
func (a *StyleAgent) AnalyzeCSS(ctx context.Context, content string) (*StyleAnalysisResult, error) {
	var allViolations []StyleViolation

	// Run all rules on CSS
	for _, rule := range a.rules {
		violations := rule.Check(content)
		allViolations = append(allViolations, violations...)
	}

	score := a.calculateScore(allViolations)

	result := &StyleAnalysisResult{
		Valid:      len(allViolations) == 0,
		Violations: allViolations,
		Score:      score,
		Summary:    a.generateSummary(allViolations, score),
	}

	return result, nil
}

// Rule check implementations

func (a *StyleAgent) checkBlueColors(content string) []StyleViolation {
	var violations []StyleViolation

	// Check for blue color codes
	bluePatterns := []string{
		`#[0-9a-f]{0,2}[0-9a-f]{2}[0-9a-f]{2}[fF]{2}`, // Hex blues ending in high blue
		`rgb\(\s*\d+\s*,\s*\d+\s*,\s*2[0-9]{2}\s*\)`,   // RGB with high blue value
		`rgba\(\s*59\s*,\s*130\s*,\s*246`,              // Specific blue (59, 130, 246)
		`#3b82f6`,                                       // Common blue hex
	}

	for _, pattern := range bluePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringIndex(content, -1)

		for _, match := range matches {
			// Check if it's in a team color context
			contextStart := max(0, match[0]-200)
			contextEnd := min(len(content), match[1]+200)
			context := content[contextStart:contextEnd]

			if !strings.Contains(context, "btn-team") && !strings.Contains(context, "team-color") {
				violations = append(violations, StyleViolation{
					Type:       "color",
					Severity:   "error",
					Message:    "Blue color detected outside of team context",
					Suggestion: "Use monochrome colors (black/white/gray). Only use colors in team-specific contexts.",
					RuleID:     "no-blue-colors",
				})
			}
		}
	}

	return violations
}

func (a *StyleAgent) checkColoredShadows(content string) []StyleViolation {
	var violations []StyleViolation

	// Check for colored box-shadows
	shadowPattern := regexp.MustCompile(`box-shadow[^;]*rgba\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)`)
	matches := shadowPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 4 {
			r, g, b := match[1], match[2], match[3]
			// If RGB values are not all equal (not gray), it's colored
			if r != g || g != b || r != b {
				// Exception for button glow effects
				if strings.Contains(content[max(0, strings.Index(content, match[0])-100):], "btn-") {
					continue
				}

				violations = append(violations, StyleViolation{
					Type:       "shadow",
					Severity:   "warning",
					Message:    "Colored shadow detected",
					Suggestion: "Use black shadows: rgba(0, 0, 0, 0.04-0.12)",
					RuleID:     "no-colored-shadows",
				})
			}
		}
	}

	return violations
}

func (a *StyleAgent) checkFocusStates(content string) []StyleViolation {
	var violations []StyleViolation

	// Check for blue focus states
	focusPattern := regexp.MustCompile(`:focus[^}]*box-shadow[^}]*rgba\(\s*59\s*,\s*130\s*,\s*246`)
	if focusPattern.MatchString(content) {
		violations = append(violations, StyleViolation{
			Type:       "focus-state",
			Severity:   "error",
			Message:    "Blue focus glow detected",
			Suggestion: "Use black focus glow: box-shadow: 0 0 0 4px rgba(0, 0, 0, 0.08)",
			RuleID:     "monochrome-focus",
		})
	}

	return violations
}

func (a *StyleAgent) checkFontFamily(content string) []StyleViolation {
	var violations []StyleViolation

	// Check for non-Outfit fonts in body/inputs/general UI
	fontPattern := regexp.MustCompile(`font-family:\s*['"]?([^'";\n]+)['"]?`)
	matches := fontPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			font := strings.ToLower(strings.TrimSpace(match[1]))

			// Check if context is a button
			matchIndex := strings.Index(content, match[0])
			contextStart := max(0, matchIndex-100)
			contextEnd := min(len(content), matchIndex+100)
			context := content[contextStart:contextEnd]

			isButton := strings.Contains(context, "btn-") || strings.Contains(context, "button")

			if !isButton && !strings.Contains(font, "outfit") &&
			   font != "sans-serif" && font != "inherit" {
				violations = append(violations, StyleViolation{
					Type:       "typography",
					Severity:   "warning",
					Message:    fmt.Sprintf("Non-Outfit font detected: %s", match[1]),
					Suggestion: "Use 'Outfit', sans-serif for all body text and UI elements",
					RuleID:     "font-family-outfit",
				})
			}
		}
	}

	return violations
}

func (a *StyleAgent) checkPixelifySans(content string) []StyleViolation {
	var violations []StyleViolation

	// Check if Pixelify Sans is used outside buttons
	pixelifyPattern := regexp.MustCompile(`font-family[^;]*[Pp]ixelify`)
	matches := pixelifyPattern.FindAllStringIndex(content, -1)

	for _, match := range matches {
		contextStart := max(0, match[0]-100)
		contextEnd := min(len(content), match[1]+100)
		context := content[contextStart:contextEnd]

		if !strings.Contains(context, "btn-") && !strings.Contains(context, "button") {
			violations = append(violations, StyleViolation{
				Type:       "typography",
				Severity:   "error",
				Message:    "Pixelify Sans used outside of button context",
				Suggestion: "Pixelify Sans should only be used for buttons",
				RuleID:     "pixelify-buttons-only",
			})
		}
	}

	return violations
}

func (a *StyleAgent) checkBackdropFilter(content string) []StyleViolation {
	var violations []StyleViolation

	// Find elements with opacity < 1
	opacityPattern := regexp.MustCompile(`(?s)([\w-]+)\s*\{[^}]*opacity:\s*0\.\d+[^}]*\}`)
	matches := opacityPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			blockContent := match[0]

			// Check if it has backdrop-filter
			if !strings.Contains(blockContent, "backdrop-filter") {
				violations = append(violations, StyleViolation{
					Type:       "glassmorphism",
					Severity:   "warning",
					Message:    fmt.Sprintf("Element '%s' has opacity but no backdrop-filter", match[1]),
					Suggestion: "Add backdrop-filter: blur(10px) to translucent elements",
					RuleID:     "backdrop-filter-required",
				})
			}
		}
	}

	return violations
}

func (a *StyleAgent) checkButtonStructure(content string) []StyleViolation {
	var violations []StyleViolation

	// Check for buttons without proper structure
	buttonPattern := regexp.MustCompile(`<button[^>]*class="[^"]*btn-skeu[^"]*"[^>]*>([^<]+)</button>`)
	matches := buttonPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			buttonContent := match[1]

			// Check for btn-cap and btn-text
			if !strings.Contains(buttonContent, "btn-cap") || !strings.Contains(buttonContent, "btn-text") {
				violations = append(violations, StyleViolation{
					Type:       "component",
					Severity:   "error",
					Message:    "Button missing proper 3-layer structure",
					Suggestion: "Use: <button class=\"btn-skeu\"><span class=\"btn-cap\"><span class=\"btn-text\">Text</span></span></button>",
					RuleID:     "button-structure",
				})
			}
		}
	}

	return violations
}

func (a *StyleAgent) checkShadowOpacity(content string) []StyleViolation {
	var violations []StyleViolation

	// Check shadow opacity values
	shadowPattern := regexp.MustCompile(`box-shadow[^;]*rgba\([^,]+,[^,]+,[^,]+,\s*(0\.\d+)`)
	matches := shadowPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			opacity := match[1]
			opacityFloat := 0.0
			fmt.Sscanf(opacity, "%f", &opacityFloat)

			if opacityFloat > 0.15 {
				violations = append(violations, StyleViolation{
					Type:       "shadow",
					Severity:   "info",
					Message:    fmt.Sprintf("Shadow opacity too high: %s", opacity),
					Suggestion: "Keep shadow opacity between 0.04-0.12 for subtle effects",
					RuleID:     "shadow-opacity",
				})
			}
		}
	}

	return violations
}

func (a *StyleAgent) checkBorderRadius(content string) []StyleViolation {
	var violations []StyleViolation

	// Check for non-standard border-radius values
	radiusPattern := regexp.MustCompile(`border-radius:\s*(\d+)px`)
	matches := radiusPattern.FindAllStringSubmatch(content, -1)

	standardValues := map[string]bool{
		"8": true, "10": true, "12": true, "14": true, "16": true, "18": true, "24": true, "50": true,
	}

	for _, match := range matches {
		if len(match) >= 2 {
			value := match[1]
			if !standardValues[value] {
				violations = append(violations, StyleViolation{
					Type:       "spacing",
					Severity:   "info",
					Message:    fmt.Sprintf("Non-standard border-radius: %spx", value),
					Suggestion: "Use standard values: 12px (small), 14px (medium), 16px (large), 24px (cards)",
					RuleID:     "border-radius-scale",
				})
			}
		}
	}

	return violations
}

// Helper functions

func (a *StyleAgent) calculateScore(violations []StyleViolation) float64 {
	if len(violations) == 0 {
		return 100.0
	}

	score := 100.0
	for _, v := range violations {
		switch v.Severity {
		case "error":
			score -= 10.0
		case "warning":
			score -= 5.0
		case "info":
			score -= 2.0
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (a *StyleAgent) generateSummary(violations []StyleViolation, score float64) string {
	if len(violations) == 0 {
		return "✓ All style guidelines followed. Perfect score!"
	}

	errors := 0
	warnings := 0
	infos := 0

	for _, v := range violations {
		switch v.Severity {
		case "error":
			errors++
		case "warning":
			warnings++
		case "info":
			infos++
		}
	}

	return fmt.Sprintf("Score: %.1f/100 | Errors: %d, Warnings: %d, Info: %d", score, errors, warnings, infos)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateStyleFix generates automatic fixes for violations
func (a *StyleAgent) GenerateStyleFix(ctx context.Context, violation StyleViolation, content string) (string, error) {
	// This would use the AI service to generate intelligent fixes
	// For now, return the suggestion
	return violation.Suggestion, nil
}

// ToJSON converts result to JSON
func (r *StyleAnalysisResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
