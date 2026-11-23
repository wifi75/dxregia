# Changelog

All notable changes to DXRegia (the `dxregia/` folder) are recorded here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and versions follow [Semantic Versioning](https://semver.org/). The upstream engine keeps its own history in `Changes`.

## [Unreleased]

## [0.15.8] - 2026-11-23

### Changed

- Dashboard: the list of connected amateurs shows eight rows per page instead of four, so a busy day fits on one page.

## [0.15.7] - 2026-11-23

### Changed

- Dashboard: the six status cards sit on two rows of three (two per row below 900 px, one on a phone) instead of one row, and their text wraps instead of being cut off with an ellipsis, so the latest events and the connection records can be read in full.

## [0.15.6] - 2026-11-23

### Added

- Dashboard card "Record collegati insieme": the highest number of users connected at the same time over Telnet and over the DXSpider web client, with the date it was reached. The records never reset (no daily or monthly window) and are stored in `local_data/dxregia-records.json`, so they survive restarts and are included in backups. The count comes from the presence snapshot every 5 seconds, with the same rule as the list of connected amateurs. New sysop-only endpoint `GET /api/records`.

## [0.15.5] - 2026-11-23

### Changed

- Updates page: a commit without a version name that changes only documentation (`docs/`, `*.md`) is no longer offered as an update, so the page no longer says "Aggiornamento disponibile" with a bare commit id after a docs-only push to `main`; those changes arrive with the next released version. A released version, or any change to the panel or the service, is offered as before.
- The "Disponibile" line is shown only when there is something to install. It shows the version name (for example `v0.15.5`) with the commit id in brackets; a commit with code that is not yet released reads "Versione non ancora rilasciata".

## [0.15.4] - 2026-11-23

### Changed

- Dashboard, "Radioamatori collegati al tuo nodo": the country name now sits under the callsign inside the same box (long names are shortened with an ellipsis, the full name is in the tooltip). The country used to appear only when hovering the flag.

## [0.15.3] - 2026-11-23

### Fixed

### Changed

- Dashboard, "Radioamatori collegati al tuo nodo": the newest connections come first, so whoever has been connected longest moves to the end. A callsign that has just connected (last 12 seconds) flashes, with a green LED running around the border of its box; with reduced motion enabled it shows a still green outline instead.

- The engine hook (`local/Local.pm`) was not upgraded on nodes that were already installed: the update only replaced it when its SHA-256 was in a fixed list, and the list had not been extended since 2026-11-20. Nodes kept an old hook, so the live reload of the "nodes and users blocked at login" lists (0.12.0) and the "Partner silenziosi" card (0.15.0) did not work, with no error shown. The hook now carries a `DXREGIA-MANAGED-HOOK` marker on its second line and any file with it is replaced on update; the missing versions were added to the list too. A file the sysop wrote (no marker, unknown digest) is still left alone. Tests cover fresh install, upgrade and custom file.

## [0.15.2] - 2026-11-23

### Changed

- Engine aligned to DXSpider Mojo build 931 (upstream commit `30e57466`): two upstream changes are applied on top of build 929. DXUser can rebuild a missing `dxusers.db` from the weekly `user_json` export; the bad* lists are validated and repaired at startup, with backups in `local_data/badlist-backups/`, and IPv6 CIDR handling no longer prints spurious warnings.
- The engine now converts `badnode.local` and `badspotter.local` from the old Perl hash format to plain text (one callsign per line) the first time it starts. The panel reads both formats and saves plain text, so the lists stay visible and are not overwritten after the update. Comments (`#`) at the top of a list are kept; the insertion date of an entry is no longer stored.
- Public page footer: "DXSpider Mojo" links to the upstream branch, and shows the engine build.
- The first update to this version restarts the engine, because `engine/` changes.

## [0.15.1] - 2026-11-23

### Fixed

- Admin panel: choosing a page from the menu (or using the browser's back button) now scrolls the window to the top, so the page's buttons are visible at once. Before, the page stayed at the scroll position of the menu entry that was clicked.

## [0.15.0] - 2026-11-23

### Added

- Dashboard, first row: a "Partner silenziosi" card lists connected partner nodes from which no DX spot has arrived for more than 30 minutes (the green LED does not show this). A "Ultimi eventi" card shows the latest link and start events of the engine (connections, restarts, take-overs).
- The engine hook (`presence.pm`) now records when each neighbouring node last sent a spot; the presence snapshot carries it as `SpotAt` on node channels. Because the hook is embedded in the binary, the first update to this version restarts the engine.

### Changed

- Dashboard, "Nodi partner" card: the country sits on the same line as the callsign, separated by a dot, and the nodes are laid out on three columns (18 per page), so the card is much lower.

## [0.14.2] - 2026-11-22

### Changed

- Dashboard: the "Nodi collegati in ingresso" card is gone. The "Nodi partner" card already lists every connected node, including inbound ones that are not registered as partners, with the IN tag, so the second card repeated the same information.

## [0.14.1] - 2026-11-22

### Changed

- The engine's startup log no longer shows the harmless `cannot stat '/spider/initial_data/*'` and `not a git repository` lines: `engine/perl/cluster.pl` skips the copy when the folder does not exist and asks git only when a repository exists (change dated in the file).
- Safety net in the service unit: after a clean stop, without a recent `keep-engine` file, a leftover engine process is stopped.

## [0.14.0] - 2026-11-22

### Changed

- Updates that do not touch the engine (DXSpider files, `COPYING`, `LICENSE`, `presence.pm`) no longer interrupt Telnet and the partner links: the web service restarts and the running engine stays up; the new service finds it and takes it over. When the engine changes, the update restarts everything as before. The first update that installs this version still restarts the engine, because the old service unit stops it; the following ones keep it.
- On Linux the engine now runs in its own process group, started by a small shell that records its process number and exit code. The web service stops it when the service stops, unless the update script left a recent `keep-engine` file. A crash of an engine taken over by a new service is still detected (automatic restart).
- The panel's diagnostic log shows when the new service took over the running engine and when a stopping service left it running.


### Fixed

- The admin panel stayed blank after sign-in since 0.12.0: the Backup page component was used but not imported (`BackupNode is not defined`). The build does not report a missing component, so it went unnoticed. Anyone who updated to 0.12.0 or 0.13.0 should update to 0.13.1.

## [0.13.0] - 2026-11-22

### Added

- Change the callsign of a partner node from its card, to fix a wrong callsign. It is possible only while the node has not been enabled in DXSpider and is not connected; after that the callsign also lives in the engine, so the node is deleted and registered again. Saved credentials move with the node and a note records the change.
- Preview before updating (Gestione > Aggiornamenti): the changes between the installed and the available version, taken from GitHub, and whether the update restarts the service or only replaces the panel files.

### Changed

- The last four browser confirmation windows (delete a contact request, delete a filter, revoke a registration, Telnet password required) now use the panel's own dialog, shared in `ConfirmDialog.svelte`.

## [0.12.0] - 2026-11-22

### Added

- Backup of the running node from the panel (Gestione > Backup): create one now or every day at a chosen UTC hour, keep the last 1-30 copies, download or delete them. The copy does not interrupt Telnet or the links; the user database is copied with SQLite's online backup and checked. Downloading and deleting need a full session (password and second factor). Restoring from a backup is still done from the terminal.
- Optional daily summary email (after 06:00 UTC) with the partners' status and the routing events, refused logins and automatic restarts of the last 24 hours. It is switched on in the email settings.
- The block lists page also edits the nodes and users blocked at login (`badnode` and `badspotter`). The entries are saved in the same `.local` files that `set/badnode` and `set/badspotter` write, and the node reloads them within seconds without a restart.

### Changed

- `backup_node.py backup` accepts `--online` for a running node. Without it the tool still requires a stopped node.

## [0.11.0] - 2026-11-22

### Added

- The DXSpider log shows two new categories: **Routing** (spots and PC92 packets the engine discarded, route changes rolled back) and **Accessi** (logins refused by `badip`, `badnode` or `badspotter`, too many connections, logins too close together, a second login refused). The existing search finds a partner by callsign. Private details (user IPs, packet contents) are not shown.
- The engine restarts by itself after an abnormal exit, up to five times in ten minutes; a clean exit (for example the sysop `shutdown` command) does not restart it. Each automatic restart is recorded in the DXSpider log.
- The node diagnostics show how long the service and the engine have been running and how many automatic restarts happened.

### Changed

- The DXSpider log moves to a new **Diagnostica** menu group, which leaves the Nodo group shorter.

## [0.10.0] - 2026-11-22

### Changed

- The DXSpider engine is aligned with upstream 1.57 build 929 (it was build 688): PC92 routing fixes with rollback, limits and bad-node/bad-spotter checks at login, new `load/badnode` and `load/badspotter` commands, automatic compression of old logs, many warnings and loops fixed. DXWeb and DXWeb Admin are not included.
- The panel console answers the new DXSpider question before replacing an older session of the same callsign.
- The laboratory callsigns of the Linux tests no longer start with Q, which upstream now rejects, and the test spot travels over a public-looking internal subnet because upstream drops PC61 from non-public addresses.

## [0.9.0] - 2026-11-22

### Changed

- Updates that change only the panel's pages and the documentation no longer restart the service: the files are replaced in place, so Telnet and the partner links stay connected. If the check on the new files fails, the previous pages are restored. Any change to the program, the engine or the system files still does the full update with a restart. `DXREGIA_FULL=1` forces the full update.

### Fixed

- The admin panel now shows the DXSpider version, the project link and the author in its footer.
- The Updates page names the versions (for example v0.8.0) next to the commits, instead of a possibly outdated "latest release". The admin sidebar showed a fixed v0.7.0: it now follows the real version.

## [0.8.0] - 2026-11-22

### Added

- Update DXRegia from the admin panel (Management > Updates). The page shows the installed and the available version, checks GitHub on demand and starts the official update after the sysop confirms in a dialog. No password is asked again: the sign-in with the two-step check is enough. Progress and the result are shown live, and the page reconnects by itself after the restart. Native Linux installs only. Installations made before this change need one last update from the terminal (`dxregia-update`) to install the two systemd units.

### Changed

- The partner nodes card on the admin status page shows the country flag and name of each node, a divider between the two columns, only the LED for active nodes and `IN` or `OUT` in the Direction column.

## [0.7.4] - 2026-11-21

### Fixed

- After a reboot, if the network or DNS is not ready yet, DXSpider Mojo no longer stays off: the service waits for the network (`network-online.target`, `nss-lookup.target`) and DXRegia retries the engine start every 10 seconds, up to every 2 minutes, until it answers.

### Changed

- The partner nodes card on the admin status page shows the nodes in two compact columns, 16 per page.

## [0.7.3] - 2026-11-21

### Changed

- In the partner list, "Unlock this node" and "Delete node" now ask for confirmation in a DXRegia dialog instead of the browser's own alert.

## [0.7.2] - 2026-11-21

### Fixed

- Public portal pages are now revalidated on every visit, so a browser never keeps showing an older version after an update. Hashed files in `/assets` stay cached. A page that still cannot find a file after an update reloads itself once.
- The portal footer (visits, DXSpider version, project link, author) stays at the bottom of the page when the content is short.

## [0.7.1] - 2026-11-21

### Changed

- The sysop two-step check is now remembered for 10 days on each browser. After the six-digit code, a persistent trusted-browser cookie lets the next password sign-in skip the code, even after closing the browser or restarting the service. Trust ends when the password or the second factor changes. Session length can now be set to 10 days.

## [0.7.0] - 2026-11-20

### Added

- Optional Reverse Beacon Network integration for CW and RTTY. Sysops can connect to the official feed from the admin panel. The public RBN tab shows the most recent reports, with DX or skimmer search, mode filters, SNR and quality. Telnet users can subscribe to the modes they want.

## [0.6.4] - 2026-11-20

### Changed

- Place the disk, RAM and node diagnostics cards on the same row on desktop. Disk and RAM show current usage as colored pie charts.
- Show the localized country or territory name when hovering over a connected station’s DXCC flag.

## [0.6.3] - 2026-11-20

### Changed

- Combine disk and RAM usage into one compact row on the node dashboard. Each resource has a ring chart for current usage and shows used, total and available space. Reduce the height of the diagnostics panel.

## [0.6.2] - 2026-11-20

### Changed

- Make the connected radio amateurs grid easier to read with larger callsigns and DXCC flags, white text in dark mode and no per-station clock. The desktop grid retains ten columns and four rows per page.
- Show the solar and radio indices on square, color-coded cards. The public Propagation page focuses on solar conditions, HF forecasts, VHF and the real-time MUF map.

## [0.6.1] - 2026-11-20

### Changed

- Show up to 40 connected radio amateurs in a ten-column, four-row grid on wide screens. Each station retains its DXCC flag, callsign, connection type and start time.

## [0.6.0] - 2026-11-20

### Added

- Propagation overview with HamQSL solar indices, HF conditions by day and night, VHF phenomena, the KC2G MUF map and observed activity by band and region.
- Graphical editor for DXSpider spot filters in the sysop panel.
- Telnet setup guide for seven loggers on the public registration page, in all six portal languages.
- Public portal and sysop dashboard screenshots in all six README translations.

### Changed

- The node dashboard has compact status cards, paginated partner nodes and a full-width grid with up to 30 connected radio amateurs per page on desktop. Callsign, DXCC flag, connection type and time share one line.
- Live spots use a compact table with a dedicated LoTW and eQSL AG column and a consolidated toolbar. Removed the Announcements page, the QSL service selector and new-entity highlighting.
- Reuse encoded JSON spot snapshots across requests until the archive changes.

### Fixed

- Recover missed spots from the archive after the live feed reconnects.
- Refresh the DXRegia-managed presence hook when upgrading from 0.5.0 so existing nodes report DXCC values for Telnet sessions.

## [0.5.0] - 2026-11-19

### Added

- Browser-local radio alert rules, opted-in system notifications, quiet hours, grouped reports, daily summaries and saved filter expiry.
- Saved web spot filters, compact and band views, a side-by-side band comparison, pause with archive recovery, corroboration, comment operators and filtered CSV export.
- DXCC directory, callsign suggestions and history, prefix exceptions, and geographic filters from DXSpider spot records.
- Node diagnostics for session refresh delay, disk space, log growth and managed process IDs.
- Latest spot age, web arrival timing and conditional spot snapshots, plus resumable live dashboard updates.
- Opt-in public status at /stato and configurable retention for event archives.
- Operator and ham quick-start guides, multilingual peering examples, migration and Raspberry Pi notes, and issue templates.
- Passwordless inbound access for CC Cluster peers by public IP or DNS hostname.
- Partner onboarding emails after local setup and a separate confirmation when an inbound session becomes active.
- Inbound session overview in the dashboard and a dedicated filter in the partner list.

- Public DX Cluster statistics, partner email history with redacted credentials, SMTP preferences and partner alerts.
- Persistent Telnet web console with command history, separate spot output and command suggestions.
- Prebuilt Linux installation packages and SHA-256 checksums for amd64 and arm64.
- Admin disk and RAM graphs with up to one hour of in-memory samples.

### Changed

- Consolidated the public navigation and spot search into the main header and the existing search field.
- Two-factor authentication is optional and verified web sessions persist across service restarts.
- Native first setup has LAN access, and radio registration sets the local cluster as Home Node.

### Fixed

- Saved Telnet banner translations are selected before login. Public IPs use the country database, local access uses the node’s callsign, and English is the fallback.
- The portal header stays in place when switching between spots, DXCC search and radio alerts.
- Updated the Linux checks for optional 2FA and persistent web sessions. Documentation-only changes skip those jobs.

## [0.4.0] - 2026-11-18

### Added
- Admin panel, Node menu: tooltips on the Telnet options, and a plain explanation that registration decides who can send spots.
- The portal can be installed as an app on a phone (manifest, service worker, install button; an instruction for Safari on iPhone).
- Telnet banner: the message of the day is shown before the login prompt, and its box grows to fit the text.
- User registration from the portal: a web form (callsign, email, password) with an email confirmation link; confirmed users are created in DXSpider with a bcrypt password and can send spots, and Telnet then asks for that password. Sysop page "Registered users" to enable it, set the public site address and revoke users.
- Access history page and a dashboard that separates partner nodes from the users connected to the node.
- Portal: dismissible Aurora sponsor banner on the spots page (six languages; closing it hides it for 7 days) and the author credit in the footer.
- Admin panel translated into the six portal languages (the language button is now in the admin header too). Texts live in `dxregia/web/src/adminTranslations.ts`; `scripts/missing_admin_strings.py` lists texts without a translation.
- Partner card: shows whether the partner is connected to this node (incoming) and when it first did.
- Portal footer: DXSpider version in use.
- Admin panel: the sysop session ends with a return to the login page, and its length can be chosen (Password page).
- Admin panel, Node menu: message of the day, crontab, startup script, block lists (words, DX callsigns, IP addresses), hop counts and filter inventory, each with tooltips. See `docs/NODE-SETTINGS.md`.

### Changed
- Spots page on a phone: a compact table with the flags always in view, multi-select drop-down filters (bands, modes), a slim sponsor bar and a header on one row.
- `update-native.sh` keeps only the last three previous versions after a successful update.
- The DXSpider Mojo engine now lives in `engine/` (original structure kept); `COPYING` and `LICENSE` stay at the root. Installers, the Docker image, the bundle builder and the CI were updated.
- `update-native.sh` accepts `DXREGIA_BRANCH` to update from a branch other than `main`.

### Fixed
- The node now reports the real DXSpider build (1.57, build 688 of upstream commit 9b55f03f) instead of a stale one.
- Portuguese texts for the spot filters and the sound button sat in the Spanish catalogue; the Portuguese portal showed them in English.
- A flat close button on the sponsor banner (it inherited the 3-D button style).

## [0.3.0] - 2026-11-18

### Added
- New logo: the word DXREGIA drawn by hand in very heavy blocks with an amber X and a pointed A, plus the DX icon for the favicon, phone and installable-site icons, in high definition (SVG and PNG, light and dark).
- Analytics center in the admin panel: portal audience (visits, unique visitors without cookies, language, devices, hours) and cluster activity (spots per day and hour, bands, modes, entities, callsigns).
- Rebuilt public spots page: multi-select filters on one row, search, a lens on callsigns, new spots highlighted with an optional sound, mode inferred from the frequency, light and dark theme, phone and tablet layout, a slim footer with the visit count and a project link.
- Announced DX operations (public NG3K ADXO feed) with a popup next to the callsign.
- Portal and emails in six languages (Italian, English, Spanish, German, French, Portuguese; English for any other language); HTML emails to partners with a confirmation at the first connection.
- Automatic partner setup from the Contact form, with a switch and a daily limit.
- Deletion of contact requests and partner nodes (a deleted node is disconnected and its access revoked); live dashboard with status LEDs; sign-in with user and password; admin panel moved to `/admin`.
- Documentation in six languages: README and installation and configuration guide (English first).

### Changed
- The sign-in step before the two-factor code now lasts 20 minutes.

## [0.2.0] - 2026-11-18

### Added
- Status dashboard with LEDs, counters, spots per band and latest spots.
- Partner nodes page in two cards (outgoing and incoming links), with "Enable and email" and a resend of the instructions.
- Management → Mail server panel with a test message.
- Flags for every DXCC entity on the web cluster.

### Fixed
- Saving a partner's record no longer erases its remote password.
- A Mojo account without a password no longer blocks the outgoing link.

## [0.1.0] - 2026-11-18

First published version, still in testing.

### Added
- Public spots page with filters, DXCC names and a mobile layout.
- Sysop area with protected sign-in, a mandatory change of the temporary password and TOTP confirmation.
- First-run wizard: configures, initialises and starts Mojo from the browser.
- Partner node management: live status, remote password, a preview of the peering script, Mojo account provisioning and automatic reconnection.
- Authenticated local Telnet console and Telnet options managed from the panel.
- Native Linux install and Docker, with updates through `dxregia-update`.

### Changed
- A partner's remote password is optional when saving; turning off reconnection keeps the password.
