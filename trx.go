package m2mclient
import (
	"log"
	"net"
	"time"
	"bufio"
	"encoding/json"
	"strconv"
	"io"
)

func sender(conn net.Conn, outbox <-chan Cmd, stopSend <-chan bool, sendStopped chan<- bool){
	for {
		select{
		case <- stopSend:
			sendStopped <- true
			log.Print("sender stopping.")
			return
		case command:= <- outbox :
			msg, err := json.Marshal(command)
			if err != nil {
				log.Print("Error marshaling msg")
				sendStopped <- true
				return
			}
			// sending message header, the data size.
			dataSize := []byte(strconv.Itoa(len(msg)))
			dataSize = append(dataSize, byte('\n'))
			w := bufio.NewWriterSize(conn, len(dataSize))

			bytesSent, err := w.Write(dataSize)
			if err != nil || bytesSent < len(dataSize) {
				log.Print("Error buffering msg len")
				sendStopped <- true
				return
			}
			err = w.Flush()
			if err != nil {
				log.Print("Error sending msg")
				sendStopped <- true
				return
			}
			log.Print("bytes sent (header): ", bytesSent)

			// sending the data
			w = bufio.NewWriterSize(conn, len(msg))

			bytesSent, err = w.Write(msg)
			if err != nil || bytesSent < len(msg){
				log.Print("Error buffering msg")
				sendStopped <- true
				return
			}
			err = w.Flush()
			if err != nil {
				log.Print("Error sending msg")
				sendStopped <- true
				return
			}
			log.Print("bytes sent: ", bytesSent)
		}
	}
}

func receiver(conn net.Conn, inbox chan<- Cmd, stopRcv <-chan bool, rcvStopped chan<- bool){
	chunkSize := 4*1024
	// don't recreate the reader buffer, or else you loose unread buffered data.
	r := bufio.NewReaderSize(conn, chunkSize)
	for {
		select {
		case <- stopRcv:
			rcvStopped <- true
			log.Print("receiver stopping.")
			return
		default:
		}

		conn.SetDeadline(time.Now().Add(1e9))
		// read a chunk and extract following data size to allow us to reassemble the frames.
		dataSizeByte, err := r.ReadBytes('\n')
		if err != nil {
			// expect reader to timeout and re run the loop when no data is in queue.
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Print("Error reading:", err.Error())
			return
			rcvStopped <- true
		}
		dataSizeByte = dataSizeByte[:len(dataSizeByte)-1] // strip the last byte \n
		dataSize, err := strconv.Atoi(string(dataSizeByte))
		if err != nil {
			log.Print("data size not integer ", dataSizeByte, err)
			continue
		}
		log.Print("dataSize ", dataSize)
		// holder for data
		data := make([]byte, 0)
		for {
			// read only enough data, less or equal to chunkSize, but not too much
			// risking reading data that doesn't belong to us.
			buffSize := dataSize
			if dataSize > chunkSize{
				buffSize = chunkSize
			}
			if buffSize == 0 {
				log.Print("data read complete.")
				break
			}
			log.Print("dataSize: ", dataSize, "buffSize ", buffSize)
			buf := make([]byte, buffSize) //the chunk size
			n, err := r.Read(buf) //loading chunk into buffer
			log.Print("read ", n, " bytes.")
			// update the remaining bytes we expect to read.
			dataSize -= n
			buf = buf[:n] // trim the read buffer
			if n == 0 || err != nil{
				if err == io.EOF {
					// this will probably never be the case.
					log.Print("read complete")
					break
				} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
					log.Print("read time out")
					break
				} else if err != nil {
					log.Print("very bad error ", n, err)
					break
				}
				log.Print("unknown error ", n, err)
				break
			} else {
				// reassemble the data.
				data = append(data, buf...)
			}
		}

		var cmd Cmd
		err = json.Unmarshal (data , &cmd)
		if err != nil {
			log.Print("Error unmarshaling msg:", err)
			rcvStopped <- true
			return
		}
		inbox <- cmd
	}
}
