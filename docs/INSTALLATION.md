# Installation and configuration guide

**English** · [Italiano](INSTALLATION.it.md) · [Español](INSTALLATION.es.md) · [Deutsch](INSTALLATION.de.md) · [Français](INSTALLATION.fr.md) · [Português](INSTALLATION.pt.md)

[← Back to the project page](../README.md)

## 1. What you need

- A Linux server or PC with systemd: Debian or Ubuntu (tested on Ubuntu 24.04, amd64 and arm64). A virtual machine or an unprivileged LXC container also works.
- For the install from source: Git, Go 1.22+, Node.js 22.12+ with npm, curl and sudo (or root). Docker is the alternative and needs only Docker and curl.
- Two TCP ports: one for the web panel (default 8080) and one for Telnet (default 7300). If other nodes or users must reach you from the Internet, forward the Telnet port on your router.

## 2. Install

**Prebuilt release.**

```sh
curl -fL -o install-release.sh https://raw.githubusercontent.com/wifi75/dxregia/v0.6.3/dxregia/deploy/install-release.sh
sh install-release.sh v0.6.3
```

The script verifies the SHA-256 checksum, installs runtime dependencies and starts the service. It prints the panel address and a temporary password for the first sign-in.

**Native install from source.** Run these commands on the server:

```sh
git clone https://github.com/wifi75/dxregia.git
cd dxregia
sh dxregia/deploy/install-native.sh
```

The script builds the panel and the backend, installs the Perl dependencies, creates a dedicated user and starts the `dxregia` systemd service. The program lives in `/opt/dxregia` and the data in `/var/lib/dxregia`. At the end it prints the panel address and, on the first run, a temporary password. Running it again starts the installed service and keeps your data. It also installs the `dxregia-update` command.

To choose the ports, set them first:

```sh
DXREGIA_WEB_PORT=8787 DXREGIA_TELNET_PORT=7373 sh dxregia/deploy/install-native.sh
```

**Docker.** With Docker and curl already installed:

```sh
sh dxregia/deploy/install-local.sh
```

It builds the image and starts a container with a persistent volume, then prints the address and the temporary password. Running it again keeps the container and the volume. The same two port variables apply.

## 3. Open the panel

Open a LAN URL printed by the native installer, or use `http://127.0.0.1:8080/admin` on the server. A fresh native installation listens on all IPv4 interfaces. For Docker, use an SSH tunnel because the admin port is published on the host loopback.

```sh
ssh -L 8080:127.0.0.1:8080 USER@SERVER
```

## 4. First sign-in

1. Open `/admin`. Sign in with the user `sysop` and the temporary password printed by the installer.
2. DXRegia asks for a new password (at least 12 characters). The temporary one stops working.
3. Two-factor authentication is optional and off by default. To turn it on, go to **Management → Password → Two-step verification**, confirm your web password, scan the QR code and save the recovery codes.

If you missed the temporary password, read it on the server before the first password change:

```sh
sudo cat /var/lib/dxregia/sysop-initial
```

From then on every sign-in asks for the user, the password and the 6-digit code. The user is your sysop callsign as soon as you have completed the setup (before that it is `sysop`); upper or lower case does not matter.

## 5. First setup (wizard)

After signing in, open Node → Configuration (Nodo → Configurazione) and fill in: node callsign with SSID (for example `IU3WJQ-7`), sysop callsign, locator, Telnet port, and optionally your name, email, QTH, coordinates and a public host name. The public host name is only a reference: it does not change your DNS or your router. If you already run DXSpider, you can import `DXVars.pm` and `Listeners.pm` from the same page. Apply the settings: DXRegia writes the files Mojo needs and starts the node, and the panel confirms when Telnet answers.

## 6. Configure the rest

- **Telnet and firewall.** Make sure the Telnet port is open in your firewall. To be reachable from the Internet, forward it on your router to the server. Options are in Node → Telnet options (Opzioni Telnet).
- **Mail server.** Management → Mail server (Gestione → Server di posta). Pick a preset (Gmail, Aruba, Office 365) or type server and port, encryption, sender address, user and password, then Save and press Send test (Invia prova). The password stays on the server in a private file. With Gmail you need an app password. Mail is used for notices to you and for the instructions sent to partners.
- **Partner nodes.** Links → Partner nodes → Add node (Collegamenti → Nodi partner → Aggiungi nodo): callsign, host, port, and the remote password if the other node asks for one. Press Save and connect (Salva e connetti). Tick the retry box to reconnect automatically after a drop.
- **Backups.** See section 9.
- **Language, theme and sound.** The icons at the top right of the public pages.

### Partner requests and inbound links

- **Public form.** Under **Links → Received requests**, enable automatic setup. With SMTP and the local node’s public Telnet hostname and port configured, DXRegia creates eligible new partners. Once DXSpider enables inbound access, DXRegia emails the partner the connection details and a dedicated password. The default limit is 10 automatic setups in 24 hours. With automation off, open the request and prepare the partner record manually.
- **Passwordless CC Cluster access.** Under **Links → Partner nodes**, select **Inbound authentication → Without password**, confirm the public source IP or hostname, then save and send the instructions. DXRegia suggests the node’s saved hostname and refreshes its DNS addresses. The partner must connect from one of those addresses. You can also save without sending email.
- **Connection confirmation.** After the instructions are sent, DXRegia watches for an active session initiated by the partner for up to seven days and sends a second email when it sees one. An outbound session does not trigger this email. Check SMTP acceptance and message history under **Links → Partner email history**. If an existing outgoing session occupies the link, **Test inbound now** opens a five-minute window for the partner to connect. The automatic test is optional and repeats that window every 30 minutes while inbound is disconnected.

## 7. Check that it works

- Open `http://SERVER:8080/spots`: the live spots page.
- Node → Status and sessions (Stato e sessioni): the four tiles should be green.
- From a PC, test Telnet, then log in with a Telnet client or a logger and type `show/dx`:

```sh
python3 dxregia/deploy/check_telnet.py HOST 7300
```

Telnet sends data in clear text: use admin commands only on a trusted network.

## 8. Update

On a native install, run this from any directory:

```sh
dxregia-update
```

It downloads the latest version, builds it, keeps the previous copy and rolls back by itself if the panel or Telnet stop answering. Your data, credentials, ports and customisations are kept. `dxregia-update --check` only shows the installed and the available version. The service restarts: links drop for about a minute and reconnect by themselves, and you sign in to the panel again. Installs made before this command existed can fetch `dxregia/deploy/update-native.sh` and run it with `sh update-native.sh`. For Docker use `sh update-docker.sh`.

## 9. Backup and restore

The offline tool needs Python 3.10+. Stop the service first so the copy is consistent:

```sh
sudo systemctl stop dxregia
python3 dxregia/deploy/backup_node.py backup /var/lib/dxregia/mojo /path/to/node.zip
python3 dxregia/deploy/backup_node.py verify /path/to/node.zip
sudo systemctl start dxregia
```

It saves the engine and node data, users, filters, partners, settings and access data: treat the archive as confidential. To check an archive from the panel, use Management → Backup (Gestione → Backup). To restore, use a new, empty directory:

```sh
python3 dxregia/deploy/backup_node.py restore /path/to/node.zip /path/to/restored-node
```

## 10. Move an existing Mojo node

On a stopped node with an SQLite user database, make a separate copy and start DXRegia on it. Replace the running node only after you have checked Telnet, filters and links on the copy:

```sh
python3 -B dxregia/deploy/migrate_mojo.py /path/to/mojo /path/to/mojo-dxregia
```

## 11. Troubleshooting

- **I cannot sign in.** The user is your sysop callsign (or `sysop` before setup), not the node callsign. After a wrong attempt wait a second.
- **The two-factor code is rejected.** The clocks must agree: turn on automatic time on the phone and check the server with `timedatectl status` (for an LXC container, check the host). Use the next code if one was reported as already used. As a last resort use a recovery code.
- **The panel does not open from another PC.** Follow section 3 and check the firewall for the web port.
- **Telnet is not reachable.** Check that the port is listening (`ss -ltn | grep 7300`), open it in the firewall and, from the Internet, forward it on the router.
- **A partner does not connect.** Check host, port and password, and which side starts the link: in the partner card, "N2KI connects to you" needs your Telnet port to be reachable from outside.
- **Logs.** On the server:

```sh
sudo journalctl -u dxregia -f
```

The engine log is also in Node → DXSpider log (Log DXSpider).

## More

- [Linux installer reference](../dxregia/deploy/README.md) (Italian)
- [Sysop operations](OPERATIONS.md) (Italian)
- [Changelog](../CHANGELOG.md)

## DXRegia 0.6.3

Propagation shows HamQSL solar indices on square, color-coded cards, daytime and nighttime HF conditions, VHF activity and the KC2G MUF map. The public registration page explains Telnet setup for seven loggers. The sysop dashboard shows up to 40 connected radio amateurs per page on wide screens, with DXCC flags.

[DXRegia 0.6.3](https://github.com/wifi75/dxregia/releases/tag/v0.6.3)
