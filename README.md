# M2M Client

A Go Object Request Broker

This is a server client library that lets you send commands in json format to execute remote code
 to multiple clients/workers in the form of broadcast, multicast, individual.
This is the client part.

You can find the server part here [https://github.com/ausrasul/m2mserver](https://github.com/ausrasul/m2mserver)

## When to use Go ORB
- To run functions or OS commands on multiple clients and receive responses.
- To run the above as a scheduled task (like cron job)

## Features:
- Add clients at any time.
- Manage authentication yourself.
- Commands are in json format.
- Immediate or scheduled one time execution, or scheduled repeated task.
- Allowed commands by client and server are user defined.
- Command execution is user implemented.

## Work mechanism:

![alt text](https://github.com/ausrasul/m2mclient/blob/master/image.jpg?raw=true)

The server starts at a given port, and listen to connections.
When a client tries to connect, it is authenticated.
The client or the server can initiate communication (defined by the user)
When the server sends a command to client, the client lookup the command name in a map of handlers.
The corresponding handler is then called and the result is sent back together with the command tracking id (serial number)
The response sent back by the client is considered a "command" by the server, there is no difference.

The server can process or leave that command depends if the user have defined a handler for it.

Server have also a list of handlers mapped per command name.

The server and client have heartbeat, if that is broken, the client is considered disconnected.
The client will automatically try to reconnect after a given time.
The user can see how many clients are defined and how many clients are actually connected to the server.

## Dependencies
None

## Usage

```

package main
import "github.com/ausrasul/m2mclient"
import "time"
import "log"

func main() {
	cnf := &m2mclient.Config{        // configure the client
		Ip:   "localhost",           // server ip address
		Port: "7000",                // server port
		Ttl:  5,                     // heartbeat timeout in seconds
		Uid:  "000c29620510",        // Identification used even for verification at the server.
		CmdTtl: 2,                   // command timeout when sent to server. in seconds.
	}
	c := m2mclient.New(cnf)          // Configure the client

	c.AddHandler("test", test)       // Add supported commands and their handlers.

	c.Run()                          // Start the client, it'll wait for commands from server.
	                                 // It is non blocking, so you have to keep the program running.
	time.Sleep(time.Second * 10)
	command := m2mclient.Cmd{
		Name: "modem_added",
		Param: "test234",
	}
	c.SendCmd(command)               // you can also send commands/responses back to server
	                                 // you should also include the tracking number of the command so
									 // the server can correlate to which command the response is.

	time.Sleep(time.Second * 1000)
	c.Stop()                         // you can also stop the client.
}

func test(c *m2mclient.Client, param string){
	// This is an command handler, Do something with the command's parameters
	log.Print("Tested!!", param, "--")
	c.SendCmd(m2mclient.Cmd{Name: "test_ack", Param: param})
}




```
