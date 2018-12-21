# smpl-msg
A simple AMQP abstraction to easily initialize consumers and/or publishers for a go app. This is simply a bootstrap project to use for MVPs or to extend, which is the focus of this implementation.

## Consumers and publishers
There are 3 possible consumers or publishers you can use,

```go
// Subscriber is a message topology which can only consume messages on a given route
type Subscriber interface {
  Consume(routingKey string) (<-chan amqp.Delivery, <-chan error)
  Close() error
}

// Publisher is a message topology which can only publish messages on a given route
type Publisher interface {
  Publish(routingKey string, headers amqp.Table, payload []byte) error
  Close() error
}

Okey what is
this or smth

// PublisherSubscriber is a message topology which can both publish and consume messages on a given route
type PublisherSubscriber interface {
  Consume(routingKey string) (<-chan amqp.Delivery, <-chan error)
  Publish(routingKey string, headers amqp.Table, payload []byte) error
  Close() error
}
```

## Usage
### Publishing

```go
c, err := smplmsg.NewPublisher("localhost:321", "exchangeName", "clientID", SetContentType("application/json"))
if err != nil {
  return err
}

// amqp.Table{} are the headers of the publisher message
err := c.Publish("routingKey", amqp.Table{}, []byte("Hello world"))
if err != nil {
  return err
}

```

### Consuming
```go
c, err := smplmsg.NewConsumer("localhost:321", "exchangeName", "clientID")
if err != nil {
  return err
}

msgCh, errCh := c.Consume("routingKey")
for {
    select {
    case err := <-errCh:
        return err
    case msg := <-msgCh:
        fmt.Printf("msg: %s", string(msg.Body))
    }
}
```

## Reconnection
There might be a hicup on the network and the AMQP library does not implement any automatic reconnection functionality. A simple implementation of a one-time retry to re-establish the connection is available.
