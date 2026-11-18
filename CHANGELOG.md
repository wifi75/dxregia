# Changelog

All notable changes to DXRegia (the `dxregia/` folder) are recorded here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and versions follow [Semantic Versioning](https://semver.org/). The upstream engine keeps its own history in `Changes`.

## [Unreleased]

### Added
- Admin panel: the sysop session ends with a return to the login page, and its length can be chosen (Password page).
- Admin panel, Node menu: message of the day, crontab, startup script, block lists (words, DX callsigns, IP addresses), hop counts and filter inventory, each with tooltips. See `docs/NODE-SETTINGS.md`.

### Changed
- The DXSpider Mojo engine now lives in `engine/` (original structure kept); `COPYING` and `LICENSE` stay at the root. Installers, the Docker image, the bundle builder and the CI were updated.
- `update-native.sh` accepts `DXREGIA_BRANCH` to update from a branch other than `main`.

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
