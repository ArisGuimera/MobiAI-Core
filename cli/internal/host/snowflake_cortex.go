package host

// newSnowflakeCortex returns a HostAdapter for Snowflake Cortex Code.
// Tier-3: speculative path; community-confirmable.
func newSnowflakeCortex() HostAdapter {
	return &genericAdapter{
		id:         "snowflake-cortex",
		name:       "Snowflake Cortex Code",
		homepage:   "https://www.snowflake.com",
		homeSubdir: ".snowflake/cortex-code",
		caps:       Caps{Skills: true},
	}
}
