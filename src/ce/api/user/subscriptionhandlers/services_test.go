package subscriptionhandlers_test

import (
	"reflect"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/user/subscriptionhandlers"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
)

func TestServices(t *testing.T) {
	r := shttp.NewRouter()
	s := r.RegisterService(subscriptionhandlers.Services)

	if s == nil {
		t.Fatalf("Was expecting service not to be nil")
	}

	handlers := []string{
		"POST:/user/subscription/update",
	}

	if reflect.DeepEqual(s.Handlers(), handlers) == false {
		t.Fatalf("Handlers are not registered correctly")
	}
}
