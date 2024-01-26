package scheduler

import (
	"container/heap"
	"github.com/LL-res/CRM/common/key"
)

type Item struct {
	value    key.WithModelKey // with model key,could be used to find the predictor
	priority float64          // The priority of the item in the queue.aka ,the loss of the predictor
	index    int              // The index of the item in the heap.
}

type PriorityQueue []*Item

// Key : no model key, Val : all predictors for that metric
type MetricToPredictors map[key.NoModelKey]*PriorityQueue

//var sortedPredictors MetricToPredictors

func NewMetricToPredictors() MetricToPredictors {
	res := make(map[key.NoModelKey]*PriorityQueue)
	return res
}
func (m *MetricToPredictors) GetHeap(nmk key.NoModelKey) *PriorityQueue {
	if _, ok := (*m)[nmk]; !ok {
		pq := make(PriorityQueue, 0)
		heap.Init(&pq)
		(*m)[nmk] = &pq
	}
	return (*m)[nmk]
}
func (pq *PriorityQueue) Len() int { return len(*pq) }

func (pq *PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return (*pq)[i].priority < (*pq)[j].priority
}

func (pq *PriorityQueue) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
	(*pq)[i].index = i
	(*pq)[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, value key.WithModelKey, priority float64) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}
func (pq *PriorityQueue) Peek() key.WithModelKey {
	n := len(*pq)
	if n == 0 {
		return key.WithModelKey{}
	}
	return (*pq)[0].value
}
func (pq *PriorityQueue) find(wmk key.WithModelKey) *Item {
	for _, item := range *pq {
		if item.value == wmk {
			return item
		}
	}
	return nil
}
