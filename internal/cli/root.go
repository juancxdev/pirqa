package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pirqa",
	Short: "Pirqa — gestor de rumis para generación de código",
	Long: color.CyanString(`
██████╗ ██╗██████╗  ██████╗  █████╗ 
██╔══██╗██║██╔══██╗██╔═══██╗██╔══██╗
██████╔╝██║██████╔╝██║   ██║███████║
██╔═══╝ ██║██╔══██╗██║▄▄ ██║██╔══██║
██║     ██║██║  ██║╚██████╔╝██║  ██║
╚═╝     ╚═╝╚═╝  ╚═╝ ╚══▀▀═╝ ╚═╝  ╚═╝
`) + `
Gestor de rumis inspirado en la arquitectura inca.
Construye tu código bloque a bloque — como los incas construyeron el Tahuantinsuyo.

Comandos disponibles:
  pirqa init              Inicializa pirqa en el proyecto actual
  pirqa new <nombre>      Crea un nuevo rumi
  pirqa add <nombre>      Agrega un rumi al proyecto
  pirqa make <nombre>     Genera código desde un rumi
  pirqa list              Lista los rumis disponibles
`,
	Version: "0.1.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(makeCmd)
	rootCmd.AddCommand(listCmd)
}
