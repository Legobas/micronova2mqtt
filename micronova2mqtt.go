package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"micronova2mqtt/appdata"
	"micronova2mqtt/micronova"
	"micronova2mqtt/mqtt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	appName = "Micronova2MQTT"
)

var (
	//Version import - build with -ldflags "-X main.Version=${VERSION}"
	Version  string = "dev"
	dataMgr  *appdata.DataManager
	mqttConn *mqtt.MqttConnection
)

func init() {
	// Setup logging
	out := zerolog.NewConsoleWriter()
	out.NoColor = true
	out.FormatLevel = func(i any) string {
		return strings.ToUpper(fmt.Sprintf("%-6s", i))
	}
	out.PartsExclude = []string{zerolog.TimestampFieldName, zerolog.CallerFieldName}
	log.Logger = log.Output(out)

	switch strings.ToLower(os.Getenv("LOGLEVEL")) {
	case "debug":
		log.Logger = log.With().Caller().Logger()
		log.Logger = log.Level(zerolog.DebugLevel)
	case "trace":
		log.Logger = log.With().Caller().Logger()
		log.Logger = log.Level(zerolog.TraceLevel)
	default:
		log.Logger = log.Level(zerolog.InfoLevel)
	}
}

func startHealthServer(addr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Info().Msgf("Health server started on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Health server failed")
			os.Exit(1)
		}
	}()

	return srv
}

func shutdownHealthServer(srv *http.Server) {
	log.Info().Msg("Shutting down health server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("error shutting down health server")
	}
}

func checkHealth(healthCheckParam *bool) {
	if *healthCheckParam {
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("http://127.0.0.1:8080/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Error().Msgf("Healthcheck: %s did not respond", appName)
			os.Exit(1)
		}
		os.Exit(0)
	}
}

func main() {
	param := flag.Bool("healthcheck", false, "Run Docker healthcheck")
	flag.Parse()
	checkHealth(param)

	log.Info().Msgf("%s %s", appName, Version)

	// Initialize Config
	var err error
	dataMgr, err = appdata.NewDataManager()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize config")
	}

	var mqttProperties mqtt.MqttProperties
	mqttProperties.ClientId = fmt.Sprintf("%s_%s", appName, fmt.Sprintf("%d", time.Now().UnixMicro()))
	mqttProperties.Url = dataMgr.Config.Mqtt.Url
	mqttProperties.User = dataMgr.Config.Mqtt.Username
	mqttProperties.Password = dataMgr.Config.Mqtt.Password
	mqttProperties.Qos = byte(dataMgr.Config.Mqtt.Qos)
	mqttProperties.Retain = dataMgr.Config.Mqtt.Retain
	mqttProperties.BaseTopic = dataMgr.Config.Mqtt.BaseTopic
	mqttConn, err = mqtt.NewMqttConnection(mqttProperties)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize MQTT")
	}

	customerCode, apiDomain, err := dataMgr.GetBrand(dataMgr.Config.Micronova.Brand)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get brand data")
	}

	go micronova.Run(dataMgr, customerCode, apiDomain, mqttConn)

	srv := startHealthServer("127.0.0.1:8080")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	shutdownHealthServer(srv)
	mqttConn.Close()
	log.Info().Msgf("%s stopped", strings.ToUpper(appName))
}
