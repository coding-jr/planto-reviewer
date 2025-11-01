-- Database schema for Developer KPI Tracking System
-- MySQL 8.0+

CREATE DATABASE IF NOT EXISTS code_quality_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE code_quality_dev;

-- Organizations (Multi-tenant)
CREATE TABLE organizations (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    github_org_name VARCHAR(255) UNIQUE NOT NULL,
    github_token VARCHAR(512) NOT NULL COMMENT 'Encrypted token',
    settings JSON,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_github_org (github_org_name),
    INDEX idx_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Repositories
CREATE TABLE repositories (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    github_id BIGINT UNSIGNED UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    full_name VARCHAR(512) NOT NULL,
    language VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    last_synced_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    INDEX idx_org_repo (organization_id, name),
    INDEX idx_github_id (github_id),
    INDEX idx_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Developers
CREATE TABLE developers (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    github_username VARCHAR(255) NOT NULL,
    github_id BIGINT UNSIGNED,
    email VARCHAR(255),
    name VARCHAR(255),
    avatar_url VARCHAR(512),
    total_prs INT DEFAULT 0,
    total_commits INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    UNIQUE KEY uk_org_username (organization_id, github_username),
    INDEX idx_github_username (github_username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Pull Requests
CREATE TABLE pull_requests (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    repository_id BIGINT UNSIGNED NOT NULL,
    developer_id BIGINT UNSIGNED NOT NULL,
    github_id BIGINT UNSIGNED UNIQUE NOT NULL,
    pr_number INT NOT NULL,
    title VARCHAR(512) NOT NULL,
    state VARCHAR(50) NOT NULL,
    base_branch VARCHAR(255),
    head_branch VARCHAR(255),
    lines_added INT DEFAULT 0,
    lines_deleted INT DEFAULT 0,
    files_changed INT DEFAULT 0,
    commits_count INT DEFAULT 0,
    opened_at TIMESTAMP NOT NULL,
    closed_at TIMESTAMP NULL,
    merged_at TIMESTAMP NULL,
    needs_review BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (developer_id) REFERENCES developers(id) ON DELETE CASCADE,
    INDEX idx_org_repo (organization_id, repository_id),
    INDEX idx_developer (developer_id),
    INDEX idx_state (state),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Reviews
CREATE TABLE reviews (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    pull_request_id BIGINT UNSIGNED NOT NULL,
    review_type VARCHAR(50) NOT NULL,
    model_used VARCHAR(100) NOT NULL,
    tokens_used INT DEFAULT 0,
    cost_usd DECIMAL(10, 6) DEFAULT 0.000000,
    status VARCHAR(50) NOT NULL,
    summary_text TEXT,
    completed_at TIMESTAMP NULL,
    error_message TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests(id) ON DELETE CASCADE,
    INDEX idx_org_pr (organization_id, pull_request_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Issues Found (KPI tracking)
CREATE TABLE issues_found (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    review_id BIGINT UNSIGNED NOT NULL,
    pull_request_id BIGINT UNSIGNED NOT NULL,
    developer_id BIGINT UNSIGNED NOT NULL,
    file_path VARCHAR(1024) NOT NULL,
    line_number INT,
    issue_type VARCHAR(100) NOT NULL COMMENT 'null_check, logic_error, scalability, security, etc',
    severity VARCHAR(50) NOT NULL COMMENT 'critical, high, medium, low',
    title VARCHAR(512) NOT NULL,
    description TEXT NOT NULL,
    suggestion TEXT,
    code_snippet TEXT,
    is_resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (review_id) REFERENCES reviews(id) ON DELETE CASCADE,
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests(id) ON DELETE CASCADE,
    FOREIGN KEY (developer_id) REFERENCES developers(id) ON DELETE CASCADE,
    INDEX idx_org_developer (organization_id, developer_id),
    INDEX idx_issue_type (issue_type),
    INDEX idx_severity (severity),
    INDEX idx_created_at (created_at),
    INDEX idx_resolved (is_resolved)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Developer Metrics (aggregated for fast queries)
CREATE TABLE developer_metrics (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    developer_id BIGINT UNSIGNED NOT NULL,
    metric_date DATE NOT NULL,
    total_prs INT DEFAULT 0,
    total_issues_found INT DEFAULT 0,
    critical_issues INT DEFAULT 0,
    high_issues INT DEFAULT 0,
    medium_issues INT DEFAULT 0,
    low_issues INT DEFAULT 0,
    null_check_issues INT DEFAULT 0,
    logic_error_issues INT DEFAULT 0,
    scalability_issues INT DEFAULT 0,
    security_issues INT DEFAULT 0,
    performance_issues INT DEFAULT 0,
    lines_added INT DEFAULT 0,
    lines_deleted INT DEFAULT 0,
    avg_pr_size DECIMAL(10, 2) DEFAULT 0.00,
    code_quality_score DECIMAL(5, 2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (developer_id) REFERENCES developers(id) ON DELETE CASCADE,
    UNIQUE KEY uk_dev_date (developer_id, metric_date),
    INDEX idx_org_date (organization_id, metric_date),
    INDEX idx_code_quality (code_quality_score)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Organization Metrics
CREATE TABLE organization_metrics (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    metric_date DATE NOT NULL,
    total_prs INT DEFAULT 0,
    total_developers INT DEFAULT 0,
    total_issues_found INT DEFAULT 0,
    avg_issues_per_pr DECIMAL(5, 2) DEFAULT 0.00,
    avg_code_quality_score DECIMAL(5, 2) DEFAULT 0.00,
    total_cost_usd DECIMAL(10, 2) DEFAULT 0.00,
    total_tokens_used BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    UNIQUE KEY uk_org_date (organization_id, metric_date),
    INDEX idx_date (metric_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;