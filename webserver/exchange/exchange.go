package exchange

import "sync"

type Message struct {
	Topic   string
	Content interface{}
	Sender  int
}

type MessageExchange interface {
	Publish(topic string, message interface{}, sender int) int
	Subscribe(topics []string) <-chan Message
	Shutdown()
}

type subscriber []chan Message

type messageExchange struct {
	mutex         sync.RWMutex
	subscriptions map[string]subscriber
}

// NewMessageExchange creates a new exchange for broadcasting messages
func NewMessageExchange() MessageExchange {
	return &messageExchange{
		subscriptions: make(map[string]subscriber),
	}
}

// Publish broadcasts a message about a specific topic and returns the number of recipients
func (exchange *messageExchange) Publish(topic string, message interface{}, sender int) int {
	var subs []chan Message

	exchange.mutex.RLock()
	defer exchange.mutex.RUnlock()

	subs, ok := exchange.subscriptions[topic]
	if !ok {
		return 0
	}

	for _, subscriber := range subs {
		subscriber <- Message{
			Topic:   topic,
			Content: message,
			Sender:  sender,
		}
	}
	return len(subs)
}

// Subscribe allows a subscriber to receive messages of specific topics
func (exchange *messageExchange) Subscribe(topics []string) <-chan Message {
	exchange.mutex.Lock()
	defer exchange.mutex.Unlock()

	channel := make(chan Message, 100)

	for _, topic := range topics {
		subscriber, ok := exchange.subscriptions[topic]
		if !ok {
			subscriber = make([]chan Message, 0, 1)
		}

		exchange.subscriptions[topic] = append(subscriber, channel)
	}
	return channel
}

// Shutdown closes all subscription channels and resets the message exchange
func (exchange *messageExchange) Shutdown() {
	exchange.mutex.Lock()
	defer exchange.mutex.Unlock()

	for _, subs := range exchange.subscriptions {
		for _, subscriber := range subs {
			close(subscriber)
		}
	}
	exchange.subscriptions = make(map[string]subscriber)
}
