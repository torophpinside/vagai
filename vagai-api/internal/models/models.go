package models

import (
	"time"

	"gorm.io/gorm"
)

type JobStatus string

const (
	JobStatusNew       JobStatus = "new"
	JobStatusMatched   JobStatus = "matched"
	JobStatusAnalyzed  JobStatus = "analyzed"
	JobStatusUnmatched JobStatus = "unmatched"
	JobStatusIgnored   JobStatus = "ignored"
)

type UserRole string

const (
	RoleOwner  UserRole = "owner"
	RoleAdmin  UserRole = "admin"
	RoleMember UserRole = "member"
)

type SubscriptionStatus string

const (
	SubStatusTrial    SubscriptionStatus = "trial"
	SubStatusActive   SubscriptionStatus = "active"
	SubStatusPastDue  SubscriptionStatus = "past_due"
	SubStatusCanceled SubscriptionStatus = "canceled"
	SubStatusExpired  SubscriptionStatus = "expired"
)

type Organization struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	Name             string         `gorm:"size:200;not null" json:"name"`
	Slug             string         `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Plan             string         `gorm:"size:50;default:free" json:"plan"`
	StripeCustomerID string         `gorm:"size:255" json:"-"`
	TrialEndsAt      *time.Time     `json:"trial_ends_at"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

type User struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"size:200;not null" json:"name"`
	Email           string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash    string         `gorm:"size:255;not null" json:"-"`
	Avatar          string         `gorm:"size:500" json:"avatar"`
	Timezone        string         `gorm:"size:50;default:America/Sao_Paulo" json:"timezone"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type Membership struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	OrganizationID uint         `gorm:"uniqueIndex:idx_user_org;not null" json:"organization_id"`
	UserID         uint         `gorm:"uniqueIndex:idx_user_org;not null" json:"user_id"`
	Role           UserRole     `gorm:"size:20;default:member" json:"role"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	Organization   Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
}

type Plan struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"size:100;not null" json:"name"`
	Slug            string    `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	PriceMonthly    int       `gorm:"default:0" json:"price_monthly"`
	PriceYearly     int       `gorm:"default:0" json:"price_yearly"`
	MaxJobs         int       `gorm:"default:100" json:"max_jobs"`
	MaxResumes      int       `gorm:"default:3" json:"max_resumes"`
	MaxSites        int       `gorm:"default:5" json:"max_sites"`
	MaxCrawlsPerDay int       `gorm:"default:10" json:"-"`
	Features        string    `gorm:"type:json" json:"features"`
	StripePriceID   string    `gorm:"size:255" json:"-"`
	CreatedAt       time.Time `json:"created_at"`
}

type Subscription struct {
	ID                   uint               `gorm:"primaryKey" json:"id"`
	OrganizationID       uint               `gorm:"uniqueIndex;not null" json:"organization_id"`
	PlanID               uint               `gorm:"index" json:"plan_id"`
	StripeSubscriptionID string             `gorm:"size:255;uniqueIndex" json:"-"`
	Status               SubscriptionStatus `gorm:"size:20;default:trial" json:"status"`
	CurrentPeriodStart   *time.Time         `json:"current_period_start"`
	CurrentPeriodEnd     *time.Time         `json:"current_period_end"`
	CancelAtPeriodEnd    bool               `gorm:"default:false" json:"cancel_at_period_end"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
	DeletedAt            gorm.DeletedAt     `gorm:"index" json:"-"`
}

type ApiKey struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	OrganizationID uint       `gorm:"index;not null" json:"organization_id"`
	Name           string     `gorm:"size:100;not null" json:"name"`
	KeyHash        string     `gorm:"size:255;not null" json:"-"`
	LastFour       string     `gorm:"size:4" json:"last_four"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

type AuditLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrganizationID uint      `gorm:"index;not null" json:"organization_id"`
	UserID         uint      `gorm:"index" json:"user_id"`
	Action         string    `gorm:"size:200;not null" json:"action"`
	Resource       string    `gorm:"size:100" json:"resource"`
	ResourceID     uint      `json:"resource_id"`
	IP             string    `gorm:"size:45" json:"ip"`
	UserAgent      string    `gorm:"size:500" json:"-"`
	Details        string    `gorm:"type:json" json:"details"`
	CreatedAt      time.Time `json:"created_at"`
}

type Site struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	OrganizationID    uint       `gorm:"index;not null" json:"organization_id"`
	Name              string     `gorm:"size:100;not null" json:"name"`
	URL               string     `gorm:"size:500;not null" json:"url"`
	SelectorLinks     string     `gorm:"size:255" json:"selector_links"`
	SelectorCompany   string     `gorm:"size:255" json:"selector_company"`
	SelectorDescription string   `gorm:"size:255" json:"selector_description"`
	DelaySeconds      int        `gorm:"default:2" json:"delay_seconds"`
	RespectRobots     bool       `gorm:"default:true" json:"respect_robots"`
	Active            bool       `gorm:"default:true" json:"active"`
	LastCrawl         *time.Time `json:"last_crawl"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Job struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	OrganizationID uint       `gorm:"index;not null" json:"organization_id"`
	SiteID         *uint      `gorm:"index" json:"site_id"`
	Site           *Site      `gorm:"foreignKey:SiteID" json:"site,omitempty"`
	Title          string     `gorm:"size:255" json:"title"`
	Company        string     `gorm:"size:255" json:"company"`
	Description    string     `gorm:"type:text" json:"description"`
	URL            string     `gorm:"size:500;uniqueIndex:idx_job_url_org" json:"url"`
	PostedDate     *time.Time `json:"posted_date"`
	CollectedAt    time.Time  `gorm:"autoCreateTime" json:"collected_at"`
	Status         JobStatus  `gorm:"size:20;default:new" json:"status"`
}

type Resume struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrganizationID uint      `gorm:"index;not null" json:"organization_id"`
	Name           string    `gorm:"size:100" json:"name"`
	FilePath       string    `gorm:"size:500" json:"file_path"`
	Content        string    `gorm:"type:text" json:"content"`
	Version        int       `json:"version"`
	UploadedAt     time.Time `json:"uploaded_at"`
}

type Match struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	OrganizationID  uint       `gorm:"index;not null" json:"organization_id"`
	JobID           uint       `gorm:"index" json:"job_id"`
	Job             Job        `gorm:"foreignKey:JobID" json:"job,omitempty"`
	ResumeID        uint       `gorm:"index" json:"resume_id"`
	Resume          Resume     `gorm:"foreignKey:ResumeID" json:"resume,omitempty"`
	SimilarityScore float64    `gorm:"type:decimal(5,2)" json:"similarity_score"`
	KeywordsMatched string     `gorm:"type:json" json:"keywords_matched"`
	AIReason        string     `gorm:"type:text" json:"ai_reason"`
	Applied         bool       `gorm:"default:false" json:"applied"`
	AppliedAt       *time.Time `json:"applied_at"`
	AnalyzedAt      time.Time  `gorm:"autoCreateTime" json:"analyzed_at"`
}

type ResumeAnalysis struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrganizationID uint      `gorm:"index;not null" json:"organization_id"`
	ResumeID       *uint     `gorm:"index" json:"resume_id"`
	FileName       string    `gorm:"size:255" json:"file_name"`
	FullAnalysis   string    `gorm:"type:text" json:"full_analysis"`
	Strengths      string    `gorm:"type:json" json:"strengths"`
	Weaknesses     string    `gorm:"type:json" json:"weaknesses"`
	Suggestions    string    `gorm:"type:json" json:"suggestions"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type AgentLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrganizationID uint      `gorm:"index;not null" json:"organization_id"`
	AgentName      string    `gorm:"size:50" json:"agent_name"`
	Action         string    `gorm:"size:255" json:"action"`
	Details        string    `gorm:"type:json" json:"details"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}
