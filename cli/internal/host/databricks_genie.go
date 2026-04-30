package host

// newDatabricksGenie returns a HostAdapter for Databricks Genie Code.
// Tier-3: speculative path; community-confirmable.
func newDatabricksGenie() HostAdapter {
	return &genericAdapter{
		id:         "databricks-genie",
		name:       "Databricks Genie Code",
		homepage:   "https://www.databricks.com",
		homeSubdir: ".databricks/genie-code",
		caps:       Caps{Skills: true},
	}
}
