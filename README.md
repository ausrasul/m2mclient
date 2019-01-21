'''M2m Client package for Go''''
''Usage example''

	package main
	import "github.com/ausrasul/m2mclient"
	import "time"
	import "log"

	func main() {
		cnf := &m2mclient.Config{
			Ip:   "localhost",
			Port: "7000",
			Ttl:  5,
			Uid:  "000c29620510",
			CmdTtl: 2,
		}
		c := m2mclient.New(cnf)
		c.AddHandler("test", Test)
		c.Run()
		time.Sleep(time.Second * 1000)
		c.Stop()
	}

	func Test(c *m2mclient.Client, param string){
		// This is an RPC handler, Do something with the RPC parameters
		log.Print("Tested!!", param, "--")
		c.SendCmd(m2mclient.Cmd{Name: "test_ack", Param: param})
	}
