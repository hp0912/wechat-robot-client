package skills

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseSKILLMD 解析 SKILL.md 文件，返回元数据和 body 指令
func ParseSKILLMD(filePath string) (*SkillMetadata, string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read SKILL.md: %w", err)
	}

	return ParseSKILLMDContent(data)
}

// ParseSKILLMDContent 解析 SKILL.md 内容字节
func ParseSKILLMDContent(data []byte) (*SkillMetadata, string, error) {
	frontmatter, body, err := splitFrontmatter(data)
	if err != nil {
		return nil, "", err
	}

	var meta SkillMetadata
	if err := yaml.Unmarshal(frontmatter, &meta); err != nil {
		return nil, "", fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// 校验必填字段
	if err := validateMetadata(&meta); err != nil {
		return nil, "", err
	}

	return &meta, strings.TrimSpace(string(body)), nil
}

// splitFrontmatter 分割 YAML frontmatter 和 markdown body
func splitFrontmatter(data []byte) ([]byte, []byte, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	// 寻找开头的 ---
	foundStart := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "---" {
			foundStart = true
			break
		}
		return nil, nil, fmt.Errorf("SKILL.md must start with YAML frontmatter (---)")
	}

	if !foundStart {
		return nil, nil, fmt.Errorf("SKILL.md must contain YAML frontmatter")
	}

	// 收集 frontmatter 内容直到遇到结尾的 ---
	var frontmatterBuf bytes.Buffer
	foundEnd := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			foundEnd = true
			break
		}
		frontmatterBuf.WriteString(line)
		frontmatterBuf.WriteString("\n")
	}

	if !foundEnd {
		return nil, nil, fmt.Errorf("SKILL.md frontmatter not properly closed with ---")
	}

	// 剩余的都是 body
	var bodyBuf bytes.Buffer
	for scanner.Scan() {
		bodyBuf.WriteString(scanner.Text())
		bodyBuf.WriteString("\n")
	}

	return frontmatterBuf.Bytes(), bodyBuf.Bytes(), nil
}

// namePattern 合法的 skill name 正则
var namePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// validateMetadata 校验 Skill 元数据
func validateMetadata(meta *SkillMetadata) error {
	// name 校验
	if meta.Name == "" {
		return fmt.Errorf("skill name is required")
	}
	if len(meta.Name) > 64 {
		return fmt.Errorf("skill name must be at most 64 characters")
	}
	if !namePattern.MatchString(meta.Name) {
		return fmt.Errorf("skill name must contain only lowercase letters, numbers, and hyphens, must not start/end with hyphen")
	}
	if strings.Contains(meta.Name, "--") {
		return fmt.Errorf("skill name must not contain consecutive hyphens")
	}

	// description 校验
	if meta.Description == "" {
		return fmt.Errorf("skill description is required")
	}
	if len(meta.Description) > 1024 {
		return fmt.Errorf("skill description must be at most 1024 characters")
	}

	// compatibility 校验
	if len(meta.Compatibility) > 500 {
		return fmt.Errorf("skill compatibility must be at most 500 characters")
	}

	return nil
}

// DiscoverSkills 在指定目录下发现所有有效的 Skill（包含 SKILL.md 的子目录）
func DiscoverSkills(baseDir string) ([]string, error) {
	var skillPaths []string

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read skills directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillMDPath := filepath.Join(baseDir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillMDPath); err == nil {
			skillPaths = append(skillPaths, filepath.Join(baseDir, entry.Name()))
		}
	}

	return skillPaths, nil
}

// LoadSkillMetadataOnly 仅加载元数据（轻量，用于启动时注入 prompt）
func LoadSkillMetadataOnly(skillDir string) (*SkillMetadata, error) {
	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	meta, _, err := ParseSKILLMD(skillMDPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse skill at %s: %w", skillDir, err)
	}
	return meta, nil
}

// LoadSkillFull 加载完整的 Skill（包含 instructions body）
func LoadSkillFull(skillDir string) (*Skill, error) {
	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	meta, body, err := ParseSKILLMD(skillMDPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse skill at %s: %w", skillDir, err)
	}

	return &Skill{
		SkillMetadata: *meta,
		Instructions:  body,
		Path:          skillDir,
	}, nil
}

// ReadSkillResource 读取 Skill 中的附属文件（scripts/、references/、assets/ 等）
func ReadSkillResource(skillDir, relativePath string) (string, error) {
	// 安全检查：防止路径穿越
	absPath := filepath.Join(skillDir, relativePath)
	absSkillDir, _ := filepath.Abs(skillDir)
	absFile, _ := filepath.Abs(absPath)

	if !strings.HasPrefix(absFile, absSkillDir) {
		return "", fmt.Errorf("path traversal detected: %s", relativePath)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read resource %s: %w", relativePath, err)
	}

	return string(data), nil
}
