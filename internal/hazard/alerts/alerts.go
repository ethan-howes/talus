package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func EvaluateAndFire(ctx context.Context, pool *pgxpool.Pool, routeID int, riskScore float64, freezeThawActive bool, summary string) error {

	sql :=
		`
	SELECT id, risk_threshold, freeze_thaw_trigger, webhook_url
	FROM alert_configs
	WHERE route_id = $1 AND enabled = true
	`

	rows, err := pool.Query(ctx, sql, routeID)
	if err != nil {
		return fmt.Errorf("failed to query alery configs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var configID int
		var riskThreshold float64
		var freezeThawTrigger bool
		var webhookURL string

		err := rows.Scan(&configID, &riskThreshold, &freezeThawTrigger, &webhookURL)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		if riskScore >= riskThreshold {

			payload, _ := json.Marshal(map[string]interface{}{
				"route_id":           routeID,
				"risk_score":         riskScore,
				"freeze_thaw_active": freezeThawActive,
				"summary":            summary,
			})
			http.Post(webhookURL, "application/json", bytes.NewReader(payload))

			_, err = pool.Exec(ctx,
				`INSERT INTO alert_events (config_id, route_id, triggered_at, risk_score, summary, freeze_thaw_active) VALUES ($1, $2, NOW(), $3, $4, $5)`,
				configID, routeID, riskScore, summary, freezeThawActive)
			if err != nil {
				return fmt.Errorf("failed to insert alery event: %w", err)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}
	return nil
}
