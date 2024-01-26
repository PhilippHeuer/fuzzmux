package lookup

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/rs/zerolog/log"
)

type Project struct {
	Name         string   // Name is the name of the project
	Path         string   // Path is the absolute path to the project
	RelativePath string   // RelativePath is the path relative to the source directory
	Tags         []string // Tags are the tags of the project
}

type ScanResult struct {
	Projects []Project
	Error    error
}

func ScanForProjects(sources []config.SourceDirectory, checks []string) ([]Project, error) {
	log.Debug().Interface("directories", sources).Msg("searching for project directories")
	var (
		wg      sync.WaitGroup
		results = make(chan ScanResult, len(sources))
	)

	for _, source := range sources {
		wg.Add(1)
		go func(source config.SourceDirectory) {
			defer wg.Done()

			projects, err := scanDirectory(source, checks)
			results <- ScanResult{Projects: projects, Error: err}
		}(source)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var projects []Project
	for result := range results {
		if result.Error != nil {
			return nil, result.Error
		}
		projects = append(projects, result.Projects...)
	}

	return projects, nil
}

func scanDirectory(source config.SourceDirectory, checks []string) ([]Project, error) {
	var projects []Project

	// Compile regex patterns for exclusion
	var excludePatterns []*regexp.Regexp
	for _, pattern := range source.Exclude {
		excludePattern, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to compile exclude pattern '%s': %w", pattern, err)
		}
		excludePatterns = append(excludePatterns, excludePattern)
	}

	err := filepath.WalkDir(source.Directory, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip files
		if !info.IsDir() {
			return nil
		}

		// exclusion patterns
		for _, pattern := range excludePatterns {
			log.Trace().Str("path", path).Str("pattern", pattern.String()).Bool("match", pattern.MatchString(info.Name())).Msg("checking for matches with exclude pattern")
			if pattern.MatchString(info.Name()) {
				return filepath.SkipDir
			}
		}

		// check depth
		rel, err := filepath.Rel(source.Directory, path)
		if err != nil {
			return err
		}
		depth := strings.Count(filepath.ToSlash(rel), "/") + 1
		if depth > source.Depth {
			return filepath.SkipDir
		}

		// check
		for _, check := range checks {
			if _, err := os.Stat(filepath.Join(path, check)); err == nil {
				projects = append(projects, Project{
					Name:         filepath.Base(path),
					Path:         path,
					RelativePath: filepath.Base(source.Directory) + "/" + filepath.ToSlash(rel),
					Tags:         source.Tags,
				})
				return filepath.SkipDir
			}
		}

		return nil
	})

	return projects, err
}
