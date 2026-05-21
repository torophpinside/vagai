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

type Site struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	OrganizationID      uint       `gorm:"index;default:0" json:"organization_id"`
	Name                string     `gorm:"size:100;not null" json:"name"`
	URL                 string     `gorm:"size:500;not null" json:"url"`
	SelectorLinks       string     `gorm:"size:255" json:"selector_links"`
	SelectorCompany     string     `gorm:"size:255" json:"selector_company"`
	SelectorDescription string     `gorm:"size:255" json:"selector_description"`
	Active              bool       `gorm:"default:true" json:"active"`
	DelaySeconds        int        `gorm:"default:0" json:"delay_seconds"`
	RespectRobots       bool       `gorm:"default:true" json:"respect_robots"`
	LastCrawl           *time.Time `json:"last_crawl"`
}

type Job struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	OrganizationID uint       `gorm:"index;default:0" json:"organization_id"`
	SiteID         *uint      `gorm:"index" json:"site_id"`
	Site           *Site      `gorm:"foreignKey:SiteID" json:"site,omitempty"`
	Title          string     `gorm:"size:255" json:"title"`
	Company        string     `gorm:"size:255" json:"company"`
	Description    string     `gorm:"type:text" json:"description"`
	URL            string     `gorm:"size:500;uniqueIndex" json:"url"`
	PostedDate     *time.Time `json:"posted_date"`
	CollectedAt    time.Time  `gorm:"autoCreateTime" json:"collected_at"`
	Status         JobStatus  `gorm:"size:20;default:new" json:"status"`
}

type Resume struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrganizationID uint      `gorm:"index;default:0" json:"organization_id"`
	Name           string    `gorm:"size:100" json:"name"`
	FilePath       string    `gorm:"size:500" json:"file_path"`
	Content        string    `gorm:"type:text" json:"content"`
	Version        int       `json:"version"`
	UploadedAt     time.Time `json:"uploaded_at"`
}

type Match struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	OrganizationID  uint      `gorm:"index;default:0" json:"organization_id"`
	JobID           uint      `gorm:"index" json:"job_id"`
	Job             Job       `gorm:"foreignKey:JobID" json:"job,omitempty"`
	ResumeID        uint      `gorm:"index" json:"resume_id"`
	Resume          Resume    `gorm:"foreignKey:ResumeID" json:"resume,omitempty"`
	SimilarityScore float64   `gorm:"type:decimal(5,2)" json:"similarity_score"`
	KeywordsMatched string    `gorm:"type:json" json:"keywords_matched"`
	AIReason        string    `gorm:"type:text" json:"ai_reason"`
	AnalyzedAt      time.Time `gorm:"autoCreateTime" json:"analyzed_at"`
}

type AgentLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrganizationID uint      `gorm:"index;default:0" json:"organization_id"`
	AgentName      string    `gorm:"size:50" json:"agent_name"`
	Action         string    `gorm:"size:255" json:"action"`
	Details        string    `gorm:"type:json" json:"details"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Schedule struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	OrganizationID uint       `gorm:"index;default:0" json:"organization_id"`
	Name           string     `gorm:"size:100;uniqueIndex" json:"name"`
	Command        string     `gorm:"size:500;not null" json:"command"`
	Schedule       string     `gorm:"size:100;not null" json:"schedule"`
	Active         bool       `gorm:"default:true" json:"active"`
	LastRun        *time.Time `json:"last_run"`
	NextRun        *time.Time `json:"next_run"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

type ScheduleLog struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	ScheduleID uint       `gorm:"index" json:"schedule_id"`
	Status     string     `gorm:"size:20" json:"status"`
	Output     string     `gorm:"type:text" json:"output"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
}
