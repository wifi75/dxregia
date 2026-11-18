package main

// Email di conferma della registrazione, nella lingua scelta da chi si registra (inglese per le altre).

type registerTexts struct {
	Subject, Title, Intro, Button, Expiry, Ignore string
}

var registerMailTexts = map[string]registerTexts{
	"it": {
		Subject: "Conferma la registrazione su {node}", Title: "Conferma la tua registrazione",
		Intro:  "Hai chiesto di registrare l’indicativo {call} per inviare spot sul cluster. Conferma che l’email è tua cliccando il pulsante: da quel momento potrai collegarti via Telnet con l’indicativo e la password scelti.",
		Button: "Conferma la registrazione", Expiry: "Il link vale 24 ore e si usa una sola volta.",
		Ignore: "Se non sei stato tu, ignora questo messaggio: senza la conferma non viene registrato nulla.",
	},
	"en": {
		Subject: "Confirm your registration on {node}", Title: "Confirm your registration",
		Intro:  "You asked to register the callsign {call} so you can send spots on the cluster. Confirm the email is yours by clicking the button: from then on you can connect by Telnet with that callsign and the password you chose.",
		Button: "Confirm registration", Expiry: "The link is valid for 24 hours and works once.",
		Ignore: "If it was not you, ignore this message: nothing is registered without the confirmation.",
	},
	"es": {
		Subject: "Confirma tu registro en {node}", Title: "Confirma tu registro",
		Intro:  "Has pedido registrar el indicativo {call} para enviar spots en el clúster. Confirma que el correo es tuyo pulsando el botón: desde ese momento podrás conectarte por Telnet con ese indicativo y la contraseña elegida.",
		Button: "Confirmar el registro", Expiry: "El enlace vale 24 horas y se usa una sola vez.",
		Ignore: "Si no has sido tú, ignora este mensaje: sin la confirmación no se registra nada.",
	},
	"de": {
		Subject: "Bestätige deine Registrierung auf {node}", Title: "Bestätige deine Registrierung",
		Intro:  "Du hast darum gebeten, das Rufzeichen {call} zu registrieren, um Spots im Cluster zu senden. Bestätige mit einem Klick auf den Button, dass die E-Mail dir gehört: Danach kannst du dich per Telnet mit diesem Rufzeichen und dem gewählten Passwort verbinden.",
		Button: "Registrierung bestätigen", Expiry: "Der Link gilt 24 Stunden und nur einmal.",
		Ignore: "Wenn du das nicht warst, ignoriere diese Nachricht: Ohne Bestätigung wird nichts registriert.",
	},
	"fr": {
		Subject: "Confirmez votre inscription sur {node}", Title: "Confirmez votre inscription",
		Intro:  "Vous avez demandé d’enregistrer l’indicatif {call} pour envoyer des spots sur le cluster. Confirmez que l’adresse e-mail est la vôtre en cliquant sur le bouton : vous pourrez ensuite vous connecter en Telnet avec cet indicatif et le mot de passe choisi.",
		Button: "Confirmer l’inscription", Expiry: "Le lien est valable 24 heures et ne sert qu’une fois.",
		Ignore: "Si ce n’était pas vous, ignorez ce message : sans confirmation, rien n’est enregistré.",
	},
	"pt": {
		Subject: "Confirme o seu registo em {node}", Title: "Confirme o seu registo",
		Intro:  "Pediu para registar o indicativo {call} para enviar spots no cluster. Confirme que o email é seu clicando no botão: a partir daí poderá ligar-se por Telnet com esse indicativo e a palavra-passe escolhida.",
		Button: "Confirmar o registo", Expiry: "A ligação vale 24 horas e só pode ser usada uma vez.",
		Ignore: "Se não foi você, ignore esta mensagem: sem a confirmação nada é registado.",
	},
}

func mailRegistrationConfirm(lang, call, node, link string) mailContent {
	texts, ok := registerMailTexts[normalizeLang(lang)]
	if !ok {
		texts = registerMailTexts["en"]
	}
	s := textsFor(lang)
	v := mailValues(call, node, call, "", 0)
	return renderMail(fill(texts.Subject, v), mailBlocks{
		Preheader: fill(texts.Title, v), Title: texts.Title, Hello: fill(s.Hello, v), Intro: fill(texts.Intro, v),
		ButtonLabel: texts.Button, ButtonURL: link, Notice: texts.Expiry + " " + texts.Ignore,
		Tone: "info", Lang: lang, Footer: s.Footer, Auto: fill(s.AutoNote, v),
	})
}
