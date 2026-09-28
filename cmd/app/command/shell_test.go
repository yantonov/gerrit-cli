package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These all fail on dispatch, before touching $SHELL or the filesystem, so
// they are safe to run anywhere.
func TestShellRejectsBadSubcommands(t *testing.T) {
	cases := map[string][]string{
		"no subcommand":      {},
		"unknown subcommand": {"uninstall"},
		"two subcommands":    {"show", "install"},
	}

	for name, args := range cases {
		if err := shellCommand.run(nil, args); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestCurrentShellDetectsBash(t *testing.T) {
	cases := []string{"/bin/bash", "/usr/bin/bash.exe"}

	for _, value := range cases {
		t.Setenv("SHELL", value)

		name, err := currentShell()
		if err != nil {
			t.Errorf("SHELL=%q: unexpected error: %v", value, err)
		}
		if name != "bash" {
			t.Errorf("SHELL=%q: currentShell() = %q, want bash", value, name)
		}
	}
}

func TestCurrentShellRejectsUnsupportedOrMissingShell(t *testing.T) {
	t.Setenv("SHELL", "")
	if _, err := currentShell(); err == nil {
		t.Error("expected an error: SHELL is not set")
	}

	t.Setenv("SHELL", "/bin/zsh")
	_, err := currentShell()
	if err == nil {
		t.Fatal("expected an error: zsh is not supported")
	}
	if !strings.Contains(err.Error(), "zsh") {
		t.Errorf("error = %q, want it to name the unsupported shell", err)
	}
}

func TestBashCompletionScriptNamesTheProgram(t *testing.T) {
	script := bashCompletionScript("gerrit-cli")

	if !strings.Contains(script, "complete -o default -F _gerrit_cli_completion gerrit-cli") {
		t.Error("script does not register completion for gerrit-cli")
	}

	for _, name := range commandNames {
		if !strings.Contains(script, name) {
			t.Errorf("script does not mention command %q", name)
		}
	}
}

func TestShellShowRejectsUnsupportedShell(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")

	if err := shellCommand.run(nil, []string{"show"}); err == nil {
		t.Error("expected an error: zsh is not supported")
	}
}

func TestBashCompletionTargetDetectsGitBash(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	t.Setenv("MSYSTEM", "MINGW64")
	target, err := bashCompletionTarget("gerrit-cli")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := filepath.Join(home, "bash_completion.d", "gerrit-cli.bash"); target != want {
		t.Errorf("MSYSTEM set: bashCompletionTarget() = %q, want %q", target, want)
	}

	t.Setenv("MSYSTEM", "")
	target, err = bashCompletionTarget("gerrit-cli")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := filepath.Join(home, ".local", "share", "bash-completion", "completions", "gerrit-cli"); target != want {
		t.Errorf("MSYSTEM unset: bashCompletionTarget() = %q, want %q", target, want)
	}
}

func TestShellInstallWritesTheCompletionScript(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	t.Setenv("MSYSTEM", "")

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	if err := shellCommand.run(nil, []string{"install"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// shellCommand.run names the script after the running binary, which under
	// `go test` is the compiled test binary rather than gerrit-cli.
	program := programName(os.Args)
	target, err := bashCompletionTarget(program)
	if err != nil {
		t.Fatalf("bashCompletionTarget: %v", err)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("completion script was not written to %s: %v", target, err)
	}

	if string(content) != bashCompletionScript(program) {
		t.Error("installed script does not match shell show's own output")
	}
}

func TestShellInstallReplacesAnExistingFile(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	t.Setenv("MSYSTEM", "")

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	target, err := bashCompletionTarget(programName(os.Args))
	if err != nil {
		t.Fatalf("bashCompletionTarget: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("failed to prepare completions dir: %v", err)
	}
	if err := os.WriteFile(target, []byte("stale"), 0o644); err != nil {
		t.Fatalf("failed to seed a stale completion file: %v", err)
	}

	if err := shellCommand.run(nil, []string{"install"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("completion script was not written to %s: %v", target, err)
	}
	if string(content) == "stale" {
		t.Error("install did not replace the existing file")
	}
}
