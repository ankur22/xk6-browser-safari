package browser

import (
	"testing"
)

func TestParseSelector(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		want     ParsedSelector
	}{
		{
			name:     "CSS selector (default)",
			selector: "button.submit",
			want:     ParsedSelector{StrategyCSSSelector, "button.submit", true},
		},
		{
			name:     "XPath with prefix",
			selector: "xpath=//button[@type='submit']",
			want:     ParsedSelector{StrategyXPath, "//button[@type='submit']", true},
		},
		{
			name:     "XPath without prefix",
			selector: "//button[@type='submit']",
			want:     ParsedSelector{StrategyXPath, "//button[@type='submit']", true},
		},
		{
			name:     "Text selector",
			selector: "text=Submit Form",
			want:     ParsedSelector{StrategyText, "Submit Form", false},
		},
		{
			name:     "Visible text selector",
			selector: "visible-text=Submit",
			want:     ParsedSelector{StrategyVisibleText, "Submit", false},
		},
		{
			name:     "Data test ID",
			selector: "data-testid=submit-button",
			want:     ParsedSelector{StrategyDataTestID, "submit-button", false},
		},
		{
			name:     "ARIA label",
			selector: "aria-label=Close dialog",
			want:     ParsedSelector{StrategyAriaLabel, "Close dialog", false},
		},
		{
			name:     "ARIA role",
			selector: "role=button",
			want:     ParsedSelector{StrategyRole, "button", false},
		},
		{
			name:     "ID selector",
			selector: "id=submitBtn",
			want:     ParsedSelector{StrategyID, "submitBtn", true},
		},
		{
			name:     "Class selector",
			selector: "class=submit-button",
			want:     ParsedSelector{StrategyClassName, "submit-button", true},
		},
		{
			name:     "Tag selector",
			selector: "tag=button",
			want:     ParsedSelector{StrategyTagName, "button", true},
		},
		{
			name:     "Link text",
			selector: "link=Click Here",
			want:     ParsedSelector{StrategyLinkText, "Click Here", true},
		},
		{
			name:     "Partial link text",
			selector: "partial-link=Click",
			want:     ParsedSelector{StrategyPartialLinkText, "Click", true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSelector(tt.selector)
			if got.Strategy != tt.want.Strategy {
				t.Errorf("ParseSelector(%q).Strategy = %v, want %v", tt.selector, got.Strategy, tt.want.Strategy)
			}
			if got.Value != tt.want.Value {
				t.Errorf("ParseSelector(%q).Value = %v, want %v", tt.selector, got.Value, tt.want.Value)
			}
			if got.IsNative != tt.want.IsNative {
				t.Errorf("ParseSelector(%q).IsNative = %v, want %v", tt.selector, got.IsNative, tt.want.IsNative)
			}
		})
	}
}

func TestGenerateSelectorScript(t *testing.T) {
	tests := []struct {
		name          string
		strategy      SelectorStrategy
		value         string
		wantSubstring string
	}{
		{
			name:          "Text selector",
			strategy:      StrategyText,
			value:         "Submit",
			wantSubstring: "textContent.trim() === \"Submit\"",
		},
		{
			name:          "Visible text selector",
			strategy:      StrategyVisibleText,
			value:         "Submit",
			wantSubstring: "offsetWidth",
		},
		{
			name:          "Data test ID",
			strategy:      StrategyDataTestID,
			value:         "submit-btn",
			wantSubstring: "[data-testid=\"submit-btn\"]",
		},
		{
			name:          "ARIA label",
			strategy:      StrategyAriaLabel,
			value:         "Close",
			wantSubstring: "[aria-label=\"Close\"]",
		},
		{
			name:          "ARIA role",
			strategy:      StrategyRole,
			value:         "button",
			wantSubstring: "[role=\"button\"]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSelectorScript(tt.strategy, tt.value)
			if !contains(got, tt.wantSubstring) {
				t.Errorf("generateSelectorScript(%v, %q) = %v, want to contain %v", tt.strategy, tt.value, got, tt.wantSubstring)
			}
		})
	}
}

func TestIsRegex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Valid regex", "/.*test.*/", true},
		{"Valid regex simple", "/test/", true},
		{"Not regex - no slashes", "test", false},
		{"Not regex - single slash", "/test", false},
		{"Not regex - empty", "", false},
		{"Not regex - only slashes", "/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRegex(tt.input)
			if got != tt.want {
				t.Errorf("IsRegex(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseRegex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid regex", "/test.*/", false},
		{"Valid complex regex", "/[a-z]+\\d{2,}/", false},
		{"Invalid - not regex format", "test", true},
		{"Invalid - bad regex pattern", "/[/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRegex(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRegex(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
