package main

import "embed"

// VSCodeExtFS holds the bundled VS Code extension files.
// They are extracted at install time by "speek install-vscode".
//
//go:embed vscode-ext
var VSCodeExtFS embed.FS

// SpeekLogoICO holds the embedded .ico used for Windows file-type registration.
//
//go:embed public/speekLogo.ico
var SpeekLogoICO []byte
