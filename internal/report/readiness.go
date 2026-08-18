package report

import (
	"fmt"
	"sort"
)

type ReadinessItem struct {
	Name     string `json:"name"`
	Ready    bool   `json:"ready"`
	Current  int    `json:"current"`
	Required int    `json:"required"`
	Detail   string `json:"detail"`
}

type Readiness struct {
	Ready   bool            `json:"ready"`
	Percent int             `json:"percent"`
	Items   []ReadinessItem `json:"items"`
}

func BuildReadiness(operations Operations) Readiness {
	items := []ReadinessItem{
		countItem("approved_campaign", Find(operations.Campaigns, "approved")+Find(operations.Campaigns, "in_progress"), 1),
		countItem("active_permit", Find(operations.Permits, "approved")+Find(operations.Permits, "activated"), 1),
		countItem("assigned_work", Find(operations.WorkOrders, "assigned")+Find(operations.WorkOrders, "started"), 1),
		countItem("reserved_dispatch", Find(operations.Dispatches, "reserved")+Find(operations.Dispatches, "departed"), 1),
		inverseItem("outbox_backlog", operations.PendingOutbox, 20),
		inverseItem("resource_locks", operations.OpenLocks, 50),
	}
	sort.Slice(items, func(left, right int) bool { return items[left].Name < items[right].Name })
	readyCount := 0
	for _, item := range items {
		if item.Ready {
			readyCount++
		}
	}
	percent := 0
	if len(items) > 0 {
		percent = readyCount * 100 / len(items)
	}
	return Readiness{Ready: readyCount == len(items), Percent: percent, Items: items}
}

func countItem(name string, current, required int) ReadinessItem {
	ready := current >= required
	detail := fmt.Sprintf("%d of %d required", current, required)
	return ReadinessItem{Name: name, Ready: ready, Current: current, Required: required, Detail: detail}
}

func inverseItem(name string, current, maximum int) ReadinessItem {
	ready := current <= maximum
	detail := fmt.Sprintf("%d observed, maximum %d", current, maximum)
	return ReadinessItem{Name: name, Ready: ready, Current: current, Required: maximum, Detail: detail}
}

func BlockingItems(readiness Readiness) []ReadinessItem {
	result := make([]ReadinessItem, 0)
	for _, item := range readiness.Items {
		if !item.Ready {
			result = append(result, item)
		}
	}
	return result
}
