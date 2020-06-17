package m2mclient

import "time"

type task struct {
	active bool
	stop   chan bool
}

var tasks = make(map[string]task)

func taskDel(serial string){
	t, ok := tasks[serial]
	if !ok{
		return
	}
	t.stop <- true
	return
}
func taskDo(callback func(*Client, string), c *Client, cmd Cmd) {
	tasks[cmd.Serial] = task{active: true, stop: make(chan bool)}
	go doBg(callback, c, cmd, tasks[cmd.Serial].stop)
}
func doBg(callback func(*Client, string), c *Client, cmd Cmd, stop <-chan bool) {
	if cmd.SchTime.Before(time.Now()) && cmd.SchType > 0 {
		delete(tasks, cmd.Serial)
		return
	}
	firstRun := cmd.SchTime.Sub(time.Now())

	/*runAt, err := time.Parse("2006-01-02 15:04:05", cmd.SchTime)
	if err != nil && cmd.SchType > 0{
		log.Print("incorrect time received for a task")
	}
	period = time.Duration(cmd.SchPeriod)*/
	if cmd.SchType == 0 {
		// Run immedietly once.
		callback(c, cmd.Param)
		delete(tasks, cmd.Serial)
		return
	}
	if cmd.SchType == 1 {
		// Run once at specific tiime
		select {
		case <-stop:
			delete(tasks, cmd.Serial)
			return
		case <-time.After(firstRun):
			callback(c, cmd.Param)
			delete(tasks, cmd.Serial)
			return
		}
	}
	if cmd.SchType == 2 {
		// Run once at specific tiime
		select {
		case <-stop:
			delete(tasks, cmd.Serial)
			return
		case <-time.After(firstRun):
			for{
				callback(c, cmd.Param)
				select{
				case <-stop:
					delete(tasks, cmd.Serial)
					return
				case <-time.After(cmd.SchPeriod):
					continue
				}
			}

		}

	}
}
