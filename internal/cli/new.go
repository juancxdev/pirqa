package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/juancxdev/pirqa/internal/rumi"
	"github.com/spf13/cobra"
)

var newOutputDir string

var newCmd = &cobra.Command{
	Use:   "new <nombre>",
	Short: "Crea un nuevo rumi",
	Long:  "Genera la estructura base de un nuevo rumi con su rumi.yaml, templates y hooks.",
	Args:  cobra.ExactArgs(1),
	RunE:  runNew,
}

func init() {
	newCmd.Flags().StringVarP(&newOutputDir, "output", "o", "", "Directorio de salida (por defecto: directorio actual)")
}

func runNew(cmd *cobra.Command, args []string) error {
	rumiName := args[0]

	// Preguntar tecnología
	technology, err := promptTechnology()
	if err != nil {
		return err
	}

	// Determinar directorio de salida
	outputDir := newOutputDir
	if outputDir == "" {
		outputDir = "."
	}
	rumiDir := filepath.Join(outputDir, rumiName)

	// Verificar que no exista ya
	if _, err := os.Stat(rumiDir); err == nil {
		return fmt.Errorf("ya existe un directorio '%s'. Elige otro nombre o directorio", rumiDir)
	}

	start := time.Now()

	// Generar el rumi
	files, err := rumi.Generate(rumiName, technology, rumiDir)
	if err != nil {
		return fmt.Errorf("error generando rumi: %w", err)
	}

	elapsed := time.Since(start)

	// Output estilo Mason
	fmt.Println()
	color.Green("✓ Generated %d file(s). (%s)", len(files), elapsed.Round(time.Millisecond))
	for _, f := range files {
		color.White("  created %s", f)
	}
	fmt.Println()
	color.Cyan("Próximos pasos:")
	fmt.Printf("  1. Edita %s/rumi.yaml para definir tus variables\n", rumiDir)
	fmt.Printf("  2. Agrega tus templates en %s/__rumi__/\n", rumiDir)
	fmt.Printf("  3. Personaliza los hooks en %s/hooks/\n", rumiDir)
	fmt.Printf("  4. Ejecuta 'pirqa make %s' para probar tu rumi\n", rumiName)
	fmt.Println()

	return nil
}

func promptTechnology() (string, error) {
	technologies := []string{
		"Golang",
		"Flutter",
		"Python",
	}

	var selected string
	prompt := &survey.Select{
		Message: "¿Para qué tecnología es este rumi?",
		Options: technologies,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return "", fmt.Errorf("selección cancelada")
	}

	// Normalizar a minúsculas internas
	switch selected {
	case "Golang":
		return "golang", nil
	case "Flutter":
		return "flutter", nil
	case "Python":
		return "python", nil
	default:
		return "golang", nil
	}
}
