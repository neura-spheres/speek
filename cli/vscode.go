package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

// extensionID must match "publisher.name-version" exactly -- VS Code uses this to identify the folder
const extensionID = "speek-lang.speek-0.1.0"

func InstallVSCode(extFS fs.FS, force bool) int {
	if extFS == nil {
		fmt.Println("VS Code extension update requires re-running: speek install-vscode --force")
		return 0
	}
	dirs := findVSCodeExtDirs()
	if len(dirs) == 0 {
		fmt.Println("No VS Code installation found.")
		fmt.Println("Install VS Code first, then run: speek install-vscode")
		return 1
	}

	anyFailed := false
	for _, extDir := range dirs {
		dest := filepath.Join(extDir, extensionID)
		installed, err := installExtension(extFS, dest, force)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to install into %s: %v\n", dest, err)
			anyFailed = true
		} else if installed {
			fmt.Printf("Installed -> %s\n", dest)
		}
	}

	if anyFailed {
		return 1
	}

	fmt.Println()
	fmt.Println("Done! Restart VS Code and open any .spk file to see syntax highlighting.")
	return 0
}

func installExtension(extFS fs.FS, dest string, force bool) (bool, error) {
	if _, err := os.Stat(dest); err == nil {
		if !force {
			fmt.Printf("  Skipped %s (already installed, use --force to update)\n", dest)
			return false, nil
		}
		if err := os.RemoveAll(dest); err != nil {
			return false, fmt.Errorf("cannot remove existing installation: %w", err)
		}
	}

	err := fs.WalkDir(extFS, "vscode-ext", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel("vscode-ext", path)
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := fs.ReadFile(extFS, path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func findVSCodeExtDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	candidates := []string{
		filepath.Join(home, ".vscode", "extensions"),
		filepath.Join(home, ".vscode-insiders", "extensions"),
		filepath.Join(home, ".vscode-server", "extensions"),
		filepath.Join(home, ".cursor", "extensions"),
		filepath.Join(home, ".vscodium", "extensions"),
		filepath.Join(home, ".vscode-oss", "extensions"),
	}

	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			candidates = append(candidates,
				filepath.Join(appData, "Code", "User", "extensions"),
				filepath.Join(appData, "Code - Insiders", "User", "extensions"),
				filepath.Join(appData, "Cursor", "User", "extensions"),
			)
		}
	}

	if runtime.GOOS == "darwin" {
		candidates = append(candidates,
			filepath.Join(home, "Library", "Application Support", "Code", "User", "extensions"),
			filepath.Join(home, "Library", "Application Support", "Cursor", "User", "extensions"),
		)
	}

	var found []string
	for _, dir := range candidates {
		// the extensions folder might not exist yet, but the parent app data dir must
		parent := filepath.Dir(dir)
		if _, err := os.Stat(parent); err == nil {
			found = append(found, dir)
		}
	}
	return found
}
