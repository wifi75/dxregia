package main

import "strings"

// Lingue supportate dal portale pubblico e dalle email. Per ogni altra lingua si usa l'inglese.
var mailLanguages = []string{"it", "en", "es", "de", "fr", "pt"}

func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if i := strings.IndexAny(lang, "-_"); i > 0 {
		lang = lang[:i]
	}
	for _, supported := range mailLanguages {
		if lang == supported {
			return lang
		}
	}
	return ""
}

func validLang(lang string) bool { return lang == "" || normalizeLang(lang) == lang }

// mailStrings raccoglie tutti i testi delle email per una lingua. I segnaposto
// {call}, {node}, {host}, {port} e {name} vengono sostituiti prima dell'invio.
type mailStrings struct {
	Hello, Footer, AutoNote                                       string
	InstrSubject, InstrTitle, InstrIntro                          string
	DataTitle, LNode, LHost, LPort, LCall, LPass, LProfile, VProf string
	StepsTitle                                                    string
	Steps                                                         [4]string
	ScriptTitle, ScriptNote                                       string
	AfterTitle, AfterText, SecurityNote                           string
	OKSubject, OKTitle, OKText, OKNext                            string
	NoSubject, NoTitle, NoText                                    string
	NoChecks                                                      [3]string
	RecvSubject, RecvTitle, RecvText, RecvNext                    string
	YourData                                                      string
}

var mailTexts = map[string]mailStrings{
	"it": {
		Hello: "Ciao {name},", AutoNote: "Messaggio automatico di DXRegia. Per qualsiasi dubbio rispondi a questa email o scrivi al sysop del nodo {node}.",
		Footer:       "DXRegia · cluster DX per radioamatori",
		InstrSubject: "DXRegia · le istruzioni per collegare {call} al nodo {node}",
		InstrTitle:   "Il tuo nodo {call} è stato abilitato",
		InstrIntro:   "Abbiamo preparato l’accesso del tuo nodo al cluster {node}. Qui sotto trovi tutti i dati e i passi da seguire per completare il collegamento.",
		DataTitle:    "I dati per collegarti", LNode: "Nodo a cui collegarti", LHost: "Indirizzo (host)", LPort: "Porta Telnet", LCall: "Indicativo da usare al login", LPass: "Password dedicata", LProfile: "Profilo del collegamento", VProf: "Nodo DXSpider (peering)",
		StepsTitle: "Cosa fare, passo per passo",
		Steps: [4]string{
			"Sul tuo nodo imposta il collegamento verso {host}, porta {port}, come nodo di tipo DXSpider.",
			"Quando richiesto, accedi con l’indicativo {call} e la password che trovi qui sopra.",
			"Avvia il collegamento. Se il tuo nodo riprova da solo dopo una caduta, ancora meglio.",
			"Controlla sul tuo nodo che compaia {node} tra i collegamenti attivi.",
		},
		ScriptTitle: "Esempio di script di connessione (DXSpider)", ScriptNote: "Adattalo al tuo nodo: è solo un modello di partenza.",
		AfterTitle: "Cosa succede dopo", AfterText: "Appena il sistema vede il tuo nodo collegato, ti arriva una seconda email di conferma. Se dopo un po’ il collegamento non risulta attivo, ti scriviamo per aiutarti a capire il motivo.",
		SecurityNote: "Conserva questa password con cura e non condividerla: serve solo al tuo nodo.",
		OKSubject:    "DXRegia · {call} è collegato: tutto a posto", OKTitle: "Collegamento riuscito",
		OKText:    "Il tuo nodo {call} è collegato al cluster {node} e funziona correttamente. Da ora scambiate spot e informazioni in tempo reale.",
		OKNext:    "Non devi fare altro. Se il collegamento dovesse cadere, il tuo nodo può riprovare da solo; per qualsiasi problema rispondi a questa email.",
		NoSubject: "DXRegia · il collegamento di {call} non è ancora attivo", NoTitle: "Non vediamo ancora il tuo nodo collegato",
		NoText:      "Dopo l’invio delle istruzioni il collegamento tra {call} e {node} non risulta ancora attivo. Può essere solo questione di tempo, ma conviene controllare questi punti:",
		NoChecks:    [3]string{"Host e porta sono quelli indicati nell’email delle istruzioni.", "Il login usa l’indicativo {call} e la password dedicata, senza spazi.", "Il tuo nodo è impostato come DXSpider e la porta è raggiungibile da Internet."},
		RecvSubject: "DXRegia · abbiamo ricevuto la tua richiesta", RecvTitle: "Richiesta ricevuta",
		RecvText: "Grazie {name}, abbiamo ricevuto la richiesta per {call}. Il sysop del nodo {node} la esamina e ti risponde a questo indirizzo.",
		RecvNext: "Se la richiesta riguarda un collegamento tra nodi, riceverai i dati di accesso con una nuova email non appena sarà pronto.", YourData: "Riepilogo della richiesta",
	},
	"en": {
		Hello: "Hello {name},", AutoNote: "Automatic message from DXRegia. If in doubt, reply to this email or write to the sysop of node {node}.",
		Footer:       "DXRegia · DX cluster for radio amateurs",
		InstrSubject: "DXRegia · instructions to connect {call} to node {node}",
		InstrTitle:   "Your node {call} has been enabled",
		InstrIntro:   "We have prepared your node’s access to the {node} cluster. Below you will find all the details and the steps to complete the link.",
		DataTitle:    "Your connection details", LNode: "Node to connect to", LHost: "Address (host)", LPort: "Telnet port", LCall: "Callsign to use at login", LPass: "Dedicated password", LProfile: "Link profile", VProf: "DXSpider node (peering)",
		StepsTitle: "What to do, step by step",
		Steps: [4]string{
			"On your node, set up a link to {host}, port {port}, as a DXSpider-type node.",
			"When prompted, log in with the callsign {call} and the password shown above.",
			"Start the link. If your node retries by itself after a drop, even better.",
			"Check on your node that {node} appears among the active links.",
		},
		ScriptTitle: "Example connection script (DXSpider)", ScriptNote: "Adapt it to your node: it is only a starting template.",
		AfterTitle: "What happens next", AfterText: "As soon as the system sees your node connected, you will get a second confirmation email. If the link is still not active after a while, we will write to help you find out why.",
		SecurityNote: "Keep this password safe and do not share it: it is only for your node.",
		OKSubject:    "DXRegia · {call} is connected: all good", OKTitle: "Link established",
		OKText:    "Your node {call} is connected to the {node} cluster and working correctly. From now on you exchange spots and information in real time.",
		OKNext:    "There is nothing else to do. If the link drops, your node can retry by itself; for any problem just reply to this email.",
		NoSubject: "DXRegia · the link for {call} is not active yet", NoTitle: "We do not see your node connected yet",
		NoText:      "After the instructions were sent, the link between {call} and {node} is still not active. It may just be a matter of time, but please check these points:",
		NoChecks:    [3]string{"Host and port are those given in the instructions email.", "The login uses the callsign {call} and the dedicated password, without spaces.", "Your node is set as DXSpider and the port is reachable from the Internet."},
		RecvSubject: "DXRegia · we have received your request", RecvTitle: "Request received",
		RecvText: "Thank you {name}, we have received the request for {call}. The sysop of node {node} will review it and reply to this address.",
		RecvNext: "If your request is about a link between nodes, you will receive the access details in a new email as soon as it is ready.", YourData: "Summary of your request",
	},
	"es": {
		Hello: "Hola {name},", AutoNote: "Mensaje automático de DXRegia. Si tienes dudas, responde a este correo o escribe al sysop del nodo {node}.",
		Footer:       "DXRegia · clúster DX para radioaficionados",
		InstrSubject: "DXRegia · instrucciones para conectar {call} al nodo {node}",
		InstrTitle:   "Tu nodo {call} ha sido habilitado",
		InstrIntro:   "Hemos preparado el acceso de tu nodo al clúster {node}. Abajo encontrarás todos los datos y los pasos para completar el enlace.",
		DataTitle:    "Tus datos de conexión", LNode: "Nodo al que conectarte", LHost: "Dirección (host)", LPort: "Puerto Telnet", LCall: "Indicativo para el login", LPass: "Contraseña dedicada", LProfile: "Perfil del enlace", VProf: "Nodo DXSpider (peering)",
		StepsTitle: "Qué hacer, paso a paso",
		Steps: [4]string{
			"En tu nodo configura el enlace hacia {host}, puerto {port}, como nodo de tipo DXSpider.",
			"Cuando se solicite, entra con el indicativo {call} y la contraseña indicada arriba.",
			"Inicia el enlace. Si tu nodo reintenta solo tras una caída, mejor aún.",
			"Comprueba en tu nodo que {node} aparece entre los enlaces activos.",
		},
		ScriptTitle: "Ejemplo de script de conexión (DXSpider)", ScriptNote: "Adáptalo a tu nodo: es solo una plantilla de partida.",
		AfterTitle: "Qué sucede después", AfterText: "En cuanto el sistema vea tu nodo conectado, recibirás un segundo correo de confirmación. Si pasado un tiempo el enlace no está activo, te escribiremos para ayudarte a averiguar el motivo.",
		SecurityNote: "Guarda esta contraseña con cuidado y no la compartas: es solo para tu nodo.",
		OKSubject:    "DXRegia · {call} está conectado: todo correcto", OKTitle: "Enlace establecido",
		OKText:    "Tu nodo {call} está conectado al clúster {node} y funciona correctamente. A partir de ahora intercambiáis spots e información en tiempo real.",
		OKNext:    "No tienes que hacer nada más. Si el enlace se cae, tu nodo puede reintentar solo; ante cualquier problema responde a este correo.",
		NoSubject: "DXRegia · el enlace de {call} aún no está activo", NoTitle: "Todavía no vemos tu nodo conectado",
		NoText:      "Tras enviar las instrucciones, el enlace entre {call} y {node} todavía no está activo. Puede ser cuestión de tiempo, pero conviene revisar estos puntos:",
		NoChecks:    [3]string{"El host y el puerto son los indicados en el correo de instrucciones.", "El login usa el indicativo {call} y la contraseña dedicada, sin espacios.", "Tu nodo está configurado como DXSpider y el puerto es accesible desde Internet."},
		RecvSubject: "DXRegia · hemos recibido tu solicitud", RecvTitle: "Solicitud recibida",
		RecvText: "Gracias {name}, hemos recibido la solicitud de {call}. El sysop del nodo {node} la revisará y te responderá a esta dirección.",
		RecvNext: "Si la solicitud es para un enlace entre nodos, recibirás los datos de acceso en un nuevo correo en cuanto estén listos.", YourData: "Resumen de tu solicitud",
	},
	"de": {
		Hello: "Hallo {name},", AutoNote: "Automatische Nachricht von DXRegia. Bei Fragen antworte auf diese E-Mail oder schreibe dem Sysop des Knotens {node}.",
		Footer:       "DXRegia · DX-Cluster für Funkamateure",
		InstrSubject: "DXRegia · Anleitung zum Verbinden von {call} mit dem Knoten {node}",
		InstrTitle:   "Dein Knoten {call} wurde freigeschaltet",
		InstrIntro:   "Wir haben den Zugang deines Knotens zum Cluster {node} vorbereitet. Unten findest du alle Daten und die Schritte, um die Verbindung herzustellen.",
		DataTitle:    "Deine Verbindungsdaten", LNode: "Knoten, mit dem du dich verbindest", LHost: "Adresse (Host)", LPort: "Telnet-Port", LCall: "Rufzeichen für den Login", LPass: "Eigenes Passwort", LProfile: "Verbindungsprofil", VProf: "DXSpider-Knoten (Peering)",
		StepsTitle: "Was zu tun ist, Schritt für Schritt",
		Steps: [4]string{
			"Richte auf deinem Knoten die Verbindung zu {host}, Port {port}, als Knoten vom Typ DXSpider ein.",
			"Melde dich auf Anfrage mit dem Rufzeichen {call} und dem oben angegebenen Passwort an.",
			"Starte die Verbindung. Wenn dein Knoten nach einem Abbruch selbst neu verbindet, umso besser.",
			"Prüfe auf deinem Knoten, ob {node} unter den aktiven Verbindungen erscheint.",
		},
		ScriptTitle: "Beispiel für ein Verbindungsskript (DXSpider)", ScriptNote: "Passe es an deinen Knoten an: es ist nur eine Vorlage.",
		AfterTitle: "Wie es weitergeht", AfterText: "Sobald das System deinen Knoten verbunden sieht, erhältst du eine zweite Bestätigungs-E-Mail. Ist die Verbindung nach einer Weile nicht aktiv, melden wir uns, um die Ursache zu finden.",
		SecurityNote: "Bewahre dieses Passwort sicher auf und gib es nicht weiter: es gilt nur für deinen Knoten.",
		OKSubject:    "DXRegia · {call} ist verbunden: alles in Ordnung", OKTitle: "Verbindung hergestellt",
		OKText:    "Dein Knoten {call} ist mit dem Cluster {node} verbunden und funktioniert einwandfrei. Ab jetzt tauscht ihr Spots und Informationen in Echtzeit aus.",
		OKNext:    "Du musst nichts weiter tun. Bricht die Verbindung ab, kann dein Knoten selbst neu verbinden; bei Problemen antworte einfach auf diese E-Mail.",
		NoSubject: "DXRegia · die Verbindung von {call} ist noch nicht aktiv", NoTitle: "Wir sehen deinen Knoten noch nicht verbunden",
		NoText:      "Nach dem Versand der Anleitung ist die Verbindung zwischen {call} und {node} noch nicht aktiv. Vielleicht ist es nur eine Frage der Zeit, prüfe aber bitte diese Punkte:",
		NoChecks:    [3]string{"Host und Port entsprechen den Angaben in der Anleitungs-E-Mail.", "Der Login verwendet das Rufzeichen {call} und das eigene Passwort, ohne Leerzeichen.", "Dein Knoten ist als DXSpider eingerichtet und der Port ist aus dem Internet erreichbar."},
		RecvSubject: "DXRegia · wir haben deine Anfrage erhalten", RecvTitle: "Anfrage erhalten",
		RecvText: "Danke {name}, wir haben die Anfrage für {call} erhalten. Der Sysop des Knotens {node} prüft sie und antwortet an diese Adresse.",
		RecvNext: "Betrifft die Anfrage eine Verbindung zwischen Knoten, erhältst du die Zugangsdaten in einer neuen E-Mail, sobald sie bereit sind.", YourData: "Zusammenfassung deiner Anfrage",
	},
	"fr": {
		Hello: "Bonjour {name},", AutoNote: "Message automatique de DXRegia. En cas de doute, réponds à cet e-mail ou écris au sysop du nœud {node}.",
		Footer:       "DXRegia · cluster DX pour radioamateurs",
		InstrSubject: "DXRegia · instructions pour connecter {call} au nœud {node}",
		InstrTitle:   "Ton nœud {call} a été activé",
		InstrIntro:   "Nous avons préparé l’accès de ton nœud au cluster {node}. Tu trouveras ci-dessous toutes les données et les étapes pour terminer la liaison.",
		DataTitle:    "Tes données de connexion", LNode: "Nœud auquel te connecter", LHost: "Adresse (hôte)", LPort: "Port Telnet", LCall: "Indicatif à utiliser à la connexion", LPass: "Mot de passe dédié", LProfile: "Profil de la liaison", VProf: "Nœud DXSpider (peering)",
		StepsTitle: "Que faire, étape par étape",
		Steps: [4]string{
			"Sur ton nœud, configure la liaison vers {host}, port {port}, comme nœud de type DXSpider.",
			"Quand c’est demandé, connecte-toi avec l’indicatif {call} et le mot de passe indiqué ci-dessus.",
			"Lance la liaison. Si ton nœud réessaie seul après une coupure, c’est encore mieux.",
			"Vérifie sur ton nœud que {node} apparaît parmi les liaisons actives.",
		},
		ScriptTitle: "Exemple de script de connexion (DXSpider)", ScriptNote: "Adapte-le à ton nœud : ce n’est qu’un modèle de départ.",
		AfterTitle: "Ce qui se passe ensuite", AfterText: "Dès que le système voit ton nœud connecté, tu reçois un second e-mail de confirmation. Si la liaison n’est toujours pas active après un moment, nous t’écrivons pour t’aider à comprendre pourquoi.",
		SecurityNote: "Conserve ce mot de passe précieusement et ne le partage pas : il ne sert qu’à ton nœud.",
		OKSubject:    "DXRegia · {call} est connecté : tout va bien", OKTitle: "Liaison établie",
		OKText:    "Ton nœud {call} est connecté au cluster {node} et fonctionne correctement. À partir de maintenant, vous échangez spots et informations en temps réel.",
		OKNext:    "Tu n’as rien d’autre à faire. Si la liaison tombe, ton nœud peut réessayer seul ; en cas de problème, réponds simplement à cet e-mail.",
		NoSubject: "DXRegia · la liaison de {call} n’est pas encore active", NoTitle: "Nous ne voyons pas encore ton nœud connecté",
		NoText:      "Après l’envoi des instructions, la liaison entre {call} et {node} n’est pas encore active. Ce n’est peut-être qu’une question de temps, mais vérifie ces points :",
		NoChecks:    [3]string{"L’hôte et le port sont ceux indiqués dans l’e-mail d’instructions.", "La connexion utilise l’indicatif {call} et le mot de passe dédié, sans espaces.", "Ton nœud est configuré comme DXSpider et le port est joignable depuis Internet."},
		RecvSubject: "DXRegia · nous avons reçu ta demande", RecvTitle: "Demande reçue",
		RecvText: "Merci {name}, nous avons reçu la demande pour {call}. Le sysop du nœud {node} l’examine et te répond à cette adresse.",
		RecvNext: "Si la demande concerne une liaison entre nœuds, tu recevras les données d’accès dans un nouvel e-mail dès qu’elles seront prêtes.", YourData: "Récapitulatif de ta demande",
	},
	"pt": {
		Hello: "Olá {name},", AutoNote: "Mensagem automática do DXRegia. Em caso de dúvida, responde a este e-mail ou escreve ao sysop do nó {node}.",
		Footer:       "DXRegia · cluster DX para radioamadores",
		InstrSubject: "DXRegia · instruções para ligar {call} ao nó {node}",
		InstrTitle:   "O teu nó {call} foi ativado",
		InstrIntro:   "Preparámos o acesso do teu nó ao cluster {node}. Abaixo encontras todos os dados e os passos para concluir a ligação.",
		DataTitle:    "Os teus dados de ligação", LNode: "Nó ao qual te ligas", LHost: "Endereço (host)", LPort: "Porta Telnet", LCall: "Indicativo a usar no login", LPass: "Palavra-passe dedicada", LProfile: "Perfil da ligação", VProf: "Nó DXSpider (peering)",
		StepsTitle: "O que fazer, passo a passo",
		Steps: [4]string{
			"No teu nó, configura a ligação a {host}, porta {port}, como nó do tipo DXSpider.",
			"Quando pedido, entra com o indicativo {call} e a palavra-passe indicada acima.",
			"Inicia a ligação. Se o teu nó tentar de novo sozinho após uma queda, ainda melhor.",
			"Verifica no teu nó que {node} aparece entre as ligações ativas.",
		},
		ScriptTitle: "Exemplo de script de ligação (DXSpider)", ScriptNote: "Adapta-o ao teu nó: é apenas um modelo de partida.",
		AfterTitle: "O que acontece a seguir", AfterText: "Assim que o sistema vir o teu nó ligado, recebes um segundo e-mail de confirmação. Se passado algum tempo a ligação não estiver ativa, escrevemos-te para ajudar a perceber o motivo.",
		SecurityNote: "Guarda esta palavra-passe com cuidado e não a partilhes: serve só para o teu nó.",
		OKSubject:    "DXRegia · {call} está ligado: tudo certo", OKTitle: "Ligação estabelecida",
		OKText:    "O teu nó {call} está ligado ao cluster {node} e funciona corretamente. A partir de agora trocam spots e informações em tempo real.",
		OKNext:    "Não tens mais nada a fazer. Se a ligação cair, o teu nó pode tentar de novo sozinho; para qualquer problema responde a este e-mail.",
		NoSubject: "DXRegia · a ligação de {call} ainda não está ativa", NoTitle: "Ainda não vemos o teu nó ligado",
		NoText:      "Depois do envio das instruções, a ligação entre {call} e {node} ainda não está ativa. Pode ser só uma questão de tempo, mas convém verificar estes pontos:",
		NoChecks:    [3]string{"O host e a porta são os indicados no e-mail de instruções.", "O login usa o indicativo {call} e a palavra-passe dedicada, sem espaços.", "O teu nó está configurado como DXSpider e a porta é acessível a partir da Internet."},
		RecvSubject: "DXRegia · recebemos o teu pedido", RecvTitle: "Pedido recebido",
		RecvText: "Obrigado {name}, recebemos o pedido para {call}. O sysop do nó {node} analisa-o e responde para este endereço.",
		RecvNext: "Se o pedido for para uma ligação entre nós, receberás os dados de acesso num novo e-mail assim que estiverem prontos.", YourData: "Resumo do teu pedido",
	},
}

func textsFor(lang string) mailStrings {
	if t, ok := mailTexts[normalizeLang(lang)]; ok {
		return t
	}
	// Lingua sconosciuta o non supportata (ad esempio il cinese): si risponde sempre in inglese.
	return mailTexts["en"]
}

func fill(text string, values map[string]string) string {
	for key, value := range values {
		text = strings.ReplaceAll(text, "{"+key+"}", value)
	}
	return text
}
