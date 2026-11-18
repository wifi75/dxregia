# Node settings in the admin panel

DXSpider keeps many settings in plain files that a sysop used to edit by hand. DXRegia exposes them as pages in the **Node** menu of the admin panel (sysop login with two-step verification). Each page has a `?` tooltip next to every field.

Every page writes the original DXSpider file, validates the input first and keeps any line it does not understand untouched. Nothing typed in the panel is ever run as free Perl code.

| Page | DXSpider file | Takes effect |
| --- | --- | --- |
| Message and schedule | `motd`, `local_cmd/crontab` | Message: next connection. Schedule: DXSpider re-reads the file by itself |
| Startup script | `scripts/startup` | Next start of the node |
| Block lists | `local_data/badword.new`, `local_data/baddx`, `local_data/badip.local` | Within a few seconds, no restart |
| Hops and filters | `local/dxregia-hops.json` (applied to `$DXProt::def_hopcount` and `%DXProt::hopcount`), `filter/` | Hops: within a few seconds. Filters: next load of the filter |

## Message of the day

Text shown to Telnet users right after they connect, up to 2000 bytes. An empty text removes the file.

## Schedule (crontab)

One row per command: minute, hour, day of month, month, day of week, command. DXSpider evaluates the command column as Perl, so the panel only writes `run_cmd("command")` and only accepts letters, digits, spaces and `. / : = , + -`. Quotes, `$`, `@` and `\` are refused. Existing rows in another form (for example `Spot::daily()`) are kept as they are and shown below the table.

## Startup script

One command per line. Accepted forms:

- `set/…`, `unset/…`, `load/…`
- `connect CALLSIGN`
- `set/var $Package::name = value`, where the value is a number or a short plain text

Anything else is refused. Existing lines in other forms stay in the file.

## Block lists

- **Forbidden words**: letters and digits only. Advanced regular-expression rules already present in `badword.new` stay as they are.
- **Blocked DX callsigns**: spots with these callsigns are neither shown nor forwarded.
- **Blocked IP addresses**: single addresses (`192.0.2.1`) or networks (`192.0.2.0/24`), IPv4 and IPv6.

After saving, the panel hook reloads the lists inside the running node (`BadWords::load`, `DXCIDR::reload` and a fresh `baddx` hash), so no restart is needed.

## Hops and filters

The default hop count (1–99, DXSpider uses 30) applies to every packet type without its own row. Exceptions are set per PC number (1–999), for example PC16/PC17 and PC19/PC21 start at 15.

The old `@nodx_node`, `@noann_node` and `@nowwv_node` lists of `DXProtVars.pm` are not read by this engine, so the panel does not offer them: use filters instead.

Filters are listed (with their content, read-only) and can be deleted. To create or change one, use the `accept/…` and `reject/…` commands in **Console Telnet**.

## Where the data lives

All paths are relative to the node directory (`/var/lib/dxregia/mojo` on a native install). Back them up with the node backup.

## Session length

**Management → Password → Session length** sets how long a panel login lasts (1 hour to 7 days, default 8 hours). When the session ends, the panel returns to the login page on the same section.

## Access history

**Node → Node accesses** lists who connected to the node (radio amateurs or partner nodes), with callsign, source address and duration, for the last 24 hours, 7 or 30 days. It is built from the DXSpider log, so it only goes back as far as the log files do. The dashboard shows who is connected right now.

## Telnet options

**Node → Telnet options** has a tooltip on every field. The key choice is **User registration**:

- *Everyone can send spots* (default): anyone who connects can send spots, announcements and chat.
- *Only users registered by the sysop send spots*: anyone can connect and read, but spots, announcements, chat and messages need a registered user. Register by hand with `set/register CALLSIGN`, or let users register themselves (below).
- *Everyone except those deregistered by the sysop*: registered by default, with a blacklist.

## Registration of users

**Links → Registered users** turns on self-registration from the portal. Visitors open `/registrazione`, give a callsign, an email and a password (8–72 printable characters, no spaces) and confirm a link sent by email (valid 24 hours, usable once). Only then is the callsign created in DXSpider as a registered user with that password, stored as a bcrypt hash; Telnet then asks for the password. The page also lists the registrations and can revoke any of them.

To use it:

1. Set up the mail server (Management → Mail server).
2. In *Registered users*, tick the switch and write the public address of the site (for example `https://dxregia.example.org`). It is set by hand on purpose: the links in the emails never take their address from the request.
3. Choose *Only users registered by the sysop send spots* (the page offers a button).

Limits: 40 requests a day, 3 per callsign or email. A callsign that already has a password set by the sysop cannot be taken over. Registration does not prove that the person owns the callsign, only the email address.

## Portal extras

- The spots page has a dismissible sponsor banner (Aurora), the DXSpider version and the author in the footer.
- The portal can be installed on a phone (manifest and service worker; HTTPS required). The service worker never caches the API or `/admin`.
