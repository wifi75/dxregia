<h1><picture><source media="(prefers-color-scheme: dark)" srcset="dxregia/web/public/brand/dxregia-logo-dark.svg"><img alt="DXRegia" src="dxregia/web/public/brand/dxregia-logo.svg" height="48"></picture></h1>

<p><a href="https://github.com/wifi75/dxregia/releases"><img src="https://img.shields.io/github/v/release/wifi75/dxregia?display_name=tag&amp;color=147d72" alt="Version"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/actions/workflows/linux-smoke.yml"><img src="https://img.shields.io/github/actions/workflow/status/wifi75/dxregia/linux-smoke.yml?branch=main&amp;label=Test%20Linux&amp;logo=githubactions&amp;logoColor=white" alt="Test Linux"></a>&thinsp;<a href="LICENSE"><img src="https://img.shields.io/badge/License-Artistic%202.0-0b6e99" alt="License: Artistic 2.0"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Status-in%20development-e0a23b" alt="Status: in development"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/commits/main"><img src="https://img.shields.io/github/last-commit/wifi75/dxregia?color=475569" alt="Last commit"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Maintained-yes%202026-2ea44f" alt="Maintained: yes 2026"></a><br>
<a href="dxregia/backend/README.md"><img src="https://img.shields.io/badge/Backend-Go%201.22-00ADD8?logo=go&amp;logoColor=white" alt="Backend: Go 1.22"></a>&thinsp;<a href="dxregia/web/README.md"><img src="https://img.shields.io/badge/Panel-Svelte%205-FF3E00?logo=svelte&amp;logoColor=white" alt="Panel: Svelte 5"></a>&thinsp;<a href="https://www.typescriptlang.org/"><img src="https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&amp;logoColor=white" alt="TypeScript 5.9"></a>&thinsp;<a href="https://vite.dev/"><img src="https://img.shields.io/badge/Vite-7-646CFF?logo=vite&amp;logoColor=white" alt="Vite 7"></a>&thinsp;<a href="https://github.com/EA3CV/dxspider/tree/mojo"><img src="https://img.shields.io/badge/Engine-DXSpider%20Mojo-147d72" alt="Engine: DXSpider Mojo"></a>&thinsp;<a href="https://www.perl.org/"><img src="https://img.shields.io/badge/Perl-5.38-39457E?logo=perl&amp;logoColor=white" alt="Perl 5.38"></a>&thinsp;<a href="docs/INTERFACES.md"><img src="https://img.shields.io/badge/Console-Telnet-334155" alt="Console: Telnet"></a><br>
<a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Platform-Linux-334155?logo=linux&amp;logoColor=white" alt="Platform: Linux"></a>&thinsp;<a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Tested-Ubuntu%2024.04-E95420?logo=ubuntu&amp;logoColor=white" alt="Tested on Ubuntu 24.04, amd64 and arm64"></a>&thinsp;<a href="dxregia/deploy/README.md"><img src="https://img.shields.io/badge/Docker-supported-2496ED?logo=docker&amp;logoColor=white" alt="Docker supported"></a>&thinsp;<a href="dxregia/web/src/translations.ts"><img src="https://img.shields.io/badge/Languages-IT%20%C2%B7%20EN%20%C2%B7%20ES%20%C2%B7%20DE%20%C2%B7%20FR%20%C2%B7%20PT-0f766e" alt="Languages: IT EN ES DE FR PT"></a><br>
<a href="https://semver.org/"><img src="https://img.shields.io/badge/SemVer-2.0.0-3F4551?logo=semver&amp;logoColor=white" alt="SemVer 2.0.0"></a>&thinsp;<a href="CHANGELOG.md"><img src="https://img.shields.io/badge/Changelog-Keep%20a%20Changelog-E05735?logo=keepachangelog&amp;logoColor=white" alt="Keep a Changelog"></a></p>

[English](README.md) · [Italiano](README.it.md) · **Español** · [Deutsch](README.de.md) · [Français](README.fr.md) · [Português](README.pt.md)

**Un nodo DX Cluster para Linux con panel web, basado en DXSpider Mojo.**

DXRegia facilita instalar, gestionar y compartir un nodo DX Cluster. Mantiene DXSpider Mojo como motor de red (acceso Telnet clásico, comandos y enlaces entre nodos) y añade un panel web moderno y una página pública con los spots en directo.

## Portal público y administración

Capturas de DXRegia con datos de ejemplo. Haz clic en una imagen para ampliarla.

### Spots en directo · portal público

En `/spots`, los radioaficionados siguen los spots DX en tiempo real, filtran por banda o modo e identifican las entidades por sus banderas DXCC.

[![Spots en directo · portal público](docs/screenshots/spots.png)](docs/screenshots/spots.png)

### Estado del nodo y sesiones · panel del sysop

En `/admin`, el sysop puede consultar el estado de Telnet y del clúster, la dirección de los enlaces, las sesiones activas y los gráficos de recursos del servidor.

[![Estado del nodo y sesiones · panel del sysop](docs/screenshots/admin.png)](docs/screenshots/admin.png)

## Qué hace

- **Página de spots en directo** — pública y en tiempo real. Filtros por banda, modo e indicativo (varios a la vez), banderas DXCC, resaltado amarillo y un breve sonido por cada spot nuevo. Funciona en móvil, tableta y ordenador, con tema claro u oscuro.
- **Spots más inteligentes** — el modo se deduce de la frecuencia cuando el comentario no lo indica, y las operaciones DX anunciadas (NG3K ADXO) aparecen junto al indicativo.
- **Seis idiomas** — italiano, inglés, español, alemán, francés y portugués; cualquier otro idioma usa inglés.
- **Panel de administración en `/admin`** — configuración guiada, acceso con contraseña y 2FA opcional, panel en directo con LED de estado y centro de analítica.
- **Ajustes del nodo en el panel** — mensaje del día, comandos programados, script de inicio, listas de bloqueo (palabras, indicativos, direcciones IP), saltos, filtros y opciones de Telnet, cada uno con su tooltip. El panel habla los seis idiomas y muestra quién se ha conectado al nodo.
- **Registro para enviar spots** — los visitantes se registran en el portal (indicativo, correo, contraseña) y confirman por correo; solo los registrados pueden enviar spots por Telnet. Opcional, se activa en el panel.
- **Instalable como app** — el portal se instala en la pantalla de inicio del móvil y se abre también sin conexión.
- **Nodos socios** — añade, conecta y reconecta automáticamente; ve quién inició cada enlace; elimina un nodo para desconectarlo y revocar su acceso.
- **Solicitudes de enlace en automático** — otros sysops lo piden desde tu página de Contacto; DXRegia puede crear su cuenta, enviarles por correo tus datos de conexión y confirmar cuando se conectan (interruptor y límite diario en el panel). Puedes autorizar la entrada sin contraseña de un nodo CC Cluster por IP pública o nombre DNS. El segundo correo confirma una sesión entrante activa.
- **Correos en HTML** — instrucciones y confirmaciones en seis idiomas mediante tu propio servidor SMTP, configurado y probado desde el panel.
- **Acceso clásico conservado** — Telnet y comandos DXSpider para loggers y usuarios, compatibles con los nodos existentes.
- **Seguro de operar** — copias de seguridad sin conexión con verificación, actualizaciones con un solo comando y reversión automática, estadísticas de visitas sin cookies ni direcciones guardadas.
- **Fácil de instalar** — un solo script en Debian o Ubuntu (systemd, amd64 o arm64), o Docker.

- Guarda y compara vistas de spots, consulta DXCC e historiales por indicativo y configura alertas mientras el portal está abierto.
- **Propagación y QSL** — consulta índices solares y condiciones HF de día y noche y VHF de HamQSL, el mapa MUF de KC2G, spots por región y banda y boletines WWV/WCY. La tabla de spots tiene una columna QSL para LoTW y eQSL AG. La página de registro explica la conexión Telnet desde siete programas de log documentados. El sysop edita las reglas en Saltos y filtros.

## Inicio rápido

En Debian o Ubuntu con systemd, instala el paquete precompilado para amd64 o arm64. Necesitas curl y sudo o acceso como root.

```sh
curl -fL -o install-release.sh https://raw.githubusercontent.com/wifi75/dxregia/v0.9.0/dxregia/deploy/install-release.sh
sh install-release.sh v0.9.0
```

1. Abre la dirección LAN que muestra el instalador, por ejemplo `http://SERVIDOR:8080/admin`. Una instalación nativa nueva escucha en todas las interfaces IPv4.
2. Entra como `sysop` con la contraseña temporal del instalador y elige una nueva contraseña. El 2FA es opcional y está desactivado por defecto. Actívalo en **Gestión → Contraseña**, confirma la contraseña web, escanea el código QR y guarda los códigos de recuperación.
3. Completa el asistente de primera configuración: indicativo del nodo, indicativo del sysop, locator y puerto Telnet. Al terminar, la página de spots está en `/spots` y Telnet responde en el puerto 7300.

La guía también explica Docker y la instalación desde el código fuente.

La guía paso a paso explica puertos, correo, nodos socios, actualizaciones, copias de seguridad y resolución de problemas: [**Guía de instalación y configuración**](docs/INSTALLATION.es.md).

## Más documentación

_Los documentos técnicos siguientes están escritos en italiano._

- [docs/OPERATIONS.md](docs/OPERATIONS.md) — Operaciones del sysop: nodos socios, registros, copias
- [docs/INTERFACES.md](docs/INTERFACES.md) — Página pública, panel de administración, Telnet y marca
- [dxregia/deploy/README.md](dxregia/deploy/README.md) — Referencia del instalador de Linux
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — Arquitectura
- [TASK.md](TASK.md) — Plan de trabajo
- [CHANGELOG.md](CHANGELOG.md) — Registro de cambios (en inglés)
- [Guías rápidas y protocolos (en italiano)](docs/QUICKSTART.md)
- [Ejemplos de peering en seis idiomas](docs/PEERING.md)

**Estructura del repositorio:** `dxregia/` contiene el código propio de DXRegia (backend, panel, instaladores), `engine/` el motor DXSpider Mojo con su estructura original y `docs/` la documentación.

## Estado

DXRegia está en desarrollo activo (versión 0.9.0). La instalación, el primer acceso, Telnet, la copia de seguridad y el reinicio están probados en Ubuntu 24.04 (amd64 y arm64). Siguen pendientes: la prueba en Proxmox LXC, la entrega de correo con más proveedores y un paquete de versión firmado. Los informes y las ideas son bienvenidos en [GitHub Issues](https://github.com/wifi75/dxregia/issues).

## Créditos y licencia

DXRegia es un proyecto independiente basado en DXSpider Mojo de Dirk Koopman G1TLH y sus colaboradores (el [espejo de EA3CV](https://github.com/EA3CV/dxspider/tree/mojo)). El repositorio conserva los archivos originales [COPYING](COPYING) y [LICENSE](LICENSE) (Artistic License 2.0). Las operaciones DX anunciadas proceden de [NG3K ADXO](https://www.ng3k.com/Misc/adxo.html); las banderas, de [flag-icons](https://github.com/lipis/flag-icons) (MIT).

[Licencias y avisos de terceros](docs/LICENSING.md)
