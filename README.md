# 🏔️ Pirqa

> *Pirqa (quechua): muro de piedras que encajan perfectamente sin mortero.*

Gestor de **rumis** para generación de código, inspirado en la arquitectura del Imperio Inca. Construye tu código bloque a bloque — como los incas construyeron el Tahuantinsuyo.

---

## Instalación

```bash
go install github.com/juancxdev/pirqa/cmd@latest
```

## Comandos

```bash
pirqa init              # Inicializa pirqa en el proyecto actual
pirqa new <nombre>      # Crea un nuevo rumi
pirqa add <nombre>      # Agrega un rumi al proyecto
pirqa make <nombre>     # Genera código desde un rumi
pirqa list              # Lista los rumis disponibles
```

## Uso rápido

```bash
# 1. Inicializar pirqa en tu proyecto
pirqa init

# 2. Agregar un rumi oficial
pirqa add module

# 3. Generar código
pirqa make module
# ¿Cómo se llama el módulo? user
# ✓ Generated 4 file(s). (12ms)
#   created internal/user/handler.go
#   created internal/user/service.go
#   created internal/user/repository.go
#   created internal/user/routes.go
```

## Rumis oficiales

| Rumi | Descripción | Tecnología |
|------|-------------|------------|
| `project` | Estructura base de proyecto | Golang, Flutter |
| `module` | Módulo con handler, service y repository | Golang, Flutter |

Visita [pirqa-rumis](https://github.com/juancxdev/pirqa-rumis) para ver todos los rumis disponibles.

## Crear tu propio rumi

```bash
pirqa new mi-rumi
# ¿Para qué tecnología es este rumi?
# > Golang
# ✓ Generated 8 file(s). (15ms)
```

Esto genera:

```
mi-rumi/
├── rumi.yaml          ← configuración y variables
├── README.md
├── CHANGELOG.md
├── LICENSE
├── __rumi__/          ← tus templates aquí
│   └── HELLO.md
└── hooks/
    ├── pre_gen.go     ← lógica antes de generar
    ├── post_gen.go    ← lógica después de generar
    └── go.mod
```

## Vocabulario

| Quechua | Significado | En pirqa |
|---------|-------------|----------|
| **pirqa** | Muro de piedras | La herramienta CLI |
| **rumi** | Piedra | Template / bloque de código |

---

Desarrollado con 🏔️ desde Perú.
