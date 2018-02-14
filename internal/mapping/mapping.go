package mapping

import (
	"sync"

	"github.com/aws/aws-lambda-go/lambda/messages"
)

var mut = new(sync.Mutex)

type Map map[interface{}]*messages.InvokeRequest

func Set(i interface{}, r *messages.InvokeRequest) {
	mut.Lock()
	defer mut.Unlock()
	M[i] = r
}

func Get(i interface{}) *messages.InvokeRequest {
	mut.Lock()
	defer mut.Unlock()
	return M[i]
}

func del(req *messages.InvokeRequest) {
	mut.Lock()
	defer mut.Unlock()
	for k, v := range M {
		if v == req {
			delete(M, k)
			return
		}
	}
}

func Delete(req *messages.InvokeRequest) {
	go del(req)
}

var M Map
