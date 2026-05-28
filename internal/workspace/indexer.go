package workspace

import (
	"context"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/redhotvengeance/promptd/internal/promptd"
)

var ignoredDirs = map[string]bool{
	".git":   true,
	"bin":    true,
	"build":  true,
	"vendor": true,
}

var ignoredExts = map[string]bool{
	".exe": true,
	".dll": true,
	".zip": true,
	".tar": true,
	".gz":  true,
}

type Indexer struct {
	llm   promptd.LLM
	store promptd.Datastore
}

func NewIndexer(llm promptd.LLM, store promptd.Datastore) *Indexer {
	return &Indexer{
		llm:   llm,
		store: store,
	}
}

func (i *Indexer) DiscoverFiles(ctx context.Context, rootPath string) ([]string, error) {
	var targetFiles []string

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		name := d.Name()

		if d.IsDir() {
			if ignoredDirs[name] {
				return filepath.SkipDir
			}

			return nil
		}

		ext := strings.ToLower(filepath.Ext(name))
		if ignoredExts[ext] {
			return nil
		}

		targetFiles = append(targetFiles, path)

		return nil
	})

	return targetFiles, err
}

func (i *Indexer) IndexWorkspace(ctx context.Context, rootPath string) error {
	log.Printf("=> Starting index of workspace: %s", rootPath)

	files, err := i.DiscoverFiles(ctx, rootPath)
	if err != nil {
		return err
	}

	var totalChunks int
	for _, file := range files {
		chunks, err := i.ParseFile(ctx, file)
		if err != nil {
			log.Printf("Warning: failed to parse %s: %v", file, err)

			continue
		}

		var embeddedChunks []promptd.Chunk
		for _, chunk := range chunks {
			vector, err := i.llm.Embed(ctx, chunk.Content, "")
			if err != nil {
				log.Printf("Warning: failed to embed chunk in %s: %v", file, err)

				continue
			}

			chunk.ID = uuid.NewString()
			chunk.Embedding = vector
			chunk.WorkspacePath = rootPath

			embeddedChunks = append(embeddedChunks, chunk)
		}

		if len(embeddedChunks) > 0 {
			err = i.store.Workspaces().InsertChunks(ctx, embeddedChunks)
			if err != nil {
				log.Printf("Warning: failed to save chunks to DB: %v", err)
			} else {
				totalChunks += len(embeddedChunks)
			}
		}
	}

	log.Printf("=> Indexing complete! Embedded %d chunks.", totalChunks)

	return nil
}
