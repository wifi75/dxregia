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
