# Guide d'installation et de configuration

[English](INSTALLATION.md) · [Italiano](INSTALLATION.it.md) · [Español](INSTALLATION.es.md) · [Deutsch](INSTALLATION.de.md) · **Français** · [Português](INSTALLATION.pt.md)

[← Retour à la page du projet](../README.fr.md)

> Le panneau d'administration est pour l'instant en italien. Quand ce guide nomme un menu ou un bouton, le libellé affiché à l'écran est donné entre parenthèses.

## 1. Ce qu'il faut

- Un serveur ou un PC Linux avec systemd : Debian ou Ubuntu (testé sur Ubuntu 24.04, amd64 et arm64). Une machine virtuelle ou un conteneur LXC non privilégié convient aussi.
- Pour l'installation depuis les sources : Git, Go 1.22+, Node.js 22.12+ avec npm, curl et sudo (ou root). L'alternative est Docker, qui ne demande que Docker et curl.
- Deux ports TCP : un pour le panneau web (8080 par défaut) et un pour Telnet (7300 par défaut). Si d'autres nœuds ou utilisateurs doivent vous joindre depuis Internet, redirigez le port Telnet sur votre routeur.

## 2. Installation

**Installation native depuis les sources (recommandée).** Lancez ces commandes sur le serveur :

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
3. Activez la connexion à deux facteurs : scannez le code QR avec une application d'authentification (ou saisissez la clé à la main) et entrez le code à 6 chiffres. **Conservez les codes de récupération** en lieu sûr : ils permettent d'associer une nouvelle application si vous perdez le téléphone.

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
- **Laisser les autres demander une liaison.** Votre page Contact propose la demande « Relier mon nœud ». Dans Liaisons → Demandes reçues (Richieste ricevute) un interrupteur décide : activé, DXRegia crée le partenaire, lui envoie vos données de connexion par e-mail dans sa langue et écrit de nouveau quand il se connecte (limite par défaut 10 par jour) ; désactivé, vous approuvez chaque demande à la main avec « Préparer la fiche partenaire ».
- **Sauvegardes.** Voir la section 9.
- **Langue, thème et son.** Les icônes en haut à droite des pages publiques.

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
