package relay

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type Channel struct {
	ID            int64
	Name          string
	Provider      string
	BaseURL       string
	APIKey        string
	Models        []string
	Status        int
	Priority      int
	Weight        int
	CooldownUntil *time.Time
}

type Router struct {
	cooldown CooldownStore
}

func NewRouter(cooldown CooldownStore) *Router {
	return &Router{cooldown: cooldown}
}

func (r *Router) SelectChannel(channels []Channel, model string) (*Channel, error) {
	available := r.filterAvailable(channels, model)
	if len(available) == 0 {
		return nil, fmt.Errorf("no available channel for model %s", model)
	}

	grouped := r.groupByPriority(available)
	topGroup := grouped[0]

	selected := r.weightedRandom(topGroup)
	return selected, nil
}

func (r *Router) filterAvailable(channels []Channel, model string) []Channel {
	now := time.Now()
	var result []Channel

	for i := range channels {
		ch := &channels[i]

		if ch.Status != 1 {
			continue
		}

		if ch.CooldownUntil != nil && now.Before(*ch.CooldownUntil) {
			continue
		}

		if r.cooldown != nil && r.cooldown.IsCoolingDown(context.Background(), ch.ID) {
			continue
		}

		if !r.supportsModel(ch, model) {
			continue
		}

		result = append(result, *ch)
	}

	return result
}

func (r *Router) supportsModel(ch *Channel, model string) bool {
	for _, m := range ch.Models {
		if m == model || m == "*" {
			return true
		}
	}
	return false
}

func (r *Router) groupByPriority(channels []Channel) [][]Channel {
	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Priority > channels[j].Priority
	})

	var groups [][]Channel
	var current []Channel
	currentPriority := channels[0].Priority

	for _, ch := range channels {
		if ch.Priority != currentPriority {
			groups = append(groups, current)
			current = nil
			currentPriority = ch.Priority
		}
		current = append(current, ch)
	}
	if len(current) > 0 {
		groups = append(groups, current)
	}

	return groups
}

func (r *Router) weightedRandom(channels []Channel) *Channel {
	if len(channels) == 1 {
		return &channels[0]
	}

	totalWeight := 0
	for _, ch := range channels {
		totalWeight += ch.Weight
	}

	pick := rand.Intn(totalWeight)
	cumulative := 0
	for i := range channels {
		cumulative += channels[i].Weight
		if pick < cumulative {
			return &channels[i]
		}
	}

	return &channels[0]
}
