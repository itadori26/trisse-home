package domain

import "testing"

func TestContainerHealthAndSummary(t *testing.T) {
	cs := []Container{
		{Name: "web", State: "running", Status: "Up 2 hours (healthy)"},
		{Name: "db", State: "running", Status: "Up 2 hours (unhealthy)"},
		{Name: "job", State: "exited", Status: "Exited (0) 5 minutes ago"},
		{Name: "app", State: "running", Status: "Up 1 second (health: starting)"},
	}
	if cs[0].Health() != HealthHealthy || cs[1].Health() != HealthUnhealthy || cs[2].Health() != HealthNone || cs[3].Health() != HealthStarting {
		t.Fatal("health parsing wrong")
	}
	if s := Summarise(cs); s != (Summary{Total: 4, Running: 3, Stopped: 1, Unhealthy: 1}) {
		t.Fatalf("summary %+v", s)
	}
	SortForDisplay(cs)
	if cs[0].Name != "db" || cs[1].Name != "job" || cs[2].Name != "app" {
		t.Fatalf("order %v %v %v", cs[0].Name, cs[1].Name, cs[2].Name)
	}
}
