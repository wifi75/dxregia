# Changelog

All notable changes to DXRegia (the `dxregia/` folder) are recorded here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and versions follow [Semantic Versioning](https://semver.org/). The upstream engine keeps its own history in `Changes`.

## [Unreleased]

### Fixed

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
