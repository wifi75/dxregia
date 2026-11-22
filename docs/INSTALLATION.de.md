# Anleitung zu Installation und Konfiguration

[English](INSTALLATION.md) · [Italiano](INSTALLATION.it.md) · [Español](INSTALLATION.es.md) · **Deutsch** · [Français](INSTALLATION.fr.md) · [Português](INSTALLATION.pt.md)

[← Zurück zur Projektseite](../README.de.md)

## 1. Was du brauchst

- Einen Linux-Server oder -PC mit systemd: Debian oder Ubuntu (getestet auf Ubuntu 24.04, amd64 und arm64). Eine virtuelle Maschine oder ein unprivilegierter LXC-Container geht ebenfalls.
- Für die Installation aus dem Quelltext: Git, Go 1.22+, Node.js 22.12+ mit npm, curl und sudo (oder root). Die Alternative ist Docker, das nur Docker und curl braucht.
- Zwei TCP-Ports: einen für das Web-Panel (Standard 8080) und einen für Telnet (Standard 7300). Sollen andere Knoten oder Benutzer dich aus dem Internet erreichen, leite den Telnet-Port am Router weiter.

## 2. Installation

**Vorgefertigtes Paket.**

```sh
curl -fL -o install-release.sh https://raw.githubusercontent.com/wifi75/dxregia/v0.12.0/dxregia/deploy/install-release.sh
sh install-release.sh v0.12.0
```

Das Skript prüft die SHA-256-Prüfsumme, installiert die Laufzeitabhängigkeiten und startet den Dienst. Es zeigt die Adresse des Panels und ein temporäres Passwort für die erste Anmeldung an.

**Native Installation aus dem Quelltext.** Führe diese Befehle auf dem Server aus:

```sh
git clone https://github.com/wifi75/dxregia.git
cd dxregia
sh dxregia/deploy/install-native.sh
```

Das Skript baut Panel und Backend, installiert die Perl-Abhängigkeiten, legt einen eigenen Benutzer an und startet den systemd-Dienst `dxregia`. Das Programm liegt in `/opt/dxregia`, die Daten in `/var/lib/dxregia`. Am Ende zeigt es die Panel-Adresse und beim ersten Lauf ein vorläufiges Passwort. Ein erneuter Lauf startet den installierten Dienst und behält deine Daten. Es installiert auch den Befehl `dxregia-update`.

Um die Ports zu wählen, setze sie vorher:

```sh
DXREGIA_WEB_PORT=8787 DXREGIA_TELNET_PORT=7373 sh dxregia/deploy/install-native.sh
```

**Docker.** Mit bereits installierten Docker und curl:

```sh
sh dxregia/deploy/install-local.sh
```

Es baut das Image, startet einen Container mit persistentem Volume und zeigt Adresse und vorläufiges Passwort. Ein erneuter Lauf behält Container und Volume. Es gelten dieselben zwei Port-Variablen.

## 3. Das Panel öffnen

Öffne eine vom nativen Installer angezeigte LAN-Adresse oder direkt auf dem Server `http://127.0.0.1:8080/admin`. Eine neue native Installation lauscht auf allen IPv4-Schnittstellen. Bei Docker ist die Admin-Portfreigabe auf Loopback beschränkt. Verwende einen SSH-Tunnel.

```sh
ssh -L 8080:127.0.0.1:8080 USER@SERVER
```

## 4. Erste Anmeldung

1. Öffne `/admin`. Melde dich mit dem Benutzer `sysop` und dem vom Installer angezeigten vorläufigen Passwort an.
2. DXRegia verlangt ein neues Passwort (mindestens 12 Zeichen). Das vorläufige funktioniert dann nicht mehr.
3. Zwei-Faktor-Anmeldung ist optional und standardmäßig aus. Unter **Verwaltung → Passwort** aktivieren, Web-Passwort bestätigen, QR-Code scannen und Wiederherstellungscodes sichern.

Das temporäre Passwort kannst du vor der ersten Änderung auf dem Server lesen:

```sh
sudo cat /var/lib/dxregia/sysop-initial
```

Danach verlangt jede Anmeldung Benutzer, Passwort und den 6-stelligen Code. Der Benutzer ist dein Sysop-Rufzeichen, sobald die Einrichtung abgeschlossen ist (vorher `sysop`); Groß- und Kleinschreibung spielt keine Rolle.

## 5. Ersteinrichtung (Assistent)

Öffne nach der Anmeldung Knoten → Konfiguration (Nodo → Configurazione) und fülle aus: Rufzeichen des Knotens mit SSID (zum Beispiel `IU3WJQ-7`), Sysop-Rufzeichen, Locator, Telnet-Port und optional Name, E-Mail, QTH, Koordinaten und einen öffentlichen Hostnamen. Der öffentliche Hostname ist nur ein Verweis: er ändert weder DNS noch Router. Wenn du schon DXSpider betreibst, kannst du `DXVars.pm` und `Listeners.pm` auf derselben Seite importieren. Übernimm die Einstellungen: DXRegia schreibt die Dateien, die Mojo braucht, und startet den Knoten; das Panel bestätigt, sobald Telnet antwortet.

## 6. Den Rest einrichten

- **Telnet und Firewall.** Stelle sicher, dass der Telnet-Port in der Firewall offen ist. Um aus dem Internet erreichbar zu sein, leite ihn am Router an den Server weiter. Die Optionen stehen unter Knoten → Telnet-Optionen (Opzioni Telnet).
- **Mailserver.** Verwaltung → Mailserver (Gestione → Server di posta). Wähle eine Vorlage (Gmail, Aruba, Office 365) oder gib Server und Port, Verschlüsselung, Absender, Benutzer und Passwort ein, speichere und drücke Test senden (Invia prova). Das Passwort bleibt in einer privaten Datei auf dem Server. Bei Gmail brauchst du ein App-Passwort. Die Mail dient für Hinweise an dich und für die Anleitungen an Partner.
- **Partnerknoten.** Verbindungen → Partnerknoten → Knoten hinzufügen (Collegamenti → Nodi partner → Aggiungi nodo): Rufzeichen, Host, Port und, falls der andere Knoten eines verlangt, das Remote-Passwort. Drücke Speichern und verbinden (Salva e connetti). Setze das Häkchen für den Wiederholungsversuch, um nach einem Abbruch automatisch neu zu verbinden.
- **Backups.** Siehe Abschnitt 9.
- **Sprache, Design und Ton.** Die Symbole oben rechts auf den öffentlichen Seiten.

### Partneranfragen und eingehende Verbindungen

- **Öffentliches Formular.** Aktiviere die automatische Einrichtung unter **Verbindungen → Eingegangene Anfragen**. Wenn SMTP sowie der öffentliche Telnet-Hostname und -Port eingerichtet sind, legt DXRegia geeignete neue Partner an. Sobald DXSpider den eingehenden Zugang freischaltet, sendet DXRegia die Verbindungsdaten und ein eigenes Passwort. Standardmäßig sind zehn Einrichtungen innerhalb von 24 Stunden möglich. Bei deaktivierter Automatik öffnest du die Anfrage und legst den Partner manuell an.
- **Passwortloser Zugang für CC Cluster.** Wähle unter **Verbindungen → Partnerknoten** bei der eingehenden Authentifizierung die Option ohne Passwort. Bestätige die öffentliche Quell-IP oder den Hostnamen und speichere und versende die Anleitung. DXRegia schlägt den gespeicherten Hostnamen vor und aktualisiert seine DNS-Adressen. Der Partner muss sich von einer dieser Adressen verbinden. Du kannst auch ohne E-Mail speichern.
- **Verbindungsbestätigung.** Nach der Anleitung wartet DXRegia bis zu sieben Tage auf eine aktive Sitzung, die der Partner selbst gestartet hat. Sobald sie erkannt wird, folgt eine zweite E-Mail. Eine ausgehende Sitzung löst diese Bestätigung nicht aus. SMTP-Annahme und Nachrichten stehen im **E-Mail-Verlauf für Partner**. Wenn eine ausgehende Sitzung die Verbindung belegt, öffnet der Eingangstest ein Zeitfenster von fünf Minuten. Der automatische Test ist optional und wiederholt das Fenster alle 30 Minuten, solange keine eingehende Sitzung besteht.

## 7. Prüfen, ob es funktioniert

- Öffne `http://SERVER:8080/spots`: die Live-Spot-Seite.
- Knoten → Status und Sitzungen (Stato e sessioni): die vier Kacheln sollten grün sein.
- Teste von einem PC aus Telnet, melde dich dann mit einem Telnet-Client oder Logger an und gib `show/dx` ein:

```sh
python3 dxregia/deploy/check_telnet.py HOST 7300
```

Telnet überträgt im Klartext: nutze Admin-Befehle nur in einem vertrauenswürdigen Netzwerk.

## 8. Aktualisieren

Bei einer nativen Installation führe aus einem beliebigen Verzeichnis aus:

```sh
dxregia-update
```

Es lädt die neueste Version, baut sie, behält die vorherige Kopie und geht selbst zurück, wenn Panel oder Telnet nicht mehr antworten. Daten, Zugangsdaten, Ports und Anpassungen bleiben erhalten. `dxregia-update --check` zeigt nur die installierte und die verfügbare Version. Der Dienst startet neu: Verbindungen fallen etwa eine Minute aus und stellen sich selbst wieder her, und du meldest dich neu am Panel an. Installationen, die vor diesem Befehl entstanden, können `dxregia/deploy/update-native.sh` holen und mit `sh update-native.sh` ausführen. Mit Docker nutze `sh update-docker.sh`.

Du kannst auch im Admin-Panel aktualisieren: **Verwaltung > Updates**. Die Seite zeigt die installierte und die verfügbare Version, **Jetzt prüfen** fragt GitHub ab und **Jetzt aktualisieren** startet nach einer Bestätigung dasselbe Update. Der Fortschritt erscheint live und die Seite verbindet sich nach dem Neustart wieder. Telnet und die Partnerverbindungen sind etwa eine Minute unterbrochen. Wenn du vor dieser Funktion installiert hast, führe einmal `dxregia-update` im Terminal aus, um sie zu aktivieren. Wenn sich nur die Seiten des Panels oder die Dokumentation ändern, werden die Dateien ohne Neustart ersetzt und Telnet bleibt verbunden.

## 9. Backup und Wiederherstellung

Das Offline-Werkzeug braucht Python 3.10+. Halte zuerst den Dienst an, damit die Kopie konsistent ist:

```sh
sudo systemctl stop dxregia
python3 dxregia/deploy/backup_node.py backup /var/lib/dxregia/mojo /path/to/node.zip
python3 dxregia/deploy/backup_node.py verify /path/to/node.zip
sudo systemctl start dxregia
```

Es sichert Engine und Knotendaten, Benutzer, Filter, Partner, Einstellungen und Zugangsdaten: behandle das Archiv vertraulich. Um ein Archiv im Panel zu prüfen, nutze Verwaltung → Backup (Gestione → Backup). Zum Wiederherstellen verwende ein neues, leeres Verzeichnis:

```sh
python3 dxregia/deploy/backup_node.py restore /path/to/node.zip /path/to/restored-node
```

## 10. Einen bestehenden Mojo-Knoten übernehmen

Erstelle bei einem gestoppten Knoten mit SQLite-Benutzerdatenbank eine getrennte Kopie und starte DXRegia darauf. Ersetze den laufenden Knoten erst, nachdem du Telnet, Filter und Verbindungen auf der Kopie geprüft hast:

```sh
python3 -B dxregia/deploy/migrate_mojo.py /path/to/mojo /path/to/mojo-dxregia
```

## 11. Fehlersuche

- **Ich kann mich nicht anmelden.** Der Benutzer ist dein Sysop-Rufzeichen (vor der Einrichtung `sysop`), nicht das Knoten-Rufzeichen. Warte nach einem Fehlversuch eine Sekunde.
- **Der Zwei-Faktor-Code wird abgelehnt.** Die Uhren müssen übereinstimmen: aktiviere die automatische Uhrzeit am Telefon und prüfe den Server mit `timedatectl status` (bei einem LXC-Container den Host). Wurde ein Code als bereits benutzt gemeldet, nimm den nächsten. Als letzte Möglichkeit nutze einen Wiederherstellungscode.
- **Das Panel öffnet sich nicht von einem anderen PC.** Folge Abschnitt 3 und prüfe die Firewall für den Web-Port.
- **Telnet ist nicht erreichbar.** Prüfe, ob der Port lauscht (`ss -ltn | grep 7300`), öffne ihn in der Firewall und leite ihn aus dem Internet am Router weiter.
- **Ein Partner verbindet sich nicht.** Prüfe Host, Port und Passwort sowie welche Seite die Verbindung startet: auf der Partnerkarte verlangt „verbindet sich mit deinem Knoten“, dass dein Telnet-Port von außen erreichbar ist.
- **Logs.** Auf dem Server:

```sh
sudo journalctl -u dxregia -f
```

Das Engine-Log findest du auch unter Knoten → DXSpider-Log (Log DXSpider).

## Mehr

- [Referenz des Linux-Installers](../dxregia/deploy/README.md) (Italienisch)
- [Sysop-Betrieb](OPERATIONS.md) (Italienisch)
- [Änderungsprotokoll](../CHANGELOG.md) (Englisch)

## DXRegia 0.7.0

Die Seite zur Ausbreitung zeigt HamQSL-Solarindizes in quadratischen, farbigen Karten, HF-Bedingungen für Tag und Nacht, VHF-Aktivität und die MUF-Karte von KC2G. Die öffentliche Registrierung erklärt die Telnet-Einrichtung für sieben Logbuchprogramme. Die Sysop-Übersicht zeigt auf breiten Bildschirmen bis zu 40 verbundene Funkamateure pro Seite mit DXCC-Flaggen.

RBN ist standardmäßig deaktiviert. Öffne **Reverse Beacon Network** im Sysop-Menü und wähle **RBN aktivieren** für die ausgehende CW/RTTY-Verbindung. Im öffentlichen Portal erscheinen die empfangenen Spots unter **RBN**. In Telnet oder im Logbuchprogramm aktiviert `SET/SKIMMER CW RTTY` den Empfang und `UNSET/SKIMMER` beendet ihn.

[DXRegia 0.7.0](https://github.com/wifi75/dxregia/releases/tag/v0.7.0)
