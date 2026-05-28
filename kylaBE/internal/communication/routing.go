package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// Router evaluates routing rules and assigns conversations.
type Router struct {
	store     *RoutingStore
	convStore *ConversationStore
	agentLB   *AgentLoadBalancer
}

// NewRouter constructs a Router.
func NewRouter(
	store *RoutingStore,
	convStore *ConversationStore,
	agentLB *AgentLoadBalancer,
) *Router {
	return &Router{
		store:     store,
		convStore: convStore,
		agentLB:   agentLB,
	}
}

// Route evaluates rules and assigns a conversation.
func (r *Router) Route(ctx context.Context, conv *Conversation) error {
	rules, err := r.store.FindActiveRules(conv.OrgID, conv.WorkspaceID)
	if err != nil {
		return fmt.Errorf("load routing rules: %w", err)
	}

	for _, rule := range rules {
		if r.evaluate(rule, conv) {
			if err := r.execute(ctx, rule, conv); err != nil {
				log.Printf("[routing] execute rule=%s conv=%s error: %v", rule.ID, conv.ID, err)
				continue
			}
			log.Printf("[routing] applied rule=%s (%s) to conv=%s", rule.ID, rule.Name, conv.ID)
			return nil
		}
	}

	log.Printf("[routing] no matching rule for conv=%s", conv.ID)
	return nil
}

func (r *Router) evaluate(rule *RoutingRule, conv *Conversation) bool {
	var conditions []struct {
		Field    string      `json:"field"`
		Operator string      `json:"op"`
		Value    interface{} `json:"value"`
	}

	if err := json.Unmarshal(rule.Conditions, &conditions); err != nil {
		log.Printf("[routing] unmarshal conditions rule=%s: %v", rule.ID, err)
		return false
	}

	for _, cond := range conditions {
		if !r.evaluateCondition(cond.Field, cond.Operator, cond.Value, conv) {
			return false
		}
	}

	return true
}

func (r *Router) evaluateCondition(field, op string, value interface{}, conv *Conversation) bool {
	var actual string

	switch field {
	case "channel":
		actual = conv.Channel
	case "priority":
		actual = conv.Priority
	case "status":
		actual = conv.Status
	default:
		log.Printf("[routing] unknown condition field: %s", field)
		return false
	}

	expected, ok := value.(string)
	if !ok {
		return false
	}

	switch op {
	case "eq", "equals":
		return actual == expected
	case "ne", "not_equals":
		return actual != expected
	default:
		log.Printf("[routing] unknown operator: %s", op)
		return false
	}
}

func (r *Router) execute(ctx context.Context, rule *RoutingRule, conv *Conversation) error {
	var actions []struct {
		Type     string `json:"type"`
		TargetID string `json:"target_id"`
	}

	if err := json.Unmarshal(rule.Actions, &actions); err != nil {
		return fmt.Errorf("unmarshal actions: %w", err)
	}

	for _, action := range actions {
		switch action.Type {
		case "assign_team":
			agentID, err := r.agentLB.PickAgent(ctx, action.TargetID, rule.Strategy)
			if err != nil {
				log.Printf("[routing] pick agent team=%s: %v", action.TargetID, err)
				continue
			}
			if _, err := r.convStore.AssignTo(conv.ID, conv.OrgID, agentID, action.TargetID); err != nil {
				return fmt.Errorf("assign to team: %w", err)
			}

		case "assign_agent":
			if _, err := r.convStore.AssignTo(conv.ID, conv.OrgID, action.TargetID, ""); err != nil {
				return fmt.Errorf("assign to agent: %w", err)
			}

		case "set_priority":
			if _, err := r.convStore.SetPriority(conv.ID, conv.OrgID, action.TargetID); err != nil {
				return fmt.Errorf("set priority: %w", err)
			}

		default:
			log.Printf("[routing] unknown action type: %s", action.Type)
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// AgentLoadBalancer
// ─────────────────────────────────────────────────────────────────────────────

// AgentLoadBalancer picks agents using round-robin or skill-based strategies.
type AgentLoadBalancer struct {
	redis *redis.Client
}

// NewAgentLoadBalancer constructs an AgentLoadBalancer.
func NewAgentLoadBalancer(redis *redis.Client) *AgentLoadBalancer {
	return &AgentLoadBalancer{redis: redis}
}

// PickAgent selects an agent from the team using the specified strategy.
func (lb *AgentLoadBalancer) PickAgent(ctx context.Context, teamID, strategy string) (string, error) {
	// TODO: Query agentops.AgentStatusStore for online agents in the team
	// For now, return a placeholder
	onlineAgents := []string{"agent-1", "agent-2", "agent-3"}

	if len(onlineAgents) == 0 {
		return "", fmt.Errorf("no online agents in team %s", teamID)
	}

	switch strategy {
	case "round_robin":
		return lb.roundRobin(ctx, teamID, onlineAgents)
	case "skill_based":
		// TODO: Filter by skills, pick agent with least active conversations
		return onlineAgents[0], nil
	case "direct":
		return onlineAgents[0], nil
	default:
		return lb.roundRobin(ctx, teamID, onlineAgents)
	}
}

func (lb *AgentLoadBalancer) roundRobin(ctx context.Context, teamID string, agents []string) (string, error) {
	key := fmt.Sprintf("kyla:lb:%s:idx", teamID)
	idx, err := lb.redis.Incr(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("redis incr: %w", err)
	}
	return agents[int(idx-1)%len(agents)], nil
}
