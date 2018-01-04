1. terminal start nsqd
```
mkdir -p /tmp/nsqd; nsqd -data-path=/tmp/nsqd; rm -rf /tmp/nsqd
```

2. terminal run example
```
go run main.go
```

3. terminal watch the data in nsq
```
nsq_tail --nsqd-tcp-address=localhost:4150 --topic=hello_world
```
