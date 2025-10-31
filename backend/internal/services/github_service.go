package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coding-jr/planto-reviewer/backend/internal/models"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type GitHubService struct {
	db *gorm.DB
}

func NewGitHubService(db *gorm.DB) *GitHubService {
	return &GitHubService{db: db}
}

// FetchNewPRs fetches new or updated PRs for all active organizations
func (s *GitHubService) FetchNewPRs() error {
	var orgs []models.Organization
	if err := s.db.Where("is_active = ?", true).Find(&orgs).Error; err != nil {
		return fmt.Errorf("failed to fetch organizations: %w", err)
	}

	for _, org := range orgs {
		if err := s.fetchOrgPRs(&org); err != nil {
			// Log error but continue with other orgs
			fmt.Printf("❌ Failed to fetch PRs for org %s: %v\n", org.GithubOrgName, err)
			continue
		}
	}

	return nil
}

// fetchOrgPRs fetches PRs for a single organization
func (s *GitHubService) fetchOrgPRs(org *models.Organization) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: org.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Get all repositories for this org
	var repos []models.Repository
	if err := s.db.Where("organization_id = ? AND is_active = ?", org.ID, true).Find(&repos).Error; err != nil {
		return fmt.Errorf("failed to fetch repositories: %w", err)
	}

	for _, repo := range repos {
		if err := s.fetchRepoPRs(ctx, client, org, &repo); err != nil {
			fmt.Printf("❌ Failed to fetch PRs for repo %s/%s: %v\n", org.GithubOrgName, repo.Name, err)
			continue
		}
	}

	return nil
}

// fetchRepoPRs fetches PRs for a single repository
func (s *GitHubService) fetchRepoPRs(ctx context.Context, client *github.Client, org *models.Organization, repo *models.Repository) error {
	// Fetch open and recently closed PRs
	opts := &github.PullRequestListOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 50},
	}

	prs, _, err := client.PullRequests.List(ctx, org.GithubOrgName, repo.Name, opts)
	if err != nil {
		return fmt.Errorf("failed to list PRs: %w", err)
	}

	for _, ghPR := range prs {
		if err := s.savePR(ctx, client, org, repo, ghPR); err != nil {
			fmt.Printf("❌ Failed to save PR #%d: %v\n", *ghPR.Number, err)
			continue
		}
	}

	return nil
}

// savePR saves or updates a PR in the database
func (s *GitHubService) savePR(ctx context.Context, client *github.Client, org *models.Organization, repo *models.Repository, ghPR *github.PullRequest) error {
	// Check if PR already exists
	var existingPR models.PullRequest
	err := s.db.Where("repository_id = ? AND pr_number = ?", repo.ID, *ghPR.Number).First(&existingPR).Error

	if err == gorm.ErrRecordNotFound {
		// Create new PR
		return s.createPR(ctx, client, org, repo, ghPR)
	} else if err != nil {
		return fmt.Errorf("failed to check existing PR: %w", err)
	}

	// Update if PR was updated after our last check
	if ghPR.UpdatedAt != nil && existingPR.GithubUpdatedAt.Before(*ghPR.UpdatedAt) {
		return s.updatePR(ctx, client, org, &existingPR, ghPR)
	}

	return nil
}

// createPR creates a new PR record
func (s *GitHubService) createPR(ctx context.Context, client *github.Client, org *models.Organization, repo *models.Repository, ghPR *github.PullRequest) error {
	// Get or create developer
	dev, err := s.getOrCreateDeveloper(org, ghPR.User)
	if err != nil {
		return fmt.Errorf("failed to get/create developer: %w", err)
	}

	// Fetch PR diff
	diff, err := s.fetchPRDiff(ctx, client, org.GithubOrgName, repo.Name, *ghPR.Number)
	if err != nil {
		return fmt.Errorf("failed to fetch PR diff: %w", err)
	}

	// Get stats
	additions := 0
	deletions := 0
	changedFiles := 0
	if ghPR.Additions != nil {
		additions = *ghPR.Additions
	}
	if ghPR.Deletions != nil {
		deletions = *ghPR.Deletions
	}
	if ghPR.ChangedFiles != nil {
		changedFiles = *ghPR.ChangedFiles
	}

	pr := &models.PullRequest{
		RepositoryID:     repo.ID,
		OrganizationID:   org.ID,
		DeveloperID:      dev.ID,
		PRNumber:         uint(*ghPR.Number),
		Title:            *ghPR.Title,
		State:            *ghPR.State,
		GithubCreatedAt:  *ghPR.CreatedAt,
		GithubUpdatedAt:  *ghPR.UpdatedAt,
		LinesAdded:       additions,
		LinesDeleted:     deletions,
		FilesChanged:     changedFiles,
		PRDiff:           &diff,
		NeedsReview:      true,
	}

	if ghPR.MergedAt != nil {
		pr.GithubMergedAt = ghPR.MergedAt
	}

	if err := s.db.Create(pr).Error; err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	fmt.Printf("✅ Created PR #%d: %s\n", pr.PRNumber, pr.Title)
	return nil
}

// updatePR updates an existing PR record
func (s *GitHubService) updatePR(ctx context.Context, client *github.Client, org *models.Organization, pr *models.PullRequest, ghPR *github.PullRequest) error {
	// Fetch latest diff if PR was updated
	diff, err := s.fetchPRDiff(ctx, client, org.GithubOrgName, "", int(*ghPR.Number))
	if err != nil {
		return fmt.Errorf("failed to fetch PR diff: %w", err)
	}

	updates := map[string]interface{}{
		"state":              *ghPR.State,
		"github_updated_at":  *ghPR.UpdatedAt,
		"pr_diff":            diff,
		"needs_review":       true, // Mark for re-review
	}

	if ghPR.MergedAt != nil {
		updates["github_merged_at"] = *ghPR.MergedAt
	}

	if err := s.db.Model(pr).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update PR: %w", err)
	}

	fmt.Printf("✅ Updated PR #%d\n", pr.PRNumber)
	return nil
}

// fetchPRDiff fetches the diff for a PR
func (s *GitHubService) fetchPRDiff(ctx context.Context, client *github.Client, owner, repo string, prNumber int) (string, error) {
	diff, _, err := client.PullRequests.GetRaw(ctx, owner, repo, prNumber, github.RawOptions{Type: github.Diff})
	if err != nil {
		return "", fmt.Errorf("failed to fetch raw diff: %w", err)
	}

	return diff, nil
}

// getOrCreateDeveloper gets or creates a developer record
func (s *GitHubService) getOrCreateDeveloper(org *models.Organization, user *github.User) (*models.Developer, error) {
	var dev models.Developer
	err := s.db.Where("organization_id = ? AND github_username = ?", org.ID, *user.Login).First(&dev).Error

	if err == gorm.ErrRecordNotFound {
		// Create new developer
		githubID := uint64(*user.ID)
		dev = models.Developer{
			OrganizationID: org.ID,
			GithubUsername: *user.Login,
			GithubID:       &githubID,
			Email:          user.Email,
			Name:           user.Name,
		}

		if err := s.db.Create(&dev).Error; err != nil {
			return nil, fmt.Errorf("failed to create developer: %w", err)
		}

		fmt.Printf("✅ Created developer: %s\n", dev.GithubUsername)
		return &dev, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check existing developer: %w", err)
	}

	return &dev, nil
}

// GetPRsNeedingReview returns PRs that need AI review
func (s *GitHubService) GetPRsNeedingReview(limit int) ([]models.PullRequest, error) {
	var prs []models.PullRequest
	err := s.db.
		Where("needs_review = ? AND pr_diff IS NOT NULL", true).
		Order("github_updated_at DESC").
		Limit(limit).
		Find(&prs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch PRs needing review: %w", err)
	}

	return prs, nil
}

// MarkPRReviewed marks a PR as reviewed
func (s *GitHubService) MarkPRReviewed(prID uint64) error {
	return s.db.Model(&models.PullRequest{}).
		Where("id = ?", prID).
		Update("needs_review", false).
		Error
}
