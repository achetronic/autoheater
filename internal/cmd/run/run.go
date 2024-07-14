package run

import (
	"fmt"
	"log"
	_ "net/http/pprof"
	"time"

	"github.com/achetronic/autoheater/api/v1alpha1"
	"github.com/achetronic/autoheater/internal/config"
	"github.com/achetronic/autoheater/internal/schedules"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	descriptionShort = `execute the commands from the autoheater config file`

	descriptionLong = `
	Run execute the command list in the hosts specified in the autoheater config file.`

	ConfigFlagErrorMessage       = "impossible to get flag --config: %s"
	LogLevelFlagErrorMessage     = "impossible to get flag --log-level: %s"
	DisableTraceFlagErrorMessage = "impossible to get flag --disable-trace: %s"
	ConfigNotParsedErrorMessage  = "impossible to parse config file: %s"
	EnvNotParsedErrorMessage     = "impossible to parse environment variables: %s"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "run",
		DisableFlagsInUseLine: true,
		Short:                 descriptionShort,
		Long:                  descriptionLong,

		Run: RunCommand,
	}

	cmd.Flags().String("config", "autoheater.yaml", "Path to the YAML config file")
	cmd.Flags().String("log-level", "info", "Verbosity level for logs")
	cmd.Flags().Bool("disable-trace", false, "Disable showing traces in logs")

	return cmd
}

// RunCommand executes the main actions of your application
// https://open-meteo.com/en/docs#latitude=28.0930127&longitude=-16.6357443&hourly=temperature_2m,relativehumidity_2m,apparent_temperature&timezone=Europe%2FLondon&forecast_days=1
func RunCommand(cmd *cobra.Command, args []string) {
	var err error
	var ctx v1alpha1.Context

	// Check the flags for this command
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		log.Fatalf(ConfigFlagErrorMessage, err)
	}

	logLevelFlag, err := cmd.Flags().GetString("log-level")
	if err != nil {
		log.Fatalf(LogLevelFlagErrorMessage, err)
	}

	disableTraceFlag, err := cmd.Flags().GetBool("disable-trace")
	if err != nil {
		log.Fatalf(DisableTraceFlagErrorMessage, err)
	}

	//
	logLevel, _ := zap.ParseAtomicLevel(logLevelFlag)

	// Initialize the logger
	loggerConfig := zap.NewProductionConfig()
	if disableTraceFlag {
		loggerConfig.DisableStacktrace = true
		loggerConfig.DisableCaller = true
	}

	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	loggerConfig.Level.SetLevel(logLevel.Level())

	// Configure the logger
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatal(err)
	}
	sugarLogger := logger.Sugar()

	// Configure application's context
	ctx = v1alpha1.Context{
		Config: &v1alpha1.ConfigSpec{},
		Logger: sugarLogger,
	}

	// Get and parse the config
	configContent, err := config.ReadFile(configPath)
	if err != nil {
		ctx.Logger.Fatalf(fmt.Sprintf(ConfigNotParsedErrorMessage, err))
	}

	// Set the configuration inside the global context
	ctx.Config = &configContent

	//
	schedules.RunScheduler(&ctx)
	//select {}
}
