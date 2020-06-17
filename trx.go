package m2mclient
import (
	"log"
	"net"
	"time"
	"bufio"
	"encoding/json"
)

func sender(conn net.Conn, outbox <-chan Cmd, stopSend <-chan bool, sendStopped chan<- bool){

	for {
		select{
		case <- stopSend:
			sendStopped <- true
			log.Print("sender stopping.")
			return
		case txt:= <- outbox :
			msg, err := json.Marshal(txt)
			if err != nil {
				log.Print("Error marshaling msg")
				sendStopped <- true
				return
			}
			w   := bufio.NewWriter(conn)

			bytesSent, err := w.Write(append(msg, byte('\n')))
			w.Flush()
			if err != nil || bytesSent < 1{
				log.Print("Error sending msg")
				sendStopped <- true
				return
			}
			log.Print("bytes sent: ", bytesSent)
		}
	}
}

func receiver(conn net.Conn, inbox chan<- Cmd, stopRcv <-chan bool, rcvStopped chan<- bool){
	// Make a buffer to hold incoming data.
	//buf := make([]byte, 1024)
	//r,_   := bufio.NewReader(conn).ReadString('\n')

	// Read the incoming connection into the buffer.
	for {
		select {
		case <- stopRcv:
			rcvStopped <- true
			log.Print("receiver stopping.")
			return
		default:
		}
		conn.SetDeadline(time.Now().Add(1e9))
		buf, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Print("Error reading:", err.Error())
			rcvStopped <- true
			return
		}
		log.Print(":", buf)
		var cmd Cmd
		err = json.Unmarshal ([]byte(buf) , &cmd)
		if err != nil {
			log.Print("Error unmarshaling msg:", err)
			rcvStopped <- true
			return
		}
		inbox <- cmd
	}
}
