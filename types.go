package modrinth

import (
	"net/url"
	"regexp"
)

// Project is one Modrinth project as the API describes it.
type Project struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// Body is the project's page in Markdown.
	Body       string   `json:"body"`
	Categories []string `json:"categories"`
	// AdditionalCategories are the ones the project picks beyond its type's primary list.
	AdditionalCategories []string `json:"additional_categories"`
	// Status is the project's moderation state; "archived" means the author has retired it.
	Status string `json:"status"`
}

// Version is one published build of a project.
type Version struct {
	ID            string       `json:"id"`
	ProjectID     string       `json:"project_id"`
	VersionNumber string       `json:"version_number"`
	DatePublished string       `json:"date_published"`
	Files         []File       `json:"files"`
	Dependencies  []Dependency `json:"dependencies"`
}

// PrimaryFile is the version's primary file, else its first; ok is false when it has none.
func (v Version) PrimaryFile() (File, bool) {
	if len(v.Files) == 0 {
		return File{}, false
	}
	for _, f := range v.Files {
		if f.Primary {
			return f, true
		}
	}
	return v.Files[0], true
}

// File is one downloadable file of a version.
type File struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	// Hashes is keyed by algorithm, "sha1" and "sha512" among them; a missing algorithm is absent.
	Hashes  map[string]string `json:"hashes"`
	Primary bool              `json:"primary"`
}

// Dependency is a project a version builds on. Type is "required", "optional", "incompatible" or "embedded"; a
// required one names the project it needs and, when pinned, the version.
type Dependency struct {
	ProjectID string `json:"project_id"`
	VersionID string `json:"version_id"`
	Type      string `json:"dependency_type"`
}

// SearchHit is one project in a search answer.
type SearchHit struct {
	ProjectID   string `json:"project_id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Author      string `json:"author"`
	IconURL     string `json:"icon_url"`
	Downloads   int    `json:"downloads"`
	Follows     int    `json:"follows"`
	// Modified is the project's last change, RFC 3339.
	Modified string `json:"date_modified"`
	// Versions are the ids of the project's versions, and Latest is the newest of them.
	Versions []string `json:"versions"`
	Latest   string   `json:"latest_version"`
}

// SearchResult is one page of search hits.
type SearchResult struct {
	// Total is the number of matching projects, not the number of hits.
	Total int         `json:"total_hits"`
	Hits  []SearchHit `json:"hits"`
}

// Category is one project category tag, such as "fabric" or "adventure".
type Category struct {
	Name string `json:"name"`
	// ProjectType is the kind of project the category groups, such as "mod" or "modpack".
	ProjectType string `json:"project_type"`
}

// idPattern is a project id or slug: Modrinth ids are base-62 and slugs add dots, underscores and hyphens.
var idPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// ValidID reports whether id is a project id or slug, so it is safe to put in a request path.
func ValidID(id string) bool { return idPattern.MatchString(id) }

// PageURL is the project's page address; projectType is the URL segment such as "mod" or "modpack", and slug is the
// project's slug or id.
func PageURL(projectType, slug string) string {
	return "https://modrinth.com/" + url.PathEscape(projectType) + "/" + url.PathEscape(slug)
}
