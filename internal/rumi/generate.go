package rumi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config representa el rumi.yaml
type Config struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Version     string             `yaml:"version"`
	Technology  string             `yaml:"technology"`
	Environment string             `yaml:"environment"`
	Vars        map[string]VarDef  `yaml:"vars"`
}

type VarDef struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Default     string `yaml:"default,omitempty"`
	Prompt      string `yaml:"prompt"`
}

// Generate crea la estructura de un rumi nuevo
func Generate(name, technology, outputDir string) ([]string, error) {
	var createdFiles []string

	// Crear estructura de directorios
	dirs := []string{
		outputDir,
		filepath.Join(outputDir, "__rumi__"),
		filepath.Join(outputDir, "hooks"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("error creando directorio %s: %w", d, err)
		}
	}

	// Generar archivos según tecnología
	files := buildFileMap(name, technology)
	for path, content := range files {
		fullPath := filepath.Join(outputDir, path)
		// Crear subdirectorio si es necesario
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("error creando %s: %w", path, err)
		}
		createdFiles = append(createdFiles, filepath.Join(filepath.Base(outputDir), path))
	}

	return createdFiles, nil
}

func buildFileMap(name, technology string) map[string]string {
	year := fmt.Sprintf("%d", time.Now().Year())

	files := map[string]string{
		"rumi.yaml":        rumiYaml(name, technology),
		"README.md":        readmeMd(name),
		"CHANGELOG.md":     changelogMd(),
		"LICENSE":          license(year),
		"__rumi__/HELLO.md": helloMd(name),
	}

	// Agregar hooks según tecnología
	hookFiles := hooksForTechnology(technology)
	for path, content := range hookFiles {
		files[path] = content
	}

	return files
}

func rumiYaml(name, technology string) string {
	return fmt.Sprintf(`name: %s
description: "Descripción de tu rumi"
version: 0.1.0
technology: %s

# Variables que el rumi acepta al ejecutar 'pirqa make %s'
# Cada variable tiene: type, description, default (opcional), prompt
vars:
  module_name:
    type: string
    description: "Nombre del módulo a generar"
    prompt: "¿Cómo se llama el módulo?"

  # Ejemplo de variable booleana:
  # with_tests:
  #   type: boolean
  #   description: "¿Incluir archivos de test?"
  #   default: "true"
  #   prompt: "¿Generar archivos de test?"
`, name, technology, name)
}

func readmeMd(name string) string {
	return fmt.Sprintf(`# %s

Un rumi de pirqa para generación de código.

## Uso

` + "```bash\n" + `pirqa make %s
` + "```" + `

## Variables

| Variable | Tipo | Descripción | Default |
|----------|------|-------------|---------|
| module_name | string | Nombre del módulo a generar | — |

## Descripción

Describe aquí qué genera este rumi y cómo usarlo.
`, name, name)
}

func changelogMd() string {
	return `# Changelog

## 0.1.0

- Versión inicial
`
}

func license(year string) string {
	return fmt.Sprintf(`MIT License

Copyright (c) %s

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`, year)
}

func helloMd(name string) string {
	return fmt.Sprintf(`# Hello from %s!

Este archivo es un template de ejemplo.
Reemplázalo con tus propios templates en __rumi__/

Puedes usar variables así: {{.module_name}}
`, name)
}

func hooksForTechnology(technology string) map[string]string {
	switch strings.ToLower(technology) {
	case "golang":
		return golangHooks()
	case "flutter":
		return flutterHooks()
	case "python":
		return pythonHooks()
	default:
		return golangHooks()
	}
}

func golangHooks() map[string]string {
	return map[string]string{
		"hooks/pre_gen.go": `package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Variables disponibles como variables de entorno con prefijo PIRQA_VAR_
	moduleName := os.Getenv("PIRQA_VAR_MODULE_NAME")
	outputDir := os.Getenv("PIRQA_OUTPUT_DIR")

	fmt.Printf("⚙️  Pre-generación del módulo: %s\n", moduleName)

	// Ejemplo: verificar que existe go.mod en el proyecto destino
	goModPath := filepath.Join(outputDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		fmt.Println("⚠️  Advertencia: no se encontró go.mod en el directorio destino.")
		fmt.Println("   Asegúrate de ejecutar 'go mod init' antes de usar este rumi.")
	}

	// Agrega aquí tu lógica de pre-generación:
	// - Validar dependencias
	// - Leer configuración del proyecto destino
	// - Preparar variables dinámicas
}
`,
		"hooks/post_gen.go": `package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	moduleName := os.Getenv("PIRQA_VAR_MODULE_NAME")
	outputDir  := os.Getenv("PIRQA_OUTPUT_DIR")

	fmt.Printf("✅  Post-generación del módulo: %s\n", moduleName)

	// Ejemplo: ejecutar go fmt sobre los archivos generados
	fmt.Println("🔧  Formateando archivos generados...")
	cmd := exec.Command("go", "fmt", "./...")
	cmd.Dir = outputDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️  No se pudo ejecutar go fmt: %v\n", err)
	}

	// Agrega aquí tu lógica de post-generación:
	// - Ejecutar go mod tidy
	// - Instalar dependencias
	// - Formatear código generado
}
`,
		"hooks/go.mod": `module hooks

go 1.23
`,
	}
}

func flutterHooks() map[string]string {
	return map[string]string{
		"hooks/pre_gen.dart": `import 'dart:io';
import 'package:mason/mason.dart';

// Hook de pre-generación para Flutter
// Se ejecuta ANTES de que pirqa genere los archivos del template
void run(HookContext context) async {
  final moduleName = context.vars['module_name'] as String;
  
  context.logger.info('⚙️  Pre-generación del módulo: $moduleName');

  // Ejemplo: leer el nombre del proyecto desde pubspec.yaml
  final pubspecFile = File('pubspec.yaml');
  if (!pubspecFile.existsSync()) {
    context.logger.warn(
      'No se encontró pubspec.yaml. '
      'Asegúrate de ejecutar este rumi desde la raíz de tu proyecto Flutter.',
    );
    return;
  }

  // Agrega aquí tu lógica de pre-generación:
  // - Leer pubspec.yaml para obtener el nombre del proyecto
  // - Validar que las dependencias necesarias estén en pubspec.yaml
  // - Preparar variables dinámicas para los templates
}
`,
		"hooks/post_gen.dart": `import 'dart:io';
import 'package:mason/mason.dart';

// Hook de post-generación para Flutter
// Se ejecuta DESPUÉS de que pirqa genere los archivos del template
void run(HookContext context) async {
  final moduleName = context.vars['module_name'] as String;

  context.logger.info('✅  Post-generación del módulo: $moduleName');

  // Ejemplo: ejecutar flutter pub get después de generar
  final progress = context.logger.progress('Ejecutando flutter pub get...');
  
  final result = await Process.run('flutter', ['pub', 'get']);
  
  if (result.exitCode == 0) {
    progress.complete('Dependencias instaladas correctamente');
  } else {
    progress.fail('Error ejecutando flutter pub get');
    context.logger.err(result.stderr.toString());
  }

  // Agrega aquí tu lógica de post-generación:
  // - Ejecutar build_runner
  // - Formatear archivos con dart format
  // - Registrar el módulo en el inyector de dependencias
}
`,
		"hooks/pubspec.yaml": `name: hooks
description: Hooks para rumi de pirqa
version: 0.1.0

environment:
  sdk: '>=3.0.0 <4.0.0'

dependencies:
  mason: any
`,
	}
}

func pythonHooks() map[string]string {
	return map[string]string{
		"hooks/pre_gen.py": `#!/usr/bin/env python3
"""
Hook de pre-generación para Python.
Se ejecuta ANTES de que pirqa genere los archivos del template.
"""
import os
import sys

def main():
    # Variables disponibles como variables de entorno con prefijo PIRQA_VAR_
    module_name = os.environ.get("PIRQA_VAR_MODULE_NAME", "")
    output_dir  = os.environ.get("PIRQA_OUTPUT_DIR", ".")

    print(f"⚙️  Pre-generación del módulo: {module_name}")

    # Ejemplo: verificar que existe requirements.txt o pyproject.toml
    has_requirements = os.path.exists(os.path.join(output_dir, "requirements.txt"))
    has_pyproject    = os.path.exists(os.path.join(output_dir, "pyproject.toml"))

    if not has_requirements and not has_pyproject:
        print("⚠️  Advertencia: no se encontró requirements.txt ni pyproject.toml.")
        print("   Considera inicializar tu proyecto Python antes de usar este rumi.")

    # Agrega aquí tu lógica de pre-generación:
    # - Validar dependencias del proyecto
    # - Leer configuración desde pyproject.toml
    # - Preparar variables dinámicas

if __name__ == "__main__":
    main()
`,
		"hooks/post_gen.py": `#!/usr/bin/env python3
"""
Hook de post-generación para Python.
Se ejecuta DESPUÉS de que pirqa genere los archivos del template.
"""
import os
import subprocess
import sys

def main():
    module_name = os.environ.get("PIRQA_VAR_MODULE_NAME", "")
    output_dir  = os.environ.get("PIRQA_OUTPUT_DIR", ".")

    print(f"✅  Post-generación del módulo: {module_name}")

    # Ejemplo: formatear archivos generados con black
    print("🔧  Formateando archivos con black...")
    result = subprocess.run(
        ["black", "."],
        cwd=output_dir,
        capture_output=True,
        text=True
    )

    if result.returncode == 0:
        print("   Archivos formateados correctamente.")
    else:
        print(f"⚠️  No se pudo ejecutar black: {result.stderr}")
        print("   Instálalo con: pip install black")

    # Agrega aquí tu lógica de post-generación:
    # - Instalar dependencias con pip
    # - Ejecutar linters
    # - Registrar el módulo en __init__.py

if __name__ == "__main__":
    main()
`,
		"hooks/requirements.txt": `# Dependencias para los hooks de este rumi
# Agrega aquí lo que necesites para tus scripts de pre/post generación
`,
	}
}
