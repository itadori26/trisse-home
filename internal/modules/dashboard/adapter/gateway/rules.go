package gateway

import (
	"database/sql"
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

// mobileRow is the stored form of a phone layout entry (#92).
type mobileRow struct {
	ID     string `json:"id"`
	Hidden bool   `json:"hidden,omitempty"`
	Rows   int    `json:"rows"`
}

// encodeMobile returns NULL for a derived layout.
func encodeMobile(m []domain.MobileEntry) any {
	if m == nil {
		return nil
	}
	rows := make([]mobileRow, 0, len(m))
	for _, e := range m {
		rows = append(rows, mobileRow{ID: e.WidgetID, Hidden: e.Hidden, Rows: e.Rows})
	}
	b, _ := json.Marshal(rows) // plain structs, cannot fail
	return string(b)
}

// decodeMobile treats NULL and unreadable values as "derived".
func decodeMobile(raw sql.NullString) []domain.MobileEntry {
	if !raw.Valid {
		return nil
	}
	var rows []mobileRow
	if json.Unmarshal([]byte(raw.String), &rows) != nil {
		return nil
	}
	out := make([]domain.MobileEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.MobileEntry{WidgetID: r.ID, Hidden: r.Hidden, Rows: r.Rows})
	}
	return out
}
