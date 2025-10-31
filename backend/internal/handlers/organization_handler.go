package handlers

import (
	"encoding/json"
	"github.com/coding-jr/planto-reviewer/backend/internal/models"
	"github.com/coding-jr/planto-reviewer/backend/internal/repositories"
	"github.com/gofiber/fiber/v2"
)

type OrganizationHandler struct {
	repo *repositories.OrganizationRepository
}

func NewOrganizationHandler(repo *repositories.OrganizationRepository) *OrganizationHandler {
	return &OrganizationHandler{repo: repo}
}

// POST /api/organizations
func (h *OrganizationHandler) Create(c *fiber.Ctx) error {
	var req struct {
		Name          string   `json:"name"`
		GithubOrgName string   `json:"github_org_name"`
		GithubToken   string   `json:"github_token"`
		Repos         []string `json:"repos"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" || req.GithubOrgName == "" || req.GithubToken == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "name, github_org_name, and github_token are required",
		})
	}

	// Build settings
	settings := models.OrganizationSettings{
		Repos:      req.Repos,
		AutoReview: true,
	}
	settingsJSON, _ := json.Marshal(settings)

	org := &models.Organization{
		Name:          req.Name,
		GithubOrgName: req.GithubOrgName,
		GithubToken:   req.GithubToken, // TODO: Encrypt in production
		Settings:      string(settingsJSON),
		IsActive:      true,
	}

	if err := h.repo.Create(org); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create organization: " + err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Organization created successfully",
		"data":    org,
	})
}

// GET /api/organizations
func (h *OrganizationHandler) List(c *fiber.Ctx) error {
	orgs, err := h.repo.FindAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch organizations",
		})
	}

	return c.JSON(fiber.Map{
		"data": orgs,
	})
}

// GET /api/organizations/:id
func (h *OrganizationHandler) Get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	org, err := h.repo.FindByID(uint64(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Organization not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": org,
	})
}

// PUT /api/organizations/:id
func (h *OrganizationHandler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	org, err := h.repo.FindByID(uint64(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Organization not found",
		})
	}

	var req struct {
		Name     string `json:"name"`
		IsActive *bool  `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name != "" {
		org.Name = req.Name
	}
	if req.IsActive != nil {
		org.IsActive = *req.IsActive
	}

	if err := h.repo.Update(org); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update organization",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Organization updated successfully",
		"data":    org,
	})
}

// DELETE /api/organizations/:id
func (h *OrganizationHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	if err := h.repo.Delete(uint64(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete organization",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Organization deleted successfully",
	})
}
