package host

// newWorkshop returns a HostAdapter for Workshop.
// Tier-3: speculative path; community-confirmable.
func newWorkshop() HostAdapter {
	return &genericAdapter{
		id:         "workshop",
		name:       "Workshop",
		homepage:   "https://workshop.codes",
		homeSubdir: ".workshop",
		caps:       Caps{Skills: true},
	}
}
