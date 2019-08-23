package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/pprof"
	"os"
	"syscall"
	"time"

	"github.com/go-logr/glogr"
	"github.com/golang/glog"
	canaryv1 "github.com/iyacontrol/shareit/pkg/apis/canary/v1"
	cronhpav1 "github.com/iyacontrol/shareit/pkg/apis/cronhpa/v1"
	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/iyacontrol/prophet/version"
)

const (
	// High enough QPS to fit all expected use cases. QPS=0 is not set here, because
	// client code is overriding it.
	defaultQPS = 1e6
	// High enough Burst to fit all expected use cases. Burst=0 is not set here, because
	// client code is overriding it.
	defaultBurst = 1e6
)

func main() {
	logf.SetLogger(glogr.New())
	rand.Seed(time.Now().UnixNano())
	fmt.Println(version.String())

	image := os.Getenv(EnvProphetImage)
	if image == "" {
		glog.Fatal("image must supply!")
	}
	account := os.Getenv(EnvProphetAccount)
	if account == "" {
		glog.Fatal("account must supply!")
	}


	options, err := getOptions()
	if err != nil {
		glog.Fatal(err)
	}
	if options.ShowVersion {
		os.Exit(0)
	}

	restCfg, err := buildRestConfig(options)
	if err != nil {
		glog.Fatal(err)
	}

	mgr, err := manager.New(restCfg, manager.Options{
		Namespace:               options.WatchNamespace,
		SyncPeriod:              &options.SyncPeriod,
		LeaderElection:          options.LeaderElection,
		LeaderElectionID:        options.LeaderElectionID,
		LeaderElectionNamespace: options.LeaderElectionNamespace,
	})
	if err != nil {
		glog.Fatal(err)
	}

	if err := canaryv1.AddToScheme(mgr.GetScheme()); err != nil {
		glog.Fatal(err)
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&canaryv1.Canary{}).
		Complete(&canaryReconciler{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		})
	if err != nil {
		glog.Fatal(err)
	}

	if err := cronhpav1.AddToScheme(mgr.GetScheme()); err != nil {
		glog.Fatal(err)
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&cronhpav1.CronHpa{}).
		Complete(&cronhpaReconciler{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),

			Image: image,
			Account: account,
		})
	if err != nil {
		glog.Fatal(err)
	}

	mux := http.NewServeMux()
	if options.ProfilingEnabled {
		registerProfiler(mux)
	}
	registerHealthz(mux)
	registerHandlers(mux)

	go startHTTPServer(options.HealthzPort, mux)

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		glog.Fatal(err)
	}

}

func buildRestConfig(options *Options) (*rest.Config, error) {
	restCfg, err := clientcmd.BuildConfigFromFlags(options.APIServerHost, options.KubeConfigFile)
	if err != nil {
		return nil, err
	}
	restCfg.QPS = defaultQPS
	restCfg.Burst = defaultBurst
	return restCfg, nil
}

func registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/build", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		b, _ := json.Marshal(version.String())
		_, _ = w.Write(b)
	})

	mux.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		if err != nil {
			glog.Errorf("Unexpected error: %v", err)
		}
	})
}

func registerHealthz(mux *http.ServeMux) {
	healthz.InstallHandler(mux, healthz.PingHealthz)
}

func registerProfiler(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/heap", pprof.Index)
	mux.HandleFunc("/debug/pprof/mutex", pprof.Index)
	mux.HandleFunc("/debug/pprof/goroutine", pprof.Index)
	mux.HandleFunc("/debug/pprof/threadcreate", pprof.Index)
	mux.HandleFunc("/debug/pprof/block", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func startHTTPServer(port int, mux *http.ServeMux) {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%v", port),
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      300 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	glog.Fatal(server.ListenAndServe())
}
