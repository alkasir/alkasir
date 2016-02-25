package client

import (
	"testing"
	"time"
)

func TestSingleUseKeystore(t *testing.T) {
	{
		s := singleUseAuthKeyStore{
			entries: make(map[string]time.Time, 0),
			ttl:     time.Hour,
		}
		if s.Authenticate("notvalid") {
			t.Errorf("notvalid should not be a valid key")
		}

		key := s.New()

		if len(s.entries) != 1 {
			t.Errorf("expeted only one entry in the store")
		}
		s.Cleanup()

		if len(s.entries) != 1 {
			t.Errorf("key should not have expired")
		}

		if !s.Authenticate(key) {
			t.Errorf("key should have been validated")
		}

		if len(s.entries) != 0 {
			t.Errorf("key should have expired due to use")
		}
		key = s.New()
		s.New()
		s.New()

		if len(s.entries) != 3 {
			t.Errorf("there should be 3 keys in the store")
		}
		s.Cleanup()
		if len(s.entries) != 3 {
			t.Errorf("there should be 3 keys in the store")
		}

		s.Authenticate(key)
		if len(s.entries) != 2 {
			t.Errorf("there should be 2 keys in the store")
		}

	}

	{
		s := singleUseAuthKeyStore{
			entries: make(map[string]time.Time, 0),
			ttl:     time.Nanosecond,
		}
		if s.Authenticate("notvalid") {
			t.Errorf("notvalid should not be a valid key")
		}
		key := s.New()

		if len(s.entries) != 1 {
			t.Errorf("expeted only one entry in the store")
		}
		time.Sleep(time.Microsecond)
		s.Cleanup()

		if len(s.entries) != 0 {
			t.Errorf("key should have expired")
		}

		if s.Authenticate(key) {
			t.Errorf("key should have expired")
		}

		key = s.New()
		s.New()
		s.New()
		if s.Authenticate(key) {
			t.Errorf("key should have expired")
		}

		if len(s.entries) != 2 {
			t.Errorf("expected 2 items in store")
		}
		s.Cleanup()
		if len(s.entries) != 0 {
			t.Errorf("all keys should have expired")
		}

	}

}
