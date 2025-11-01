package services

import (
	"context"
	"fmt"

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
	if ghPR.UpdatedAt != nil && existingPR.GithubUpdatedAt.Before(ghPR.UpdatedAt.Time) {
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
		GithubID:         uint64(*ghPR.ID),
		PRNumber:         uint(*ghPR.Number),
		Title:            *ghPR.Title,
		State:            *ghPR.State,
		GithubCreatedAt:  ghPR.CreatedAt.Time,
		GithubUpdatedAt:  ghPR.UpdatedAt.Time,
		OpenedAt:         ghPR.CreatedAt.Time,
		LinesAdded:       additions,
		LinesDeleted:     deletions,
		FilesChanged:     changedFiles,
		PRDiff:           &diff,
		NeedsReview:      true,
	}

	if ghPR.MergedAt != nil {
		mergedAt := ghPR.MergedAt.Time
		pr.GithubMergedAt = &mergedAt
		pr.MergedAt = &mergedAt
	}
	if ghPR.ClosedAt != nil {
		closedAt := ghPR.ClosedAt.Time
		pr.ClosedAt = &closedAt
	}

	if err := s.db.Create(pr).Error; err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	fmt.Printf("✅ Created PR #%d: %s\n", pr.PRNumber, pr.Title)
	return nil
}

// updatePR updates an existing PR record
func (s *GitHubService) updatePR(ctx context.Context, client *github.Client, org *models.Organization, pr *models.PullRequest, ghPR *github.PullRequest) error {
	// Get repo name
	var repo models.Repository
	if err := s.db.First(&repo, pr.RepositoryID).Error; err != nil {
		return fmt.Errorf("failed to fetch repository: %w", err)
	}

	// Fetch latest diff if PR was updated
	diff, err := s.fetchPRDiff(ctx, client, org.GithubOrgName, repo.Name, int(*ghPR.Number))
	if err != nil {
		return fmt.Errorf("failed to fetch PR diff: %w", err)
	}

	updates := map[string]interface{}{
		"state":              *ghPR.State,
		"github_updated_at":  ghPR.UpdatedAt.Time,
		"pr_diff":            diff,
		"needs_review":       true, // Mark for re-review
	}

	if ghPR.MergedAt != nil {
		mergedAt := ghPR.MergedAt.Time
		updates["github_merged_at"] = mergedAt
		updates["merged_at"] = mergedAt
	}

	if ghPR.ClosedAt != nil {
		updates["closed_at"] = ghPR.ClosedAt.Time
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

// PostReviewComment posts a review comment on a GitHub PR
func (s *GitHubService) PostReviewComment(org *models.Organization, repo *models.Repository, prNumber uint, reviewSummary string, issues []models.Issue) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: org.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Format the review comment
	comment := formatReviewComment(reviewSummary, issues)

	// Post comment on PR
	prComment := &github.IssueComment{
		Body: github.String(comment),
	}

	_, _, err := client.Issues.CreateComment(ctx, org.GithubOrgName, repo.Name, int(prNumber), prComment)
	if err != nil {
		return fmt.Errorf("failed to post comment: %w", err)
	}

	fmt.Printf("✅ Posted review comment on PR #%d\n", prNumber)
	return nil
}

// formatReviewComment formats the review into a GitHub comment
func formatReviewComment(summary string, issues []models.Issue) string {
	comment := "## 🤖 AI Code Review\n\n"
	comment += "### Summary\n"
	comment += summary + "\n\n"

	if len(issues) > 0 {
		comment += fmt.Sprintf("### Issues Found (%d)\n\n", len(issues))

		// Group by severity
		criticalIssues := []models.Issue{}
		highIssues := []models.Issue{}
		mediumIssues := []models.Issue{}
		lowIssues := []models.Issue{}

		for _, issue := range issues {
			switch issue.Severity {
			case "critical":
				criticalIssues = append(criticalIssues, issue)
			case "high":
				highIssues = append(highIssues, issue)
			case "medium":
				mediumIssues = append(mediumIssues, issue)
			case "low":
				lowIssues = append(lowIssues, issue)
			}
		}

		// Format by severity
		if len(criticalIssues) > 0 {
			comment += "#### 🔴 Critical Issues\n"
			for _, issue := range criticalIssues {
				comment += formatIssue(issue)
			}
		}

		if len(highIssues) > 0 {
			comment += "#### 🟠 High Priority Issues\n"
			for _, issue := range highIssues {
				comment += formatIssue(issue)
			}
		}

		if len(mediumIssues) > 0 {
			comment += "#### 🟡 Medium Priority Issues\n"
			for _, issue := range mediumIssues {
				comment += formatIssue(issue)
			}
		}

		if len(lowIssues) > 0 {
			comment += "#### 🟢 Low Priority Issues\n"
			for _, issue := range lowIssues {
				comment += formatIssue(issue)
			}
		}
	} else {
		comment += "### ✅ No Issues Found\n\nGreat job! The code looks good.\n"
	}

	comment += "\n---\n*Powered by AWS Bedrock Claude Sonnet 4.5*"
	return comment
}

// formatIssue formats a single issue
func formatIssue(issue models.Issue) string {
	var formatted string
	formatted += fmt.Sprintf("- **%s** (%s)\n", issue.Title, issue.IssueType)
	
	if issue.FilePath != "" {
		formatted += fmt.Sprintf("  - **File:** `%s`", issue.FilePath)
		if issue.LineNumber != nil {
			formatted += fmt.Sprintf(" (Line %d)", *issue.LineNumber)
		}
		formatted += "\n"
	}
	
	formatted += fmt.Sprintf("  - **Description:** %s\n", issue.Description)
	
	if issue.Suggestion != nil && *issue.Suggestion != "" {
		formatted += fmt.Sprintf("  - **Suggestion:** %s\n", *issue.Suggestion)
	}
	
	if issue.CodeSnippet != nil && *issue.CodeSnippet != "" {
		formatted += fmt.Sprintf("  - **Code:**\n```\n%s\n```\n", *issue.CodeSnippet)
	}
	
	formatted += "\n"
	return formatted
}
