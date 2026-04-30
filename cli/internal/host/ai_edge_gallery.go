package host

// newAIEdgeGallery returns a HostAdapter for Google AI Edge Gallery.
// Tier-3: speculative path; community-confirmable.
func newAIEdgeGallery() HostAdapter {
	return &genericAdapter{
		id:         "ai-edge-gallery",
		name:       "Google AI Edge Gallery",
		homepage:   "https://ai.google.dev",
		homeSubdir: ".google/ai-edge-gallery",
		caps:       Caps{Skills: true},
	}
}
