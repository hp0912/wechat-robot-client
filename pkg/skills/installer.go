package skills

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Installer 负责从外部源安装 Skill
type Installer struct {
	baseDir string
}

// NewInstaller 创建安装器
func NewInstaller(baseDir string) *Installer {
	return &Installer{baseDir: baseDir}
}

// InstallFromGit 从 Git 仓库安装 Skill
// repoURL: 仓库地址 (e.g. https://github.com/anthropics/skills)
// subPath: 仓库中的子路径 (e.g. skills/pptx)
// ref: Git ref (branch/tag/commit, e.g. main)
// 返回安装后的 Skill 本地目录
func (inst *Installer) InstallFromGit(repoURL, subPath, ref string) (string, error) {
	if ref == "" {
		ref = "main"
	}

	// 提取 Skill 名称（子路径的最后一段）
	skillName := filepath.Base(subPath)
	if skillName == "" || skillName == "." {
		return "", fmt.Errorf("invalid subPath: %s", subPath)
	}

	targetDir := filepath.Join(inst.baseDir, skillName)

	// 如果目标目录已存在，先移除（更新安装）
	if _, err := os.Stat(targetDir); err == nil {
		log.Printf("[Skills] Removing existing skill directory: %s", targetDir)
		if err := os.RemoveAll(targetDir); err != nil {
			return "", fmt.Errorf("failed to remove existing skill: %w", err)
		}
	}

	// 使用 sparse checkout 只拉取需要的子目录
	tmpDir, err := os.MkdirTemp("", "skill-install-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// git clone --depth=1 --filter=blob:none --sparse
	cloneCmd := exec.Command("git", "clone",
		"--depth=1",
		"--filter=blob:none",
		"--sparse",
		"--branch", ref,
		repoURL,
		tmpDir,
	)
	cloneCmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git clone failed: %w\n%s", err, string(output))
	}

	// git sparse-checkout set <subPath>
	sparseCmd := exec.Command("git", "sparse-checkout", "set", subPath)
	sparseCmd.Dir = tmpDir
	if output, err := sparseCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git sparse-checkout failed: %w\n%s", err, string(output))
	}

	// 确认 SKILL.md 存在
	srcSkillDir := filepath.Join(tmpDir, subPath)
	skillMDPath := filepath.Join(srcSkillDir, "SKILL.md")
	if _, err := os.Stat(skillMDPath); os.IsNotExist(err) {
		return "", fmt.Errorf("SKILL.md not found at %s in repository", subPath)
	}

	// 复制到目标目录
	if err := copyDir(srcSkillDir, targetDir); err != nil {
		return "", fmt.Errorf("failed to copy skill: %w", err)
	}

	log.Printf("[Skills] Installed from git: %s/%s@%s -> %s", repoURL, subPath, ref, targetDir)
	return targetDir, nil
}

// InstallFromLocal 从本地路径安装 Skill（复制到管理目录）
func (inst *Installer) InstallFromLocal(srcDir string) (string, error) {
	// 确认 SKILL.md 存在
	skillMDPath := filepath.Join(srcDir, "SKILL.md")
	if _, err := os.Stat(skillMDPath); os.IsNotExist(err) {
		return "", fmt.Errorf("SKILL.md not found in %s", srcDir)
	}

	// 加载元数据获取名称
	meta, err := LoadSkillMetadataOnly(srcDir)
	if err != nil {
		return "", err
	}

	targetDir := filepath.Join(inst.baseDir, meta.Name)

	// 如果源和目标相同，不需要复制
	absSrc, _ := filepath.Abs(srcDir)
	absTgt, _ := filepath.Abs(targetDir)
	if absSrc == absTgt {
		return targetDir, nil
	}

	// 复制
	if err := os.RemoveAll(targetDir); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to clean target: %w", err)
	}

	if err := copyDir(srcDir, targetDir); err != nil {
		return "", fmt.Errorf("failed to copy skill: %w", err)
	}

	return targetDir, nil
}

// copyDir 递归复制目录
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// 跳过 .git 目录
		if entry.Name() == ".git" {
			continue
		}

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile 复制单个文件
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, srcInfo.Mode())
}

// ExtractRepoAndSubPath 从 Git 托管平台 URL 中提取仓库地址和子路径
// 支持格式：
//   - https://github.com/anthropics/skills/tree/main/skills/pptx  (GitHub)
//   - https://git.houhoukang.com/owner/repo/src/branch/main/skills/kfc  (Gitea)
//   - https://github.com/anthropics/skills  + subPath: skills/pptx
func ExtractRepoAndSubPath(url string) (repoURL, subPath, ref string) {
	// 尝试解析 Gitea src/branch URL（如 git.houhoukang.com）
	if repoURL, remaining, ok := strings.Cut(url, "/src/branch/"); ok {
		ref, subPath, _ = strings.Cut(remaining, "/")
		return repoURL, subPath, ref
	}

	// 尝试解析 GitHub tree URL
	if repoURL, remaining, ok := strings.Cut(url, "/tree/"); ok {
		ref, subPath, _ = strings.Cut(remaining, "/")
		return repoURL, subPath, ref
	}

	// 普通仓库 URL
	repoURL = url
	ref = "main"
	return
}
