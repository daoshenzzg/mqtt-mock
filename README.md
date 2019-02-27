# MQTT-Mock: MQTT Benchmark Tool
这是一个使用[golang](https://golang.org)实现的mqtt客户端压测工具. 目前可支持Publish、Subscribe以及模拟指定数量客户端连接。
用于压测socket单broker能支撑多少连接数、下行消息、上行消息推送能力。

# 开始准备
Use go get and go build
```$xslt
go get github.com/daoshenzzg/mqtt-mock
cd ../mqtt-mock/src; go build mqtt-mock.go

```
# 使用说明
```$xslt
Usage: ./mqtt-mock [-broker <broker>] [-action <action>] [-topic <topic>] [-c <client number, default 20>] [-n <message number, default 100>] [-size <message size, default 1024>] [-username <username>] [-password <password>] [-qos <qos, default 0>] [-debug <debug, default false>]
  -action string
    	Publish or Subscribe or Subscribe(with publishing) (required) (default "pub or sub")
  -broker string
    	URI(tcp://{ip}:{port}) of MQTT broker (required)
  -c int
    	Number of clients (default 20)
  -debug
    	Debug mode
  -n int
    	Number of message to Publish or receive (default 100)
  -password string
    	Password for connecting to the MQTT broker (default "123456")
  -qos int
    	MQTT QoS(0|1|2)
  -size int
    	Message size per publish (byte) (default 1024)
  -topic string
    	Base topic (default "mqtt-mock/benchmark/")
  -username string
    	Username for connecting to the MQTT broker (default "admin")
```

# 模拟 Subscribe
```$xslt
./mqtt-mock -broker "tcp://127.0.0.1:8000" -c 2000 -n 500000 -action sub
Mock Info:
	broker:       tcp://127.0.0.1:8000
	c:            2000
	n:            500000
	username:     admin
	password:     123456
	topic:        mqtt-mock/benchmark/
	qos:          0
	debug:        false
2019/01/30 16:46:42 Throughput=16357.00(messages/sec)
2019/01/30 16:46:43 Throughput=39642.00(messages/sec)
2019/01/30 16:46:44 Throughput=35999.00(messages/sec)
...
2019/01/30 16:46:52 Throughput=38009.00(messages/sec)
2019/01/30 16:46:53 Throughput=36383.00(messages/sec)
2019/01/30 16:46:54 Throughput=37597.00(messages/sec)
2019/01/30 16:46:54 Finish subscribe mock! total=500000 cost=13s Throughput=38461.54(messages/sec)

```

# 模拟 Publish
```$xslt
./mqtt-mock -broker "tcp://127.0.0.1:8000" -c 200 -n 500000 -size 64 -action pub
Mock Info:
	broker:       tcp://127.0.0.1:8000
	c:            200
	n:            500000
	size:         1
	username:     admin
	password:     123456
	topic:        mqtt-mock/benchmark/
	qos:          0
	debug:        false
2019/01/31 11:10:56 100000 messages has been published.
2019/01/31 11:10:59 200000 messages has been published.
2019/01/31 11:11:02 300000 messages has been published.
2019/01/31 11:11:04 400000 messages has been published.
2019/01/31 11:11:07 500000 messages has been published.
2019/01/31 11:11:07 Finish publish mock! total=500000 cost=14s Throughput=35714.29(messages/sec)
```

# 模拟 Connection clients
```$xslt
./mqtt-mock -broker "tcp://127.0.0.1:8000" -c 10000 -n 1 -action sub -debug true
Mock Info:
	broker:       tcp://127.0.0.1:8000
	c:            10000
	n:            1
	username:     admin
	password:     123456
	topic:        mqtt-mock/benchmark/
	qos:          0
	debug:        true
2019/01/28 15:33:06 Connected : clientId= mqttmock-15ea9-1
2019/01/28 15:33:06 Connected : clientId= mqttmock-15ea9-2
...
2019/01/28 15:33:06 Connected : clientId= mqttmock-15ea9-9999
2019/01/28 15:33:06 Connected : clientId= mqttmock-15ea9-10000

```

# 压测Server
* https://github.com/daoshenzzg/socket-mqtt

# 参考项目
* https://github.com/mrkt/Ali_LMQ_SDK
* https://github.com/takanorig/mqtt-bench
