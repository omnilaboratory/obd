package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/lightningnetwork/lnd"
	"github.com/lightningnetwork/lnd/signal"
	"github.com/omnilaboratory/obd/proxy"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/lightclient"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
)

func initObdLog() {
	_dir := "log"
	_ = tool.PathExistsAndCreate(_dir)
	path := "log/obdServer"
	writer, err := rotatelogs.New(
		path+".%Y%m%d%H%M.log",
		rotatelogs.WithMaxAge(30*34*time.Hour),
		rotatelogs.WithRotationTime(4*time.Hour),
	)

	if err != nil {
		panic(err)
	}
	writers := []io.Writer{
		os.Stdout,
		writer,
	}
	fileAndStdoutWriter := io.MultiWriter(writers...)
	log.SetOutput(fileAndStdoutWriter)
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

// gox compile  https://blog.csdn.net/han0373/article/details/81391455
// gox -os "windows linux darwin" -arch amd64
// gox -os "linux" -arch amd64
func main() {
	config.Init()
	initObdLog()
	//tracker
	err := lightclient.ConnectToTracker()
	if err != nil {
		log.Println("because fail to connect to tracker, obd fail to start")
		return
	}

	routersInit := lightclient.InitRouter()
	addr := ":" + strconv.Itoa(config.ServerPort)
	server := &http.Server{
		Addr:           addr,
		Handler:        routersInit,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	service.Start()

	// Timer
	service.ScheduleService.StartSchedule()

	go proxy.StartGrpcServer()

	log.Println("obd " + tool.GetObdNodeId() + " start in " + config.ChainNodeType)
	log.Println("wsAddress: " + bean.CurrObdNodeInfo.WebsocketLink)

	//https://blog.csdn.net/Binary2014/article/details/103919365
	//server.ListenAndServeTLS()
	//err = server.ListenAndServeTLS("cert.pem", "key.pem")
	//if err != nil {
	//	log.Println(err)
	//}
	//startLnd()
	log.Fatal(server.ListenAndServe())
}

func startLnd() {
	go func() {
		defer os.Exit(1)

		log.Println("lnd server will open 10 second later")
		//sleep 10 second wait for obd startup complete
		time.Sleep(10 * time.Second)
		// Hook interceptor for os signals.
		shutdownInterceptor, err := signal.Intercept()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Load the configuration, and parse any command line options. This
		// function will also set up logging properly.
		loadedConfig, err := lnd.LoadConfig(shutdownInterceptor)
		if err != nil {
			if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
				// Print error if not due to help request.
				err = fmt.Errorf("failed to load config: %w", err)
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			// Help was requested, exit normally.
			os.Exit(0)
		}
		implCfg := loadedConfig.ImplementationConfig(shutdownInterceptor)

		// Call the "real" main in a nested manner so the defers will properly
		// be executed in the case of a graceful shutdown.
		if err = lnd.Main(
			loadedConfig, lnd.ListenerCfg{}, implCfg, shutdownInterceptor,
		); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()
}