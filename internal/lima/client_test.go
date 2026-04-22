package lima

import (
	"testing"
)

func TestParseListOutput_SingleVM(t *testing.T) {
	input := `{"name":"myvm","status":"Running","cpus":2,"memory":2147483648,"disk":64424509440,"mounts":[{"location":"/Users/test","writable":false}]}`
	vms, err := parseListOutput([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vms) != 1 {
		t.Fatalf("expected 1 VM, got %d", len(vms))
	}
	vm := vms[0]
	if vm.Name != "myvm" {
		t.Errorf("expected name 'myvm', got %q", vm.Name)
	}
	if vm.Status != "Running" {
		t.Errorf("expected status 'Running', got %q", vm.Status)
	}
	if vm.CPUs != 2 {
		t.Errorf("expected 2 CPUs, got %d", vm.CPUs)
	}
	if len(vm.Mounts) != 1 {
		t.Fatalf("expected 1 mount, got %d", len(vm.Mounts))
	}
	if vm.Mounts[0].Location != "/Users/test" {
		t.Errorf("expected mount location '/Users/test', got %q", vm.Mounts[0].Location)
	}
	if vm.Mounts[0].Writable {
		t.Error("expected mount to be read-only")
	}
}

func TestParseListOutput_MultipleVMs(t *testing.T) {
	input := `{"name":"vm1","status":"Running","cpus":2,"memory":2147483648,"disk":64424509440,"mounts":[]}
{"name":"vm2","status":"Stopped","cpus":4,"memory":4294967296,"disk":107374182400,"mounts":[]}`
	vms, err := parseListOutput([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vms) != 2 {
		t.Fatalf("expected 2 VMs, got %d", len(vms))
	}
	if vms[0].Name != "vm1" || vms[1].Name != "vm2" {
		t.Errorf("unexpected VM names: %q, %q", vms[0].Name, vms[1].Name)
	}
}

func TestParseListOutput_Empty(t *testing.T) {
	vms, err := parseListOutput([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error on empty input: %v", err)
	}
	if len(vms) != 0 {
		t.Errorf("expected 0 VMs, got %d", len(vms))
	}
}

func TestParseListOutput_InvalidJSON(t *testing.T) {
	_, err := parseListOutput([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestCreateVMCommand(t *testing.T) {
	c := NewClient()
	cmd := c.CreateVM("testvm", 2, 4, 60, []string{"/Users/test/projects"}, "/tmp/template.yaml")
	args := cmd.Args
	// Check key args are present
	found := false
	for _, a := range args {
		if a == "testvm" {
			found = true
		}
	}
	if !found {
		t.Error("expected VM name 'testvm' in command args")
	}
}

func TestShellVMCommand(t *testing.T) {
	c := NewClient()
	cmd := c.ShellVM("myvm")
	if cmd.Path == "" {
		t.Error("expected non-empty command path")
	}
	// Should contain limactl shell myvm
	found := false
	for _, a := range cmd.Args {
		if a == "myvm" {
			found = true
		}
	}
	if !found {
		t.Error("expected VM name 'myvm' in shell command args")
	}
}
