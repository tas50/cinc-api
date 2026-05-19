package cinc

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// CookbookVersion is one version entry in a cookbook list.
type CookbookVersion struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}

// CookbookListEntry is the list response value for one cookbook name.
type CookbookListEntry struct {
	URL      string            `json:"url"`
	Versions []CookbookVersion `json:"versions"`
}

// CookbookFileRef is one file reference in a cookbook version manifest.
type CookbookFileRef struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Specificity string `json:"specificity"`
	Checksum    string `json:"checksum"`
	URL         string `json:"url"`
}

// Cookbook is a single cookbook version's manifest as returned by the server.
// The nine file-segment slices are populated by Get/Download.
type Cookbook struct {
	CookbookName string            `json:"cookbook_name"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Files        []CookbookFileRef `json:"files"`
	Definitions  []CookbookFileRef `json:"definitions"`
	Libraries    []CookbookFileRef `json:"libraries"`
	Attributes   []CookbookFileRef `json:"attributes"`
	Recipes      []CookbookFileRef `json:"recipes"`
	Providers    []CookbookFileRef `json:"providers"`
	Resources    []CookbookFileRef `json:"resources"`
	RootFiles    []CookbookFileRef `json:"root_files"`
	Templates    []CookbookFileRef `json:"templates"`
}

// AllFiles flattens all nine file-segment slices into a single slice.
func (cb *Cookbook) AllFiles() []CookbookFileRef {
	all := make([]CookbookFileRef, 0,
		len(cb.Files)+len(cb.Definitions)+len(cb.Libraries)+
			len(cb.Attributes)+len(cb.Recipes)+len(cb.Providers)+
			len(cb.Resources)+len(cb.RootFiles)+len(cb.Templates))
	for _, seg := range [][]CookbookFileRef{
		cb.Files, cb.Definitions, cb.Libraries, cb.Attributes,
		cb.Recipes, cb.Providers, cb.Resources, cb.RootFiles, cb.Templates,
	} {
		all = append(all, seg...)
	}
	return all
}

// cookbookFile is one file belonging to a cookbook being uploaded.
type cookbookFile struct {
	name     string // path relative to the cookbook root, e.g. recipes/default.rb
	content  []byte
	checksum string // hex MD5
}

// LocalCookbook is a cookbook assembled from disk, ready to upload.
// When Identifier is set the manifest is emitted as a cookbook artifact
// version (chef_type "cookbook_artifact_version") rather than a plain version.
type LocalCookbook struct {
	Name       string
	Version    string
	Identifier string // set for cookbook artifact uploads
	files      []cookbookFile
}

// CookbooksService accesses the /cookbooks endpoints.
type CookbooksService struct{ client *Client }

// List returns all cookbooks and their available versions.
func (s *CookbooksService) List(ctx context.Context) (map[string]CookbookListEntry, *Response, error) {
	return do[map[string]CookbookListEntry](ctx, s.client, "GET",
		s.client.orgPath("/cookbooks"), nil)
}

// Get retrieves a single cookbook version manifest.
func (s *CookbooksService) Get(ctx context.Context, name, version string) (*Cookbook, *Response, error) {
	cb, resp, err := do[Cookbook](ctx, s.client, "GET",
		s.client.orgPath("/cookbooks/"+name+"/"+version), nil)
	return ptrOrNil(cb, err), resp, err
}

// Delete removes a single cookbook version.
func (s *CookbooksService) Delete(ctx context.Context, name, version string) (*Response, error) {
	_, resp, err := do[map[string]any](ctx, s.client, "DELETE",
		s.client.orgPath("/cookbooks/"+name+"/"+version), nil)
	return resp, err
}

// Upload uploads a LocalCookbook: sandbox -> file PUTs -> version manifest PUT.
func (s *CookbooksService) Upload(ctx context.Context, cb *LocalCookbook) error {
	return uploadCookbook(ctx, s.client, "/cookbooks", cb)
}

// Download fetches a cookbook version manifest and writes every file in all
// nine segments to destDir, recreating the path hierarchy. version may be the
// literal string "_latest". File content is fetched from pre-signed bookshelf
// URLs using a plain (unsigned) HTTP GET, matching the upload path in sandboxes.go.
func (s *CookbooksService) Download(ctx context.Context, name, version, destDir string) error {
	cb, _, err := s.Get(ctx, name, version)
	if err != nil {
		return fmt.Errorf("cinc: get cookbook manifest: %w", err)
	}
	for _, ref := range cb.AllFiles() {
		dest := filepath.Join(destDir, filepath.FromSlash(ref.Path))
		rel, err := filepath.Rel(destDir, dest)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("cinc: unsafe file path in cookbook manifest: %q", ref.Path)
		}
		if err := s.client.downloadFile(ctx, ref.URL, dest); err != nil {
			return fmt.Errorf("cinc: download %s: %w", ref.Path, err)
		}
	}
	return nil
}

// downloadFile GETs a pre-signed bookshelf URL (no Chef signing) and writes
// the body to dest, creating parent directories as needed.
func (c *Client) downloadFile(ctx context.Context, fileURL, dest string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return fmt.Errorf("cinc: build download request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cinc: fetch file: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return newErrorResponse("GET", fileURL, resp.StatusCode, body)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cinc: read file body: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("cinc: create dirs: %w", err)
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return fmt.Errorf("cinc: write file: %w", err)
	}
	return nil
}

// uploadCookbook implements the three-step upload, shared with cookbook_artifacts.
func uploadCookbook(ctx context.Context, c *Client, base string, cb *LocalCookbook) error {
	hexes := make([]string, 0, len(cb.files))
	for _, f := range cb.files {
		hexes = append(hexes, f.checksum)
	}
	sb, _, err := c.createSandbox(ctx, hexes)
	if err != nil {
		return fmt.Errorf("cinc: create sandbox: %w", err)
	}
	for _, f := range cb.files {
		entry, needed := sb.Checksums[f.checksum]
		if !needed || !entry.NeedsUpload {
			continue
		}
		if err := c.uploadFile(ctx, entry.URL, f.content); err != nil {
			return fmt.Errorf("cinc: upload %s: %w", f.name, err)
		}
	}
	if _, err := c.commitSandbox(ctx, sb.ID); err != nil {
		return fmt.Errorf("cinc: commit sandbox: %w", err)
	}
	manifest := cookbookManifest(cb)
	// Use Identifier for artifact uploads; Version for regular cookbooks.
	slug := cb.Version
	if cb.Identifier != "" {
		slug = cb.Identifier
	}
	_, _, err = do[map[string]any](ctx, c, "PUT",
		c.orgPath(base+"/"+cb.Name+"/"+slug), manifest)
	if err != nil {
		return fmt.Errorf("cinc: put cookbook manifest: %w", err)
	}
	return nil
}

// cookbookManifest builds the version manifest body for an upload.
// When cb.Identifier is set it emits a cookbook artifact manifest
// (chef_type "cookbook_artifact_version"); otherwise a plain version manifest.
func cookbookManifest(cb *LocalCookbook) map[string]any {
	all := make([]map[string]any, 0, len(cb.files))
	for _, f := range cb.files {
		all = append(all, map[string]any{
			"name": filepath.Base(f.name), "path": f.name,
			"checksum": f.checksum, "specificity": "default",
		})
	}
	if cb.Identifier != "" {
		return map[string]any{
			"cookbook_name": cb.Name,
			"name":          cb.Name,
			"identifier":    cb.Identifier,
			"all_files":     all,
			"chef_type":     "cookbook_artifact_version",
		}
	}
	return map[string]any{
		"cookbook_name": cb.Name,
		"name":          cb.Name + "-" + cb.Version,
		"version":       cb.Version,
		"all_files":     all,
		"chef_type":     "cookbook_version",
	}
}

// cookbookFromDir walks a cookbook directory into an uploadable LocalCookbook.
func cookbookFromDir(dir, version string) (*LocalCookbook, error) {
	cb := &LocalCookbook{Name: filepath.Base(dir), Version: version}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		cb.files = append(cb.files, cookbookFile{
			name: filepath.ToSlash(rel), content: content, checksum: md5Hex(content),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cinc: read cookbook dir: %w", err)
	}
	if len(cb.files) == 0 {
		return nil, fmt.Errorf("cinc: no files found in %s", dir)
	}
	return cb, nil
}
