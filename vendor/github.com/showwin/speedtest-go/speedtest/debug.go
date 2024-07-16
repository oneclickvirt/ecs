package speedtest

import (
	"log"
	"os"
)

type Debug struct {
	dbg  *log.Logger
	flag bool
}

func NewDebug() *Debug {
	return &Debug{dbg: log.New(os.Stdout, "[DBG]", log.Ldate|log.Ltime)}
}

func (d *Debug) Enable() {
	d.flag = true
}

func (d *Debug) Println(v ...any) {
	if d.flag {
		d.dbg.Println(v...)
	}
}

func (d *Debug) Printf(format string, v ...any) {
	if d.flag {
		d.dbg.Printf(format, v...)
	}
}

var dbg = NewDebug()
