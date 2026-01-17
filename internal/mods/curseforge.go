package mods

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"HyPrism/internal/util/download"
)

const (
	// CurseForge API endpoints
	curseForgeBaseURL = "https://api.curseforge.com/v1"
	hytaleGameID      = 70216 // Hytale game ID on CurseForge (verified via API)
	
	// CurseForge API key (public key for mod browsing)
	// Note: For production, this should be in environment variables
	cfAPIKey = "$2a$10$bL4bIL5pUWqfcO7KQtnMReakwtfHbNKh6v1uTpKlzhwoueEJQnPnm"
)

// CurseForgeResponse represents a CurseForge API response
type CurseForgeResponse struct {
	Data       json.RawMessage `json:"data"`
	Pagination *Pagination     `json:"pagination,omitempty"`
}

// Pagination represents pagination info
type Pagination struct {
	Index       int `json:"index"`
	PageSize    int `json:"pageSize"`
	ResultCount int `json:"resultCount"`
	TotalCount  int `json:"totalCount"`
}

// CurseForgeMod represents a mod from CurseForge
type CurseForgeMod struct {
	ID             int             `json:"id"`
	GameID         int             `json:"gameId"`
	Name           string          `json:"name"`
	Slug           string          `json:"slug"`
	Summary        string          `json:"summary"`
	DownloadCount  int             `json:"downloadCount"`
	DateCreated    string          `json:"dateCreated"`   // ISO 8601 format
	DateModified   string          `json:"dateModified"`  // ISO 8601 format
	DateReleased   string          `json:"dateReleased"`  // ISO 8601 format
	Logo           *ModLogo        `json:"logo"`
	Screenshots    []ModScreenshot `json:"screenshots"`
	Categories     []ModCategory   `json:"categories"`
	Authors        []ModAuthor     `json:"authors"`
	LatestFiles    []ModFile       `json:"latestFiles"`
	MainFileID     int             `json:"mainFileId"`
	AllowModDistribution bool      `json:"allowModDistribution"`
}

// ModLogo represents mod logo
type ModLogo struct {
	ID           int    `json:"id"`
	ModID        int    `json:"modId"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ThumbnailURL string `json:"thumbnailUrl"`
	URL          string `json:"url"`
}

// ModCategory represents a mod category
type ModCategory struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	URL      string `json:"url"`
	IconURL  string `json:"iconUrl"`
}

// ModAuthor represents a mod author
type ModAuthor struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ModScreenshot represents a mod screenshot
type ModScreenshot struct {
	ID           int    `json:"id"`
	ModID        int    `json:"modId"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ThumbnailURL string `json:"thumbnailUrl"`
	URL          string `json:"url"`
}

// ModFile represents a mod file
type ModFile struct {
	ID          int    `json:"id"`
	ModID       int    `json:"modId"`
	DisplayName string `json:"displayName"`
	FileName    string `json:"fileName"`
	FileLength  int64  `json:"fileLength"`
	DownloadURL string `json:"downloadUrl"`
	FileDate    string `json:"fileDate"` // ISO 8601 format
	ReleaseType int    `json:"releaseType"` // 1=Release, 2=Beta, 3=Alpha
}

// SearchModsParams represents search parameters
type SearchModsParams struct {
	Query      string
	CategoryID int
	SortField  string // 1=Featured, 2=Popularity, 3=LastUpdated, 4=Name, 5=Author, 6=TotalDownloads
	SortOrder  string // asc, desc
	PageSize   int
	Index      int
}

// SearchResult represents search results
type SearchResult struct {
	Mods       []CurseForgeMod `json:"mods"`
	TotalCount int             `json:"totalCount"`
	PageIndex  int             `json:"pageIndex"`
	PageSize   int             `json:"pageSize"`
}

// SearchMods searches for mods on CurseForge
func SearchMods(ctx context.Context, params SearchModsParams) (*SearchResult, error) {
	baseURL := fmt.Sprintf("%s/mods/search", curseForgeBaseURL)
	
	u, _ := url.Parse(baseURL)
	q := u.Query()
	q.Set("gameId", strconv.Itoa(hytaleGameID))
	
	if params.Query != "" {
		q.Set("searchFilter", params.Query)
	}
	if params.CategoryID > 0 {
		q.Set("categoryId", strconv.Itoa(params.CategoryID))
	}
	if params.SortField != "" {
		q.Set("sortField", params.SortField)
	}
	if params.SortOrder != "" {
		q.Set("sortOrder", params.SortOrder)
	}
	if params.PageSize > 0 {
		q.Set("pageSize", strconv.Itoa(params.PageSize))
	} else {
		q.Set("pageSize", "20")
	}
	if params.Index > 0 {
		q.Set("index", strconv.Itoa(params.Index))
	}
	
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", cfAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search mods: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CurseForge API error: %d - %s", resp.StatusCode, string(body))
	}

	var cfResp CurseForgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var mods []CurseForgeMod
	if err := json.Unmarshal(cfResp.Data, &mods); err != nil {
		return nil, fmt.Errorf("failed to decode mods: %w", err)
	}

	result := &SearchResult{
		Mods:       mods,
		TotalCount: 0,
		PageIndex:  params.Index,
		PageSize:   params.PageSize,
	}
	
	if cfResp.Pagination != nil {
		result.TotalCount = cfResp.Pagination.TotalCount
	}

	return result, nil
}

// GetModDetails gets detailed info about a specific mod
func GetModDetails(ctx context.Context, modID int) (*CurseForgeMod, error) {
	url := fmt.Sprintf("%s/mods/%d", curseForgeBaseURL, modID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", cfAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mod not found: %d", modID)
	}

	var cfResp CurseForgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return nil, err
	}

	var mod CurseForgeMod
	if err := json.Unmarshal(cfResp.Data, &mod); err != nil {
		return nil, err
	}

	return &mod, nil
}

// GetModFiles gets available files for a mod
func GetModFiles(ctx context.Context, modID int) ([]ModFile, error) {
	url := fmt.Sprintf("%s/mods/%d/files", curseForgeBaseURL, modID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", cfAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var cfResp CurseForgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return nil, err
	}

	var files []ModFile
	if err := json.Unmarshal(cfResp.Data, &files); err != nil {
		return nil, err
	}

	return files, nil
}

// DownloadMod downloads and installs a mod
func DownloadMod(ctx context.Context, cfMod CurseForgeMod, progressCallback func(progress float64, message string)) error {
	if len(cfMod.LatestFiles) == 0 {
		return fmt.Errorf("no files available for mod %s", cfMod.Name)
	}

	// Get the latest file
	latestFile := cfMod.LatestFiles[0]
	for _, f := range cfMod.LatestFiles {
		if f.FileDate > latestFile.FileDate {
			latestFile = f
		}
	}

	if latestFile.DownloadURL == "" {
		return fmt.Errorf("download not available for this mod (author disabled distribution)")
	}

	modsDir := GetModsDir()
	if err := os.MkdirAll(modsDir, 0755); err != nil {
		return err
	}

	destPath := filepath.Join(modsDir, latestFile.FileName)

	if progressCallback != nil {
		progressCallback(0, fmt.Sprintf("Downloading %s...", cfMod.Name))
	}

	// Download the file
	if err := download.DownloadFile(ctx, latestFile.DownloadURL, destPath, func(downloaded, total int64, speed string) {
		if progressCallback != nil && total > 0 {
			progress := float64(downloaded) / float64(total) * 100
			progressCallback(progress, fmt.Sprintf("Downloading %s... %.1f%%", cfMod.Name, progress))
		}
	}); err != nil {
		os.Remove(destPath)
		return fmt.Errorf("failed to download mod: %w", err)
	}

	// Get author name
	authorName := "Unknown"
	if len(cfMod.Authors) > 0 {
		authorName = cfMod.Authors[0].Name
	}

	// Get category
	category := "General"
	if len(cfMod.Categories) > 0 {
		category = cfMod.Categories[0].Name
	}

	// Get icon URL
	iconURL := ""
	if cfMod.Logo != nil {
		iconURL = cfMod.Logo.ThumbnailURL
	}

	// Add to manifest
	mod := Mod{
		ID:           fmt.Sprintf("cf-%d", cfMod.ID),
		Name:         cfMod.Name,
		Version:      latestFile.DisplayName,
		Author:       authorName,
		Description:  cfMod.Summary,
		DownloadURL:  latestFile.DownloadURL,
		CurseForgeID: cfMod.ID,
		Enabled:      true,
		InstalledAt:  time.Now().Format(time.RFC3339),
		UpdatedAt:    time.Now().Format(time.RFC3339),
		FilePath:     destPath,
		IconURL:      iconURL,
		Downloads:    cfMod.DownloadCount,
		Category:     category,
	}

	if err := AddMod(mod); err != nil {
		return err
	}

	if progressCallback != nil {
		progressCallback(100, fmt.Sprintf("Installed %s successfully!", cfMod.Name))
	}

	return nil
}

// GetCategories gets available mod categories for Hytale
func GetCategories(ctx context.Context) ([]ModCategory, error) {
	url := fmt.Sprintf("%s/categories?gameId=%d", curseForgeBaseURL, hytaleGameID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", cfAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var cfResp CurseForgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return nil, err
	}

	var categories []ModCategory
	if err := json.Unmarshal(cfResp.Data, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

// CheckForUpdates checks if any installed mods have updates
func CheckForUpdates(ctx context.Context) ([]Mod, error) {
	mods, err := GetInstalledMods()
	if err != nil {
		return nil, err
	}

	var modsWithUpdates []Mod

	for _, mod := range mods {
		if mod.CurseForgeID == 0 {
			continue
		}

		cfMod, err := GetModDetails(ctx, mod.CurseForgeID)
		if err != nil {
			continue
		}

		// Check if there's a newer file
		for _, file := range cfMod.LatestFiles {
			if file.FileDate > mod.UpdatedAt {
				modsWithUpdates = append(modsWithUpdates, mod)
				break
			}
		}
	}

	return modsWithUpdates, nil
}
