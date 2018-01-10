### motivation

In [minus5](https://minus5.hr) products we use nsq for last few years and transferring billions of messages each day. It serves us very well and become essential part of our infrastructure.
The other important component in our infrastructure is consul (for service discovery).
Together they give us nice decoupling between services which produce and services which consume messages. 
For consuming services all needed is connection to the consul client. There it will find lookupds and from lookupds location of producers (nsqd) of some topic.
Producing services always have collocated nsqd (on the same docker host), so they only publish to that local nsqd.
Nsqd-s get location of lookupds from consul and notifies them about topics thay have.

It all works fine for one direction streaming. But what for rpc where client needs to send request to some service and get response. We were using http rest with service discovery from consul. I was investigating grpc, rpcx and standard library rpc. They all have some nice features and I have some compliments about all of them. 
So we tried to use nsq transport for rpc.
Idea is simple; clients sends message to server topic and listens for response on his private topic. In request client adds topic on which to send response. 
Most interesting part of this project are examples so please check them.

### examples

Running examples requires some part of infrastructure. At least running nsqd. To have all required infrastructure started use script in example directory:
```
cd example
./start
```
This will run nsqd, nsqlookupd, nsqadmin and consul.
It will configure consul with nsqd and nsqlookupd locations. 
Consul is needed for examples which uses it for lookupd discovery. In other cases we need only nsqd.

Script will also start nsq_tail for all topics which are used in examples. So you can use that terminal window to examine messages which are exchanged through nsq.

# hello_world

This is basic example of sending and receiving message through nsq.
In real world server and client parts will probably be in different applications.
Here they are bot in one file:
```
go run hello_world/main.go
```

### simple_rpc

This is example of one idea of how to implement rpc over nsq. Nsq is mostly one direction message streaming platform. But with some coordination between server and client we could use it also for rpc.
In this example we need two topics. Client sends messages to the server topic (service.req). On each message we add envelope. In that envelope we have attributes: 
 * which server method are we calling
 * on which topic client is waiting for responses 
 * correlationID which connects request and response, response has same correlationID as request
 * expiration, optional
After server receives message it removes envelope. Checks expiration to see weather message is still valid. If it is expired stops processing. Valid messages are passed to the application. In application we pass method and the body of the message. Application should know what to do with that; how to decode body, and create response. Application returns response. Server rpc layer will create envelope with correlationID same as request correlationID and send it to the client topic.
Client unpacks envelope finds code which is waiting (reading from chan) for that correlationID and passes response to the code which started request.

If application code responds with error on the server side. Then the error is sent back to the client and to the application code which started request.

First start server:
```
go run simple_rpc/server.go
```

then run client:
```
go run simple_rpc/client.go
```

Watch the terminal where start script is running to get the feeling about messages which are exchanged between client and server. You should see something like this:

> {"m":"Add","r":"response","c":470217093,"x":1515589846}
> {"X":2,"Y":3}
< {"c":470217093}
< {"Z":5}

Lines with > are message parts send to the client. Each message consists of envelope, new line, and body. First two lines is a message to the server. First line is the envelope and second line is body. From envelope we can see that we are calling method Add, accepting response on the 'response' topic, c attribute is correlationID (you can see that it is same in request and response), and the last attribute is unix timestamp for message expiration (message is only valid until that time).
Lines with < prefix represents response parts (from server to client). First line is again envelope and second is body of the response. In the envelope we have only correlationID.


### rpc_with_code_generator

This example represents full idea of rpc over nsq.
After implementing something like previous example for few times I realized that there are lots of repeating code. So I tried to generate that repeating code. In this example all *\_gen.go files are actually generated:
  service/service\_gen.go
  service/api/api\_gen.go
  service/api/nsq/nsq\_gen.go

They are generated from code in service/service.go and service/api/dto.go
service.go is the definition of server side service. dto.go has definition of data transfer structures. They are in api package. The idea of api package is that it is shared definition between server and client. It don't depend on neither client nor server, only on few packages from standard library. 

gen.go is glue between application and code generator. That file is not part of the resulting application (build ignore directive). It is only used to configure code generator. In this example it starts code generator for _service.Service_ type.

So it all starts from the server side service definition, _Service_ type in service go.
Code generator examines that type and search for methods which has specific signature:
  * two method arguments
  * first argument is context
  * two response arguments, 
  * first response type is pointer; it could be nil in case of error
  * second response type is error
  * all request and response types should be declared in api/dto.go or be built in types
  * application specific errors should be declared in api/dto.go (overflow in this example)

On the top of the service.go there are few go:generate directives.
First will delete all *\_gen.go files recursively.
Second will install api package, that is needed for finding errors declared in that package later in code generator.
Third actual starts gen.go, which configures and calls code generator.
We can run code generator by:
```
cd rpc_with_code_generator/service
go generate
```

After we have generated code starting server is one-liner:
```
srv, err := nsq.Server(cfg, service.New())
```
and client:
```
cli, err := nsq.Client(cfg)
```
where nsq is package from service/api/nsq.

I'm calling it nsq because it is nsq implementation of the rpc. I could imagine any other transport protocol (http, tpc, gprc, kafka,...) with everything other staying same.


It is interesting to see that application errors are transferred from server to client. So on client side we could use typed errors (look at showError func in main.go):
```
if err == api.Overflow {
```


### tools 
If your are on the Mac this would be sufficient:
``` 
brew install consul nsq
```
  
### references
 * nsq rpc Erlang [ensq_rpc](https://github.com/project-fifo/ensq_rpc)
