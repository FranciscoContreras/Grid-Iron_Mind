package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/francisco/gridironmind/internal/styleagent"
	"github.com/francisco/gridironmind/pkg/response"
)

// StyleAgentHandler handles style checking requests
type StyleAgentHandler struct {
	agent *styleagent.StyleAgent
}

// NewStyleAgentHandler creates a new handler
func NewStyleAgentHandler() *StyleAgentHandler {
	return &StyleAgentHandler{
		agent: styleagent.NewStyleAgent(),
	}
}

// HandleStyleCheck handles POST /api/v1/style/check
func (h *StyleAgentHandler) HandleStyleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, 405, "METHOD_NOT_ALLOWED", "Only POST is allowed")
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.BadRequest(w, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	// Parse request
	var req struct {
		Content     string `json:"content"`
		ContentType string `json:"content_type"` // "html", "css", "jsx"
		File        string `json:"file,omitempty"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		response.BadRequest(w, "Invalid JSON")
		return
	}

	if req.Content == "" {
		response.BadRequest(w, "Content is required")
		return
	}

	// Analyze content
	var result *styleagent.StyleAnalysisResult

	switch req.ContentType {
	case "css":
		result, err = h.agent.AnalyzeCSS(r.Context(), req.Content)
	case "html", "jsx", "":
		result, err = h.agent.AnalyzeHTML(r.Context(), req.Content)
	default:
		response.BadRequest(w, "Invalid content_type. Use 'html' or 'css'")
		return
	}

	if err != nil {
		response.InternalError(w, "Style analysis failed")
		return
	}

	response.Success(w, result)
}

// HandleStyleGuide serves the style guide HTML
func (h *StyleAgentHandler) HandleStyleGuide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, 405, "METHOD_NOT_ALLOWED", "Only GET is allowed")
		return
	}

	// Serve the style guide HTML file
	http.ServeFile(w, r, "./dashboard/style-guide.html")
}

// HandleStyleRules returns all available style rules
func (h *StyleAgentHandler) HandleStyleRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, 405, "METHOD_NOT_ALLOWED", "Only GET is allowed")
		return
	}

	rules := []map[string]string{
		{
			"id":          "no-blue-colors",
			"name":        "No Blue in Generic UI",
			"description": "Blue colors should only appear in team-specific contexts",
			"severity":    "error",
		},
		{
			"id":          "no-colored-shadows",
			"name":        "Monochrome Shadows",
			"description": "Shadows should use black rgba, not colored",
			"severity":    "warning",
		},
		{
			"id":          "monochrome-focus",
			"name":        "Monochrome Focus States",
			"description": "Focus states should use black glow, not blue",
			"severity":    "error",
		},
		{
			"id":          "font-family-outfit",
			"name":        "Outfit Font for Body",
			"description": "Use Outfit font for all text except buttons",
			"severity":    "warning",
		},
		{
			"id":          "pixelify-buttons-only",
			"name":        "Pixelify Sans for Buttons Only",
			"description": "Pixelify Sans should only be used on buttons",
			"severity":    "error",
		},
		{
			"id":          "backdrop-filter-required",
			"name":        "Backdrop Filter on Translucent Elements",
			"description": "Elements with opacity < 1 should have backdrop-filter: blur(10px)",
			"severity":    "warning",
		},
		{
			"id":          "button-structure",
			"name":        "Button 3-Layer Structure",
			"description": "Buttons must have btn-skeu > btn-cap > btn-text structure",
			"severity":    "error",
		},
		{
			"id":          "shadow-opacity",
			"name":        "Low Opacity Shadows",
			"description": "Shadow opacity should be 0.04-0.12 for subtle effects",
			"severity":    "info",
		},
		{
			"id":          "border-radius-scale",
			"name":        "Consistent Border Radius",
			"description": "Use standard border-radius values: 12px, 14px, 16px, 24px",
			"severity":    "info",
		},
	}

	response.Success(w, map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}

// HandleStyleExample provides example HTML/CSS
func (h *StyleAgentHandler) HandleStyleExample(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, 405, "METHOD_NOT_ALLOWED", "Only GET is allowed")
		return
	}

	exampleType := r.URL.Query().Get("type")

	examples := map[string]string{
		"button": `<button class="btn-skeu">
  <span class="btn-cap">
    <span class="btn-text">Button Text</span>
  </span>
</button>`,
		"button-rainbow": `<button class="btn-skeu btn-rainbow">
  <span class="btn-cap">
    <span class="btn-text">Add to Cart</span>
  </span>
</button>`,
		"button-team": `<button class="btn-skeu btn-team-chiefs">
  <span class="btn-cap">
    <span class="btn-text">View Chiefs Roster</span>
  </span>
</button>`,
		"input": `.input-skeu {
  width: 100%;
  padding: 14px 18px;
  font-family: 'Outfit', sans-serif;
  background: rgb(255 255 255 / 0.8);
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 14px;
  backdrop-filter: blur(10px);
}

.input-skeu:focus {
  border-color: rgba(0, 0, 0, 0.2);
  box-shadow: 0 0 0 4px rgba(0, 0, 0, 0.08);
}`,
		"card": `.card {
  background: rgb(255 255 255 / 0.8);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 24px;
  padding: 32px;
  box-shadow:
    0 4px 12px rgba(0, 0, 0, 0.06),
    0 2px 6px rgba(0, 0, 0, 0.04);
  backdrop-filter: blur(10px);
}`,
	}

	if exampleType == "" {
		response.Success(w, map[string]interface{}{
			"available": []string{"button", "button-rainbow", "button-team", "input", "card"},
		})
		return
	}

	example, ok := examples[exampleType]
	if !ok {
		response.BadRequest(w, "Invalid example type")
		return
	}

	response.Success(w, map[string]string{
		"type":    exampleType,
		"example": example,
	})
}
