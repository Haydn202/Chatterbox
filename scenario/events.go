package scenario

import (
	"fmt"
	"strings"

	"github.com/Haydn202/Chatterbox/fuzz"
)

// patchState holds runtime overrides for one service during an event phase.
type patchState struct {
	rateMultiplier      float64
	levelWeights        map[string]float64
	messageFuzzer       fuzz.Fuzzer
	useSharedCorrelation bool
}

func defaultPatch() patchState {
	return patchState{rateMultiplier: 1}
}

// Event applies and removes scenario patches for a named incident.
type Event interface {
	ID() string
	// DefaultTargets returns preferred service names when a phase omits services (nil = all).
	DefaultTargets() []string
	Apply(services []string, states map[string]*serviceRuntime)
	Remove(services []string, states map[string]*serviceRuntime)
}

var eventRegistry = map[string]Event{
	"baseline":                  baselineEvent{},
	"redis_latency_spike":       redisLatencySpikeEvent{},
	"retry_storm":               retryStormEvent{},
	"pod_restarts":              podRestartsEvent{},
	"db_connection_exhaustion":  dbConnectionExhaustionEvent{},
}

func lookupEvent(id string) (Event, error) {
	id = strings.ToLower(strings.TrimSpace(id))
	ev, ok := eventRegistry[id]
	if !ok {
		return nil, fmt.Errorf("scenario: unknown event %q", id)
	}
	return ev, nil
}

type baselineEvent struct{}

func (baselineEvent) ID() string { return "baseline" }

func (baselineEvent) DefaultTargets() []string { return nil }

func (baselineEvent) Apply(services []string, states map[string]*serviceRuntime) {
	for _, name := range services {
		if st, ok := states[name]; ok {
			st.patch = defaultPatch()
		}
	}
}

func (baselineEvent) Remove([]string, map[string]*serviceRuntime) {}

type redisLatencySpikeEvent struct{}

func (redisLatencySpikeEvent) ID() string { return "redis_latency_spike" }

func (redisLatencySpikeEvent) DefaultTargets() []string {
	return []string{"redis", "api"}
}

func (redisLatencySpikeEvent) Apply(services []string, states map[string]*serviceRuntime) {
	msgs := []any{
		"redis command slowlog threshold exceeded",
		"GET session:key latency 2400ms",
		"upstream redis timeout after 5s",
		"cache miss storm on hot key",
	}
	for _, name := range services {
		st, ok := states[name]
		if !ok {
			continue
		}
		p := defaultPatch()
		p.rateMultiplier = 1.5
		switch name {
		case "redis":
			p.levelWeights = map[string]float64{"info": 0.1, "warn": 0.5, "error": 0.4}
			p.messageFuzzer = fuzz.Choice(msgs...)
		default:
			p.levelWeights = map[string]float64{"info": 0.3, "warn": 0.4, "error": 0.3}
			p.messageFuzzer = fuzz.Choice(
				"dependency redis timeout",
				"session store unreachable",
				"retrying redis connection",
			)
		}
		st.patch = p
	}
}

func (redisLatencySpikeEvent) Remove(services []string, states map[string]*serviceRuntime) {
	baselineEvent{}.Apply(services, states)
}

type retryStormEvent struct{}

func (retryStormEvent) ID() string { return "retry_storm" }

func (retryStormEvent) DefaultTargets() []string {
	return []string{"api", "auth"}
}

func (retryStormEvent) Apply(services []string, states map[string]*serviceRuntime) {
	for _, name := range services {
		st, ok := states[name]
		if !ok {
			continue
		}
		p := defaultPatch()
		p.rateMultiplier = 2.5
		p.useSharedCorrelation = true
		p.levelWeights = map[string]float64{"info": 0.2, "warn": 0.5, "error": 0.3}
		p.messageFuzzer = fuzz.Choice(
			"retry attempt 3/5 backing off 200ms",
			"circuit half-open allowing probe",
			"upstream call failed will retry",
			"rate limiter queued request",
		)
		st.patch = p
	}
}

func (retryStormEvent) Remove(services []string, states map[string]*serviceRuntime) {
	baselineEvent{}.Apply(services, states)
}

type podRestartsEvent struct{}

func (podRestartsEvent) ID() string { return "pod_restarts" }

func (podRestartsEvent) DefaultTargets() []string {
	return []string{"api", "auth", "redis"}
}

func (podRestartsEvent) Apply(services []string, states map[string]*serviceRuntime) {
	for _, name := range services {
		st, ok := states[name]
		if !ok {
			continue
		}
		p := defaultPatch()
		p.rateMultiplier = 2.0
		p.levelWeights = map[string]float64{"info": 0.2, "warn": 0.3, "error": 0.5}
		p.messageFuzzer = fuzz.Choice(
			"CrashLoopBackOff pod restarting",
			"OOMKilled container exceeded memory limit",
			"liveness probe failed restarting container",
			"pod evicted due to resource pressure",
		)
		st.patch = p
	}
}

func (podRestartsEvent) Remove(services []string, states map[string]*serviceRuntime) {
	baselineEvent{}.Apply(services, states)
}

type dbConnectionExhaustionEvent struct{}

func (dbConnectionExhaustionEvent) ID() string { return "db_connection_exhaustion" }

func (dbConnectionExhaustionEvent) DefaultTargets() []string {
	return []string{"postgres", "api"}
}

func (dbConnectionExhaustionEvent) Apply(services []string, states map[string]*serviceRuntime) {
	for _, name := range services {
		st, ok := states[name]
		if !ok {
			continue
		}
		p := defaultPatch()
		p.rateMultiplier = 1.8
		switch name {
		case "postgres":
			p.levelWeights = map[string]float64{"info": 0.05, "warn": 0.25, "error": 0.7}
			p.messageFuzzer = fuzz.Choice(
				"FATAL: remaining connection slots reserved for superuser",
				"pool exhausted max_connections reached",
				"could not obtain connection from pool",
			)
		default:
			p.levelWeights = map[string]float64{"info": 0.2, "warn": 0.3, "error": 0.5}
			p.messageFuzzer = fuzz.Choice(
				"HTTP 503 database unavailable",
				"transaction rollback connection timeout",
				"readiness check failed db ping",
			)
		}
		st.patch = p
	}
}

func (dbConnectionExhaustionEvent) Remove(services []string, states map[string]*serviceRuntime) {
	baselineEvent{}.Apply(services, states)
}
