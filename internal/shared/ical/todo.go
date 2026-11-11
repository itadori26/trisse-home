package ical

import (
	"strings"
	"time"
)

// Todo is one VTODO.
type Todo struct {
	UID       string
	Summary   string
	Done      bool
	Completed time.Time
	Created   time.Time
}

// ParseTodos extracts the VTODOs of an iCalendar object.
func ParseTodos(data string) []Todo {
	var out []Todo
	var props map[string]icalProperty
	depth := 0
	for _, line := range unfold(data) {
		switch {
		case line == "BEGIN:VTODO":
			props, depth = map[string]icalProperty{}, 0
		case strings.HasPrefix(line, "BEGIN:") && props != nil:
			depth++
		case strings.HasPrefix(line, "END:") && props != nil && depth > 0:
			depth--
		case line == "END:VTODO" && props != nil:
			t := Todo{UID: props["UID"].value, Summary: strings.TrimSpace(textUnescaper.Replace(props["SUMMARY"].value))}
			status := strings.ToUpper(props["STATUS"].value)
			t.Done = status == "COMPLETED" || props["COMPLETED"].value != ""
			if status == "CANCELLED" {
				props = nil
				continue
			}
			if c, _, err := parseDateTime(props["COMPLETED"], time.UTC); err == nil {
				t.Completed = c
			}
			if c, _, err := parseDateTime(props["CREATED"], time.UTC); err == nil {
				t.Created = c
			}
			out = append(out, t)
			props = nil
		case props != nil && depth == 0:
			if p, ok := parseLine(line); ok {
				if _, seen := props[p.name]; !seen {
					props[p.name] = p
				}
			}
		}
	}
	return out
}

// NewTodo serialises a new VTODO.
func NewTodo(uid, summary string, now time.Time) string {
	stamp := now.UTC().Format("20060102T150405Z")
	lines := []string{
		"BEGIN:VCALENDAR", "VERSION:2.0", "PRODID:-//trisse-home//shopping//EN",
		"BEGIN:VTODO", "UID:" + uid, "DTSTAMP:" + stamp, "CREATED:" + stamp, "LAST-MODIFIED:" + stamp,
		"SUMMARY:" + EscapeText(summary), "STATUS:NEEDS-ACTION",
		"END:VTODO", "END:VCALENDAR",
	}
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(FoldLine(l))
	}
	return b.String()
}

// SetTodoDone rewrites the completion properties of the first VTODO in raw
// and keeps every other property (also those of other apps) unchanged.
func SetTodoDone(raw string, done bool, now time.Time) string {
	stamp := now.UTC().Format("20060102T150405Z")
	replace := map[string]bool{"STATUS": true, "COMPLETED": true, "PERCENT-COMPLETE": true, "LAST-MODIFIED": true, "DTSTAMP": true}
	var b strings.Builder
	inTodo, depth, handled := false, 0, false
	for _, line := range unfold(raw) {
		if line == "" {
			continue
		}
		switch {
		case line == "BEGIN:VTODO" && !handled:
			inTodo, depth = true, 0
		case inTodo && strings.HasPrefix(line, "BEGIN:"):
			depth++
		case inTodo && depth > 0 && strings.HasPrefix(line, "END:"):
			depth--
		case inTodo && depth == 0 && line == "END:VTODO":
			b.WriteString(FoldLine("DTSTAMP:" + stamp))
			b.WriteString(FoldLine("LAST-MODIFIED:" + stamp))
			if done {
				b.WriteString(FoldLine("STATUS:COMPLETED"))
				b.WriteString(FoldLine("COMPLETED:" + stamp))
				b.WriteString(FoldLine("PERCENT-COMPLETE:100"))
			} else {
				b.WriteString(FoldLine("STATUS:NEEDS-ACTION"))
			}
			inTodo, handled = false, true
		case inTodo && depth == 0:
			if p, ok := parseLine(line); ok && replace[p.name] {
				continue
			}
		}
		b.WriteString(FoldLine(line))
	}
	return b.String()
}
