package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var makeOutputDir string

var makeCmd = &cobra.Command{
	Use:   "make <nombre>",
	Short: "Genera código desde un rumi",
	Long:  "Procesa los templates del rumi y genera los archivos en el proyecto.",
	Args:  cobra.ExactArgs(1),
	RunE:  runMake,
}

func init() {
	makeCmd.Flags().StringVarP(&makeOutputDir, "output", "o", ".", "Directorio de salida")
}

func runMake(cmd *cobra.Command, args []string) error {
	rumiName := args[0]

	// 1. Buscar el rumi (local o en caché)
	rumiDir, err := findRumi(rumiName)
	if err != nil {
		return fmt.Errorf("rumi '%s' no encontrado. Ejecuta 'pirqa add %s' primero", rumiName, rumiName)
	}

	// 2. Leer rumi.yaml
	config, err := loadRumiConfig(rumiDir)
	if err != nil {
		return fmt.Errorf("error leyendo rumi.yaml: %w", err)
	}

	// 3. Recolectar variables del usuario
	vars, err := collectVars(config.Vars)
	if err != nil {
		return err
	}

	start := time.Now()

	// 4. Ejecutar pre_gen hook si existe
	runHook(rumiDir, "pre_gen", config.Technology, vars, makeOutputDir)

	// 5. Procesar templates en __rumi__/
	templateDir := filepath.Join(rumiDir, "__rumi__")
	files, err := processTemplates(templateDir, makeOutputDir, vars)
	if err != nil {
		return fmt.Errorf("error procesando templates: %w", err)
	}

	// 6. Ejecutar post_gen hook si existe
	runHook(rumiDir, "post_gen", config.Technology, vars, makeOutputDir)

	elapsed := time.Since(start)

	// Output
	fmt.Println()
	color.Green("✓ Generated %d file(s). (%s)", len(files), elapsed.Round(time.Millisecond))
	for _, f := range files {
		color.White("  created %s", f)
	}
	fmt.Println()

	return nil
}

// RumiConfig es el struct del rumi.yaml
type RumiConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Version     string            `yaml:"version"`
	Technology  string            `yaml:"technology"`
	Vars        map[string]VarDef `yaml:"vars"`
}

// ArrayItemField define un sub-campo de una variable tipo array
type ArrayItemField struct {
	Name    string `yaml:"name"`
	Prompt  string `yaml:"prompt"`
	Default string `yaml:"default,omitempty"`
}

type VarDef struct {
	Type        string           `yaml:"type"`
	Description string           `yaml:"description"`
	Default     string           `yaml:"default,omitempty"`
	Prompt      string           `yaml:"prompt"`
	ItemFields  []ArrayItemField `yaml:"item_fields,omitempty"`
	Order       int              `yaml:"order,omitempty"`
}

func loadRumiConfig(rumiDir string) (*RumiConfig, error) {
	data, err := os.ReadFile(filepath.Join(rumiDir, "rumi.yaml"))
	if err != nil {
		return nil, err
	}
	var config RumiConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func findRumi(name string) (string, error) {
	// 1. Buscar en pirqa.yaml del proyecto actual (registrado con 'pirqa add')
	if data, err := os.ReadFile("pirqa.yaml"); err == nil {
		var config PirqaConfig
		if yaml.Unmarshal(data, &config) == nil {
			for _, entry := range config.Rumis {
				if entry.Name != name {
					continue
				}
				// Rumi registrado con --path local
				if entry.Path != "" {
					if _, err := os.Stat(filepath.Join(entry.Path, "rumi.yaml")); err == nil {
						return entry.Path, nil
					}
				}
				// Rumi registrado con --git → buscar en caché ~/.pirqa/rumis/<name>
				if entry.Git != "" {
					cachePath := filepath.Join(os.Getenv("HOME"), ".pirqa", "rumis", name)
					if _, err := os.Stat(filepath.Join(cachePath, "rumi.yaml")); err == nil {
						return cachePath, nil
					}
				}
			}
		}
	}

	// 2. Fallback: buscar directorio local con ese nombre
	candidates := []string{
		name,
		filepath.Join(".", name),
		filepath.Join(os.Getenv("HOME"), ".pirqa", "rumis", name),
	}

	for _, path := range candidates {
		if _, err := os.Stat(filepath.Join(path, "rumi.yaml")); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no encontrado")
}

func collectVars(varDefs map[string]VarDef) (map[string]any, error) {
	result := make(map[string]any)

	// Ordenar las keys para un prompt consistente (por Order si existe, luego alfabético)
	keys := make([]string, 0, len(varDefs))
	for k := range varDefs {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		oi, oj := varDefs[keys[i]].Order, varDefs[keys[j]].Order
		if oi != oj {
			return oi < oj
		}
		return keys[i] < keys[j]
	})

	for _, key := range keys {
		def := varDefs[key]
		switch def.Type {
		case "boolean":
			var answer bool
			prompt := &survey.Confirm{
				Message: def.Prompt,
				Default: def.Default == "true",
			}
			if err := survey.AskOne(prompt, &answer); err != nil {
				return nil, err
			}
			result[key] = answer

		case "array":
			if len(def.ItemFields) == 0 {
				result[key] = []map[string]string{}
				continue
			}
			color.Cyan("\n%s", def.Prompt)
			color.Yellow("  (Presiona ENTER sin escribir nada, o escribe 'exit' para terminar)\n")
			var items []map[string]string
			for itemIdx := 1; ; itemIdx++ {
				color.White("  → Ítem %d:", itemIdx)
				item := map[string]string{}

				// El primer sub-campo decide si continuamos o no
				first := def.ItemFields[0]
				var firstVal string
				if err := survey.AskOne(
					&survey.Input{Message: first.Prompt, Default: first.Default},
					&firstVal,
				); err != nil {
					return nil, err
				}

				val := strings.TrimSpace(firstVal)
				// Salir si está vacío o si escribe comandos típicos de salida
				if val == "" || val == "exit" || val == "quit" || val == "q" {
					break
				}
				item[first.Name] = val

				// Resto de sub-campos
				for _, f := range def.ItemFields[1:] {
					var otherVal string
					if err := survey.AskOne(
						&survey.Input{Message: f.Prompt, Default: f.Default},
						&otherVal,
					); err != nil {
						return nil, err
					}
					item[f.Name] = strings.TrimSpace(otherVal)
				}
				items = append(items, item)
			}
			result[key] = items

		default: // string
			var answer string
			prompt := &survey.Input{
				Message: def.Prompt,
				Default: def.Default,
			}
			if err := survey.AskOne(prompt, &answer, survey.WithValidator(survey.Required)); err != nil {
				return nil, err
			}
			result[key] = answer
		}
	}

	return result, nil
}

func processTemplates(templateDir, outputDir string, vars map[string]any) ([]string, error) {
	var createdFiles []string

	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Ruta relativa desde __rumi__/
		relPath, _ := filepath.Rel(templateDir, path)

		// Resolver nombre de archivo con variables (ej: {{.module_name}}_handler.go)
		resolvedPath, err := resolvePathTemplate(relPath, vars)
		if err != nil {
			return err
		}

		outPath := filepath.Join(outputDir, resolvedPath)

		// Crear directorios necesarios
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		// Leer template
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Procesar template
		tmpl, err := template.New(relPath).Funcs(templateFuncs()).Parse(string(content))
		if err != nil {
			// Si no es un template válido, copiar tal cual
			return os.WriteFile(outPath, content, 0644)
		}

		out, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer out.Close()

		if err := tmpl.Execute(out, vars); err != nil {
			return err
		}

		createdFiles = append(createdFiles, outPath)
		return nil
	})

	return createdFiles, err
}

func resolvePathTemplate(path string, vars map[string]any) (string, error) {
	tmpl, err := template.New("path").Funcs(templateFuncs()).Parse(path)
	if err != nil {
		return path, nil
	}
	var sb strings.Builder
	if err := tmpl.Execute(&sb, vars); err != nil {
		return path, nil
	}
	return sb.String(), nil
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		// snake_case → PascalCase: "user_service" → "UserService"
		"toPascal": toPascalCase,
		// snake_case → camelCase: "user_service" → "userService"
		"toCamel": toCamelCase,
		// cualquier cosa → snake_case: "UserService" → "user_service"
		"toSnake": toSnakeCase,
		// MAYÚSCULAS
		"toUpper": strings.ToUpper,
		// minúsculas
		"toLower": strings.ToLower,
	}
}

func toPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	var result strings.Builder
	for _, w := range words {
		if len(w) > 0 {
			result.WriteString(strings.ToUpper(w[:1]) + w[1:])
		}
	}
	return result.String()
}

func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if len(pascal) == 0 {
		return pascal
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func runHook(rumiDir, hookName, technology string, vars map[string]any, outputDir string) {
	// Los hooks se ejecutan como procesos separados
	// Las variables se pasan como env vars con prefijo PIRQA_VAR_
	// Implementación completa en v0.2.0
	// Por ahora registramos que el hook existe
	var hookFile string
	switch technology {
	case "golang":
		hookFile = filepath.Join(rumiDir, "hooks", hookName+".go")
	case "flutter":
		hookFile = filepath.Join(rumiDir, "hooks", hookName+".dart")
	case "python":
		hookFile = filepath.Join(rumiDir, "hooks", hookName+".py")
	}

	if _, err := os.Stat(hookFile); err == nil {
		color.Cyan("  → ejecutando %s hook...", hookName)
	}
}
