package models

import (
	"log"

	"gorm.io/gorm"
)

// CreateAnalyticsViews cria as views denormalizadas usadas pelas queries de
// analytics (data-model.md → Denormalized Views). As views são por
// organização (coluna organization_id disponível) — o filtro de tenant é
// aplicado na query do serviço (Constitution III). SQL em dialeto MySQL 8.
func CreateAnalyticsViews(db *gorm.DB) {
	views := map[string]string{
		"matching_distribution": `
			SELECT
				organization_id,
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
			GROUP BY organization_id, bucket`,

		"org_metrics_summary": `
			SELECT
				o.id AS organization_id,
				o.name AS name,
				o.plan AS plan,
				(SELECT COUNT(*) FROM jobs j WHERE j.organization_id = o.id) AS jobs_posted,
				(SELECT COUNT(*) FROM resumes r WHERE r.organization_id = o.id) AS resumes_uploaded,
				(SELECT COUNT(*) FROM matches m WHERE m.organization_id = o.id) AS matches_made,
				(SELECT COUNT(*) FROM matches ma WHERE ma.organization_id = o.id AND ma.applied = TRUE) AS matches_applied
			FROM organizations o`,

		"funnel_summary": `
			SELECT
				o.id AS organization_id,
				(SELECT COUNT(*) FROM users u
					JOIN memberships ms ON ms.user_id = u.id AND ms.organization_id = o.id
					WHERE u.deleted_at IS NULL) AS total_users,
				(SELECT COUNT(*) FROM resumes r WHERE r.organization_id = o.id) AS resumes_uploaded,
				(SELECT COUNT(*) FROM matches m WHERE m.organization_id = o.id) AS matches_made,
				(SELECT COUNT(*) FROM matches m WHERE m.organization_id = o.id AND m.applied = TRUE) AS applications
			FROM organizations o`,
	}

	for name, query := range views {
		stmt := "CREATE OR REPLACE VIEW `" + name + "` AS" + query
		if err := db.Exec(stmt).Error; err != nil {
			log.Printf("analytics views: falha ao criar view %s: %v", name, err)
		}
	}
}
