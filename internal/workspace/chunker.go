package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/redhotvengeance/promptd/internal/promptd"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

const (
	linesPerChunk = 100
	overlapLines  = 20
)

type ParserConfig struct {
	Language *sitter.Language
	Query    string
}

var languageRegistry = map[string]ParserConfig{
	".go": {
		Language: golang.GetLanguage(),
		Query: `
			(function_declaration) @function
			(method_declaration) @method
			(type_declaration) @type
		`,
	},
}

func (i *Indexer) ParseFile(ctx context.Context, path string) ([]promptd.Chunk, error) {
	ext := filepath.Ext(path)
	config, supported := languageRegistry[ext]

	if !supported {
		return i.NaiveLineChunker(path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(config.Language)

	tree, err := parser.ParseCtx(ctx, nil, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AST: %w", err)
	}
	defer tree.Close()

	q, err := sitter.NewQuery([]byte(config.Query), config.Language)
	if err != nil {
		return nil, fmt.Errorf("failted to compile query: %w", err)
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(q, tree.RootNode())

	var chunks []promptd.Chunk
	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}

		match = qc.FilterPredicates(match, content)

		for _, capture := range match.Captures {
			node := capture.Node

			codeBlock := node.Content(content)

			chunks = append(chunks, promptd.Chunk{
				FilePath: path,
				Content:  codeBlock,
			})
		}
	}

	return chunks, nil
}

func (i *Indexer) NaiveLineChunker(path string) ([]promptd.Chunk, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")

	if len(lines) <= linesPerChunk {
		return []promptd.Chunk{
			{
				FilePath: path,
				Content:  string(content),
			},
		}, nil
	}

	var chunks []promptd.Chunk
	step := linesPerChunk - overlapLines
	if step <= 0 {
		step = 1
	}

	for start := 0; start < len(lines); start += step {
		end := min(start+linesPerChunk, len(lines))

		chunkText := strings.Join(lines[start:end], "\n")

		if strings.TrimSpace(chunkText) == "" {
			if end == len(lines) {
				break
			}

			continue
		}

		chunks = append(chunks, promptd.Chunk{
			FilePath: path,
			Content:  chunkText,
		})

		if end == len(lines) {
			break
		}
	}

	return chunks, nil
}
