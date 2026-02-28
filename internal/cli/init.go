package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// PirqaConfig representa el archivo pirqa.yaml del proyecto
type PirqaConfig struct {
	Rumis []RumiEntry `yaml:"rumis"`
}

type RumiEntry struct {
	Name string `yaml:"name"`
	Path string `yaml:"path,omitempty"`
	Git  string `yaml:"git,omitempty"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa pirqa en el proyecto actual",
	Long:  "Genera el archivo pirqa.yaml en el directorio actual.",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	const configFile = "pirqa.yaml"

	// Verificar si ya existe
	if _, err := os.Stat(configFile); err == nil {
		color.Yellow("⚠️  pirqa.yaml ya existe en este directorio.")
		return nil
	}

	start := time.Now()

	config := PirqaConfig{
		Rumis: []RumiEntry{},
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("error generando pirqa.yaml: %w", err)
	}

	// Agregar comentario de cabecera
	header := "# Archivo de configuración de pirqa\n# Ejecuta 'pirqa add <nombre>' para agregar rumis\n# Visita https://github.com/juancxdev/pirqa-rumis para rumis oficiales\n\n"
	content := header + string(data)

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("error creando pirqa.yaml: %w", err)
	}

	elapsed := time.Since(start)

	fmt.Println()
	color.Green("✓ pirqa inicializado correctamente. (%s)", elapsed.Round(time.Millisecond))
	color.White("  created %s", configFile)
	fmt.Println()
	color.Cyan("Próximos pasos:")
	fmt.Println("  pirqa add module        → agrega el rumi oficial 'module'")
	fmt.Println("  pirqa add project       → agrega el rumi oficial 'project'")
	fmt.Println("  pirqa new mi-rumi       → crea tu propio rumi")
	fmt.Println()

	return nil
}
