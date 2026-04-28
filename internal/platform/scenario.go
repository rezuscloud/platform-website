package platform

import (
	"strings"
	"time"
)

func applyCommand(state SessionState, request CommandRequest) CommandResponse {
	command := normalizeCommand(request.Command)
	if state.SessionID == "" {
		state = NewSessionState(request.SessionID)
	}

	if command == "" {
		return rejectCommand(state, "Enter one of the suggested commands to move the system.")
	}

	next := state
	now := time.Now().UTC().Format(time.RFC3339)

	switch command {
	case "rezus sync demo":
		next.Summary = SummaryState{
			Kicker:   "Cross-app proof",
			Headline: "One command moved through three services",
			Detail:   "linux-app accepted the invoke, locked the session with Redis, persisted topology in PostgreSQL V2 state, and published an event on NATS JetStream that the shell and Mac surfaces are now rendering.",
			Proofs: []Proof{
				{Label: "Invoke", Status: "delivered", Detail: "terminal-app called linux-app via service invocation"},
				{Label: "Pub/Sub", Status: "published", Detail: "linux-app emitted artifact.published on homepage.events through NATS JetStream"},
				{Label: "State", Status: "persisted", Detail: "shell-app and mac-app are reading the same PostgreSQL-backed session state"},
				{Label: "Lock", Status: "held", Detail: "linux-app serialized the session update through Redis lockstore"},
			},
		}
		next.Terminal.LastCommand = command
		next.Terminal.History = appendHistory(next.Terminal.History,
			"> "+command,
			"[invoke] terminal-app -> linux-app /internal/execute",
			"[lock] redis lockstore claimed "+SessionStateKey(request.SessionID),
			"[state] PostgreSQL v2 persisted topology under "+SessionStateKey(request.SessionID),
			"[pubsub] JetStream homepage.events <- artifact.published",
			"[shell] proof rail is now live and derived",
		)
		next.Mac.Artifact = Artifact{
			Title: "Deployment dossier",
			Lines: []string{
				"session: " + request.SessionID,
				"cluster: north-workshop",
				"artifact: topology snapshot published by linux-app via JetStream",
				"reader: mac-app is showing PostgreSQL-backed shared state, not local copy",
			},
			UpdatedAt: now,
		}
		next.Mac.Notes = []string{
			"Mac now has a real artifact to inspect.",
			"The shell headline and Linux service graph came from the same PostgreSQL write guarded by Redis locking.",
		}
		next.Linux = LinuxState{
			Mode:       "reconciled",
			LastAction: "Accepted rezus sync demo, locked the session in Redis, wrote PostgreSQL state, and published a deployment dossier over JetStream.",
			Services: []ServiceStatus{
				{Name: "shell-app", Status: "reading", Detail: "summary refreshed from PostgreSQL-backed shared state"},
				{Name: "terminal-app", Status: "invoked", Detail: "command dispatched successfully"},
				{Name: "mac-app", Status: "rendered", Detail: "artifact drawer opened against shared PostgreSQL state"},
				{Name: "linux-app", Status: "healthy", Detail: "Redis lock acquired, topology written, and artifact published"},
			},
		}
		return acceptCommand(next, SessionEvent{
			SessionID: request.SessionID,
			Type:      "artifact.published",
			Source:    LinuxAppID,
			Message:   "linux-app published a topology dossier for the shell and Mac surfaces.",
			Details:   []string{"service invocation succeeded", "shared state now contains the artifact", "homepage.events fanout completed"},
			Timestamp: now,
		})
	case "rezus fanout edge":
		next.Summary = SummaryState{
			Kicker:   "Pub/sub fanout",
			Headline: "Linux shifted edge state and broadcast the change",
			Detail:   "This flow keeps the shell focused on proof, while linux-app owns the machine state in PostgreSQL, serializes updates with Redis, and fans out an event over JetStream for observers.",
			Proofs: []Proof{
				{Label: "Invoke", Status: "delivered", Detail: "terminal-app requested an edge change"},
				{Label: "Pub/Sub", Status: "fanout", Detail: "linux-app broadcast edge.shifted over NATS JetStream"},
				{Label: "State", Status: "updated", Detail: "service health and artifact metadata were rewritten in PostgreSQL"},
				{Label: "Lock", Status: "held", Detail: "Redis lockstore serialized the edge update"},
			},
		}
		next.Terminal.LastCommand = command
		next.Terminal.History = appendHistory(next.Terminal.History,
			"> "+command,
			"[lock] redis lockstore serialized edge state",
			"[invoke] linux-app accepted edge shift",
			"[state] PostgreSQL promoted the edge gateway to active",
			"[pubsub] JetStream homepage.events <- edge.shifted",
		)
		next.Mac.Artifact = Artifact{
			Title: "Edge failover memo",
			Lines: []string{
				"primary edge: online",
				"secondary edge: promoted and publishing health",
				"artifact source: linux-app",
				"observers: shell-app, mac-app",
			},
			UpdatedAt: now,
		}
		next.Mac.Notes = []string{
			"Mac is now showing the failover memo from PostgreSQL-backed shared state.",
			"Nothing in the shell hard-coded this transition.",
		}
		next.Linux = LinuxState{
			Mode:       "edge-active",
			LastAction: "Locked the session in Redis and published edge.shifted after promoting the standby path in PostgreSQL.",
			Services: []ServiceStatus{
				{Name: "shell-app", Status: "watching", Detail: "proof rail is reflecting the new event"},
				{Name: "terminal-app", Status: "invoked", Detail: "edge shift request delivered"},
				{Name: "mac-app", Status: "updated", Detail: "memo switched to failover state"},
				{Name: "linux-app", Status: "routing", Detail: "secondary edge promoted and healthy under Redis-backed locking"},
			},
		}
		return acceptCommand(next, SessionEvent{
			SessionID: request.SessionID,
			Type:      "edge.shifted",
			Source:    LinuxAppID,
			Message:   "linux-app shifted the edge path and broadcast the new topology.",
			Details:   []string{"standby edge promoted", "artifact rewritten", "shell proof rail refreshed"},
			Timestamp: now,
		})
	case "rezus inspect dossier":
		next.Summary = SummaryState{
			Kicker:   "Inspection path",
			Headline: "Mac opened the latest dossier without copying the machine state",
			Detail:   "This flow leans on the same PostgreSQL-backed shared state, which keeps inspection separate from execution but still grounded in a real session protected by Redis locking and JetStream eventing.",
			Proofs: []Proof{
				{Label: "Invoke", Status: "reused", Detail: "inspection used the current session context"},
				{Label: "Pub/Sub", Status: "published", Detail: "linux-app emitted artifact.inspected over NATS JetStream"},
				{Label: "State", Status: "read", Detail: "mac-app rendered the latest PostgreSQL-backed dossier without forking it"},
				{Label: "Lock", Status: "held", Detail: "Redis lockstore guarded the inspection update"},
			},
		}
		next.Terminal.LastCommand = command
		next.Terminal.History = appendHistory(next.Terminal.History,
			"> "+command,
			"[lock] redis lockstore guarded dossier inspection",
			"[invoke] linux-app prepared the inspection payload",
			"[state] mac-app reading the current PostgreSQL dossier",
			"[pubsub] JetStream homepage.events <- artifact.inspected",
		)
		next.Mac.Artifact = Artifact{
			Title: "Inspection transcript",
			Lines: []string{
				"artifact lineage: terminal -> linux -> PostgreSQL state -> mac",
				"session health: stable",
				"intent: prove ownership through readable system state",
				"next: open /apps/linux for the full service surface",
			},
			UpdatedAt: now,
		}
		next.Mac.Notes = []string{
			"Inspection is intentionally separate from execution.",
			"The artifact stayed in PostgreSQL-backed shared state while the Mac surface opened it.",
		}
		next.Linux.LastAction = "Prepared the latest dossier for inspection, guarded the session with Redis locking, and published artifact.inspected over JetStream."
		return acceptCommand(next, SessionEvent{
			SessionID: request.SessionID,
			Type:      "artifact.inspected",
			Source:    LinuxAppID,
			Message:   "linux-app prepared the latest dossier for the Mac inspection surface.",
			Details:   []string{"mac-app opened the current state", "shell headline reflects the inspection path"},
			Timestamp: now,
		})
	default:
		next.Terminal.LastCommand = command
		next.Terminal.History = appendHistory(next.Terminal.History,
			"> "+command,
			"[error] unknown command",
			"Try one of: rezus sync demo, rezus fanout edge, rezus inspect dossier",
		)
		return rejectCommand(next, "Unknown command. Use one of the suggested flows.")
	}
}

func acceptCommand(state SessionState, event SessionEvent) CommandResponse {
	if event.SessionID == "" {
		event.SessionID = state.SessionID
	}
	state.Events = appendEvent(state.Events, event)
	state.UpdatedAt = event.Timestamp

	return CommandResponse{
		Accepted: true,
		Message:  event.Message,
		Event:    event,
		State:    state,
	}
}

func rejectCommand(state SessionState, message string) CommandResponse {
	now := time.Now().UTC().Format(time.RFC3339)
	event := SessionEvent{
		SessionID: state.SessionID,
		Type:      "command.rejected",
		Source:    LinuxAppID,
		Message:   message,
		Details:   []string{"The current slice only supports the scripted demo commands."},
		Timestamp: now,
	}
	state.Events = appendEvent(state.Events, event)
	state.UpdatedAt = now

	return CommandResponse{
		Accepted: false,
		Message:  message,
		Event:    event,
		State:    state,
	}
}

func observeShellEvent(state SessionState, event SessionEvent) SessionState {
	if event.SessionID == "" || event.SessionID != state.SessionID {
		return state
	}

	for i, proof := range state.Summary.Proofs {
		if proof.Label == "Pub/Sub" {
			state.Summary.Proofs[i].Status = "observed"
			state.Summary.Proofs[i].Detail = "shell-app observed " + event.Type + " on " + SessionEventsTopic + " through NATS JetStream"
			continue
		}

		if proof.Label == "Lock" {
			state.Summary.Proofs[i].Status = "released"
			state.Summary.Proofs[i].Detail = "Redis lockstore serialized the update and released the session lock"
			continue
		}
	}

	state.UpdatedAt = event.Timestamp
	return state
}

func appendHistory(history []string, lines ...string) []string {
	history = append(history, lines...)
	if len(history) > MaxTerminalHistory {
		return append([]string(nil), history[len(history)-MaxTerminalHistory:]...)
	}

	return history
}

func appendEvent(events []SessionEvent, event SessionEvent) []SessionEvent {
	events = append([]SessionEvent{event}, events...)
	if len(events) > MaxSessionEventCount {
		return append([]SessionEvent(nil), events[:MaxSessionEventCount]...)
	}

	return events
}

func normalizeCommand(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(strings.ToLower(value))), " ")
}
