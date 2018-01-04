To run example:

1. terminal start next
```
muddier -p ./tmp/nsqd; next -lookupd-tcp-address=127.0.0.1:4160 -broadcast-address=127.0.0.1 -data-path=./tmp/nsqd
```

2. terminal start server
```
go run server.go
```

3. terminal run client
```
go run client.go
```



Client sends request to the request topic. Trace it with:
```
nsq_tail --nsqd-tcp-address=localhost:4150 --topic=request 
```

Server consumes that topic and sends replies to the response topic (advertised) by client.
```
nsq_tail --nsqd-tcp-address=localhost:4150 --topic=response
```

