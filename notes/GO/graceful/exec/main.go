package main


import (
	_ "context"
	_ "flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)


var (
	gWg sync.WaitGroup
	gListen *demoListener
	gServer *http.Server
	gHookableSignals []os.Signal
)

func main() {
	fmt.Println("Hello World!")

	gServer = &http.Server{
		Addr:           "localhost:8086",
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 16,
		Handler:		demoHandler{},
	}

	var gracefulChild bool
	var netListen net.Listener
	var err error
	args := os.Args

	//init signal var
	Init()

	//入参解析
	//flag.BoolVar(&gracefulChild, "graceful", false, "listen on fd open 3 (internal use only)")
	if len(args) > 1 && args[1] == "-graceful" {
		gracefulChild = true
	} else {
		gracefulChild = false
	}

	fmt.Println("gracefulChild:", gracefulChild)

	if gracefulChild {
		//重用套接字
		log.Print("main: Listening to existing file descriptor 3.")
		f := os.NewFile(3, "")
		netListen, err = net.FileListener(f)
	} else {
		log.Print("main: Listening on a new file descriptor.")
		netListen, err = net.Listen("tcp", gServer.Addr)
	}
	if err != nil {
		log.Fatal(err)
		return
	}

	if gracefulChild {
		syscall.Kill(syscall.Getppid(), syscall.SIGTERM)
		log.Println("Graceful shutdown parent process.")
	}
	gListen = newDemoListener(netListen)

	//handle signal...
	go handleSignals()

	gServer.Serve(gListen)
	gWg.Wait()

	/*
		    // 注册http请求的处理方法
		    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Hello world!"))
			})

		    // 在8086端口启动http服务，会一直阻塞执行
		    err = http.ListenAndServe("localhost:8086", nil)
		    if err != nil {
		        log.Println(err)
		    }
	*/
}

type demoHandler struct {}

func (handler demoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	time.Sleep(20*time.Second)
	w.Write([]byte("Hello 80 Tencent!"))
}

type demoListener struct {
	net.Listener
	stopped bool
	stop    chan error
}

func newDemoListener(listen net.Listener) (demoListen *demoListener) {
	demoListen = &demoListener{
		Listener: listen,
		stop: make(chan error),
	}

	return
}

func (listen *demoListener) Accept() (conn net.Conn, err error) {
	conn, err = listen.Listener.Accept()
	if err != nil {
		return
	}

	conn = demoConn{Conn: conn}
	gWg.Add(1)
	return
}

func (listen *demoListener) Close() error {
	if listen.stopped {
		return syscall.EINVAL
	}

	listen.stopped = true
	return listen.Listener.Close() //停止接受新的连接
}

//get fd
func (listen *demoListener) File() *os.File {
	// returns a dup(2) - FD_CLOEXEC flag *not* set
	tcpListen := listen.Listener.(*net.TCPListener)
	fd, _ := tcpListen.File()
	return fd
}

type demoConn struct {
	net.Conn
}

func (conn demoConn) Close() error {
	err := conn.Conn.Close()
	if err == nil {
		gWg.Done()
	}

	return nil
}

func Init() {
	gHookableSignals = []os.Signal{
		syscall.SIGUSR2,
		syscall.SIGTERM,
	}
}

func forkProcess() error {
	var err error
	files := []*os.File{gListen.File()} //demo only one //.File()
	path := "./graceful"
	args := []string{
		"-graceful",
	}

	env := append(
		os.Environ(),
		"ENDLESS_CONTINUE=1",
	)
	env = append(env, fmt.Sprintf(`ENDLESS_SOCKET_ORDER=%s`, "0,127.0.0.1"))

	cmd := exec.Command(path, args...)
	//cmd := exec.Command(path, "-graceful", "true")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = files
	cmd.Env = env

	err = cmd.Start()
	if err != nil {
		log.Fatalf("Restart: Failed to launch, error: %v", err)
		return err
	}

	return nil
}

func shutdownProcess() error {
	gServer.SetKeepAlivesEnabled(false)
	gListen.Close()
	//server.Shutdown(context.Background())
	log.Println("shutdownProcess success.")
	return nil
}

func handleSignals() {
	var sig os.Signal
	sigChan := make(chan os.Signal)

	signal.Notify(
		sigChan,
		gHookableSignals...,
	)

	pid := syscall.Getpid()
	for {
		sig = <- sigChan
		switch sig {
		case syscall.SIGUSR2:
			log.Println(pid, "Received SIGUSR2.")
			forkProcess()
		case syscall.SIGTERM:
			log.Println(pid, "Received SIGTERM.")
			shutdownProcess()
		default:
			log.Printf("Received %v: nothing i care about...\n", sig)
		}
	}
}
