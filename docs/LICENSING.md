# Licenze e avvisi di distribuzione

Inventario verificato il 19 novembre 2026. Ambito: repository sorgente, binario Go, bundle web, installazione nativa e immagine OCI.

## Motore DXSpider

DXRegia incorpora una versione modificata di DXSpider Mojo, di Dirk Koopman (G1TLH) e dei contributori. La base del motore è il [commit upstream 9b55f03f](https://github.com/EA3CV/dxspider/commit/9b55f03ff870d50a9d3e57090dc7408035a4a2b1). I file [COPYING](../COPYING) e [LICENSE](../LICENSE) coincidono con quelli della base upstream e vengono conservati anche nell’installazione del motore.

Il commit upstream [Make license version 2](https://github.com/EA3CV/dxspider/commit/8fdda36498ce8c8b0feaf33d67f6582de5c2b03b) aggiorna esplicitamente `LICENSE` ad [Artistic License 2.0](https://opensource.org/license/Artistic-2.0). `COPYING` conserva il testo Artistic storico. Mantieni entrambi e rispetta gli avvisi specifici dei singoli file. I testi upstream riguardano il motore e i suoi derivati. La scelta esplicita di licenza per i componenti originali separati sotto `dxregia/` resta da registrare.

Differenze rispetto alla base per i file del motore effettivamente installati:

| File | Modifica DXRegia |
| --- | --- |
| `cmd/set/password.pl`, `cmd/unset/password.pl`, `cmd/sysop.pl` | Gestione delle credenziali Telnet con hash e del prompt di autenticazione. |
| `perl/DXCommandmode.pm` | Persistenza degli spot differiti prima dell’inoltro. |
| `perl/ExtMsg.pm` | Banner prima del login, autenticazione, gestione Telnet dei partner e log delle fasi senza credenziali. |
| `perl/Verify.pm` | Uso di Digest::SHA del core Perl per la compatibilità Debian. |
| `perl/Version.pm` | Identificazione della versione e della revisione della base upstream. |

Il repository sorgente comprende questi file e l’integrazione web e backend. Per un eseguibile o un’immagine pubblicati, indica la revisione esatta del sorgente corrispondente nelle note di release e conserva i riferimenti e gli avvisi richiesti. Vedi le sezioni 4–6 del testo Artistic 2.0.

## Dipendenze incluse nei binari e nel bundle

I testi originali delle licenze di queste librerie sono in [third-party-notices.txt](../dxregia/web/public/third-party-notices.txt), copiato da Vite nella distribuzione web e quindi incluso sia nel runtime nativo che nell’immagine OCI.

| Componente | Versione | Licenza | Distribuzione |
| --- | --- | --- | --- |
| Runtime e libreria standard Go | Toolchain del build | BSD-3-Clause e PATENTS | Binario Go |
| `github.com/oschwald/maxminddb-golang` | 1.13.1 | ISC | Binario Go |
| `golang.org/x/crypto` | 0.33.0 | BSD-3-Clause e PATENTS | Binario Go |
| `golang.org/x/sys` | 0.30.0 | BSD-3-Clause e PATENTS | Binario Go |
| `svelte` | 5.57.3 | MIT | JavaScript web |
| `qrcode` | 1.5.4 | MIT | JavaScript web |
| `dijkstrajs` | 1.0.3 | MIT | JavaScript web |
| `esm-env` | 1.2.2 | MIT | JavaScript web |

Inventario Go verificato con `go list -deps` su Linux amd64 e arm64 con CGO disattivato. L’elenco web proviene dai moduli con codice emesso nei chunk di produzione di Rollup/Vite. Le altre dipendenze npm, comprese quelle di compilazione e i binari opzionali per piattaforma, sono censite con versione e licenza dichiarata in [package-lock.json](../dxregia/web/package-lock.json). Aggiorna gli avvisi quando il contenuto della build cambia.

L’installatore nativo usa i pacchetti del sistema. L’immagine finale usa Debian Bookworm e installa Perl, Python, Git, procps, certificati CA e i moduli Perl elencati nel [Containerfile](../dxregia/deploy/Containerfile). Node e la toolchain Go sono negli stadi di compilazione. Per l’artefatto OCI da pubblicare, conserva l’inventario esatto dei pacchetti Debian ottenuto con `dpkg-query -W`, i relativi avvisi e l’identificativo dell’immagine base. Versioni e dipendenze transitive dipendono dal momento della build. Vedi la [documentazione Debian sui file copyright](https://www.debian.org/doc/debian-policy/ch-docs.html#copyright-information).

## Asset e dati

| Origine | Uso e stato |
| --- | --- |
| [flag-icons](https://github.com/lipis/flag-icons) | SVG distribuiti con la [licenza MIT originale](../dxregia/web/public/flags/LICENSE-flag-icons.txt). |
| [DB-IP Country Lite](https://db-ip.com/db/lite.php) | Database facoltativo scaricato durante installazione o aggiornamento nativi. Licenza CC BY 4.0. I collegamenti di attribuzione sono in **Accessi al nodo** e **Log DXSpider**, che mostrano i risultati. |
| [ADXO di NG3K](https://www.ng3k.com/Misc/aboutadxo.html) | Feed esterno consultato dal nodo durante l’esecuzione e accreditato nel portale. Il feed non è incluso nel pacchetto. La fonte conserva il copyright e la pagina consultata non conferma i termini di ripubblicazione. Da chiarire con il titolare. L’integrazione può essere disabilitata con `DXREGIA_ADXO=off`. |
| [Country Files di Jim Reisert, AD1C](https://www.country-files.com/dx-cluster/wpxloc/) | `cty.dat`, `wpxloc.raw` e tabella prefissi inclusi a partire dal motore upstream. Provenienza riscontrata negli avvisi e nella cronologia. I termini specifici di redistribuzione delle copie e dei dati derivati restano da confermare. |

## Verifiche prima della release

L’inventario applicativo e gli avvisi distribuiti sono disponibili. Restano la dichiarazione esplicita per il codice originale DXRegia, la verifica delle condizioni dei dati AD1C e ADXO e il controllo dei pacchetti e degli avvisi dentro l’immagine definitiva.

La mappa di propagazione usa le coste del dataset [Natural Earth Land 1:110m](https://www.naturalearthdata.com/downloads/110m-physical-vectors/110m-land/). I poligoni esterni provengono dal file GeoJSON nel tag `v5.1.2`, con coordinate arrotondate a 0,1° e punti consecutivi duplicati rimossi. Il dataset è di pubblico dominio. [Termini Natural Earth](https://www.naturalearthdata.com/about/terms-of-use/)
