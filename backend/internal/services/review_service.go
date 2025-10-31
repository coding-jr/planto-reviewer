package services

import (
	"fmt"
	"time"

	"github.com/coding-jr/planto-reviewer/backend/internal/models"
	"github.com/coding-jr/planto-reviewer/backend/pkg/ai"
	"gorm.io/gorm"
)

type ReviewService struct {
	db            *gorm.DB
	aiClient      *ai.Client
	githubService *GitHubService
	metricsService *MetricsService
}

func NewReviewService(db *gorm.DB, aiClient *ai.Client) *ReviewService {
	return &ReviewService{
		db:            db,
		aiClient:      aiClient,
		githubService: NewGitHubService(db),
		metricsService: NewMetricsService(db),
	}
}

// ProcessPendingReviews processes all PRs that need review
func (s *ReviewService) ProcessPendingReviews(batchSize int) error {
	// Get PRs that need review
	prs, err := s.githubService.GetPRsNeedingReview(batchSize)
	if err != nil {
		return fmt.Errorf("failed to fetch PRs needing review: %w", err)
	}

	if len(prs) == 0 {
		fmt.Println("ℹ️  No PRs need review")
		return nil
	}

	fmt.Printf("📝 Processing %d PRs for review\n", len(prs))

	for _, pr := range prs {
		if err := s.reviewPR(&pr); err != nil {
			fmt.Printf("❌ Failed to review PR #%d: %v\n", pr.PRNumber, err)
			continue
		}
		fmt.Printf("✅ Reviewed PR #%d: %s\n", pr.PRNumber, pr.Title)
	}

	return nil
}

// reviewPR reviews a single PR
func (s *ReviewService) reviewPR(pr *models.PullRequest) error {
	// Skip if no diff
	if pr.PRDiff == nil || *pr.PRDiff == "" {
		return fmt.Errorf("PR has no diff")
	}

	// Get repository and org info for context
	var repo models.Repository
	if err := s.db.First(&repo, pr.RepositoryID).Error; err != nil {
		return fmt.Errorf("failed to fetch repository: %w", err)
	}

	var org models.Organization
	if err := s.db.First(&org, pr.OrganizationID).Error; err != nil {
		return fmt.Errorf("failed to fetch organization: %w", err)
	}

	// Build context
	context := fmt.Sprintf("Repository: %s/%s\nPR #%d: %s\nAuthor: Developer ID %d",
		org.GithubOrgName, repo.Name, pr.PRNumber, pr.Title, pr.DeveloperID)

	// Call AI for review
	result, err := s.aiClient.ReviewCode(*pr.PRDiff, context)
	if err != nil {
		return fmt.Errorf("AI review failed: %w", err)
	}

	// Save review result
	if err := s.saveReview(pr, result); err != nil {
		return fmt.Errorf("failed to save review: %w", err)
	}

	// Mark PR as reviewed
	if err := s.githubService.MarkPRReviewed(pr.ID); err != nil {
		return fmt.Errorf("failed to mark PR as reviewed: %w", err)
	}

	// Update metrics
	if err := s.metricsService.UpdateDeveloperMetrics(pr.DeveloperID); err != nil {
		fmt.Printf("⚠️  Failed to update metrics for developer %d: %v\n", pr.DeveloperID, err)
	}

	if err := s.metricsService.UpdateOrganizationMetrics(pr.OrganizationID); err != nil {
		fmt.Printf("⚠️  Failed to update org metrics for org %d: %v\n", pr.OrganizationID, err)
	}

	return nil
}

// saveReview saves review result and issues to database
func (s *ReviewService) saveReview(pr *models.PullRequest, result *ai.ReviewResult) error {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create review record
	review := &models.Review{
		PullRequestID:    pr.ID,
		OrganizationID:   pr.OrganizationID,
		ReviewSummary:    result.Summary,
		CodeQualityScore: result.CodeQualityScore,
		IssuesFoundCount: len(result.Issues),
		ReviewedAt:       time.Now(),
	}

	if err := tx.Create(review).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create review: %w", err)
	}

	// Save issues
	for _, aiIssue := range result.Issues {
		issue := &models.Issue{
			ReviewID:       review.ID,
			PullRequestID:  pr.ID,
			OrganizationID: pr.OrganizationID,
			DeveloperID:    pr.DeveloperID,
			IssueType:      aiIssue.Type,
			Severity:       aiIssue.Severity,
			Title:          aiIssue.Title,
			Description:    aiIssue.Description,
			Suggestion:     aiIssue.Suggestion,
			CodeSnippet:    aiIssue.CodeSnippet,
			LineNumber:     aiIssue.LineNumber,
			FileName:       aiIssue.FileName,
			IsResolved:     false,
		}

		if err := tx.Create(issue).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create issue: %w", err)
		}
	}

	// Update PR with review info
	updates := map[string]interface{}{
		"last_review_id":    review.ID,
		"code_quality_score": result.CodeQualityScore,
		"issues_found_count": len(result.Issues),
	}

	if err := tx.Model(pr).Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update PR: %w", err)
	}

	// Update developer total PRs
	if err := tx.Model(&models.Developer{}).
		Where("id = ?", pr.DeveloperID).
		UpdateColumn("total_prs", gorm.Expr("total_prs + 1")).
		Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update developer PR count: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetReviewByPRID gets review for a PR
func (s *ReviewService) GetReviewByPRID(prID uint64) (*models.Review, error) {
	var review models.Review
	err := s.db.Where("pull_request_id = ?", prID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// GetIssuesByReviewID gets all issues for a review
func (s *ReviewService) GetIssuesByReviewID(reviewID uint64) ([]models.Issue, error) {
	var issues []models.Issue
	err := s.db.Where("review_id = ?", reviewID).Find(&issues).Error
	if err != nil {
		return nil, err
	}
	return issues, nil
}

// GetIssuesByDeveloperID gets all issues for a developer
func (s *ReviewService) GetIssuesByDeveloperID(developerID uint64, limit int) ([]models.Issue, error) {
	var issues []models.Issue
	query := s.db.Where("developer_id = ?", developerID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&issues).Error
	if err != nil {
		return nil, err
	}
	return issues, nil
}

// ResolveIssue marks an issue as resolved
func (s *ReviewService) ResolveIssue(issueID uint64) error {
	return s.db.Model(&models.Issue{}).
		Where("id = ?", issueID).
		Update("is_resolved", true).
		Error
}
