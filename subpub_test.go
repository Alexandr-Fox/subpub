package subpub

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestSubscribePublish(t *testing.T) {
	sp := NewSubPub()
	defer sp.Close(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	_, err := sp.Subscribe("test", func(msg interface{}) {
		defer wg.Done()
		if msg != "hello" {
			t.Errorf("Expected 'hello', got %v", msg)
		}
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	err = sp.Publish("test", "hello")
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	wg.Wait()
}

func TestMultipleSubscribers(t *testing.T) {
	sp := NewSubPub()
	defer sp.Close(context.Background())

	const numSubscribers = 5
	var wg sync.WaitGroup
	wg.Add(numSubscribers)

	for i := 0; i < numSubscribers; i++ {
		_, err := sp.Subscribe("multi", func(msg interface{}) {
			defer wg.Done()
			if msg != "broadcast" {
				t.Errorf("Expected 'broadcast', got %v", msg)
			}
		})
		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}
	}

	err := sp.Publish("multi", "broadcast")
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	wg.Wait()
}

func TestUnsubscribe(t *testing.T) {
	sp := NewSubPub()
	defer sp.Close(context.Background())

	called := false

	sub, err := sp.Subscribe("unsub", func(msg interface{}) {
		called = true
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	sub.Unsubscribe()

	err = sp.Publish("unsub", "should not be received")
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if called {
		t.Error("Handler was called after unsubscribe")
	}
}

func TestSlowSubscriber(t *testing.T) {
	sp := NewSubPub()
	defer sp.Close(context.Background())

	var wg sync.WaitGroup
	wg.Add(2)

	// Fast subscriber
	_, err := sp.Subscribe("slow", func(msg interface{}) {
		defer wg.Done()
		if msg != "message" {
			t.Errorf("Expected 'message', got %v", msg)
		}
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Slow subscriber
	_, err = sp.Subscribe("slow", func(msg interface{}) {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		if msg != "message" {
			t.Errorf("Expected 'message', got %v", msg)
		}
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	start := time.Now()
	err = sp.Publish("slow", "message")
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	wg.Wait()

	if time.Since(start) < 100*time.Millisecond {
		t.Error("Fast subscriber was blocked by slow one")
	}
}

func TestClose(t *testing.T) {
	sp := NewSubPub()

	// Add a subscriber that will be slow to process
	var wg sync.WaitGroup
	wg.Add(1)
	_, err := sp.Subscribe("close", func(msg interface{}) {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond)
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Publish a message that will take time to process
	err = sp.Publish("close", "data")
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Try to close with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = sp.Close(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected DeadlineExceeded, got %v", err)
	}

	// Wait for the handler to finish
	wg.Wait()

	// Now close properly
	err = sp.Close(context.Background())
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify we can't publish after close
	err = sp.Publish("close", "new data")
	if err.Error() != "SubPub closed" {
		t.Errorf("Expected closed error, got %v", err)
	}
}

func TestConcurrentPublishSubscribe(t *testing.T) {
	sp := NewSubPub()
	defer sp.Close(context.Background())

	const numMessages = 100
	const numSubscribers = 10

	var wg sync.WaitGroup
	wg.Add(numMessages * numSubscribers)

	// Start subscribers
	for i := 0; i < numSubscribers; i++ {
		_, err := sp.Subscribe("concurrent", func(msg interface{}) {
			defer wg.Done()
			if n, ok := msg.(int); !ok || n < 0 || n >= numMessages {
				t.Errorf("Unexpected message: %v", msg)
			}
		})
		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}
	}

	// Concurrently publish messages
	for i := 0; i < numMessages; i++ {
		go func(n int) {
			err := sp.Publish("concurrent", n)
			if err != nil {
				t.Errorf("Publish failed: %v", err)
			}
		}(i)
	}

	wg.Wait()
}

func TestSubscribeAfterClose(t *testing.T) {
	sp := NewSubPub()

	// Close immediately
	err := sp.Close(context.Background())
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Try to subscribe after close
	_, err = sp.Subscribe("closed", func(msg interface{}) {})
	if err.Error() != "SubPub closed" {
		t.Errorf("Expected closed error, got %v", err)
	}
}

func TestPublishNoSubscribers(t *testing.T) {
	sp := NewSubPub()
	defer sp.Close(context.Background())

	err := sp.Publish("nonexistent", "data")
	if err != nil {
		t.Errorf("Publish to nonexistent subject should not error, got %v", err)
	}
}
