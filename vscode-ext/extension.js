const vscode = require('vscode');
const path = require('path');

function activate(context) {

    // Run current .spk file in the integrated terminal
    const runCmd = vscode.commands.registerCommand('speek.runFile', () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) {
            vscode.window.showErrorMessage('No file is open.');
            return;
        }

        const file = editor.document.fileName;
        if (!file.endsWith('.spk')) {
            vscode.window.showErrorMessage('This is not a Speek file (.spk)');
            return;
        }

        editor.document.save().then(() => {
            const terminal = getOrCreateTerminal();
            terminal.show(true);
            terminal.sendText(`speek run "${file}"`);
        });
    });

    // Run with --debug flag (shows intent panel)
    const debugCmd = vscode.commands.registerCommand('speek.debugFile', () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) return;

        const file = editor.document.fileName;
        if (!file.endsWith('.spk')) return;

        editor.document.save().then(() => {
            const terminal = getOrCreateTerminal();
            terminal.show(true);
            terminal.sendText(`speek run "${file}" --debug`);
        });
    });

    // Check / validate without running
    const checkCmd = vscode.commands.registerCommand('speek.checkFile', () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) return;

        const file = editor.document.fileName;
        if (!file.endsWith('.spk')) return;

        editor.document.save().then(() => {
            const terminal = getOrCreateTerminal();
            terminal.show(true);
            terminal.sendText(`speek check "${file}"`);
        });
    });

    context.subscriptions.push(runCmd, debugCmd, checkCmd);
}

function getOrCreateTerminal() {
    const existing = vscode.window.terminals.find(t => t.name === 'Speek');
    if (existing) return existing;
    return vscode.window.createTerminal({ name: 'Speek' });
}

function deactivate() {}

module.exports = { activate, deactivate };
