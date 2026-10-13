package models

import (
	"log"

	"gorm.io/gorm"
)

// CreateAnalyticsViews cria as views denormalizadas usadas pelas queries de
// analytics (data-model.md → Denormalized Views). As views são por usuário
// (coluna user_id disponível) — o filtro de tenant é aplicado na query do
// serviço. SQL em dialeto MySQL 8.
func CreateAnalyticsViews(db *gorm.DB) {
	views := map[string]string{
		"matching_distribution": `
			SELECT
				user_id,
				CASE
					WHEN similarity_score < 0.2 THEN '0.0-0.2'
					WHEN similarity_score < 0.4 THEN '0.2-0.4'
					WHEN similarity_score < 0.6 THEN '0.4-0.6'
					WHEN similarity_score < 0.8 THEN '0.6-0.8'
					ELSE '0.8-1.0'
				END AS bucket,
				COUNT(*) AS cnt,
				AVG(similarity_score) AS avg_score
			FROM matches
			GROUP BY user_id, bucket`,

		"user_metrics_summary": `
			SELECT
				u.id AS user_id,
				u.name AS name,
				u.plan_slug AS plan,
				(SELECT COUNT(*) FROM jobs j WHERE j.user_id = u.id) AS jobs_posted,
				(SELECT COUNT(*) FROM resumes r WHERE r.user_id = u.id) AS resumes_uploaded,
				(SELECT COUNT(*) FROM matches m WHERE m.user_id = u.id) AS matches_made,
				(SELECT COUNT(*) FROM matches ma WHERE ma.user_id = u.id AND ma.applied = TRUE) AS matches_applied
			FROM users u`,

		"funnel_summary": `
			SELECT
				u.id AS user_id,
				(SELECT COUNT(*) FROM resumes r WHERE r.user_id = u.id) AS resumes_uploaded,
				(SELECT COUNT(*) FROM matches m WHERE m.user_id = u.id) AS matches_made,
				(SELECT COUNT(*) FROM matches m WHERE m.user_id = u.id AND m.applied = TRUE) AS applications
			FROM users u`,
	}

	for name, query := range views {
		stmt := "CREATE OR REPLACE VIEW `" + name + "` AS" + query
		if err := db.Exec(stmt).Error; err != nil {
			log.Printf("analytics views: falha ao criar view %s: %v", name, err)
		}
	}
}
