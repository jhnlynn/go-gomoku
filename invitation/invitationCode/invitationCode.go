package invitationCode

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

const maxSeed = 9999

var (
	ids   = [maxSeed + 1]int{}
	end   = maxSeed
	mutex sync.Mutex
)

func initCode() {
	for i := range ids {
		ids[i] = i
	}
	rand.Seed(time.Now().UnixNano())
}

/*
	Get
	returns a unique 4-digit invitation code,
	for the opposite to play with you in another command CLIs
*/
func Get() (int, error) {
	mutex.Lock()

	defer mutex.Unlock()

	if end < 0 {
		return 0, errors.New("no available Id at the time")
	}

	i := rand.Intn(end + 1)
	code := ids[i]
	ids[i], ids[end] = ids[end], ids[i]
	end -= 1
	return code, nil
}

/*
	Return
	gives back the invitation code for future use
*/
func Return(code int) {
	mutex.Lock()
	defer mutex.Unlock()

	end += 1
	ids[end] = code
}
