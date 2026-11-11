package controller

// UpdateDTO is the body of PUT /api/onboarding: true completes the tour,
// false restarts it.
type UpdateDTO struct {
	Completed bool `json:"completed"`
}
