<h1><picture><source media="(prefers-color-scheme: dark)" srcset="dxregia/web/public/brand/dxregia-logo-dark.svg"><img alt="DXRegia" src="dxregia/web/public/brand/dxregia-logo.svg" height="48"></picture></h1>

<p><a href="https://github.com/wifi75/dxregia/releases"><img src="https://img.shields.io/github/v/release/wifi75/dxregia?display_name=tag&amp;color=147d72" alt="Version"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/actions/workflows/linux-smoke.yml"><img src="https://img.shields.io/github/actions/workflow/status/wifi75/dxregia/linux-smoke.yml?branch=main&amp;label=Test%20Linux&amp;logo=githubactions&amp;logoColor=white" alt="Test Linux"></a>&thinsp;<a href="LICENSE"><img src="https://img.shields.io/badge/License-Artistic%202.0-0b6e99" alt="License: Artistic 2.0"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Status-in%20development-e0a23b" alt="Status: in development"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/commits/main"><img src="https://img.shields.io/github/last-commit/wifi75/dxregia?color=475569" alt="Last commit"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Maintained-yes%202026-2ea44f" alt="Maintained: yes 2026"></a><br>
<a href="dxregia/backend/README.md"><img src="https://img.shields.io/badge/Backend-Go%201.22-00ADD8?logo=go&amp;logoColor=white" alt="Backend: Go 1.22"></a>&thinsp;<a href="dxregia/web/README.md"><img src="https://img.shields.io/badge/Panel-Svelte%205-FF3E00?logo=svelte&amp;logoColor=white" alt="Panel: Svelte 5"></a>&thinsp;<a href="https://www.typescriptlang.org/"><img src="https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&amp;logoColor=white" alt="TypeScript 5.9"></a>&thinsp;<a href="https://vite.dev/"><img src="https://img.shields.io/badge/Vite-7-646CFF?logo=vite&amp;logoColor=white" alt="Vite 7"></a>&thinsp;<a href="https://github.com/EA3CV/dxspider/tree/mojo"><img src="https://img.shields.io/badge/Engine-DXSpider%20Mojo-147d72" alt="Engine: DXSpider Mojo"></a>&thinsp;<a href="https://www.perl.org/"><img src="https://img.shields.io/badge/Perl-5.38-39457E?logo=perl&amp;logoColor=white" alt="Perl 5.38"></a>&thinsp;<a href="docs/INTERFACES.md"><img src="https://img.shields.io/badge/Console-Telnet-334155" alt="Console: Telnet"></a><br>
<a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Platform-Linux-334155?logo=linux&amp;logoColor=white" alt="Platform: Linux"></a>&thinsp;<a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Tested-Ubuntu%2024.04-E95420?logo=ubuntu&amp;logoColor=white" alt="Tested on Ubuntu 24.04, amd64 and arm64"></a>&thinsp;<a href="dxregia/deploy/README.md"><img src="https://img.shields.io/badge/Docker-supported-2496ED?logo=docker&amp;logoColor=white" alt="Docker supported"></a>&thinsp;<a href="dxregia/web/src/translations.ts"><img src="https://img.shields.io/badge/Languages-IT%20%C2%B7%20EN%20%C2%B7%20ES%20%C2%B7%20DE%20%C2%B7%20FR%20%C2%B7%20PT-0f766e" alt="Languages: IT EN ES DE FR PT"></a><br>
<a href="https://semver.org/"><img src="https://img.shields.io/badge/SemVer-2.0.0-3F4551?logo=semver&amp;logoColor=white" alt="SemVer 2.0.0"></a>&thinsp;<a href="CHANGELOG.md"><img src="https://img.shields.io/badge/Changelog-Keep%20a%20Changelog-E05735?logo=keepachangelog&amp;logoColor=white" alt="Keep a Changelog"></a></p>

[English](README.md) · [Italiano](README.it.md) · [Español](README.es.md) · [Deutsch](README.de.md) · **Français** · [Português](README.pt.md)

**Un nœud DX Cluster pour Linux avec panneau web, basé sur DXSpider Mojo.**

DXRegia facilite l'installation, la gestion et le partage d'un nœud DX Cluster. Il conserve DXSpider Mojo comme moteur réseau (accès Telnet classique, commandes et liaisons entre nœuds) et ajoute un panneau web moderne et une page publique avec les spots en direct.

## Portail public et administration

Captures de DXRegia avec des données de démonstration. Cliquez sur une image pour l’agrandir.

### Spots en direct · portail public

Sur `/spots`, les radioamateurs suivent les spots DX en temps réel, filtrent par bande ou mode et identifient les entités grâce aux drapeaux DXCC.

[![Spots en direct · portail public](docs/screenshots/spots.png)](docs/screenshots/spots.png)

### État du nœud et sessions · panneau sysop

Sur `/admin`, les sysops consultent l’état de Telnet et du cluster, le sens des liaisons, les sessions actives et les graphiques des ressources du serveur.

[![État du nœud et sessions · panneau sysop](docs/screenshots/admin.png)](docs/screenshots/admin.png)

## Ce qu'il fait

- **Page de spots en direct** — publique et en temps réel. Filtres par bande, mode et indicatif (plusieurs à la fois), drapeaux DXCC, surbrillance jaune et un court signal sonore pour chaque nouveau spot. Fonctionne sur téléphone, tablette et ordinateur, en thème clair ou sombre.
- **Des spots plus intelligents** — le mode est déduit de la fréquence quand le commentaire ne l'indique pas, et les opérations DX annoncées (NG3K ADXO) apparaissent à côté de l'indicatif.
- **Six langues** — italien, anglais, espagnol, allemand, français et portugais ; toute autre langue bascule sur l'anglais.
- **Panneau d'administration sur `/admin`** — configuration guidée, connexion avec utilisateur, mot de passe et code à deux facteurs, tableau de bord en direct avec voyants d'état et centre d'analyse.
- **Réglages du nœud dans le panneau** — message du jour, commandes planifiées, script de démarrage, listes de blocage (mots, indicatifs, adresses IP), sauts, filtres et options Telnet, chacun avec une infobulle. Le panneau parle les six langues et montre qui s'est connecté au nœud.
- **Inscription pour envoyer des spots** — les visiteurs s'inscrivent sur le portail (indicatif, e-mail, mot de passe) et confirment par e-mail ; seuls les inscrits peuvent envoyer des spots en Telnet. Facultative, à activer dans le panneau.
- **Installable comme application** — le portail s'installe sur l'écran d'accueil du smartphone et s'ouvre même sans connexion.
- **Nœuds partenaires** — ajoutez, connectez et reconnectez automatiquement ; voyez qui a lancé chaque liaison ; supprimez un nœud pour le déconnecter et révoquer son accès.
- **Demandes de liaison en automatique** — d'autres sysops la demandent depuis votre page Contact ; DXRegia peut créer leur compte, leur envoyer vos données de connexion par e-mail et confirmer quand ils se connectent (interrupteur et limite quotidienne dans le panneau). Pour un nœud CC Cluster, l’accès entrant sans mot de passe peut être limité à une IP publique ou à un nom DNS. Un second courriel confirme la session entrante active.
- **E-mails HTML** — instructions et confirmations en six langues via votre propre serveur SMTP, configuré et testé depuis le panneau.
- **Accès classique conservé** — Telnet et commandes DXSpider pour les loggers et les utilisateurs, compatibles avec les nœuds existants.
- **Sûr à exploiter** — sauvegardes hors ligne avec vérification, mises à jour en une commande avec retour arrière automatique, statistiques de visites sans cookies ni adresses conservées.
- **Facile à installer** — un seul script sur Debian ou Ubuntu (systemd, amd64 ou arm64), ou Docker.

- Enregistrez et comparez les vues, consultez le catalogue DXCC et l’historique des indicatifs et configurez des alertes actives lorsque le portail est ouvert.
- **Propagation et QSL** — consultez les indices solaires et les conditions HF jour/nuit et VHF de HamQSL, la carte MUF de KC2G, les spots par région et bande et les bulletins WWV/WCY. Le tableau des spots propose une colonne QSL pour LoTW et eQSL AG. La page d’inscription explique la connexion Telnet avec sept logiciels de log documentés. Le sysop modifie les règles dans Sauts et filtres.

## Démarrage rapide

Sous Debian ou Ubuntu avec systemd, installez le paquet précompilé pour amd64 ou arm64. Il faut curl et sudo ou une session root.

```sh
curl -fL -o install-release.sh https://raw.githubusercontent.com/wifi75/dxregia/v0.15.5/dxregia/deploy/install-release.sh
sh install-release.sh v0.15.5
```

1. Ouvrez l’adresse LAN affichée par le programme d’installation, par exemple `http://SERVEUR:8080/admin`. Une nouvelle installation native écoute sur toutes les interfaces IPv4.
2. Connectez-vous comme `sysop` avec le mot de passe temporaire affiché et choisissez un nouveau mot de passe. Le 2FA est facultatif et désactivé par défaut. Activez-le dans **Gestion → Mot de passe**, confirmez le mot de passe web, scannez le QR et conservez les codes de récupération.
3. Remplissez l'assistant de première configuration : indicatif du nœud, indicatif du sysop, locator et port Telnet. Ensuite, la page des spots est sur `/spots` et Telnet répond sur le port 7300.

Le guide décrit aussi Docker et l’installation depuis les sources.

Le guide pas à pas couvre les ports, la messagerie, les nœuds partenaires, les mises à jour, les sauvegardes et le dépannage : [**Guide d'installation et de configuration**](docs/INSTALLATION.fr.md).

## Autre documentation

_Les documents techniques ci-dessous sont rédigés en italien._

- [docs/OPERATIONS.md](docs/OPERATIONS.md) — Opérations du sysop : nœuds partenaires, journaux, sauvegardes
- [docs/INTERFACES.md](docs/INTERFACES.md) — Page publique, panneau d'administration, Telnet et marque
- [dxregia/deploy/README.md](dxregia/deploy/README.md) — Référence de l'installateur Linux
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — Architecture
- [TASK.md](TASK.md) — Plan de travail
- [CHANGELOG.md](CHANGELOG.md) — Journal des modifications (en anglais)
- [Guides de démarrage et protocoles (en italien)](docs/QUICKSTART.md)
- [Exemples de peering en six langues](docs/PEERING.md)

**Organisation du dépôt :** `dxregia/` contient le code propre à DXRegia (backend, panneau, installateurs), `engine/` le moteur DXSpider Mojo avec sa structure d'origine et `docs/` la documentation.

## État

DXRegia est en développement actif (version 0.15.5). L'installation, la première connexion, Telnet, la sauvegarde et le redémarrage sont testés sur Ubuntu 24.04 (amd64 et arm64). Restent à faire : un test sur Proxmox LXC, la remise du courrier avec d'autres fournisseurs et un paquet de version signé. Les rapports et idées sont les bienvenus dans les [GitHub Issues](https://github.com/wifi75/dxregia/issues).

## Origine et licence

DXRegia est un projet indépendant basé sur DXSpider Mojo de Dirk Koopman G1TLH et de ses contributeurs (le [miroir EA3CV](https://github.com/EA3CV/dxspider/tree/mojo)). Le dépôt conserve les fichiers d'origine [COPYING](COPYING) et [LICENSE](LICENSE) (Artistic License 2.0). Les opérations DX annoncées viennent de [NG3K ADXO](https://www.ng3k.com/Misc/adxo.html) ; les drapeaux de [flag-icons](https://github.com/lipis/flag-icons) (MIT).

[Licences et avis de tiers](docs/LICENSING.md)
