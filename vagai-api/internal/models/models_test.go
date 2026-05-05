package models

import (
	"testing"
	"time"
)

func TestJobStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   JobStatus
		expected string
	}{
		{"New", JobStatusNew, "new"},
		{"Matched", JobStatusMatched, "matched"},
		{"Unmatched", JobStatusUnmatched, "unmatched"},
		{"Ignored", JobStatusIgnored, "ignored"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("JobStatus = %v, expected %v", tt.status, tt.expected)
			}
		})
	}
}

func TestUserRoleConstants(t *testing.T) {
	tests := []struct {
		name     string
		role     UserRole
		expected string
	}{
		{"Owner", RoleOwner, "owner"},
		{"Admin", RoleAdmin, "admin"},
		{"Member", RoleMember, "member"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.role) != tt.expected {
				t.Errorf("UserRole = %v, expected %v", tt.role, tt.expected)
			}
		})
	}
}

func TestSubscriptionStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   SubscriptionStatus
		expected string
	}{
		{"Trial", SubStatusTrial, "trial"},
		{"Active", SubStatusActive, "active"},
		{"PastDue", SubStatusPastDue, "past_due"},
		{"Canceled", SubStatusCanceled, "canceled"},
		{"Expired", SubStatusExpired, "expired"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("SubscriptionStatus = %v, expected %v", tt.status, tt.expected)
			}
		})
	}
}

func TestOrganizationModel(t *testing.T) {
	trialEnds := time.Now().Add(14 * 24 * time.Hour)
	org := Organization{
		Name:        "Test Org",
		Slug:        "test-org",
		Plan:        "free",
		TrialEndsAt: &trialEnds,
	}

	if org.Name != "Test Org" {
		t.Errorf("Organization.Name = %v, expected Test Org", org.Name)
	}
	if org.Slug != "test-org" {
		t.Errorf("Organization.Slug = %v, expected test-org", org.Slug)
	}
	if org.Plan != "free" {
		t.Errorf("Organization.Plan = %v, expected free", org.Plan)
	}
	if org.TrialEndsAt == nil {
		t.Error("Organization.TrialEndsAt should not be nil")
	}
}

func TestUserModel(t *testing.T) {
	user := User{
		Name:     "Test User",
		Email:    "test@example.com",
		PasswordHash: "hashedpassword",
		Timezone: "America/Sao_Paulo",
	}

	if user.Name != "Test User" {
		t.Errorf("User.Name = %v, expected Test User", user.Name)
	}
	if user.Email != "test@example.com" {
		t.Errorf("User.Email = %v, expected test@example.com", user.Email)
	}
	if user.PasswordHash != "hashedpassword" {
		t.Errorf("User.PasswordHash = %v, expected hashedpassword", user.PasswordHash)
	}
	if user.Timezone != "America/Sao_Paulo" {
		t.Errorf("User.Timezone = %v, expected America/Sao_Paulo", user.Timezone)
	}
}

func TestMembershipModel(t *testing.T) {
	membership := Membership{
		OrganizationID: 1,
		UserID:         1,
		Role:           RoleOwner,
	}

	if membership.OrganizationID != 1 {
		t.Errorf("Membership.OrganizationID = %v, expected 1", membership.OrganizationID)
	}
	if membership.UserID != 1 {
		t.Errorf("Membership.UserID = %v, expected 1", membership.UserID)
	}
	if membership.Role != RoleOwner {
		t.Errorf("Membership.Role = %v, expected owner", membership.Role)
	}
}

func TestJobModel(t *testing.T) {
	job := Job{
		Title:       "Software Engineer",
		Company:     "Test Corp",
		Description: "Looking for a great developer",
		URL:         "https://example.com/job/1",
		Status:      JobStatusNew,
	}

	if job.Title != "Software Engineer" {
		t.Errorf("Job.Title = %v, expected Software Engineer", job.Title)
	}
	if job.Company != "Test Corp" {
		t.Errorf("Job.Company = %v, expected Test Corp", job.Company)
	}
	if job.Status != JobStatusNew {
		t.Errorf("Job.Status = %v, expected new", job.Status)
	}
}

func TestMatchModel(t *testing.T) {
	now := time.Now()
	match := Match{
		JobID:           1,
		ResumeID:        1,
		SimilarityScore: 85.5,
		KeywordsMatched: `["go", "api"]`,
		Applied:         false,
	}

	if match.JobID != 1 {
		t.Errorf("Match.JobID = %v, expected 1", match.JobID)
	}
	if match.SimilarityScore != 85.5 {
		t.Errorf("Match.SimilarityScore = %v, expected 85.5", match.SimilarityScore)
	}
	if match.Applied != false {
		t.Errorf("Match.Applied = %v, expected false", match.Applied)
	}

	match.Applied = true
	match.AppliedAt = &now

	if match.Applied != true {
		t.Errorf("Match.Applied = %v, expected true", match.Applied)
	}
	if match.AppliedAt == nil {
		t.Error("Match.AppliedAt should not be nil after applying")
	}
}

func TestSiteModel(t *testing.T) {
	site := Site{
		Name:          "LinkedIn",
		URL:           "https://linkedin.com/jobs",
		SelectorLinks: ".job-card",
		Active:        true,
	}

	if site.Name != "LinkedIn" {
		t.Errorf("Site.Name = %v, expected LinkedIn", site.Name)
	}
	if site.URL != "https://linkedin.com/jobs" {
		t.Errorf("Site.URL = %v, expected https://linkedin.com/jobs", site.URL)
	}
	if site.Active != true {
		t.Errorf("Site.Active = %v, expected true", site.Active)
	}
}

func TestResumeModel(t *testing.T) {
	resume := Resume{
		Name:     "resume.pdf",
		FilePath: "/uploads/resumes/resume.pdf",
		Content:  "John Doe - Software Engineer...",
		Version:  1,
	}

	if resume.Name != "resume.pdf" {
		t.Errorf("Resume.Name = %v, expected resume.pdf", resume.Name)
	}
	if resume.FilePath != "/uploads/resumes/resume.pdf" {
		t.Errorf("Resume.FilePath = %v, expected /uploads/resumes/resume.pdf", resume.FilePath)
	}
	if resume.Version != 1 {
		t.Errorf("Resume.Version = %v, expected 1", resume.Version)
	}
}

func TestPlanModel(t *testing.T) {
	plan := Plan{
		Name:            "Pro",
		Slug:            "pro",
		PriceMonthly:    4900,
		PriceYearly:     49000,
		MaxJobs:         500,
		MaxResumes:      10,
		MaxSites:        20,
		MaxCrawlsPerDay: 50,
	}

	if plan.Name != "Pro" {
		t.Errorf("Plan.Name = %v, expected Pro", plan.Name)
	}
	if plan.PriceMonthly != 4900 {
		t.Errorf("Plan.PriceMonthly = %v, expected 4900", plan.PriceMonthly)
	}
	if plan.MaxJobs != 500 {
		t.Errorf("Plan.MaxJobs = %v, expected 500", plan.MaxJobs)
	}
}

func TestSubscriptionModel(t *testing.T) {
	now := time.Now()
	sub := Subscription{
		OrganizationID:     1,
		Status:             SubStatusTrial,
		CurrentPeriodStart: &now,
		CurrentPeriodEnd:   func() *time.Time { t := now.Add(14 * 24 * time.Hour); return &t }(),
		CancelAtPeriodEnd:  false,
	}

	if sub.Status != SubStatusTrial {
		t.Errorf("Subscription.Status = %v, expected trial", sub.Status)
	}
	if sub.CancelAtPeriodEnd != false {
		t.Errorf("Subscription.CancelAtPeriodEnd = %v, expected false", sub.CancelAtPeriodEnd)
	}
}

func TestApiKeyModel(t *testing.T) {
	apiKey := ApiKey{
		OrganizationID: 1,
		Name:           "Production API Key",
		KeyHash:        "hashed_key_123",
		LastFour:       "1234",
	}

	if apiKey.Name != "Production API Key" {
		t.Errorf("ApiKey.Name = %v, expected Production API Key", apiKey.Name)
	}
	if apiKey.LastFour != "1234" {
		t.Errorf("ApiKey.LastFour = %v, expected 1234", apiKey.LastFour)
	}
}

func TestAuditLogModel(t *testing.T) {
	log := AuditLog{
		OrganizationID: 1,
		UserID:         1,
		Action:         "job.created",
		Resource:       "Job",
		ResourceID:     42,
		IP:             "192.168.1.1",
		Details:        `{"title": "Dev"}`,
	}

	if log.Action != "job.created" {
		t.Errorf("AuditLog.Action = %v, expected job.created", log.Action)
	}
	if log.Resource != "Job" {
		t.Errorf("AuditLog.Resource = %v, expected Job", log.Resource)
	}
	if log.IP != "192.168.1.1" {
		t.Errorf("AuditLog.IP = %v, expected 192.168.1.1", log.IP)
	}
}

func TestAgentLogModel(t *testing.T) {
	agentLog := AgentLog{
		OrganizationID: 1,
		AgentName:      "crawler",
		Action:         "crawl.completed",
		Details:        `{"jobs_found": 10}`,
	}

	if agentLog.AgentName != "crawler" {
		t.Errorf("AgentLog.AgentName = %v, expected crawler", agentLog.AgentName)
	}
	if agentLog.Action != "crawl.completed" {
		t.Errorf("AgentLog.Action = %v, expected crawl.completed", agentLog.Action)
	}
}
