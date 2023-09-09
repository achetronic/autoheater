package run

import (
	"github.com/achetronic/autoheater/internal/config"
	"github.com/achetronic/autoheater/internal/price"
	"github.com/achetronic/autoheater/internal/weather"
	"github.com/spf13/cobra"
	"log"
)

const (
	descriptionShort = `execute the commands from the kolega config file`

	descriptionLong = `
	Run execute the command list in the hosts specified in the kolega config file.`
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

	return cmd
}

// RunCommand executes the main actions of your application
// https://open-meteo.com/en/docs#latitude=28.0930127&longitude=-16.6357443&hourly=temperature_2m,relativehumidity_2m,apparent_temperature&timezone=Europe%2FLondon&forecast_days=1
func RunCommand(cmd *cobra.Command, args []string) {
	var err error

	configPath, err := cmd.Flags().GetString("config")
	log.Print(configPath)
	if err != nil {
		log.Fatalf("La flag del fichero de config está chunga: %s", err)
	}

	// Get and parse the config
	configContent, err := config.ReadFile(configPath)
	if err != nil {
		log.Fatalf("No se pudo parsear: %v", err)
	}

	log.Print(configContent)

	//os.Exit(0)

	if configContent.Spec.Weather.Enabled {
		zorro, _ := weather.IsHeatingDay(&configContent)
		log.Print(zorro)
	}

	pepe, _ := price.GetApiData(&configContent)
	_ = pepe

	// pass params related to Open Meteo to weather module to get information

	// Pass the params

	_ = err

}
