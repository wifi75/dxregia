# Guía de instalación y configuración

[English](INSTALLATION.md) · [Italiano](INSTALLATION.it.md) · **Español** · [Deutsch](INSTALLATION.de.md) · [Français](INSTALLATION.fr.md) · [Português](INSTALLATION.pt.md)

[← Volver a la página del proyecto](../README.es.md)

## 1. Qué necesitas

- Un servidor o PC Linux con systemd: Debian o Ubuntu (probado en Ubuntu 24.04, amd64 y arm64). También sirve una máquina virtual o un contenedor LXC sin privilegios.
- Para instalar desde el código: Git, Go 1.22+, Node.js 22.12+ con npm, curl y sudo (o root). La alternativa es Docker, que solo necesita Docker y curl.
- Dos puertos TCP: uno para el panel web (por defecto 8080) y otro para Telnet (por defecto 7300). Si otros nodos o usuarios deben llegar a ti desde Internet, redirige el puerto Telnet en tu router.

## 2. Instalación

**Paquete precompilado.**

```sh
curl -fL -o install-release.sh https://raw.githubusercontent.com/wifi75/dxregia/v0.15.7/dxregia/deploy/install-release.sh
sh install-release.sh v0.15.7
```

El script verifica la suma SHA-256, instala las dependencias e inicia el servicio. Muestra la dirección del panel y una contraseña temporal para el primer acceso.

**Instalación nativa desde el código.** Ejecuta estos comandos en el servidor:

```sh
git clone https://github.com/wifi75/dxregia.git
cd dxregia
sh dxregia/deploy/install-native.sh
```

El script compila el panel y el backend, instala las dependencias de Perl, crea un usuario dedicado e inicia el servicio systemd `dxregia`. El programa está en `/opt/dxregia` y los datos en `/var/lib/dxregia`. Al final muestra la dirección del panel y, la primera vez, una contraseña temporal. Si lo ejecutas de nuevo, inicia el servicio ya instalado y conserva los datos. También instala el comando `dxregia-update`.

Para elegir los puertos, defínelos antes:

```sh
DXREGIA_WEB_PORT=8787 DXREGIA_TELNET_PORT=7373 sh dxregia/deploy/install-native.sh
```

**Docker.** Con Docker y curl ya instalados:

```sh
sh dxregia/deploy/install-local.sh
```

Construye la imagen e inicia un contenedor con un volumen persistente, y muestra la dirección y la contraseña temporal. Si lo ejecutas de nuevo, conserva el contenedor y el volumen. Valen las mismas dos variables de puertos.

## 3. Abrir el panel

Abre una de las direcciones LAN que muestra el instalador nativo o usa `http://127.0.0.1:8080/admin` en el servidor. Una instalación nativa nueva escucha en todas las interfaces IPv4. Con Docker usa un túnel SSH, ya que el puerto del panel se publica en el loopback del host.

```sh
ssh -L 8080:127.0.0.1:8080 USER@SERVER
```

## 4. Primer acceso

1. Abre `/admin`. Entra con el usuario `sysop` y la contraseña temporal que muestra el instalador.
2. DXRegia pide una contraseña nueva (al menos 12 caracteres). La temporal deja de funcionar.
3. El 2FA es opcional y está desactivado por defecto. Actívalo en **Gestión → Contraseña**, confirma la contraseña web, escanea el código QR y guarda los códigos de recuperación.

Puedes recuperar la contraseña temporal en el servidor antes de cambiarla por primera vez:

```sh
sudo cat /var/lib/dxregia/sysop-initial
```

Desde entonces cada acceso pide usuario, contraseña y el código de 6 dígitos. El usuario es tu indicativo de sysop en cuanto completes la configuración (antes es `sysop`); da igual mayúsculas o minúsculas.

## 5. Primera configuración (asistente)

Tras entrar, abre Nodo → Configuración y rellena: indicativo del nodo con SSID (por ejemplo `IU3WJQ-7`), indicativo del sysop, locator, puerto Telnet y, si quieres, tu nombre, correo, QTH, coordenadas y un nombre de host público. El host público es solo una referencia: no cambia tu DNS ni tu router. Si ya usas DXSpider, puedes importar `DXVars.pm` y `Listeners.pm` desde la misma página. Aplica: DXRegia escribe los archivos que necesita Mojo e inicia el nodo, y el panel confirma cuando Telnet responde.

## 6. Configurar el resto

- **Telnet y cortafuegos.** Comprueba que el puerto Telnet esté abierto en el cortafuegos. Para ser accesible desde Internet, redirígelo en el router hacia el servidor. Las opciones están en Nodo → Opciones Telnet.
- **Servidor de correo.** Gestión → Servidor de correo. Elige un preajuste (Gmail, Aruba, Office 365) o escribe servidor y puerto, cifrado, remitente, usuario y contraseña; guarda y pulsa Enviar prueba. La contraseña queda en el servidor en un archivo privado. Con Gmail hace falta una contraseña de aplicación. El correo sirve para los avisos que recibes y para las instrucciones que se envían a los socios.
- **Nodos socios.** Enlaces → Nodos socios → Añadir nodo: indicativo, host, puerto y, si el otro nodo la pide, la contraseña remota. Pulsa Guardar y conectar. Marca la casilla de reintento para reconectar solo tras una caída.
- **Copias de seguridad.** Mira la sección 9.
- **Idioma, tema y sonido.** Los iconos de arriba a la derecha en las páginas públicas.

### Solicitudes de socios y enlaces entrantes

- **Formulario público.** En **Enlaces → Solicitudes recibidas**, activa la configuración automática. Con SMTP y el nombre de host y puerto Telnet públicos del nodo configurados, DXRegia crea las fichas de los nuevos socios que cumplen los requisitos. Cuando DXSpider habilita la entrada, envía los datos de conexión y una contraseña dedicada. El límite predeterminado es de 10 configuraciones en 24 horas. Con la automatización desactivada, abre la solicitud y prepara la ficha manualmente.
- **Acceso de CC Cluster sin contraseña.** En **Enlaces → Nodos socios**, elige **Autenticación entrante → Sin contraseña**, confirma la IP pública o el nombre de host de origen y guarda y envía las instrucciones. DXRegia propone el host guardado y actualiza sus direcciones DNS. El otro nodo debe conectarse desde una de esas direcciones. También puedes guardar sin enviar el correo.
- **Confirmación del enlace.** Después de enviar las instrucciones, DXRegia espera hasta siete días una sesión activa iniciada por el socio y envía un segundo correo al detectarla. Una sesión saliente no genera esta confirmación. Consulta la aceptación por SMTP y el contenido en **Enlaces → Historial de correos de socios**. Si un enlace saliente está ocupando la sesión, la prueba entrante abre una ventana de cinco minutos. La prueba automática es opcional y repite la ventana cada 30 minutos mientras no haya una sesión entrante.

## 7. Comprobar que funciona

- Abre `http://SERVIDOR:8080/spots`: la página de spots en directo.
- Nodo → Estado y sesiones: los cuatro paneles deben estar en verde.
- Desde un PC prueba Telnet; luego entra con un cliente Telnet o un logger y escribe `show/dx`:

```sh
python3 dxregia/deploy/check_telnet.py HOST 7300
```

Telnet envía los datos sin cifrar: usa los comandos de administración solo en una red de confianza.

## 8. Actualizar

En una instalación nativa, ejecuta desde cualquier directorio:

```sh
dxregia-update
```

Descarga la última versión, la compila, conserva la copia anterior y vuelve atrás sola si el panel o Telnet dejan de responder. Se conservan los datos, las credenciales, los puertos y tus personalizaciones. `dxregia-update --check` solo muestra la versión instalada y la disponible. El servicio se reinicia: los enlaces caen aproximadamente un minuto y se restablecen solos, y debes entrar de nuevo en el panel. Las instalaciones hechas antes de este comando pueden descargar `dxregia/deploy/update-native.sh` y ejecutarlo con `sh update-native.sh`. Con Docker usa `sh update-docker.sh`.

También puedes actualizar desde el panel de administración: **Gestión > Actualizaciones**. La página muestra la versión instalada y la disponible, **Comprobar ahora** consulta GitHub y **Actualizar ahora** inicia la misma actualización tras una confirmación. El progreso se ve en directo y la página se reconecta tras el reinicio. Telnet y los enlaces con los socios se interrumpen durante cerca de un minuto. Si instalaste antes de esta función, ejecuta una vez `dxregia-update` desde el terminal para activarla. Cuando solo cambian las páginas del panel o la documentación, los archivos se sustituyen sin reiniciar y Telnet sigue conectado.

## 9. Copia de seguridad y restauración

La herramienta sin conexión requiere Python 3.10+. Detén antes el servicio para que la copia sea coherente:

```sh
sudo systemctl stop dxregia
python3 dxregia/deploy/backup_node.py backup /var/lib/dxregia/mojo /path/to/node.zip
python3 dxregia/deploy/backup_node.py verify /path/to/node.zip
sudo systemctl start dxregia
```

Guarda el motor y los datos del nodo, usuarios, filtros, socios, ajustes y datos de acceso: trata el archivo como confidencial. Para verificar un archivo desde el panel usa Gestión → Copia de seguridad. Para restaurar, usa un directorio nuevo y vacío:

```sh
python3 dxregia/deploy/backup_node.py restore /path/to/node.zip /path/to/restored-node
```

## 10. Mover un nodo Mojo existente

En un nodo detenido con base de usuarios SQLite, haz una copia aparte e inicia DXRegia sobre ella. Sustituye el nodo en servicio solo después de comprobar Telnet, filtros y enlaces en la copia:

```sh
python3 -B dxregia/deploy/migrate_mojo.py /path/to/mojo /path/to/mojo-dxregia
```

## 11. Resolución de problemas

- **No puedo entrar.** El usuario es tu indicativo de sysop (o `sysop` antes de la configuración), no el indicativo del nodo. Tras un intento fallido espera un segundo.
- **Se rechaza el código de dos factores.** Los relojes deben coincidir: activa la hora automática en el teléfono y comprueba el servidor con `timedatectl status` (en un contenedor LXC, comprueba el anfitrión). Si un código figura como ya usado, pasa al siguiente. Como último recurso usa un código de recuperación.
- **El panel no abre desde otro PC.** Sigue la sección 3 y revisa el cortafuegos para el puerto web.
- **Telnet no es accesible.** Comprueba que el puerto esté escuchando (`ss -ltn | grep 7300`), ábrelo en el cortafuegos y, desde Internet, redirígelo en el router.
- **Un socio no se conecta.** Revisa host, puerto y contraseña, y qué lado inicia el enlace: en la ficha del socio, «se conecta a tu nodo» exige que tu puerto Telnet sea accesible desde fuera.
- **Registros.** En el servidor:

```sh
sudo journalctl -u dxregia -f
```

El registro del motor también está en Nodo → Registro DXSpider.

## Más

- [Referencia del instalador de Linux](../dxregia/deploy/README.md) (en italiano)
- [Operaciones del sysop](OPERATIONS.md) (en italiano)
- [Registro de cambios](../CHANGELOG.md) (en inglés)

## DXRegia 0.7.0

La página de propagación muestra los índices solares de HamQSL en tarjetas cuadradas de colores, las condiciones HF diurnas y nocturnas, la actividad VHF y el mapa MUF de KC2G. El registro público explica cómo configurar Telnet en siete programas de log. El panel del sysop muestra hasta 40 radioaficionados conectados por página en pantallas anchas, con banderas DXCC.

RBN está desactivado de forma predeterminada. Abre **Reverse Beacon Network** en el menú del sysop y pulsa **Activar RBN** para iniciar la conexión saliente CW/RTTY. La pestaña pública **RBN** muestra los spots recibidos. En Telnet o en tu logger, usa `SET/SKIMMER CW RTTY` para recibirlos y `UNSET/SKIMMER` para detenerlos.

[DXRegia 0.7.0](https://github.com/wifi75/dxregia/releases/tag/v0.7.0)
