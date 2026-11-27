<h1><picture><source media="(prefers-color-scheme: dark)" srcset="dxregia/web/public/brand/dxregia-logo-dark.svg"><img alt="DXRegia" src="dxregia/web/public/brand/dxregia-logo.svg" height="48"></picture></h1>

<p><a href="https://github.com/wifi75/dxregia/releases"><img src="https://img.shields.io/github/v/release/wifi75/dxregia?display_name=tag&amp;color=147d72" alt="Version"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/actions/workflows/linux-smoke.yml"><img src="https://img.shields.io/github/actions/workflow/status/wifi75/dxregia/linux-smoke.yml?branch=main&amp;label=Test%20Linux&amp;logo=githubactions&amp;logoColor=white" alt="Test Linux"></a>&thinsp;<a href="LICENSE"><img src="https://img.shields.io/badge/License-Artistic%202.0-0b6e99" alt="License: Artistic 2.0"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Status-in%20development-e0a23b" alt="Status: in development"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/commits/main"><img src="https://img.shields.io/github/last-commit/wifi75/dxregia?color=475569" alt="Last commit"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Maintained-yes%202026-2ea44f" alt="Maintained: yes 2026"></a><br>
<a href="dxregia/backend/README.md"><img src="https://img.shields.io/badge/Backend-Go%201.22-00ADD8?logo=go&amp;logoColor=white" alt="Backend: Go 1.22"></a>&thinsp;<a href="dxregia/web/README.md"><img src="https://img.shields.io/badge/Panel-Svelte%205-FF3E00?logo=svelte&amp;logoColor=white" alt="Panel: Svelte 5"></a>&thinsp;<a href="https://www.typescriptlang.org/"><img src="https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&amp;logoColor=white" alt="TypeScript 5.9"></a>&thinsp;<a href="https://vite.dev/"><img src="https://img.shields.io/badge/Vite-7-646CFF?logo=vite&amp;logoColor=white" alt="Vite 7"></a>&thinsp;<a href="https://github.com/EA3CV/dxspider/tree/mojo"><img src="https://img.shields.io/badge/Engine-DXSpider%20Mojo-147d72" alt="Engine: DXSpider Mojo"></a>&thinsp;<a href="https://www.perl.org/"><img src="https://img.shields.io/badge/Perl-5.38-39457E?logo=perl&amp;logoColor=white" alt="Perl 5.38"></a>&thinsp;<a href="docs/INTERFACES.md"><img src="https://img.shields.io/badge/Console-Telnet-334155" alt="Console: Telnet"></a><br>
<a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Platform-Linux-334155?logo=linux&amp;logoColor=white" alt="Platform: Linux"></a>&thinsp;<a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Tested-Ubuntu%2024.04-E95420?logo=ubuntu&amp;logoColor=white" alt="Tested on Ubuntu 24.04, amd64 and arm64"></a>&thinsp;<a href="dxregia/deploy/README.md"><img src="https://img.shields.io/badge/Docker-supported-2496ED?logo=docker&amp;logoColor=white" alt="Docker supported"></a>&thinsp;<a href="dxregia/web/src/translations.ts"><img src="https://img.shields.io/badge/Languages-IT%20%C2%B7%20EN%20%C2%B7%20ES%20%C2%B7%20DE%20%C2%B7%20FR%20%C2%B7%20PT-0f766e" alt="Languages: IT EN ES DE FR PT"></a><br>
<a href="https://semver.org/"><img src="https://img.shields.io/badge/SemVer-2.0.0-3F4551?logo=semver&amp;logoColor=white" alt="SemVer 2.0.0"></a>&thinsp;<a href="CHANGELOG.md"><img src="https://img.shields.io/badge/Changelog-Keep%20a%20Changelog-E05735?logo=keepachangelog&amp;logoColor=white" alt="Keep a Changelog"></a></p>

[English](README.md) · [Italiano](README.it.md) · [Español](README.es.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · **Português**

**Um nó DX Cluster para Linux com painel web, baseado no DXSpider Mojo.**

O DXRegia torna simples instalar, gerir e partilhar um nó DX Cluster. Mantém o DXSpider Mojo como motor de rede (acesso Telnet clássico, comandos e ligações entre nós) e acrescenta um painel web moderno e uma página pública com os spots em direto.

## Portal público e administração

Capturas do DXRegia com dados de demonstração. Clique numa imagem para a ampliar.

### Spots em direto · portal público

Em `/spots`, os radioamadores acompanham os spots DX em tempo real, filtram por banda ou modo e identificam as entidades pelas bandeiras DXCC.

[![Spots em direto · portal público](docs/screenshots/spots.png)](docs/screenshots/spots.png)

### Estado do nó e sessões · painel do sysop

Em `/admin`, o sysop consulta o estado do Telnet e do cluster, o sentido das ligações, as sessões ativas e os gráficos dos recursos do servidor.

[![Estado do nó e sessões · painel do sysop](docs/screenshots/admin.png)](docs/screenshots/admin.png)

## O que faz

- **Página de spots em direto** — pública e em tempo real. Filtros por banda, modo e indicativo (vários em simultâneo), bandeiras DXCC, realce amarelo e um breve som por cada spot novo. Funciona em telemóvel, tablet e computador, com tema claro ou escuro.
- **Spots mais inteligentes** — o modo é deduzido da frequência quando o comentário não o indica, e as operações DX anunciadas (NG3K ADXO) aparecem junto ao indicativo.
- **Seis idiomas** — italiano, inglês, espanhol, alemão, francês e português; qualquer outro idioma usa inglês.
- **Painel de administração em `/admin`** — configuração guiada, acesso com palavra-passe e autenticação de dois fatores opcional, painel em direto com LED de estado e centro de análise.
- **Definições do nó no painel** — mensagem do dia, comandos agendados, script de arranque, listas de bloqueio (palavras, indicativos, endereços IP), saltos, filtros e opções de Telnet, cada uma com um tooltip. O painel fala as seis línguas e mostra quem se ligou ao nó.
- **Registo para enviar spots** — os visitantes registam-se no portal (indicativo, email, palavra-passe) e confirmam por email; só os registados podem enviar spots por Telnet. Opcional, ativa-se no painel.
- **Instalável como app** — o portal instala-se no ecrã principal do telemóvel e abre mesmo sem ligação.
- **Nós parceiros** — adicione, ligue e religue automaticamente; veja quem iniciou cada ligação; elimine um nó para o desligar e revogar o acesso.
- **Pedidos de ligação em automático** — outros sysops pedem a partir da sua página de Contacto; o DXRegia pode criar a conta deles, enviar por e-mail os seus dados de ligação e confirmar quando se ligam (interruptor e limite diário no painel). Para parceiros CC Cluster, é possível autorizar a entrada sem palavra-passe por IP público ou nome DNS. O segundo e-mail confirma uma sessão de entrada ativa.
- **E-mails em HTML** — instruções e confirmações em seis idiomas através do seu próprio servidor SMTP, configurado e testado no painel.
- **Acesso clássico mantido** — Telnet e comandos DXSpider para loggers e utilizadores, compatíveis com os nós existentes.
- **Seguro de operar** — cópias de segurança offline com verificação, atualizações com um só comando e reversão automática, estatísticas de visitas sem cookies nem endereços guardados.
- **Fácil de instalar** — um só script em Debian ou Ubuntu (systemd, amd64 ou arm64), ou Docker.

- Guarde e compare vistas, consulte o catálogo DXCC e o histórico por indicativo e configure alertas que funcionam enquanto o portal estiver aberto.
- **Propagação e QSL** — consulte os índices solares e condições HF de dia e noite e VHF do HamQSL, o mapa MUF de KC2G, spots por região e banda e boletins WWV/WCY. A tabela tem uma coluna QSL para LoTW e eQSL AG. A página de registo explica a ligação Telnet em sete loggers documentados. O sysop edita as regras em Saltos e filtros.

## Início rápido

No Debian ou Ubuntu com systemd, instale o pacote pré-compilado para amd64 ou arm64. É preciso curl e sudo ou uma sessão root.

```sh
curl -fL -o install-release.sh https://raw.githubusercontent.com/wifi75/dxregia/v0.17.1/dxregia/deploy/install-release.sh
sh install-release.sh v0.17.1
```

1. Abra o endereço LAN apresentado pelo instalador, por exemplo `http://SERVIDOR:8080/admin`. Uma nova instalação nativa escuta em todas as interfaces IPv4.
2. Entre como `sysop` com a palavra-passe temporária apresentada e escolha uma nova palavra-passe. O 2FA é opcional e vem desativado. Ative-o em **Gestão → Palavra-passe**, confirme a palavra-passe web, digitalize o QR e guarde os códigos de recuperação.
3. Preencha o assistente de primeira configuração: indicativo do nó, indicativo do sysop, locator e porta Telnet. No fim, a página de spots fica em `/spots` e o Telnet responde na porta 7300.

O guia também descreve Docker e a instalação a partir do código-fonte.

O guia passo a passo explica portas, correio, nós parceiros, atualizações, cópias de segurança e resolução de problemas: [**Guia de instalação e configuração**](docs/INSTALLATION.pt.md).

## Mais documentação

_Os documentos técnicos abaixo estão escritos em italiano._

- [docs/OPERATIONS.md](docs/OPERATIONS.md) — Operações do sysop: nós parceiros, registos, cópias
- [docs/INTERFACES.md](docs/INTERFACES.md) — Página pública, painel de administração, Telnet e marca
- [dxregia/deploy/README.md](dxregia/deploy/README.md) — Referência do instalador Linux
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — Arquitetura
- [TASK.md](TASK.md) — Plano de trabalho
- [CHANGELOG.md](CHANGELOG.md) — Registo de alterações (em inglês)
- [Guias rápidos e protocolos (em italiano)](docs/QUICKSTART.md)
- [Exemplos de peering em seis idiomas](docs/PEERING.md)

**Estrutura do repositório:** `dxregia/` contém o código próprio do DXRegia (backend, painel, instaladores), `engine/` o motor DXSpider Mojo com a sua estrutura original e `docs/` a documentação.

## Estado

O DXRegia está em desenvolvimento ativo (versão 0.17.1). A instalação, o primeiro acesso, o Telnet, a cópia de segurança e o reinício estão testados em Ubuntu 24.04 (amd64 e arm64). Falta ainda: um teste em Proxmox LXC, a entrega de correio com mais fornecedores e um pacote de versão assinado. Relatos e ideias são bem-vindos nas [GitHub Issues](https://github.com/wifi75/dxregia/issues).

## Origem e licença

O DXRegia é um projeto independente baseado no DXSpider Mojo de Dirk Koopman G1TLH e dos seus contribuidores (o [espelho EA3CV](https://github.com/EA3CV/dxspider/tree/mojo)). O repositório mantém os ficheiros originais [COPYING](COPYING) e [LICENSE](LICENSE) (Artistic License 2.0). As operações DX anunciadas vêm do [NG3K ADXO](https://www.ng3k.com/Misc/adxo.html); as bandeiras do [flag-icons](https://github.com/lipis/flag-icons) (MIT).

[Licenças e avisos de terceiros](docs/LICENSING.md)
