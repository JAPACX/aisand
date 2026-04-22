package lima

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Client wraps limactl CLI operations.
type Client struct{}

// NewClient creates a new Lima client.
func NewClient() *Client {
	return &Client{}
}

// IsLimactlInstalled returns true if limactl is found in PATH.
func (c *Client) IsLimactlInstalled() bool {
	_, err := exec.LookPath("limactl")
	return err == nil
}

// IsBrewInstalled returns true if brew is found in PATH.
func (c *Client) IsBrewInstalled() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

// GetHostCPUs returns the number of logical CPUs on the host.
func (c *Client) GetHostCPUs() int {
	out, err := exec.Command("sysctl", "-n", "hw.logicalcpu").Output()
	if err != nil {
		return runtime.NumCPU()
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return runtime.NumCPU()
	}
	return n
}

// GetHostRAMGB returns the host RAM in gigabytes.
func (c *Client) GetHostRAMGB() int {
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return 8
	}
	bytes, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 8
	}
	return int(bytes / (1024 * 1024 * 1024))
}

// limaListEntry is the raw JSON structure from `limactl list --format json`.
type limaListEntry struct {
	Name   string      `json:"name"`
	Status string      `json:"status"`
	CPUs   int         `json:"cpus"`
	Memory int64       `json:"memory"`
	Disk   int64       `json:"disk"`
	Mounts []limaMount `json:"mounts"`
}

type limaMount struct {
	Location string `json:"location"`
	Writable bool   `json:"writable"`
}

// ListVMs returns all Lima VMs by parsing `limactl list --format json`.
func (c *Client) ListVMs() ([]VM, error) {
	out, err := exec.Command("limactl", "list", "--format", "json").Output()
	if err != nil {
		return nil, fmt.Errorf("limactl list failed: %w", err)
	}
	return parseListOutput(out)
}

// parseListOutput parses the JSON output from `limactl list --format json`.
// limactl outputs one JSON object per line (NDJSON), not a JSON array.
func parseListOutput(data []byte) ([]VM, error) {
	var vms []VM
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry limaListEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("failed to parse VM entry: %w", err)
		}
		vm := VM{
			Name:   entry.Name,
			Status: entry.Status,
			CPUs:   entry.CPUs,
			Memory: entry.Memory,
			Disk:   entry.Disk,
		}
		for _, m := range entry.Mounts {
			vm.Mounts = append(vm.Mounts, Mount{
				Location: m.Location,
				Writable: m.Writable,
			})
		}
		vms = append(vms, vm)
	}
	return vms, nil
}

// StartVM returns a command to start a VM (for streaming output).
func (c *Client) StartVM(name string) *exec.Cmd {
	return exec.Command("limactl", "start", name)
}

// StopVM2 returns a command to stop a VM (for streaming output).
func (c *Client) StopVM2(name string) *exec.Cmd {
	return exec.Command("limactl", "stop", name)
}

// StopVM stops a running VM (blocking, used internally).
func (c *Client) StopVM(name string) error {
	out, err := exec.Command("limactl", "stop", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("limactl stop %s failed: %w\n%s", name, err, out)
	}
	return nil
}

// DeleteVM stops the VM if running, then deletes it (blocking, used internally).
func (c *Client) DeleteVM(name string) error {
	// Try to stop first (ignore error if already stopped)
	_ = exec.Command("limactl", "stop", name).Run()
	out, err := exec.Command("limactl", "delete", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("limactl delete %s failed: %w\n%s", name, err, out)
	}
	return nil
}

// DeleteVMCmd returns a command to delete a VM (for streaming output).
// Uses --force so limactl handles stop+delete in one pass.
func (c *Client) DeleteVMCmd(name string) *exec.Cmd {
	return exec.Command("limactl", "delete", "--force", name)
}

// CreateVM returns a command to create a VM (for streaming output).
// templatePath is the path to the Lima YAML template file.
func (c *Client) CreateVM(name string, cpus, ram, disk int, mounts []string, templatePath string) *exec.Cmd {
	args := []string{
		"create",
		"--name", name,
		fmt.Sprintf("--cpus=%d", cpus),
		fmt.Sprintf("--memory=%d", ram),
		fmt.Sprintf("--disk=%d", disk),
	}
	for _, m := range mounts {
		args = append(args, "--mount", m+":w")
	}
	args = append(args, templatePath)
	return exec.Command("limactl", args...)
}

// ShellVM returns a command to open an interactive shell in a VM.
// Use with tea.ExecProcess to suspend the TUI.
func (c *Client) ShellVM(name string) *exec.Cmd {
	return exec.Command("limactl", "shell", name)
}

// AddMount adds a host directory as a read-write mount to a VM.
func (c *Client) AddMount(name, path string) error {
	out, err := exec.Command("limactl", "edit", name, "--mount", path+":w", "--tty=false").CombinedOutput()
	if err != nil {
		return fmt.Errorf("limactl edit (add mount) for %s failed: %w\n%s", name, err, out)
	}
	return nil
}

// RemoveMount removes a mount from a VM by rewriting the mount list without the removed one.
// If remainingMounts is empty, uses --mount-none.
func (c *Client) RemoveMount(name string, remainingMounts []Mount) error {
	if len(remainingMounts) == 0 {
		out, err := exec.Command("limactl", "edit", name, "--mount-none", "--tty=false").CombinedOutput()
		if err != nil {
			return fmt.Errorf("limactl edit (remove all mounts) failed: %w\n%s", err, out)
		}
		return nil
	}
	args := []string{"edit", name, "--tty=false"}
	for _, m := range remainingMounts {
		suffix := ":r"
		if m.Writable {
			suffix = ":w"
		}
		args = append(args, "--mount", m.Location+suffix)
	}
	out, err := exec.Command("limactl", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("limactl edit (update mounts) failed: %w\n%s", err, out)
	}
	return nil
}

// InstallTool returns a command that pipes scriptContent to bash inside the VM.
func (c *Client) InstallTool(name string, scriptContent []byte) *exec.Cmd {
	cmd := exec.Command("limactl", "shell", name, "bash", "-s")
	cmd.Stdin = strings.NewReader(string(scriptContent))
	return cmd
}

// StopAllVMs stops all currently running VMs.
func (c *Client) StopAllVMs(vms []VM) error {
	var errs []string
	for _, vm := range vms {
		if vm.Status == "Running" {
			if err := c.StopVM(vm.Name); err != nil {
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors stopping VMs: %s", strings.Join(errs, "; "))
	}
	return nil
}

// WriteDefaultConfig saves default resource config to ~/.lima/_config/default.yaml.
func (c *Client) WriteDefaultConfig(cpus, ramGB, diskGB int) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := home + "/.lima/_config"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	content := fmt.Sprintf("cpus: %d\nmemory: %dGiB\ndisk: %dGiB\n", cpus, ramGB, diskGB)
	return os.WriteFile(dir+"/default.yaml", []byte(content), 0644)
}
