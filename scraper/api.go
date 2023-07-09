// Encapsulates all the logic for interacting with the Substack API
package scraper

import "time"

func buildUrl(pubName string, path string) string {
	return "https://" + pubName + ".substack.com/api/v1/" + path
}

// There are more fields than this, but these are the ones we care about
// /api/v1/posts/:slug
type postApiResponse struct {
	Id            int64     `json:"id"`
	PublicationId int       `json:"publication_id"`
	Type          string    `json:"type"`
	Title         string    `json:"title"`
	Subtitle      string    `json:"subtitle"`
	Slug          string    `json:"slug"`
	PostDate      time.Time `json:"post_date"`
	CanonicalUrl  string    `json:"canonical_url"`
	Description   string    `json:"description"`
	Body          string    `json:"body_html"`
}

// /api/v1/archive?offset=0&limit=10
type archiveApiResponse []archiveApiInnerResponse

type archiveApiInnerResponse struct {
	Slug     string    `json:"slug"`
	PostDate time.Time `json:"post_date"`
	// audience is 'everyone' or 'only_paid'
	Audience    string `json:"audience"`
	SectionSlug string `json:"section_slug"`
	SectionName string `json:"section_name"`
}
