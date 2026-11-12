package gateway

import (
	"encoding/json"

	"trisse-home/internal/modules/dashboard/domain"
)

// ruleRow is the stored form of a tab rule (#71).
type ruleRow struct {
	Days     []int    `json:"days"`
	From     int      `json:"from"`
	To       int      `json:"to"`
	Networks []string `json:"networks"`
}

func encodeRules(rules []domain.TabRule) string {
	rows := make([]ruleRow, 0, len(rules))
	for _, r := range rules {
		from, to := r.Minutes()
		rows = append(rows, ruleRow{Days: r.Days(), From: from, To: to, Networks: r.Networks()})
	}
	b, _ := json.Marshal(rows) // plain structs, cannot fail
	return string(b)
}

// decodeRules ignores unreadable rules instead of breaking the dashboard.
func decodeRules(raw string) []domain.TabRule {
	var rows []ruleRow
	if json.Unmarshal([]byte(raw), &rows) != nil {
		return nil
	}
	out := make([]domain.TabRule, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.RestoreTabRule(r.Days, r.From, r.To, r.Networks))
	}
	return out
}
