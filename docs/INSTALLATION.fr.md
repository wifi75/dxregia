# Guide d'installation et de configuration

[English](INSTALLATION.md) · [Italiano](INSTALLATION.it.md) · [Español](INSTALLATION.es.md) · [Deutsch](INSTALLATION.de.md) · **Français** · [Português](INSTALLATION.pt.md)

[← Retour à la page du projet](../README.fr.md)

## 1. Ce qu'il faut

- Un serveur ou un PC Linux avec systemd : Debian ou Ubuntu (testé sur Ubuntu 24.04, amd64 et arm64). Une machine virtuelle ou un conteneur LXC non privilégié convient aussi.
- Pour l'installation depuis les sources : Git, Go 1.22+, Node.js 22.12+ avec npm, curl et sudo (ou root). L'alternative est Docker, qui ne demande que Docker et curl.
- Deux ports TCP : un pour le panneau web (8080 par défaut) et un pour Telnet (7300 par défaut). Si d'autres nœuds ou utilisateurs doivent vous joindre depuis Internet, redirigez le port Telnet sur votre routeur.

## 2. Installation

**Paquet précompilé.**

```sh
curl -fL -o install-release.sh https://raw.githubusercontent.com/wifi75/dxregia/v0.15.1/dxregia/deploy/install-release.sh
sh install-release.sh v0.15.1
```

Le script vérifie la somme SHA-256, installe les dépendances et démarre le service. Il affiche l’adresse du panneau et un mot de passe temporaire pour la première connexion.

**Installation native depuis les sources.** Lancez ces commandes sur le serveur :

```sh
git clone https://github.com/wifi75/dxregia.git
cd dxregia
sh dxregia/deploy/install-native.sh
```

Le script compile le panneau et le backend, installe les dépendances Perl, crée un utilisateur dédié et démarre le service systemd `dxregia`. Le programme est dans `/opt/dxregia` et les données dans `/var/lib/dxregia`. À la fin il affiche l'adresse du panneau et, au premier lancement, un mot de passe temporaire. Le relancer démarre le service déjà installé et conserve les données. Il installe aussi la commande `dxregia-update`.

Pour choisir les ports, définissez-les d'abord :

```sh
DXREGIA_WEB_PORT=8787 DXREGIA_TELNET_PORT=7373 sh dxregia/deploy/install-native.sh
```

**Docker.** Avec Docker et curl déjà installés :

```sh
sh dxregia/deploy/install-local.sh
```

Il construit l'image et démarre un conteneur avec un volume persistant, puis affiche l'adresse et le mot de passe temporaire. Le relancer conserve le conteneur et le volume. Les deux mêmes variables de ports s'appliquent.

## 3. Ouvrir le panneau

Ouvrez une adresse LAN affichée par le programme d’installation natif ou utilisez `http://127.0.0.1:8080/admin` sur le serveur. Une nouvelle installation native écoute sur toutes les interfaces IPv4. Avec Docker, utilisez un tunnel SSH car le port du panneau est publié sur la boucle locale de l’hôte.

```sh
ssh -L 8080:127.0.0.1:8080 USER@SERVER
```

## 4. Première connexion

1. Ouvrez `/admin`. Connectez-vous avec l'utilisateur `sysop` et le mot de passe temporaire affiché par l'installateur.
2. DXRegia demande un nouveau mot de passe (au moins 12 caractères). Le temporaire cesse de fonctionner.
3. Le 2FA est facultatif et désactivé par défaut. Activez-le dans **Gestion → Mot de passe**, confirmez le mot de passe web, scannez le QR et conservez les codes de récupération.

Si nécessaire, lisez le mot de passe temporaire sur le serveur avant son premier changement :

```sh
sudo cat /var/lib/dxregia/sysop-initial
```

Ensuite chaque connexion demande l'utilisateur, le mot de passe et le code à 6 chiffres. L'utilisateur est votre indicatif de sysop dès que la configuration est terminée (avant, c'est `sysop`) ; majuscules et minuscules n'ont pas d'importance.

## 5. Première configuration (assistant)

Après la connexion, ouvrez Nœud → Configuration (Nodo → Configurazione) et remplissez : indicatif du nœud avec SSID (par exemple `IU3WJQ-7`), indicatif du sysop, locator, port Telnet et, si vous le souhaitez, nom, e-mail, QTH, coordonnées et un nom d'hôte public. Le nom d'hôte public n'est qu'une référence : il ne modifie ni votre DNS ni votre routeur. Si vous utilisez déjà DXSpider, vous pouvez importer `DXVars.pm` et `Listeners.pm` depuis la même page. Appliquez : DXRegia écrit les fichiers dont Mojo a besoin et démarre le nœud, et le panneau confirme quand Telnet répond.

## 6. Configurer le reste

- **Telnet et pare-feu.** Vérifiez que le port Telnet est ouvert dans le pare-feu. Pour être joignable depuis Internet, redirigez-le sur le routeur vers le serveur. Les options sont dans Nœud → Options Telnet (Opzioni Telnet).
- **Serveur de messagerie.** Gestion → Serveur de messagerie (Gestione → Server di posta). Choisissez un préréglage (Gmail, Aruba, Office 365) ou saisissez serveur et port, chiffrement, adresse d'expéditeur, utilisateur et mot de passe, enregistrez puis appuyez sur Envoyer un test (Invia prova). Le mot de passe reste sur le serveur dans un fichier privé. Avec Gmail il faut un mot de passe d'application. La messagerie sert aux avis qui vous sont destinés et aux instructions envoyées aux partenaires.
- **Nœuds partenaires.** Liaisons → Nœuds partenaires → Ajouter un nœud (Collegamenti → Nodi partner → Aggiungi nodo) : indicatif, hôte, port et, si l'autre nœud en demande un, le mot de passe distant. Appuyez sur Enregistrer et connecter (Salva e connetti). Cochez la case de nouvelle tentative pour se reconnecter seul après une coupure.
- **Sauvegardes.** Voir la section 9.
- **Langue, thème et son.** Les icônes en haut à droite des pages publiques.

### Demandes de partenaires et liaisons entrantes

- **Formulaire public.** Activez la configuration automatique dans **Liaisons → Demandes reçues**. Une fois SMTP ainsi que le nom d’hôte et le port Telnet publics configurés, DXRegia crée les fiches des nouveaux partenaires admissibles. Lorsque DXSpider autorise l’accès entrant, DXRegia envoie les paramètres de connexion et un mot de passe dédié. La limite par défaut est de dix configurations sur 24 heures. Avec l’automatisation désactivée, ouvrez la demande et préparez la fiche manuellement.
- **Accès CC Cluster sans mot de passe.** Dans **Liaisons → Nœuds partenaires**, choisissez l’authentification entrante sans mot de passe, confirmez l’adresse IP publique ou le nom d’hôte source, puis enregistrez et envoyez les instructions. DXRegia propose l’hôte enregistré et actualise ses adresses DNS. Le partenaire doit se connecter depuis une de ces adresses. Vous pouvez aussi enregistrer sans envoyer de courriel.
- **Confirmation de la liaison.** Après l’envoi des instructions, DXRegia attend jusqu’à sept jours une session active ouverte par le partenaire et envoie un second courriel dès qu’il la détecte. Une session sortante ne déclenche pas cette confirmation. Consultez l’acceptation SMTP et le contenu dans l’**Historique des courriels partenaires**. Si une liaison sortante occupe déjà la session, le test entrant ouvre une fenêtre de cinq minutes. Le test automatique est facultatif et répète cette fenêtre toutes les 30 minutes tant qu’il n’y a pas de session entrante.

## 7. Vérifier que cela fonctionne

- Ouvrez `http://SERVEUR:8080/spots` : la page des spots en direct.
- Nœud → État et sessions (Stato e sessioni) : les quatre tuiles doivent être vertes.
- Depuis un PC, testez Telnet, puis connectez-vous avec un client Telnet ou un logger et tapez `show/dx` :

```sh
python3 dxregia/deploy/check_telnet.py HOST 7300
```

Telnet transmet en clair : n'utilisez les commandes d'administration que sur un réseau de confiance.

## 8. Mettre à jour

Sur une installation native, lancez depuis n'importe quel répertoire :

```sh
dxregia-update
```

Il télécharge la dernière version, la compile, conserve la copie précédente et revient seul en arrière si le panneau ou Telnet cessent de répondre. Données, identifiants, ports et personnalisations sont conservés. `dxregia-update --check` affiche seulement la version installée et la version disponible. Le service redémarre : les liaisons tombent environ une minute et se rétablissent seules, et vous vous reconnectez au panneau. Les installations faites avant cette commande peuvent récupérer `dxregia/deploy/update-native.sh` et l'exécuter avec `sh update-native.sh`. Avec Docker, utilisez `sh update-docker.sh`.

Vous pouvez aussi mettre à jour depuis le panneau d'administration : **Gestion > Mises à jour**. La page affiche la version installée et la version disponible, **Vérifier maintenant** interroge GitHub et **Mettre à jour maintenant** lance la même mise à jour après confirmation. La progression s'affiche en direct et la page se reconnecte après le redémarrage. Telnet et les liens avec les partenaires sont coupés pendant environ une minute. Si vous avez installé avant cette fonction, exécutez une fois `dxregia-update` depuis le terminal pour l'activer. Lorsque seules les pages du panneau ou la documentation changent, les fichiers sont remplacés sans redémarrage et Telnet reste connecté.

## 9. Sauvegarde et restauration

L'outil hors ligne demande Python 3.10+. Arrêtez d'abord le service pour que la copie soit cohérente :

```sh
sudo systemctl stop dxregia
python3 dxregia/deploy/backup_node.py backup /var/lib/dxregia/mojo /path/to/node.zip
python3 dxregia/deploy/backup_node.py verify /path/to/node.zip
sudo systemctl start dxregia
```

Il enregistre le moteur et les données du nœud, les utilisateurs, filtres, partenaires, réglages et données d'accès : traitez l'archive comme confidentielle. Pour vérifier une archive depuis le panneau, utilisez Gestion → Sauvegarde (Gestione → Backup). Pour restaurer, utilisez un répertoire neuf et vide :

```sh
python3 dxregia/deploy/backup_node.py restore /path/to/node.zip /path/to/restored-node
```

## 10. Reprendre un nœud Mojo existant

Sur un nœud arrêté avec une base d'utilisateurs SQLite, faites une copie séparée et démarrez DXRegia dessus. Ne remplacez le nœud en service qu'après avoir vérifié Telnet, les filtres et les liaisons sur la copie :

```sh
python3 -B dxregia/deploy/migrate_mojo.py /path/to/mojo /path/to/mojo-dxregia
```

## 11. Dépannage

- **Je n'arrive pas à me connecter.** L'utilisateur est votre indicatif de sysop (ou `sysop` avant la configuration), pas l'indicatif du nœud. Après un essai raté, attendez une seconde.
- **Le code à deux facteurs est refusé.** Les horloges doivent concorder : activez l'heure automatique sur le téléphone et vérifiez le serveur avec `timedatectl status` (pour un conteneur LXC, vérifiez l'hôte). Si un code est signalé comme déjà utilisé, passez au suivant. En dernier recours, utilisez un code de récupération.
- **Le panneau ne s'ouvre pas depuis un autre PC.** Suivez la section 3 et vérifiez le pare-feu pour le port web.
- **Telnet n'est pas joignable.** Vérifiez que le port écoute (`ss -ltn | grep 7300`), ouvrez-le dans le pare-feu et, depuis Internet, redirigez-le sur le routeur.
- **Un partenaire ne se connecte pas.** Vérifiez hôte, port et mot de passe, et quel côté lance la liaison : dans la fiche du partenaire, « se connecte à votre nœud » exige que votre port Telnet soit joignable de l'extérieur.
- **Journaux.** Sur le serveur :

```sh
sudo journalctl -u dxregia -f
```

Le journal du moteur est aussi dans Nœud → Journal DXSpider (Log DXSpider).

## Pour aller plus loin

- [Référence de l'installateur Linux](../dxregia/deploy/README.md) (en italien)
- [Opérations du sysop](OPERATIONS.md) (en italien)
- [Journal des modifications](../CHANGELOG.md) (en anglais)

## DXRegia 0.7.0

La page Propagation affiche les indices solaires de HamQSL dans des cartes carrées et colorées, les conditions HF de jour et de nuit, l’activité VHF et la carte MUF de KC2G. L’inscription publique explique comment configurer Telnet dans sept logiciels de carnet de trafic. Le tableau de bord sysop affiche jusqu’à 40 radioamateurs connectés par page sur les grands écrans, avec leurs drapeaux DXCC.

RBN est désactivé par défaut. Ouvrez **Reverse Beacon Network** dans le menu sysop et sélectionnez **Activer RBN** pour démarrer la connexion sortante CW/RTTY. L’onglet public **RBN** affiche les spots reçus. Dans Telnet ou un logiciel de log, utilisez `SET/SKIMMER CW RTTY` pour les recevoir et `UNSET/SKIMMER` pour arrêter.

[DXRegia 0.7.0](https://github.com/wifi75/dxregia/releases/tag/v0.7.0)
