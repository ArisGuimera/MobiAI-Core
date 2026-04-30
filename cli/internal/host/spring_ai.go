package host

// newSpringAI returns a HostAdapter for Spring AI.
// Tier-3: speculative path; community-confirmable.
func newSpringAI() HostAdapter {
	return &genericAdapter{
		id:         "spring-ai",
		name:       "Spring AI",
		homepage:   "https://spring.io/projects/spring-ai",
		homeSubdir: ".spring-ai",
		caps:       Caps{Skills: true},
	}
}
