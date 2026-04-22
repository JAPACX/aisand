package embed

import _ "embed"

// AgentYAML is the embedded Lima VM template with cloud-init provisioning.
//
//go:embed opencode-agent.yaml
var AgentYAML []byte

// OpenCodeScript is the embedded opencode installation script.
//
//go:embed scripts/opencode.sh
var OpenCodeScript []byte
