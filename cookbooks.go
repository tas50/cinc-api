package cinc

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

// Cookbook is a single cookbook version's manifest as returned by the server.
type Cookbook struct {
	CookbookName string `json:"cookbook_name"`
	Name         string `json:"name"`
	Version      string `json:"version"`
}

// cookbookFile is one file belonging to a cookbook being uploaded.
type cookbookFile struct {
	name     string // path relative to the cookbook root, e.g. recipes/default.rb
	content  []byte
	checksum string // hex MD5
}

// LocalCookbook is a cookbook assembled from disk, ready to upload.
type LocalCookbook struct {
	Name    string
	Version string
	files   []cookbookFile
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
	_, _, err = do[map[string]any](ctx, c, "PUT",
		c.orgPath(base+"/"+cb.Name+"/"+cb.Version), manifest)
	if err != nil {
		return fmt.Errorf("cinc: put cookbook manifest: %w", err)
	}
	return nil
}

// cookbookManifest builds the version manifest body for an upload.
func cookbookManifest(cb *LocalCookbook) map[string]any {
	all := make([]map[string]any, 0, len(cb.files))
	for _, f := range cb.files {
		all = append(all, map[string]any{
			"name": filepath.Base(f.name), "path": f.name,
			"checksum": f.checksum, "specificity": "default",
		})
	}
	return map[string]any{
		"cookbook_name": cb.Name,
		"name":          cb.Name + "-" + cb.Version,
		"version":       cb.Version,
		"all_files":     all,
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
