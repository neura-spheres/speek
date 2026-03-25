package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func RegisterFileType(iconData []byte) int {
	installDir := filepath.Join(os.Getenv("USERPROFILE"), ".speek")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "register-filetype: cannot create install dir: %v\n", err)
		return 1
	}

	// Write icon to ~/.speek/speekLogo.ico
	iconPath := filepath.Join(installDir, "speekLogo.ico")
	if err := os.WriteFile(iconPath, iconData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "register-filetype: cannot write icon: %v\n", err)
		return 1
	}

	speekExe := filepath.Join(installDir, "speek.exe")

	type regEntry struct{ key, value, data string }
	entries := []regEntry{
		{`HKCU\Software\Classes\.spk`, "", "SpeekFile"},
		{`HKCU\Software\Classes\SpeekFile`, "", "Speek Script"},
		{`HKCU\Software\Classes\SpeekFile\DefaultIcon`, "", iconPath + ",0"},
		{`HKCU\Software\Classes\SpeekFile\shell\open\command`, "", `"` + speekExe + `" run "%1"`},
	}

	for _, e := range entries {
		args := []string{"add", e.key, "/ve", "/d", e.data, "/f"}
		if e.value != "" {
			args = []string{"add", e.key, "/v", e.value, "/d", e.data, "/f"}
		}
		if out, err := exec.Command("reg", args...).CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "register-filetype: reg add failed: %v\n%s\n", err, out)
			return 1
		}
	}

	fmt.Println("Registered .spk file type with Speek icon.")
	fmt.Println("You may need to restart Explorer for the icon to appear.")
	return 0
}
