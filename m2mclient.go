/*
import "github.com/ausrasul/m2mclient"

cnf := &m2mClient.Config{
Ip: "127.0.0.1",
Port: "7000",
Ttl: 10,
Uid: "abcd",
}
c := m2mClient.New(cnf)
c.Run()

command := m2mclient.Cmd{
	Name: "modem_added",
	Param: m.Imei,
}
c.SendCmd(command)

*/
package m2mclient

import (
	"errors"
	"log"
	"net"
	"time"
)

type Config struct {
	Ip     string
	Port   string
	Ttl    int
	CmdTtl int
	Uid    string
}

type Client struct {
	initialized bool
	conf        *Config
	stopCh      chan bool
	stoppedCh   chan bool
	running bool
	handler map[string]func(*Client, string)
	msgQ chan Cmd
}



type Cmd struct{
	Name string
	Serial string
	SchType int // 0 immediet, 1 once, 2 periodical
	SchTime time.Time
	SchPeriod time.Duration
	Param string
}
func New(c *Config) *Client {
	return &Client{
		initialized: true,
		conf:        c,
		stopCh:      make(chan bool, 1),
		stoppedCh:   make(chan bool, 1),
		running: false,
		msgQ: make(chan Cmd, 1000),
		handler: make(map[string]func(*Client, string)),
	}
}

func (c *Client) AddHandler(cmdName string, handler func(*Client, string)){
	c.handler[cmdName] = handler
}
func (c *Client) HasHandler(cmdName string) bool{
	_, ok := c.handler[cmdName]
	return ok
}

func (c *Client) Run() (chan Cmd, chan Cmd, error) {
	in := make(chan Cmd)
	out := make(chan Cmd)
	go c.run(in, out, c.stopCh, c.stoppedCh)
	return in, out, nil
}

func (c *Client) Stop() error {
	if c.running {
		c.running = false
	} else {
		return errors.New("already stopped")
	}
	c.stopCh <- true
	<-c.stoppedCh
	return nil
}

func (c *Client) SendCmd(cmd Cmd) bool{
	c.msgQ <- cmd
	log.Print("sent msg")
	return true
}

func (c *Client) run(in <-chan Cmd, out chan<- Cmd, stop <-chan bool, stopped chan<- bool) {
	if !c.initialized {
		return
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	hbTimeout := time.Duration(c.conf.Ttl + 5)
	hbRate := time.Duration(c.conf.Ttl)
	for {
		time.Sleep(time.Second)
		log.Print("Starting client")
		conn, err := net.Dial("tcp", c.conf.Ip+":"+c.conf.Port)
		if err != nil {
			log.Print("Cannot connect to server: ", err)
			continue
		}
		if !authenticate(c.conf.Uid, conn) {
			log.Print("Error authenticating client")
			continue
		}
		outbox := make(chan Cmd, 1)
		stopSend := make(chan bool, 1)
		sendStopped := make(chan bool, 1)

		inbox := make(chan Cmd, 1)
		stopRcv := make(chan bool, 1)
		rcvStopped := make(chan bool, 1)
		defer close(outbox)
		defer close(stopSend)
		defer close(sendStopped)
		defer close(inbox)
		defer close(stopRcv)
		defer close(rcvStopped)
		//probe.Esn = c.conf.Uid

		go sender(conn, outbox, stopSend, sendStopped)
		go receiver(conn, inbox, stopRcv, rcvStopped)
		timer := time.NewTimer(time.Second * hbTimeout)
		hb := time.NewTimer(time.Second * hbRate)
		ok := true
		for ok {
			select {
			case <- stop:
				log.Print("Stopping client...")
				stopRcv <- true
				stopSend <- true
				<-rcvStopped
				<-sendStopped
				stopped <- true
				return
			case <-hb.C:
				log.Print("Heartbeat -->")
				hb = time.NewTimer(time.Second * hbRate)
				var cmd Cmd
				cmd.Name = "hb"
				outbox <- cmd
			case <-timer.C:
				log.Print("No heartbeat response from server.")
				stopRcv <- true
				stopSend <- true
				<-rcvStopped
				<-sendStopped
				select {
				case <-c.msgQ:
					default:
				}
				ok = false
				continue
				// handle timeout
			case cmd := <-inbox:
				log.Print("handle response")
				if cmd.Name == "hb_ack" {
					log.Print("HB_ACK received")
					timer = time.NewTimer(time.Second * hbTimeout)
				} else if cmd.Name == "del_task"{
					taskDel(cmd.Param)
				} else {
					callback, ok := c.handler[cmd.Name]
					if ok {
						taskDo(callback, c, cmd)
						//callback(c, cmd.Param)
					}
				}
			case msg := <- c.msgQ:
				log.Print("Sending command: ", msg.Name)
				outbox <- msg
			case <-sendStopped:
				log.Print("Sender stopped.")
				stopRcv <- true
				<-rcvStopped
				ok = false
				continue
			case <-rcvStopped:
				log.Print("Receiver stopped.")
				stopSend <- true
				<-sendStopped
				ok = false
				continue
			case <-stop:
				log.Print("client stopping.")
				stopSend <- true
				stopRcv <- true
				<-sendStopped
				<-rcvStopped
				ok = false
				continue
			}
		}
		if !ok {
			log.Print("All stopped, reconnecting in 10 seconds")
		}
		time.Sleep(10 * time.Second)
	}

}
