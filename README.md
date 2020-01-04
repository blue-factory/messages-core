# Messages API

Microservice responsible to send everything notifications such as: email, sms, iot, push and hooks. The service has a planner who, through the grpc protocol, can place, update or cancel the sending of notifications.

## Service Architecture

The solution consists of a central microservice that is connected to the queue and the database. On the other hand, there are several **Channel APIs** implementations that were registered to Messages API.

[![Architecture of
microservices](./docs/images/messages-architecture.png)](./docs/images/messages-architecture.png)

## Channels

Corresponds to the type of message that could be sent, for this there must be the implementation of that channel. There are currently several channels available to use, some are [Email](https://github.com/microapis/messages-email-api), [SMS](https://github.com/microapis/messages-sms-api), [IoT](https://github.com/microapis/messages-iot-api), [Push](https://github.com/microapis/messages-push-api) and [Hook](https://github.com/microapis/messages-hook-api).

The channel corresponds to an attribute of **message**, therefore, the name will be the unique key to identify the type of channel.

## Providers

The provider corresponds and attribute of **channel** and allows to identify what types of messages are available for a specific channel.

For example, for the [Email Channel](https://github.com/microapis/messages-email-api) there are the providers of [Sendgrid](https://sendgrid.com/), [Mandrill](https://mandrill.com/) and [AWS SES](https://aws.amazon.com/ses/). To know more, you must enter the repositories associated with the channels.

## gRPC Service

```go
message Message {
  string id = 1;
  string channel = 2;
  string provider = 3;
  string content = 4;
  string status = 5;
}

message Channel {
  string name = 1;
  repeated Provider providers = 2;
  string host = 3;
  string port = 4;
}

message Provider {
  string name = 1;
  map<string, google.protobuf.Any> params = 2;
}

service MessageService {
  rpc Put(MessagePutRequest) returns (MessagePutResponse) {}
  rpc Get(MessageGetRequest) returns (MessageGetResponse) {}
  rpc Update(MessageUpdateRequest) returns (MessageUpdateResponse) {}
  rpc Cancel(MessageCancelRequest) returns (MessageCancelResponse) {}
  rpc Register(MessageRegisterRequest) returns (MessageRegisterResponse) {}
}
```

## Client

If you are already using Messages-API we recommend you to use the client in go. [[Link]](https://github.com/microapis/clients-go)

## Commands (Development)

`make build`: build user service for osx.

`make linux`: build user service for linux os.

`make docker`: build docker.

`docker run -it -p 5050:5050 messages-api`: run docker.

`PORT=<port> REDIS_HOST=<redis-host> REDIS_PORT=<redis-port> ./bin/messages-api`: run messages service.

## TODO

- [ ] Task 1.
- [ ] Task 2.
- [ ] Task 3.
