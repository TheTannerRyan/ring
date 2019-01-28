// Copyright (c) 2019 Tanner Ryan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ring

import (
	"errors"
	"math"
	"sync"
)

// Ring contains the information for a ring data store.
type Ring struct {
	size  uint64        // size of bit (bit array is size/8+1)
	bits  []uint8       // main bit array
	hash  uint64        // number of hash rounds
	mutex *sync.RWMutex // mutex for locking Add, Test, and Reset operations
}

// Init initializes and returns a new ring, or an error. Given a number of
// elements, it accurately states if data is not added. Within a falsePositive
// rate, it will indicate if the data has been added.
func Init(elements int, falsePositive float64) (*Ring, error) {
	if elements < 0 {
		return nil, errors.New("elements must be greater than zero")
	}
	if falsePositive >= 1 || falsePositive < 0 {
		return nil, errors.New("falsePositive should be between 0 and 1")
	}
	r := Ring{}
	// length of filter
	m := (-1 * float64(elements) * math.Log(falsePositive)) / math.Pow(math.Log(2), 2)
	// number of hash rounds
	k := (m / float64(elements)) * math.Log(2)

	// check parameters
	if m <= 0 || k <= 0 {
		return nil, errors.New("invalid parameters")
	}

	// ring parameters
	r.mutex = &sync.RWMutex{}
	r.size = uint64(math.Ceil(m))
	r.hash = uint64(math.Ceil(k))
	r.bits = make([]uint8, r.size/8+1)
	return &r, nil
}

// Add adds the data to the ring.
func (r *Ring) Add(data []byte) {
	// generate hashes
	hash := generateMultiHash(data)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i := uint64(0); i < r.hash; i++ {
		index := getRound(hash, i) % r.size
		// set index%8-th bit to active
		r.bits[index/8] |= (1 << (index % 8))
	}
}

// Reset clears the ring.
func (r *Ring) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// reset bits
	r.bits = make([]uint8, r.size/8+1)
}

// Test returns a bool if the data is in the ring. True indicates that the data
// may be in the ring, while false indicates that the data is not in the ring.
func (r *Ring) Test(data []byte) bool {
	// generate hashes
	hash := generateMultiHash(data)

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for i := uint64(0); i < uint64(r.hash); i++ {
		index := getRound(hash, i) % r.size
		// check if index%8-th bit is not active
		if (r.bits[index/8] & (1 << (index % 8))) == 0 {
			return false
		}
	}
	return true
}
