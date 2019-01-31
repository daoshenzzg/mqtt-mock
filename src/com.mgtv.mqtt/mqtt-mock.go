package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var Debug bool = false
var Choke chan [2]string = make(chan [2]string)

type ExecOptions struct {
	Broker      string // Broker URI
	Qos         byte   // QoS(0|1|2)
	Topic       string // Topic
	Username    string // 用户名
	Password    string // 密码
	ClientNum   int    // 启动客户端数量
	Count       int    // 消息总数(推送/接收）
	MessageSize int    // 1条消息大小(byte)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-broker <broker>] [-action <action>] [-topic <topic>] [-c <client number, default 20>] [-n <message number, default 100>] [-size <message size, default 1024>] [-username <username>] [-password <password>] [-qos <qos, default 0>] [-debug <debug, default false>]\n", os.Args[0])
		flag.PrintDefaults()
	}

	broker := flag.String("broker", "", "URI(tcp://{ip}:{port}) of MQTT broker (required)")
	action := flag.String("action", "pub or sub", "Publish or Subscribe or Subscribe(with publishing) (required)")
	topic := flag.String("topic", "mqtt-mock/benchmark/", "Base topic")
	clientNum := flag.Int("c", 20, "Number of clients")
	count := flag.Int("n", 100, "Number of message to Publish or receive")
	size := flag.Int("size", 1024, "Message size per publish (byte)")
	username := flag.String("username", "admin", "Username for connecting to the MQTT broker")
	password := flag.String("password", "123456", "Password for connecting to the MQTT broker")
	qos := flag.Int("qos", 0, "MQTT QoS(0|1|2)")
	debug := flag.Bool("debug", false, "Debug mode")

	flag.Parse()

	if len(os.Args) <= 1 {
		flag.Usage()
		return
	}

	if *broker == "" {
		fmt.Printf("Invalid argument: -broker -> %s\n", *broker)
		return
	}

	if *action != "pub" && *action != "sub" {
		fmt.Printf("Invalid argument: -action -> %s\n", *action)
		return
	}

	execOpts := ExecOptions{}
	execOpts.Broker = *broker
	execOpts.Qos = byte(*qos)
	execOpts.Topic = *topic
	execOpts.Username = *username
	execOpts.Password = *password
	execOpts.ClientNum = *clientNum
	execOpts.Count = *count
	execOpts.MessageSize = *size
	Debug = *debug

	fmt.Printf("Mock Info:\n")
	fmt.Printf("\tbroker:       %s\n", execOpts.Broker)
	fmt.Printf("\tc:            %d\n", execOpts.ClientNum)
	fmt.Printf("\tn:            %d\n", execOpts.Count)
	if *action == "pub" {
		fmt.Printf("\tsize:         %d\n", execOpts.MessageSize)
	}
	fmt.Printf("\tusername:     %s\n", execOpts.Username)
	fmt.Printf("\tpassword:     %s\n", execOpts.Password)
	fmt.Printf("\ttopic:        %s\n", execOpts.Topic)
	fmt.Printf("\tqos:          %d\n", execOpts.Qos)
	fmt.Printf("\tdebug:        %v\n", Debug)

	clients := make([]MQTT.Client, execOpts.ClientNum)
	for i := 0; i < execOpts.ClientNum; i++ {
		// Prepare client options
		clientId := GenClientId(i + 1)
		opts := MQTT.NewClientOptions()
		opts.AddBroker(execOpts.Broker)
		opts.SetClientID(clientId)
		opts.SetUsername(execOpts.Username)
		opts.SetPassword(execOpts.Password)
		opts.SetCleanSession(true)
		opts.SetKeepAlive(90 * time.Second)
		opts.SetAutoReconnect(true)
		opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
			Choke <- [2]string{msg.Topic(), string(msg.Payload())}
		})

		// Build MQTT Client
		client := MQTT.NewClient(opts)

		// Connect to MQTT Server
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Println(token.Error())
			os.Exit(1)
		}

		clients[i] = client
		if Debug {
			log.Printf("Connected: clientId=%s\n", clientId)
		}
	}

	switch *action {
	case "pub":
		DoPublish(clients, execOpts)
	case "sub":
		DoSubscribe(clients, execOpts)
	}
}

func DoPublish(clients []MQTT.Client, opts ExecOptions) {
	message := CreateFixedSizeMessage(opts.MessageSize)

	// 控制下并发数
	ch := make(chan int, opts.ClientNum)
	wg := new(sync.WaitGroup)

	start := time.Now().Unix()
	var count int64

	for i := 0; i < opts.Count; i++ {
		wg.Add(1)

		clientId := i % len(clients)
		client := clients[clientId]

		ch <- i
		go func(chan int) {
			defer wg.Done()

			topic := fmt.Sprintf(opts.Topic+"%d", clientId)
			token := client.Publish(topic, opts.Qos, false, message)
			token.Wait()

			if Debug {
				log.Printf("Published message: topic=%s, message=%s\n", topic, message)
			}

			if atomic.AddInt64(&count, 1)%100000 == 0 {
				log.Printf("%d messages has been published.\n", count)
			}

			<-ch
		}(ch)

	}

	wg.Wait()

	end := time.Now().Unix()

	totalCount := opts.Count
	throughput := float64(totalCount)
	cost := float64(end - start)
	if cost > 0 {
		throughput = float64(totalCount) / cost
	}
	log.Printf("Finish publish mock! total=%d cost=%.0fs Throughput=%.2f(messages/sec)\n", totalCount, cost, throughput)

	for _, client := range clients {
		Disconnect(client)
	}
}

func DoSubscribe(clients []MQTT.Client, opts ExecOptions) {
	for id := 0; id < len(clients); id++ {
		client := clients[id]

		// Subscribe to topic
		if token := client.Subscribe(opts.Topic, opts.Qos, nil); token.Wait() && token.Error() != nil {
			log.Println(token.Error())
			os.Exit(1)
		}
	}

	totalCount := 0
	start := time.Now().Unix()
	t1 := time.Now().Unix()
	count := 0
	for {
		data := <-Choke
		if Debug {
			log.Printf("Received message(%d): topic=%s, message=%s\n", count, data[0], data[1])
		}

		totalCount++
		if totalCount >= opts.Count {
			break
		}

		// 从接收到的第一条消息开始计算时间
		if totalCount == 1 {
			start = time.Now().Unix()
			t1 = time.Now().Unix()
		}

		var duration int64 = 3
		t2 := time.Now().Unix()
		if t2-t1 <= duration {
			count++
		} else {
			throughput := float64(count) / float64(t2-t1)
			log.Printf("Throughput=%.2f(messages/sec)\n", throughput)

			// reset
			t1 = t2
			count = 0
		}

	}
	end := time.Now().Unix()

	throughput := float64(totalCount)
	cost := float64(end - start)
	if cost > 0 {
		throughput = float64(totalCount) / cost
	}
	log.Printf("Finish subscribe mock! total=%d cost=%.0fs Throughput=%.2f(messages/sec)\n", totalCount, cost, throughput)

	for _, client := range clients {
		Disconnect(client)
	}
}

func Disconnect(client MQTT.Client) {
	client.Disconnect(10)
}

// 多个进程中，如果ClientID重复的话，Broker方面会出现问题，
// 使用过程ID，分配ID。
// mqtbench<进程ID的16进制数值>-<客户端的编号>
func GenClientId(id int) string {
	pid := strconv.FormatInt(int64(os.Getpid()), 16)
	clientId := fmt.Sprintf("mqttbench-%s-%d", pid, id)
	return clientId
}

// 生成固定大小的信息。
func CreateFixedSizeMessage(size int) string {
	var buffer bytes.Buffer
	for i := 0; i < size; i++ {
		buffer.WriteString(strconv.Itoa(i % 10))
	}

	message := buffer.String()
	return message
}
