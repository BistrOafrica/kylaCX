package ivr

import "fmt"

// Issue is a single validation finding. Severities: "error" (fatal — flow
// can't run) and "warning" (flow runs but with caveats).
type Issue struct {
	Severity string
	NodeID   string
	Code     string
	Message  string
}

// validateFlow walks the node graph and reports issues. Returns the issue
// list and a reachability map keyed by node_id (true = reachable from
// start_node_id).
//
// Checks:
//   - definition has a start_node_id
//   - start_node_id refers to a real node
//   - all branch targets resolve to existing nodes
//   - next_node_id (when non-empty) resolves to an existing node
//   - per-node config has the keys the executor requires (best-effort —
//     mirrors what executor.executeNode demands at runtime)
//   - all nodes are reachable from start_node_id
//   - menu node branches use single-character keys (DTMF digits) — warning
//     if not, since FreeSWITCH play_and_get_digits won't match longer strings
//     against a single press.
func validateFlow(def Definition) ([]Issue, map[string]bool) {
	var issues []Issue
	reachability := map[string]bool{}

	if def.StartNodeID == "" {
		issues = append(issues, Issue{
			Severity: "error", Code: "missing_start",
			Message: "definition.start_node_id is empty",
		})
		return issues, reachability
	}

	// Index nodes for fast lookup.
	byID := map[string]*Node{}
	for i := range def.Nodes {
		byID[def.Nodes[i].ID] = &def.Nodes[i]
	}

	if _, ok := byID[def.StartNodeID]; !ok {
		issues = append(issues, Issue{
			Severity: "error", Code: "missing_start",
			Message: fmt.Sprintf("start_node_id %q is not present in nodes", def.StartNodeID),
		})
		return issues, reachability
	}

	// Per-node config + branch checks.
	for _, n := range def.Nodes {
		issues = append(issues, validateNodeConfig(n)...)
		for digit, target := range n.Branches {
			if target == "" {
				issues = append(issues, Issue{
					Severity: "error", NodeID: n.ID, Code: "bad_branch",
					Message: fmt.Sprintf("branch %q has no target", digit),
				})
				continue
			}
			if _, ok := byID[target]; !ok {
				issues = append(issues, Issue{
					Severity: "error", NodeID: n.ID, Code: "bad_branch",
					Message: fmt.Sprintf("branch %q -> %q references a missing node", digit, target),
				})
			}
			if n.Type == NodeMenu && len(digit) != 1 && digit != "default" {
				issues = append(issues, Issue{
					Severity: "warning", NodeID: n.ID, Code: "menu_branch_key",
					Message: fmt.Sprintf("menu branch key %q is not a single DTMF digit", digit),
				})
			}
		}
		if n.NextNodeID != "" {
			if _, ok := byID[n.NextNodeID]; !ok {
				issues = append(issues, Issue{
					Severity: "error", NodeID: n.ID, Code: "bad_next",
					Message: fmt.Sprintf("next_node_id %q references a missing node", n.NextNodeID),
				})
			}
		}
	}

	// Reachability via BFS from start.
	frontier := []string{def.StartNodeID}
	reachability[def.StartNodeID] = true
	for len(frontier) > 0 {
		next := frontier[0]
		frontier = frontier[1:]
		node, ok := byID[next]
		if !ok {
			continue
		}
		for _, target := range node.Branches {
			if target == "" {
				continue
			}
			if !reachability[target] {
				reachability[target] = true
				frontier = append(frontier, target)
			}
		}
		if node.NextNodeID != "" && !reachability[node.NextNodeID] {
			reachability[node.NextNodeID] = true
			frontier = append(frontier, node.NextNodeID)
		}
	}

	// Report unreachable nodes (warnings — a flow can still run with dead code).
	for _, n := range def.Nodes {
		if !reachability[n.ID] && n.ID != def.StartNodeID {
			issues = append(issues, Issue{
				Severity: "warning", NodeID: n.ID, Code: "unreachable",
				Message: "node is not reachable from start_node_id",
			})
		}
	}

	return issues, reachability
}

// validateNodeConfig mirrors the executor's per-node config expectations.
// Catches missing required fields at design time instead of run time.
func validateNodeConfig(n Node) []Issue {
	var issues []Issue
	cfg := n.Config
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	missing := func(key, code, msg string) {
		if _, ok := cfg[key]; !ok {
			issues = append(issues, Issue{
				Severity: "error", NodeID: n.ID, Code: code,
				Message: msg,
			})
		} else if s, ok := cfg[key].(string); ok && s == "" {
			issues = append(issues, Issue{
				Severity: "error", NodeID: n.ID, Code: code,
				Message: msg,
			})
		}
	}
	switch n.Type {
	case NodePlayAudio:
		missing("audio_path", "missing_config", "play_audio requires audio_path")
	case NodeSay:
		missing("text", "missing_config", "say requires text")
	case NodeMenu:
		// prompt_file is recommended but not strictly required (FS accepts
		// empty prompts that just collect digits).
		if _, ok := cfg["prompt_file"]; !ok {
			issues = append(issues, Issue{
				Severity: "warning", NodeID: n.ID, Code: "missing_config",
				Message: "menu has no prompt_file — caller hears silence before timeout",
			})
		}
		if len(n.Branches) == 0 {
			issues = append(issues, Issue{
				Severity: "warning", NodeID: n.ID, Code: "menu_no_branches",
				Message: "menu has no branches — DTMF input has nowhere to route",
			})
		}
	case NodeTransfer:
		missing("target", "missing_config", "transfer requires target")
	case NodeRecord:
		missing("recording_path", "missing_config", "record requires recording_path")
	case NodeGoto:
		missing("target_node_id", "missing_config", "goto requires target_node_id")
	case NodeHangup:
		// no config needed
	default:
		issues = append(issues, Issue{
			Severity: "error", NodeID: n.ID, Code: "unknown_type",
			Message: fmt.Sprintf("unknown node type %q", n.Type),
		})
	}
	return issues
}
