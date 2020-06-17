package m2mclient

import (
	"log"
	"net"
	"time"
)

func authenticate(uid string, conn net.Conn) bool {
	if !isAccepted(conn) {
		log.Print("Connection rejected by server")
		return false
	}
	if !sendUid(conn, uid) {
		log.Print("Failed to send uid")
		return false
	}
	if !getUidConfirm(conn) {
		log.Print("Failed to authenticate uid")
		return false
	}
	return true
}
func sendUid(conn net.Conn, uid string) bool {
	bytesSent, err := conn.Write([]byte(uid))
	if err != nil || bytesSent != len(uid) {
		log.Print("Error sending uid")
		return false
	}
	return true
}

func getUidConfirm(conn net.Conn) bool {
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	conn.SetDeadline(time.Now().Add(1e9))
	reqLen, err := conn.Read(buf)
	if err != nil {
		log.Print("Error reading:", err.Error())
		return false
	}
	if string(buf[:reqLen]) == "UID_ACK" {
		log.Print("uid validated")
		return true
	}
	log.Print("Invalid packet received")
	return false
}

func isAccepted(conn net.Conn) bool {
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	conn.SetDeadline(time.Now().Add(1e9))
	reqLen, err := conn.Read(buf)
	if err != nil {
		log.Print("Error reading:", err.Error())
		return false
	}
	if string(buf[:reqLen]) == "ACCEPTED" {
		log.Print("Connection accepted")
		return true
	}
	log.Print("Invalid packet received")
	return false
}
