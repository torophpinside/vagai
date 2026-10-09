package analytics

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/spf13/cobra"
)

var funnelCmd = &cobra.Command{
	Use:   "funnel",
	Short: "Relatório de funil de conversão",
	Long: `Funil de conversão em 4 estágios (signups, currículos enviados,
matches e candidaturas) com contagem, conversão acumulada, tempo mediano
entre estágios e pontos de maior abandono.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := mustSetup(cmd)
		if err != nil {
			return err
		}
		printHeader("Funnel Analysis", user)

		// Estágio 1 — usuários cadastrados (user == 0 agrega todos os usuários).
		signupQuery := db.DB.Table("users").Where("users.deleted_at IS NULL")
		if user > 0 {
			signupQuery = signupQuery.Where("users.id = ?", user)
		}
		var signups int64
		if err := signupQuery.Count(&signups).Error; err != nil {
			return err
		}

		count := func(table, extra string) (int64, error) {
			q := db.DB.Table(table)
			if user > 0 {
				q = q.Where("user_id = ?", user)
			}
			if extra != "" {
				q = q.Where(extra)
			}
			var n int64
			err := q.Count(&n).Error
			return n, err
		}

		resumes, err := count("resumes", "")
		if err != nil {
			return err
		}
		matches, err := count("matches", "")
		if err != nil {
			return err
		}
		applied, err := count("matches", "applied = TRUE")
		if err != nil {
			return err
		}

		// Estado vazio: literais duplicados da API (o CLI não importa o pacote).
		if signups == 0 {
			fmt.Println("Nenhum usuário no funil ainda")
			fmt.Println("Cadastre usuários, envie currículos e execute o matching para começar a acompanhar o funil de conversão.")
			return nil
		}

		stages := []string{"signups", "uploaded_resumes", "matched_with_jobs", "applied_to_jobs"}
		counts := []int64{signups, resumes, matches, applied}

		fmt.Println("\nEstágios do funil:")
		for i, name := range stages {
			fmt.Printf("  %-20s %6d  (%5.1f%%)\n", name, counts[i], pct64(counts[i], signups))
		}

		// Medianas de timestamps via window functions (MySQL 8).
		signupWhere := "deleted_at IS NULL"
		var signupArgs []any
		if user > 0 {
			signupWhere += " AND id = ?"
			signupArgs = append(signupArgs, user)
		}
		userWhere := func(base string) (string, []any) {
			if user > 0 {
				return base + " AND user_id = ?", []any{user}
			}
			return base, []any{}
		}

		mSignup, err := fnFunnelMedian("created_at", "users", signupWhere, signupArgs...)
		if err != nil {
			return err
		}
		where, wArgs := userWhere("1=1")
		mResume, err := fnFunnelMedian("uploaded_at", "resumes", where, wArgs...)
		if err != nil {
			return err
		}
		where, wArgs = userWhere("1=1")
		mMatch, err := fnFunnelMedian("analyzed_at", "matches", where, wArgs...)
		if err != nil {
			return err
		}
		where, wArgs = userWhere("applied = TRUE")
		mMatchApplied, err := fnFunnelMedian("analyzed_at", "matches", where, wArgs...)
		if err != nil {
			return err
		}
		where, wArgs = userWhere("applied = TRUE AND applied_at IS NOT NULL")
		mAppliedAt, err := fnFunnelMedian("applied_at", "matches", where, wArgs...)
		if err != nil {
			return err
		}

		days := func(from, to float64) float64 {
			d := (to - from) / 86400
			if d < 0 {
				d = 0
			}
			return d
		}

		fmt.Println("\nTempo mediano por estágio (dias):")
		for _, t := range []struct {
			stage   string
			med     float64
			samples int64
		}{
			{"signup_to_resume", days(mResume, mSignup), resumes},
			{"resume_to_match", days(mMatch, mResume), matches},
			{"match_to_apply", days(mAppliedAt, mMatchApplied), applied},
		} {
			fmt.Printf("  %-20s %6.1f dias (%d amostras)\n", t.stage, t.med, t.samples)
		}

		// Abandono: razões estágio-a-estágio (estágios 2..4), 2 menores.
		type dropOff struct {
			stage string
			ratio float64
		}
		dropOffs := make([]dropOff, 0, len(stages)-1)
		for i := 1; i < len(stages); i++ {
			r := 0.0
			if counts[i-1] > 0 {
				r = float64(counts[i]) / float64(counts[i-1])
			}
			dropOffs = append(dropOffs, dropOff{stage: stages[i], ratio: r})
		}
		sort.SliceStable(dropOffs, func(a, b int) bool {
			return dropOffs[a].ratio < dropOffs[b].ratio
		})

		fmt.Println("\nMaiores pontos de abandono:")
		n := 2
		if len(dropOffs) < n {
			n = len(dropOffs)
		}
		for i := 0; i < n; i++ {
			fmt.Printf("  - %-20s retém %5.1f%% do estágio anterior\n",
				dropOffs[i].stage, dropOffs[i].ratio*100)
		}

		fmt.Printf("\nTotal de usuários: %d\n", signups)
		return nil
	},
}

// fnFunnelMedian devolve a mediana (epoch segundos) da coluna no recorte via
// window functions (MySQL 8), 0 quando o recorte está vazio.
func fnFunnelMedian(column, from, where string, args ...any) (float64, error) {
	query := fmt.Sprintf(`
		SELECT COALESCE((
			SELECT AVG(v)
			FROM (
				SELECT UNIX_TIMESTAMP(%s) AS v,
					ROW_NUMBER() OVER (ORDER BY %s) AS rn,
					COUNT(*) OVER () AS total
				FROM %s
				WHERE %s
			) ranked
			WHERE rn IN (FLOOR((total + 1) / 2), CEIL((total + 1) / 2))
		), 0) AS median`, column, column, from, where)

	var med sql.NullFloat64
	if err := db.DB.Raw(query, args...).Scan(&med).Error; err != nil {
		return 0, err
	}
	if !med.Valid {
		return 0, nil
	}
	return med.Float64, nil
}

func init() {
	Cmd.AddCommand(funnelCmd)
}
