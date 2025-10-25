package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
)

// ModerationRulesHandler handles moderation rules API requests
type ModerationRulesHandler struct {
	rulesRepo *repository.ModerationRulesRepository
}

// NewModerationRulesHandler creates a new handler
func NewModerationRulesHandler(db *sql.DB) *ModerationRulesHandler {
	return &ModerationRulesHandler{
		rulesRepo: repository.NewModerationRulesRepository(db),
	}
}

// ListRules returns a list of moderation rules with filtering and search
// GET /api/moderation-rules
// Query parameters:
//   - category: filter by category
//   - risk_level: filter by risk level (L/M/H/C)
//   - search: search by rule code or description
//   - page: page number (default 1)
//   - page_size: items per page (default 20, max 100)
func (h *ModerationRulesHandler) ListRules(c *gin.Context) {
	// Parse query parameters
	req := &models.ListModerationRulesRequest{
		Category:  c.Query("category"),
		RiskLevel: c.Query("risk_level"),
		Search:    c.Query("search"),
		Page:      1,
		PageSize:  20,
	}

	// Parse page number
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}

	// Parse page size
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = pageSize
		}
	}

	// Get rules from repository
	rules, total, err := h.rulesRepo.ListRules(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch rules",
		})
		return
	}

	// Calculate total pages
	totalPages := (total + req.PageSize - 1) / req.PageSize

	// Return response
	c.JSON(http.StatusOK, models.ListModerationRulesResponse{
		Data:       rules,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	})
}

// GetRuleByCode returns a single rule by its code
// GET /api/moderation-rules/:code
func (h *ModerationRulesHandler) GetRuleByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rule code is required",
		})
		return
	}

	rule, err := h.rulesRepo.GetRuleByCode(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch rule",
		})
		return
	}

	if rule == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Rule not found",
		})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// GetCategories returns all available categories
// GET /api/moderation-rules/categories
func (h *ModerationRulesHandler) GetCategories(c *gin.Context) {
	categories, err := h.rulesRepo.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch categories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}

// GetRiskLevels returns all available risk levels
// GET /api/moderation-rules/risk-levels
func (h *ModerationRulesHandler) GetRiskLevels(c *gin.Context) {
	levels, err := h.rulesRepo.GetRiskLevels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch risk levels",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"levels": levels,
	})
}
