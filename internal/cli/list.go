package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista los rumis disponibles en el proyecto",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	// Verificar que existe pirqa.yaml
	data, err := os.ReadFile("pirqa.yaml")
	if err != nil {
		return fmt.Errorf("no se encontró pirqa.yaml. Ejecuta 'pirqa init' primero")
	}

	var config PirqaConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error leyendo pirqa.yaml: %w", err)
	}

	fmt.Println()
	color.Cyan("Rumis registrados en este proyecto:")
	fmt.Println()

	if len(config.Rumis) == 0 {
		color.Yellow("  No hay rumis registrados.")
		fmt.Println()
		fmt.Println("  Agrega uno con:")
		fmt.Println("    pirqa add module        → rumi oficial de módulo")
		fmt.Println("    pirqa add project       → rumi oficial de proyecto")
		fmt.Println("    pirqa new mi-rumi       → crea tu propio rumi")
		fmt.Println()
		return nil
	}

	for _, r := range config.Rumis {
		color.Green("  ✓ %s", r.Name)
		if r.Path != "" {
			color.White("      local: %s", r.Path)
		}
		if r.Git != "" {
			color.White("      git:   %s", r.Git)
		}
	}

	fmt.Println()
	color.Cyan("Usa 'pirqa make <nombre>' para generar código desde un rumi")
	fmt.Println()

	return nil
}
