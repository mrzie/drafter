package tasks

type taskGroup struct {
	tasks int
	c     chan uint8
	err   error
}

type taskError struct {
	str string
}

func (t taskError) Error() string {
	return t.str
}

func taskErr(str string) error {
	return taskError{str}
}

func (t *taskGroup) Add(i int) {
	t.tasks += i
	if t.tasks < 0 {
		t.Panic(taskErr("task error: negative tasks"))
	}
}

func (t *taskGroup) Done() {
	defer func() {
		recover()
	}()

	t.Add(-1)
	t.c <- 1
}

func (t *taskGroup) Panic(e error) {
	defer func() {
		recover()
	}()

	if e == nil {
		e = taskErr("task error")
	}
	t.err = e
	t.c <- 1
}

func (t *taskGroup) Wait() error {
	defer func() {
		close(t.c)
	}()
	for {
		_, ok := <-t.c
		if !ok {
			return taskErr("task error: tasks already finished.")
		}
		if t.err != nil {
			return t.err
		}
		if t.tasks == 0 {
			return t.err
		}
	}
}

func NewTasks() *taskGroup {
	t := new(taskGroup)
	t.c = make(chan uint8)
	return t
}
