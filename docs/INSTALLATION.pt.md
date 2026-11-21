# Guia de instalação e configuração

[English](INSTALLATION.md) · [Italiano](INSTALLATION.it.md) · [Español](INSTALLATION.es.md) · [Deutsch](INSTALLATION.de.md) · [Français](INSTALLATION.fr.md) · **Português**

[← Voltar à página do projeto](../README.pt.md)

## 1. O que precisa

- Um servidor ou PC Linux com systemd: Debian ou Ubuntu (testado em Ubuntu 24.04, amd64 e arm64). Uma máquina virtual ou um contentor LXC sem privilégios também serve.
- Para instalar a partir do código: Git, Go 1.22+, Node.js 22.12+ com npm, curl e sudo (ou root). A alternativa é o Docker, que só precisa de Docker e curl.
- Duas portas TCP: uma para o painel web (por omissão 8080) e outra para o Telnet (por omissão 7300). Se outros nós ou utilizadores têm de chegar até si pela Internet, redirecione a porta Telnet no router.

## 2. Instalação

**Pacote pré-compilado.**

```sh
curl -fL -o install-release.sh https://raw.githubusercontent.com/wifi75/dxregia/v0.8.0/dxregia/deploy/install-release.sh
sh install-release.sh v0.8.0
```

O script verifica a soma SHA-256, instala as dependências e inicia o serviço. Mostra o endereço do painel e uma palavra-passe temporária para o primeiro acesso.

**Instalação nativa a partir do código.** Execute estes comandos no servidor:

```sh
git clone https://github.com/wifi75/dxregia.git
cd dxregia
sh dxregia/deploy/install-native.sh
```

O script compila o painel e o backend, instala as dependências Perl, cria um utilizador dedicado e inicia o serviço systemd `dxregia`. O programa fica em `/opt/dxregia` e os dados em `/var/lib/dxregia`. No fim mostra o endereço do painel e, na primeira vez, uma palavra-passe temporária. Executá-lo de novo inicia o serviço já instalado e mantém os dados. Instala também o comando `dxregia-update`.

Para escolher as portas, defina-as antes:

```sh
DXREGIA_WEB_PORT=8787 DXREGIA_TELNET_PORT=7373 sh dxregia/deploy/install-native.sh
```

**Docker.** Com Docker e curl já instalados:

```sh
sh dxregia/deploy/install-local.sh
```

Constrói a imagem e inicia um contentor com um volume persistente, e mostra o endereço e a palavra-passe temporária. Executá-lo de novo mantém o contentor e o volume. Valem as mesmas duas variáveis de portas.

## 3. Abrir o painel

Abra um endereço LAN apresentado pelo instalador nativo ou use `http://127.0.0.1:8080/admin` no servidor. Uma nova instalação nativa escuta em todas as interfaces IPv4. Com Docker, use um túnel SSH, pois a porta do painel é publicada na interface de loopback do anfitrião.

```sh
ssh -L 8080:127.0.0.1:8080 USER@SERVER
```

## 4. Primeiro acesso

1. Abra `/admin`. Entre com o utilizador `sysop` e a palavra-passe temporária mostrada pelo instalador.
2. O DXRegia pede uma palavra-passe nova (pelo menos 12 caracteres). A temporária deixa de funcionar.
3. O 2FA é opcional e vem desativado. Ative-o em **Gestão → Palavra-passe**, confirme a palavra-passe web, digitalize o QR e guarde os códigos de recuperação.

Pode consultar a palavra-passe temporária no servidor antes da primeira alteração:

```sh
sudo cat /var/lib/dxregia/sysop-initial
```

A partir daí cada acesso pede utilizador, palavra-passe e o código de 6 dígitos. O utilizador é o seu indicativo de sysop assim que completar a configuração (antes é `sysop`); maiúsculas ou minúsculas não importam.

## 5. Primeira configuração (assistente)

Depois de entrar, abra Nó → Configuração (Nodo → Configurazione) e preencha: indicativo do nó com SSID (por exemplo `IU3WJQ-7`), indicativo do sysop, locator, porta Telnet e, se quiser, nome, e-mail, QTH, coordenadas e um nome de host público. O host público é só uma referência: não altera o seu DNS nem o router. Se já usa o DXSpider, pode importar `DXVars.pm` e `Listeners.pm` na mesma página. Aplique: o DXRegia escreve os ficheiros de que o Mojo precisa e inicia o nó, e o painel confirma quando o Telnet responde.

## 6. Configurar o resto

- **Telnet e firewall.** Confirme que a porta Telnet está aberta na firewall. Para ser acessível pela Internet, redirecione-a no router para o servidor. As opções estão em Nó → Opções Telnet (Opzioni Telnet).
- **Servidor de correio.** Gestão → Servidor de correio (Gestione → Server di posta). Escolha uma predefinição (Gmail, Aruba, Office 365) ou escreva servidor e porta, cifragem, remetente, utilizador e palavra-passe, grave e carregue em Enviar teste (Invia prova). A palavra-passe fica no servidor num ficheiro privado. Com o Gmail precisa de uma palavra-passe de aplicação. O correio serve para os avisos que recebe e para as instruções enviadas aos parceiros.
- **Nós parceiros.** Ligações → Nós parceiros → Adicionar nó (Collegamenti → Nodi partner → Aggiungi nodo): indicativo, host, porta e, se o outro nó a pedir, a palavra-passe remota. Carregue em Guardar e ligar (Salva e connetti). Marque a caixa de nova tentativa para religar sozinho depois de uma queda.
- **Cópias de segurança.** Veja a secção 9.
- **Idioma, tema e som.** Os ícones no canto superior direito das páginas públicas.

### Pedidos de parceiros e ligações de entrada

- **Formulário público.** Em **Ligações → Pedidos recebidos**, ative a configuração automática. Com SMTP e o nome de anfitrião e a porta Telnet públicos configurados, o DXRegia cria as fichas dos novos parceiros elegíveis. Assim que o DXSpider ativar o acesso de entrada, o DXRegia envia os dados de ligação e uma palavra-passe dedicada. O limite predefinido é de dez configurações em 24 horas. Com a automatização desativada, abra o pedido e prepare manualmente a ficha do parceiro.
- **Acesso CC Cluster sem palavra-passe.** Em **Ligações → Nós parceiros**, escolha a autenticação de entrada sem palavra-passe, confirme o IP público ou o nome de anfitrião de origem e guarde e envie as instruções. O DXRegia propõe o anfitrião guardado e atualiza os respetivos endereços DNS. O parceiro deve ligar-se a partir de um desses endereços. Também pode guardar sem enviar correio.
- **Confirmação da ligação.** Depois de enviar as instruções, o DXRegia aguarda até sete dias por uma sessão ativa iniciada pelo parceiro e envia uma segunda mensagem quando a deteta. Uma sessão de saída não gera esta confirmação. Consulte a aceitação por SMTP e o conteúdo no **Histórico de e-mails dos parceiros**. Se uma ligação de saída estiver a ocupar a sessão, o teste de entrada abre uma janela de cinco minutos. O teste automático é opcional e repete a janela a cada 30 minutos enquanto não existir uma sessão de entrada.

## 7. Confirmar que funciona

- Abra `http://SERVIDOR:8080/spots`: a página de spots em direto.
- Nó → Estado e sessões (Stato e sessioni): os quatro painéis devem estar verdes.
- A partir de um PC teste o Telnet; depois entre com um cliente Telnet ou um logger e escreva `show/dx`:

```sh
python3 dxregia/deploy/check_telnet.py HOST 7300
```

O Telnet envia os dados sem cifra: use os comandos de administração só numa rede de confiança.

## 8. Atualizar

Numa instalação nativa, execute a partir de qualquer diretório:

```sh
dxregia-update
```

Descarrega a versão mais recente, compila-a, guarda a cópia anterior e volta atrás sozinho se o painel ou o Telnet deixarem de responder. Dados, credenciais, portas e personalizações mantêm-se. `dxregia-update --check` só mostra a versão instalada e a disponível. O serviço reinicia: as ligações caem cerca de um minuto e restabelecem-se sozinhas, e volta a entrar no painel. As instalações feitas antes deste comando podem obter `dxregia/deploy/update-native.sh` e executá-lo com `sh update-native.sh`. Com Docker use `sh update-docker.sh`.

Também pode atualizar a partir do painel de administração: **Gestão > Atualizações**. A página mostra a versão instalada e a disponível, **Verificar agora** consulta o GitHub e **Atualizar agora** inicia a mesma atualização após uma confirmação. O progresso aparece em direto e a página volta a ligar-se após o reinício. O Telnet e as ligações com os parceiros são interrompidos durante cerca de um minuto. Se instalou antes desta função, execute uma vez `dxregia-update` no terminal para a ativar.

## 9. Cópia de segurança e restauro

A ferramenta offline requer Python 3.10+. Pare primeiro o serviço para a cópia ficar consistente:

```sh
sudo systemctl stop dxregia
python3 dxregia/deploy/backup_node.py backup /var/lib/dxregia/mojo /path/to/node.zip
python3 dxregia/deploy/backup_node.py verify /path/to/node.zip
sudo systemctl start dxregia
```

Guarda o motor e os dados do nó, utilizadores, filtros, parceiros, definições e dados de acesso: trate o arquivo como confidencial. Para verificar um arquivo no painel use Gestão → Cópia de segurança (Gestione → Backup). Para restaurar, use um diretório novo e vazio:

```sh
python3 dxregia/deploy/backup_node.py restore /path/to/node.zip /path/to/restored-node
```

## 10. Mover um nó Mojo existente

Num nó parado com base de utilizadores SQLite, faça uma cópia separada e inicie o DXRegia nela. Só substitua o nó em serviço depois de verificar Telnet, filtros e ligações na cópia:

```sh
python3 -B dxregia/deploy/migrate_mojo.py /path/to/mojo /path/to/mojo-dxregia
```

## 11. Resolução de problemas

- **Não consigo entrar.** O utilizador é o seu indicativo de sysop (ou `sysop` antes da configuração), não o indicativo do nó. Depois de uma tentativa falhada espere um segundo.
- **O código de dois fatores é recusado.** Os relógios têm de coincidir: ative a hora automática no telemóvel e verifique o servidor com `timedatectl status` (num contentor LXC, verifique o anfitrião). Se um código foi dado como já usado, passe ao seguinte. Em último caso use um código de recuperação.
- **O painel não abre a partir de outro PC.** Siga a secção 3 e verifique a firewall para a porta web.
- **O Telnet não é acessível.** Confirme que a porta está a escutar (`ss -ltn | grep 7300`), abra-a na firewall e, a partir da Internet, redirecione-a no router.
- **Um parceiro não se liga.** Verifique host, porta e palavra-passe, e qual dos lados inicia a ligação: na ficha do parceiro, «liga-se ao seu nó» exige que a sua porta Telnet seja acessível do exterior.
- **Registos.** No servidor:

```sh
sudo journalctl -u dxregia -f
```

O registo do motor também está em Nó → Registo DXSpider (Log DXSpider).

## Mais

- [Referência do instalador Linux](../dxregia/deploy/README.md) (em italiano)
- [Operações do sysop](OPERATIONS.md) (em italiano)
- [Registo de alterações](../CHANGELOG.md) (em inglês)

## DXRegia 0.7.0

A página de propagação mostra os índices solares do HamQSL em cartões quadrados e coloridos, as condições HF diurnas e noturnas, a atividade VHF e o mapa MUF do KC2G. O registo público explica como configurar o Telnet em sete programas de log. O painel do sysop mostra até 40 radioamadores ligados por página em ecrãs largos, com bandeiras DXCC.

O RBN está desativado por omissão. Abra **Reverse Beacon Network** no menu do sysop e selecione **Ativar RBN** para iniciar a ligação de saída CW/RTTY. O separador público **RBN** mostra os spots recebidos. No Telnet ou no logger, use `SET/SKIMMER CW RTTY` para os receber e `UNSET/SKIMMER` para parar.

[DXRegia 0.7.0](https://github.com/wifi75/dxregia/releases/tag/v0.7.0)
