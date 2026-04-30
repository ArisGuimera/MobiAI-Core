package host

// newLaravelBoost returns a HostAdapter for Laravel Boost.
// Tier-3: speculative path; community-confirmable.
func newLaravelBoost() HostAdapter {
	return &genericAdapter{
		id:         "laravel-boost",
		name:       "Laravel Boost",
		homepage:   "https://laravel.com",
		homeSubdir: ".config/laravel-boost",
		caps:       Caps{Skills: true},
	}
}
