package main

import (
	"github.com/kthomas/go-zkp/common"
	"github.com/kthomas/go-zkp/zkp"
)

// TODO: import static and declare bridging header

var (
	cancelF     context.CancelFunc
	closing     uint32
	shutdownCtx context.Context
	sigs        chan os.Signal

	srv *http.Server
)

func init() {
	common.RequireJWT()
}

func main() {
	common.Log.Debugf("starting zkp API...")
	installSignalHandlers()

	runAPI()

	timer := time.NewTicker(runloopTickInterval)
	defer timer.Stop()

	for !shuttingDown() {
		select {
		case <-timer.C:
			// tick... no-op
		case sig := <-sigs:
			common.Log.Debugf("received signal: %s", sig)
			srv.Shutdown(shutdownCtx)
			shutdown()
		case <-shutdownCtx.Done():
			close(sigs)
		default:
			time.Sleep(runloopSleepInterval)
		}
	}

	common.Log.Debug("exiting go-zkp API")
	cancelF()
}

func installSignalHandlers() {
	common.Log.Debug("installing signal handlers for go-zkp API")
	sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	shutdownCtx, cancelF = context.WithCancel(context.Background())
}

func shutdown() {
	if atomic.AddUint32(&closing, 1) == 1 {
		common.Log.Debug("shutting down zkp API")
		cancelF()
	}
}

func runAPI() {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(common.CORSMiddleware())

	r.GET("/status", statusHandler)
	zkp.InstallAPI(r)

	// r.Use(token.AuthMiddleware())
	r.Use(common.AccountingMiddleware())
	r.Use(common.RateLimitingMiddleware())

	zkp.InstallAPI(r)

	srv = &http.Server{
		Addr:    common.ListenAddr,
		Handler: r,
	}

	if common.ServeTLS {
		go srv.ListenAndServeTLS(common.CertificatePath, common.PrivateKeyPath)
	} else {
		go srv.ListenAndServe()
	}

	common.Log.Debugf("listening on %s", common.ListenAddr)
}

// statusHandler for health checks and other service metadata
func statusHandler(c *gin.Context) {
	provide.Render(nil, 204, c)
}

func shuttingDown() bool {
	return (atomic.LoadUint32(&closing) > 0)
}
