package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSKILLMDContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantName string
		wantDesc string
		wantBody string
		wantErr  bool
	}{
		{
			name:     "valid basic skill",
			content:  "---\nname: pdf-processing\ndescription: Extract text and tables from PDF files.\n---\n\n# PDF Processing\n\nUse this skill for PDF files.\n",
			wantName: "pdf-processing",
			wantDesc: "Extract text and tables from PDF files.",
			wantBody: "# PDF Processing\n\nUse this skill for PDF files.",
			wantErr:  false,
		},
		{
			name:     "valid skill with all fields",
			content:  "---\nname: data-analysis\ndescription: Analyze datasets, generate charts.\nlicense: Apache-2.0\ncompatibility: Requires python3\nmetadata:\n  author: test-org\n  version: \"1.0\"\n---\n\n# Data Analysis\n\nInstructions here.\n",
			wantName: "data-analysis",
			wantDesc: "Analyze datasets, generate charts.",
			wantErr:  false,
		},
		{
			name:    "missing name",
			content: "---\ndescription: Some description\n---\n\nBody content\n",
			wantErr: true,
		},
		{
			name:    "missing description",
			content: "---\nname: test-skill\n---\n\nBody content\n",
			wantErr: true,
		},
		{
			name:    "invalid name with uppercase",
			content: "---\nname: Invalid-Skill\ndescription: A skill\n---\n\nBody\n",
			wantErr: true,
		},
		{
			name:    "name with consecutive hyphens",
			content: "---\nname: bad--name\ndescription: A skill\n---\n\nBody\n",
			wantErr: true,
		},
		{
			name:    "no frontmatter",
			content: "Just some content",
			wantErr: true,
		},
		{
			name:    "unclosed frontmatter",
			content: "---\nname: test\ndescription: test\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, body, err := ParseSKILLMDContent([]byte(tt.content))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if meta.Name != tt.wantName {
				t.Errorf("name = %q, want %q", meta.Name, tt.wantName)
			}
			if meta.Description != tt.wantDesc {
				t.Errorf("desc = %q, want %q", meta.Description, tt.wantDesc)
			}
			if tt.wantBody != "" && body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

func TestDiscoverSkills(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid skill directory
	skillDir := filepath.Join(tmpDir, "test-skill")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: test-skill\ndescription: A test skill\n---\n\nTest instructions.\n"), 0644)

	// Create invalid directory (no SKILL.md)
	os.MkdirAll(filepath.Join(tmpDir, "not-a-skill"), 0755)

	// Create a file (not dir, should skip)
	os.WriteFile(filepath.Join(tmpDir, "somefile.txt"), []byte("data"), 0644)

	paths, err := DiscoverSkills(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(paths))
	}
	if filepath.Base(paths[0]) != "test-skill" {
		t.Errorf("expected skill path to end with test-skill, got %s", paths[0])
	}
}

func TestExtractRepoAndSubPath(t *testing.T) {
	tests := []struct {
		url         string
		wantRepo    string
		wantSubPath string
		wantRef     string
	}{
		{
			url:         "https://github.com/anthropics/skills/tree/main/skills/pptx",
			wantRepo:    "https://github.com/anthropics/skills",
			wantSubPath: "skills/pptx",
			wantRef:     "main",
		},
		{
			url:         "https://github.com/anthropics/skills",
			wantRepo:    "https://github.com/anthropics/skills",
			wantSubPath: "",
			wantRef:     "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			repo, sub, ref := ExtractRepoAndSubPath(tt.url)
			if repo != tt.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
			}
			if sub != tt.wantSubPath {
				t.Errorf("subPath = %q, want %q", sub, tt.wantSubPath)
			}
			if ref != tt.wantRef {
				t.Errorf("ref = %q, want %q", ref, tt.wantRef)
			}
		})
	}
}
