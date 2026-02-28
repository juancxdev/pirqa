package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	addPath string
	addGit  string
)

var addCmd = &cobra.Command{
	Use:   "add <nombre>",
	Short: "Agrega un rumi al proyecto",
	Long: `Agrega un rumi al proyecto desde:
  - Un directorio local (--path)
  - Un repositorio de GitHub (--git)
  - El repositorio oficial pirqa-rumis (por defecto)`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().StringVar(&addPath, "path", "", "Ruta local al rumi")
	addCmd.Flags().StringVar(&addGit, "git", "", "URL del repositorio GitHub")
}

func runAdd(cmd *cobra.Command, args []string) error {
	rumiName := args[0]

	// Verificar que existe pirqa.yaml
	if _, err := os.Stat("pirqa.yaml"); os.IsNotExist(err) {
		return fmt.Errorf("no se encontró pirqa.yaml. Ejecuta 'pirqa init' primero")
	}

	var entry RumiEntry

	switch {
	case addPath != "":
		// Agregar desde ruta local
		absPath, err := filepath.Abs(addPath)
		if err != nil {
			return err
		}
		// Verificar que existe rumi.yaml en esa ruta
		if _, err := os.Stat(filepath.Join(absPath, "rumi.yaml")); os.IsNotExist(err) {
			return fmt.Errorf("no se encontró rumi.yaml en %s", absPath)
		}
		entry = RumiEntry{Name: rumiName, Path: absPath}
		color.Green("✓ Rumi '%s' agregado desde %s", rumiName, absPath)

	case addGit != "":
		// Agregar desde GitHub
		entry = RumiEntry{Name: rumiName, Git: addGit}
		color.Green("✓ Rumi '%s' agregado desde %s", rumiName, addGit)
		color.Yellow("  Ejecuta 'pirqa make %s' para descargarlo y usarlo", rumiName)

	default:
		// Agregar desde repositorio oficial pirqa-rumis
		officialURL := "https://github.com/juancxdev/pirqa-rumis"
		entry = RumiEntry{
			Name: rumiName,
			Git:  officialURL,
		}
		color.Green("✓ Rumi oficial '%s' agregado", rumiName)
		color.White("  Fuente: %s", officialURL)
	}

	// Actualizar pirqa.yaml
	if err := addRumiToConfig(entry); err != nil {
		return fmt.Errorf("error actualizando pirqa.yaml: %w", err)
	}

	fmt.Println()
	fmt.Printf("Ejecuta %s para generar código con este rumi\n",
		color.CyanString("pirqa make %s", rumiName))

	return nil
}

func addRumiToConfig(entry RumiEntry) error {
	// Leer pirqa.yaml actual
	data, err := os.ReadFile("pirqa.yaml")
	if err != nil {
		return err
	}

	var config PirqaConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	// Verificar que no esté ya registrado
	for _, r := range config.Rumis {
		if r.Name == entry.Name {
			color.Yellow("⚠️  El rumi '%s' ya está registrado en pirqa.yaml", entry.Name)
			return nil
		}
	}

	// Agregar
	config.Rumis = append(config.Rumis, entry)

	// Guardar
	out, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	header := "# Archivo de configuración de pirqa\n# Ejecuta 'pirqa add <nombre>' para agregar rumis\n# Visita https://github.com/juancxdev/pirqa-rumis para rumis oficiales\n\n"
	return os.WriteFile("pirqa.yaml", []byte(header+string(out)), 0644)
}
