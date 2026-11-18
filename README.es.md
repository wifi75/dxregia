<h1><picture><source media="(prefers-color-scheme: dark)" srcset="dxregia/web/public/brand/dxregia-logo-dark.svg"><img alt="DXRegia" src="dxregia/web/public/brand/dxregia-logo.svg" height="48"></picture></h1>

<p><a href="https://github.com/wifi75/dxregia/releases"><img src="https://img.shields.io/github/v/release/wifi75/dxregia?display_name=tag&amp;color=147d72" alt="Version"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/actions/workflows/linux-smoke.yml"><img src="https://img.shields.io/github/actions/workflow/status/wifi75/dxregia/linux-smoke.yml?branch=main&amp;label=Test%20Linux&amp;logo=githubactions&amp;logoColor=white" alt="Test Linux"></a>&thinsp;<a href="LICENSE"><img src="https://img.shields.io/badge/License-Artistic%202.0-0b6e99" alt="License: Artistic 2.0"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Status-in%20development-e0a23b" alt="Status: in development"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/commits/main"><img src="https://img.shields.io/github/last-commit/wifi75/dxregia?color=475569" alt="Last commit"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Maintained-yes%202026-2ea44f" alt="Maintained: yes 2026"></a></p>

<p><a href="dxregia/backend/README.md"><img src="https://img.shields.io/badge/Backend-Go%201.22-00ADD8?logo=go&amp;logoColor=white" alt="Backend: Go 1.22"></a>&thinsp;<a href="dxregia/web/README.md"><img src="https://img.shields.io/badge/Panel-Svelte%205-FF3E00?logo=svelte&amp;logoColor=white" alt="Panel: Svelte 5"></a>&thinsp;<a href="https://www.typescriptlang.org/"><img src="https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&amp;logoColor=white" alt="TypeScript 5.9"></a>&thinsp;<a href="https://vite.dev/"><img src="https://img.shields.io/badge/Vite-7-646CFF?logo=vite&amp;logoColor=white" alt="Vite 7"></a>&thinsp;<a href="https://github.com/EA3CV/dxspider/tree/mojo"><img src="https://img.shields.io/badge/Engine-DXSpider%20Mojo-147d72" alt="Engine: DXSpider Mojo"></a>&thinsp;<a href="https://www.perl.org/"><img src="https://img.shields.io/badge/Perl-5.38-39457E?logo=perl&amp;logoColor=white" alt="Perl 5.38"></a>&thinsp;<a href="docs/INTERFACES.md"><img src="https://img.shields.io/badge/Console-Telnet-334155" alt="Console: Telnet"></a></p>

<p><a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Platform-Linux-334155?logo=linux&amp;logoColor=white" alt="Platform: Linux"></a>&thinsp;<a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Tested-Ubuntu%2024.04-E95420?logo=ubuntu&amp;logoColor=white" alt="Tested on Ubuntu 24.04, amd64 and arm64"></a>&thinsp;<a href="dxregia/deploy/README.md"><img src="https://img.shields.io/badge/Docker-supported-2496ED?logo=docker&amp;logoColor=white" alt="Docker supported"></a>&thinsp;<a href="dxregia/web/src/translations.ts"><img src="https://img.shields.io/badge/Languages-IT%20%C2%B7%20EN%20%C2%B7%20ES%20%C2%B7%20DE%20%C2%B7%20FR%20%C2%B7%20PT-0f766e" alt="Languages: IT EN ES DE FR PT"></a></p>

<p><a href="https://semver.org/"><img src="https://img.shields.io/badge/SemVer-2.0.0-3F4551?logo=semver&amp;logoColor=white" alt="SemVer 2.0.0"></a>&thinsp;<a href="CHANGELOG.md"><img src="https://img.shields.io/badge/Changelog-Keep%20a%20Changelog-E05735?logo=keepachangelog&amp;logoColor=white" alt="Keep a Changelog"></a></p>

[English](README.md) · [Italiano](README.it.md) · **Español** · [Deutsch](README.de.md) · [Français](README.fr.md) · [Português](README.pt.md)

**Un nodo DX Cluster para Linux con panel web, basado en DXSpider Mojo.**

DXRegia facilita instalar, gestionar y compartir un nodo DX Cluster. Mantiene DXSpider Mojo como motor de red (acceso Telnet clásico, comandos y enlaces entre nodos) y añade un panel web moderno y una página pública con los spots en directo.

## Qué hace

- **Página de spots en directo** — pública y en tiempo real. Filtros por banda, modo e indicativo (varios a la vez), banderas DXCC, resaltado amarillo y un breve sonido por cada spot nuevo. Funciona en móvil, tableta y ordenador, con tema claro u oscuro.
- **Spots más inteligentes** — el modo se deduce de la frecuencia cuando el comentario no lo indica, y las operaciones DX anunciadas (NG3K ADXO) aparecen junto al indicativo.
- **Seis idiomas** — italiano, inglés, español, alemán, francés y portugués; cualquier otro idioma usa inglés.
- **Panel de administración en `/admin`** — configuración guiada, acceso con usuario, contraseña y código de dos factores, panel en directo con LED de estado y centro de analítica.
- **Nodos socios** — añade, conecta y reconecta automáticamente; ve quién inició cada enlace; elimina un nodo para desconectarlo y revocar su acceso.
- **Solicitudes de enlace en automático** — otros sysops lo piden desde tu página de Contacto; DXRegia puede crear su cuenta, enviarles por correo tus datos de conexión y confirmar cuando se conectan (interruptor y límite diario en el panel).
- **Correos en HTML** — instrucciones y confirmaciones en seis idiomas mediante tu propio servidor SMTP, configurado y probado desde el panel.
- **Acceso clásico conservado** — Telnet y comandos DXSpider para loggers y usuarios, compatibles con los nodos existentes.
- **Seguro de operar** — copias de seguridad sin conexión con verificación, actualizaciones con un solo comando y reversión automática, estadísticas de visitas sin cookies ni direcciones guardadas.
- **Fácil de instalar** — un solo script en Debian o Ubuntu (systemd, amd64 o arm64), o Docker.

## Inicio rápido

En Debian o Ubuntu con systemd. Necesitas Git, Go 1.22+, Node.js 22.12+ con npm, curl y sudo.

```sh
git clone https://github.com/wifi75/dxregia.git
cd dxregia
sh dxregia/deploy/install-native.sh
```

1. Abre el panel en `http://SERVIDOR:8080/admin`. Por defecto solo escucha en el propio servidor: desde otro PC usa un túnel SSH, o define `DXREGIA_HTTP_BIND=0.0.0.0` en `/etc/dxregia.env` y ejecuta `sudo systemctl restart dxregia`.
2. Entra con el usuario `sysop` y la contraseña temporal que muestra el instalador, elige una contraseña nueva y activa el acceso en dos pasos con una app de autenticación. Guarda los códigos de recuperación.
3. Completa el asistente de primera configuración: indicativo del nodo, indicativo del sysop, locator y puerto Telnet. Al terminar, la página de spots está en `/spots` y Telnet responde en el puerto 7300.

¿Prefieres Docker? Ejecuta en su lugar `sh dxregia/deploy/install-local.sh`.

La guía paso a paso explica puertos, correo, nodos socios, actualizaciones, copias de seguridad y resolución de problemas: [**Guía de instalación y configuración**](docs/INSTALLATION.es.md).

## Más documentación

_Los documentos técnicos siguientes están escritos en italiano._

- [docs/OPERATIONS.md](docs/OPERATIONS.md) — Operaciones del sysop: nodos socios, registros, copias
- [docs/INTERFACES.md](docs/INTERFACES.md) — Página pública, panel de administración, Telnet y marca
- [dxregia/deploy/README.md](dxregia/deploy/README.md) — Referencia del instalador de Linux
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — Arquitectura
- [TASK.md](TASK.md) — Plan de trabajo
- [CHANGELOG.md](CHANGELOG.md) — Registro de cambios (en inglés)

## Estado

DXRegia está en desarrollo activo (versión 0.3.0). La instalación, el primer acceso, Telnet, la copia de seguridad y el reinicio están probados en Ubuntu 24.04 (amd64 y arm64). Siguen pendientes: la prueba en Proxmox LXC, la entrega de correo con más proveedores y un paquete de versión firmado. Los informes y las ideas son bienvenidos en [GitHub Issues](https://github.com/wifi75/dxregia/issues).

## Créditos y licencia

DXRegia es un proyecto independiente basado en DXSpider Mojo de Dirk Koopman G1TLH y sus colaboradores (el [espejo de EA3CV](https://github.com/EA3CV/dxspider/tree/mojo)). El repositorio conserva los archivos originales [COPYING](COPYING) y [LICENSE](LICENSE) (Artistic License 2.0). Las operaciones DX anunciadas proceden de [NG3K ADXO](https://www.ng3k.com/Misc/adxo.html); las banderas, de [flag-icons](https://github.com/lipis/flag-icons) (MIT).
