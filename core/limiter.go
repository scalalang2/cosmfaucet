package core

import (
	"container/heap"
	"log"
	"time"
)

type LastRequest struct {
	ipAddress string
	time      timestamp
}

type timestamp int64
type AccountCache map[ChainId]map[string]bool
type RequestHeap []*LastRequest

func (h RequestHeap) Len() int {
	return len(h)
}

func (h RequestHeap) Less(i, j int) bool {
	return h[i].time < h[j].time
}

func (h RequestHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *RequestHeap) Push(x interface{}) {
	item := x.(*LastRequest)
	*h = append(*h, item)
}

func (h *RequestHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func (h RequestHeap) last() *LastRequest {
	return h[len(h)-1]
}

type Limiter struct {
	limitPeriod timestamp
	faucetHeap  map[ChainId]*RequestHeap
	accCache    AccountCache
}

func NewLimiter(chains []ChainId, limitPeriod int64) *Limiter {
	faucetHeap := make(map[ChainId]*RequestHeap)
	accCache := make(AccountCache)
	for _, chainId := range chains {
		accCache[chainId] = make(map[string]bool)
		reqHeap := make(RequestHeap, 0)
		heap.Init(&reqHeap)

		faucetHeap[chainId] = &reqHeap
	}

	return &Limiter{
		faucetHeap:  faucetHeap,
		accCache:    accCache,
		limitPeriod: timestamp(limitPeriod),
	}
}

// AddRequest add a request information to record the ipAddress of the user and when the request was made
func (l *Limiter) AddRequest(chainId ChainId, ipAddress string) {
	now := timestamp(time.Now().Unix())
	reqHeap, ok := l.faucetHeap[chainId]
	if !ok {
		log.Fatalf("given chainId is not registered at limiter: %s", chainId)
	}

	heap.Push(reqHeap, &LastRequest{
		ipAddress: ipAddress,
		time:      now,
	})

	if _, ok := l.accCache[chainId]; !ok {
		log.Fatalf("given chainId is not registered at limiter: %s", chainId)
	}

	l.accCache[chainId][ipAddress] = true
}

// IsAllowed checks if the ipAddress is allowed to request token
func (l *Limiter) IsAllowed(chainId ChainId, ipAddress string) bool {
	reqHeap := l.faucetHeap[chainId]
	if reqHeap.Len() == 0 {
		return true
	}

	if _, ok := l.accCache[chainId]; !ok {
		log.Fatalf("given chainId is not registered at limiter: %s", chainId)
	}

	l.removeOldCaches(reqHeap, chainId)

	_, ok := l.accCache[chainId][ipAddress]
	return !ok
}

// removeOldCaches removes a cache that is older than the `limitPeriod`
// It means that the appeared time of requests in the heap is over the `limitPeriod`,
// Therefore it must be allowed to serve the new faucet request.
func (l *Limiter) removeOldCaches(reqHeap *RequestHeap, chainId ChainId) {
	lastRequest := reqHeap.last()
	now := timestamp(time.Now().Unix())
	for reqHeap.Len() != 0 && now-lastRequest.time > l.limitPeriod {
		delete(l.accCache[chainId], lastRequest.ipAddress)
		lastRequest = reqHeap.Pop().(*LastRequest)
	}
}
