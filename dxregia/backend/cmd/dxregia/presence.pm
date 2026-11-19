package Local;

use strict;
use JSON ();

my $last_write = 0;
my $applied_profile = "";

sub process {
    my $now = time;
    return if $now - $last_write < 5;
    $last_write = $now;
    apply_node_profile();
    apply_sysop_password();
    apply_telnet_options();
    apply_hops();
    reload_lists();
    provision_users();
    provision_partners();
    connect_partners();
    remove_partners();
    my @channels;
    foreach my $channel (DXChannel::get_all()) {
        next if $channel->call eq $main::mycall;
        my $connection = $channel->conn;
        next unless $connection;
        my $kind = $channel->is_node ? 'node'
                 : $channel->is_web ? 'web'
                 : $connection->isa('ExtMsg') && $channel->is_user ? 'telnet'
                 : '';
        next unless $kind;
        push @channels, {
            Call => "" . $channel->call,
            Kind => $kind,
            State => "" . ($channel->state // ""),
            Since => int($channel->startt || 0),
            Outbound => ($kind eq 'node' && $channel->{outbound}) ? JSON::true : JSON::false,
        };
    }
    my $target = "$main::local_data/dxregia-presence.json";
    my $tmp = "$target.tmp";
    open(my $file, '>:raw', $tmp) or die "DXRegia snapshot: $!";
    print $file JSON::encode_json({ At => int($now), PartnerConnect => JSON::true, SysopPassword => JSON::true, Telnet => telnet_options(), Channels => \@channels, LockedPartners => locked_partners() }) or die "DXRegia snapshot: $!";
    close($file) or die "DXRegia snapshot: $!";
    rename($tmp, $target) or die "DXRegia snapshot: $!";
}

sub locked_partners {
    my %locked;
    open(my $directory, '<:raw', "$main::root/.dxregia-partners.json") or return \%locked;
    my $data = do { local $/; <$directory> };
    close($directory);
    my $partners = eval { JSON::decode_json($data) };
    return \%locked unless ref($partners) eq 'HASH';
    for my $call (keys %$partners) {
        next unless $call =~ /\A[A-Z0-9]+(?:-[0-9]+)?\z/ && length($call) <= 12;
        next if $call eq $main::mycall || $call eq $main::myalias;
        my $user = DXUser::get_current($call);
        next unless $user && $user->lockout;
        $locked{$call} = ($user->sort eq 'S' && ($user->priv || 0) <= 1) ? JSON::true : JSON::false;
    }
    return \%locked;
}

my $applied_telnet = '';

sub telnet_options {
    return {
        PasswordRequired => $main::passwdreq ? JSON::true : JSON::false,
        SlashCallsigns => $main::allowslashcall ? JSON::true : JSON::false,
        ReplaceSession => $main::bumpexisting ? JSON::true : JSON::false,
        Registration => int($main::reqreg // 0),
        ConcurrentUsers => int($main::maxconnect_user // 0),
        ConcurrentNodes => int($main::maxconnect_node // 0),
        MultipleSessions => int($main::allowmultiple // 0),
        ReconnectSeconds => int($main::min_reconnection_rate // 0),
        MaxSSID => int($main::max_ssid // 0),
    };
}

sub apply_telnet_options {
    my $path = "$main::root/local/dxregia-telnet.json";
    open(my $file, '<:raw', $path) or return;
    my $bytes = '';
    read($file, $bytes, 1025);
    close($file);
    return if !$bytes || length($bytes) > 1024 || $bytes eq $applied_telnet;
    my $value = eval { JSON::decode_json($bytes) };
    return unless ref($value) eq 'HASH';
    for my $key (qw(PasswordRequired SlashCallsigns ReplaceSession)) {
        return unless JSON::is_bool($value->{$key});
    }
    my %max = (Registration => 2, ConcurrentUsers => 1000, ConcurrentNodes => 1000,
               MultipleSessions => 15, ReconnectSeconds => 86400, MaxSSID => 15);
    foreach my $key (keys %max) {
        return unless defined $value->{$key} && $value->{$key} =~ /\A[0-9]+\z/ && $value->{$key} <= $max{$key};
    }
    $main::passwdreq = $value->{PasswordRequired} ? 1 : 0;
    $main::allowslashcall = $value->{SlashCallsigns} ? 1 : 0;
    $main::bumpexisting = $value->{ReplaceSession} ? 1 : 0;
    $main::reqreg = $value->{Registration};
    $main::maxconnect_user = $value->{ConcurrentUsers};
    $main::maxconnect_node = $value->{ConcurrentNodes};
    $main::allowmultiple = $value->{MultipleSessions};
    $main::min_reconnection_rate = $value->{ReconnectSeconds};
    $main::max_ssid = $value->{MaxSSID};
    $applied_telnet = $bytes;
}

my $applied_hops = '';

# Hop del protocollo: valore predefinito e tabella per numero PC (file scritto dal pannello).
sub apply_hops {
    my $path = "$main::root/local/dxregia-hops.json";
    open(my $file, '<:raw', $path) or return;
    my $bytes = '';
    read($file, $bytes, 4097);
    close($file);
    return if !$bytes || length($bytes) > 4096 || $bytes eq $applied_hops;
    my $value = eval { JSON::decode_json($bytes) };
    return unless ref($value) eq 'HASH';
    return unless defined $value->{Default} && $value->{Default} =~ /\A[0-9]+\z/ && $value->{Default} >= 1 && $value->{Default} <= 99;
    return unless ref($value->{Table}) eq 'HASH';
    my %table;
    foreach my $pc (keys %{$value->{Table}}) {
        my $hops = $value->{Table}{$pc};
        return unless $pc =~ /\A[0-9]{1,3}\z/ && defined $hops && $hops =~ /\A[0-9]+\z/ && $hops >= 1 && $hops <= 99;
        $table{$pc + 0} = $hops + 0;
    }
    $DXProt::def_hopcount = $value->{Default} + 0;
    %DXProt::hopcount = %table;
    $applied_hops = $bytes;
}

my %list_stamp;

# Liste di blocco: quando il pannello riscrive un file, Mojo lo rilegge senza riavvio.
sub reload_lists {
    my %jobs = (
        "$main::local_data/badword.new" => sub { BadWords::load() },
        "$main::local_data/badip.local" => sub { DXCIDR::reload() },
        "$main::local_data/baddx"       => sub { $DXProt::baddx = DXHash->new('baddx') },
    );
    foreach my $path (keys %jobs) {
        my $stamp = join(':', (stat $path)[9, 7]);
        if (!exists $list_stamp{$path}) {
            $list_stamp{$path} = $stamp;
            next;
        }
        next if $stamp eq $list_stamp{$path};
        $list_stamp{$path} = $stamp;
        eval { $jobs{$path}->(); 1 } or DXDebug::dbg("DXRegia list reload $path: $@");
    }
}

# Utenti registrati dal portale: l'email di conferma ha già verificato l'indirizzo, qui si crea o si
# aggiorna l'account Mojo con la password scelta (solo hash bcrypt) e lo si segna come registrato.
sub provision_users {
    my $dir = "$main::local_data/dxregia-users";
    opendir(my $dh, $dir) or return;
    my @pending = grep { /\A[A-Z0-9]{3,10}[.](?:register|revoke)\z/ } readdir($dh);
    closedir($dh);
    for my $file (@pending) {
        my ($call, $kind) = $file =~ /\A([A-Z0-9]+)[.](register|revoke)\z/;
        my $result;
        my $channel = DXChannel::get($call);
        if ($call eq $main::mycall || $call eq $main::myalias) {
            $result = 'Indicativo riservato';
        } elsif ($kind eq 'revoke') {
            my $user = DXUser::get_current($call);
            if ($user && ($user->passwd // '') =~ /\A\{DXREGIA-BCRYPT\}\$2a/ && $user->sort eq 'U' && !$user->priv) {
                $user->registered(0);
                $user->passwd('');
                $user->put;
                $channel->registered(0) if $channel;
            }
            unlink "$dir/$file";
            next;
        } else {
            my $hash = '';
            if (open(my $request, '<:raw', "$dir/$file")) {
                read($request, $hash, 2049);
                close($request);
            }
            my $profile;
            if ($hash =~ /\A\{/ && length($hash) <= 2048) {
                $profile = eval { JSON::decode_json($hash) };
                $hash = '';
                if (ref($profile) eq 'HASH' &&
                    !grep { !defined($_) || ref($_) || !length($_) || length($_) > 100 ||
                            /[\p{Cc}^|{}]/ } @{$profile}{qw(Name QTH Locator)}) {
                    if ($profile->{Locator} =~ /\A[A-R]{2}[0-9]{2}(?:[A-X]{2})?\z/) {
                        $hash = $profile->{Hash} // '';
                    }
                }
            }
            if ($hash !~ /\A\$2a\$10\$[.\/A-Za-z0-9]{53}\z/ || !ExtMsg::supports_dxregia_bcrypt($hash)) {
                $result = 'Richiesta non valida';
            } else {
                my $stored = '{DXREGIA-BCRYPT}' . $hash;
                my $user = DXUser::get_current($call);
                if ($user && ($user->lockout || $user->priv || $user->sort ne 'U')) {
                    $result = 'Account Mojo esistente da verificare';
                } elsif ($user && $user->passwd && $user->passwd !~ /\A\{DXREGIA-BCRYPT\}/) {
                    $result = 'Account Mojo esistente da verificare';
                } elsif ($user && $user->passwd && $user->passwd ne $stored) {
                    $result = 'Indicativo gia registrato';
                } else {
                    my $existing = $user;
                    unless ($existing) {
                        $user = DXUser->new($call);
                        $user->sort('U');
                        $user->priv(0);
                    }
                    if ($profile) {
                        my $qra = $profile->{Locator};
                        my ($lat, $long) = DXBearing::qratoll(length($qra) == 4 ? "${qra}LL" : $qra);
                        $user->name($profile->{Name});
                        $user->qth($profile->{QTH});
                        $user->qra($qra);
                        $user->lat($lat);
                        $user->long($long);
                        $user->homenode($main::mycall);
                    }
                    $user->passwd($stored);
                    $user->registered(1);
                    $existing ? $user->put : $user->close();
                    $channel->registered(1) if $channel;
                    $result = 'Registrato';
                }
            }
        }
        my $target = "$dir/$call.result";
        open(my $ack, '>:raw', "$target.tmp") or next;
        print $ack $result or next;
        close($ack) or next;
        rename("$target.tmp", $target) or next;
        unlink "$dir/$file";
    }
}

sub apply_sysop_password {
    my $request = "$main::local_data/.dxregia-telnet-password";
    return unless open(my $pending, '<:raw', $request);
    my $hash = '';
    my $read = read($pending, $hash, 61);
    my $modified = (stat $pending)[9];
    close($pending);
    my $status = 'failed';
    if (defined($modified) && time - $modified <= 60 &&
        defined($read) && $hash =~ /\A\$2a\$10\$[.\/A-Za-z0-9]{53}\z/) {
        my $user = DXUser::get_current($main::myalias);
        if ($user && $user->priv == 9) {
            my $old = $user->passwd;
            my $saved = eval {
                die 'bcrypt unavailable' unless ExtMsg::supports_dxregia_bcrypt($hash);
                $user->passwd('{DXREGIA-BCRYPT}' . $hash);
                $user->put;
                1;
            };
            if ($saved) {
                $status = 'applied';
            } else {
                $user->passwd($old);
            }
        }
    }
    my $result = "$main::local_data/dxregia-telnet-password.result";
    if (open(my $out, '>:raw', "$result.tmp")) {
        if (print($out $status) && close($out)) {
            rename("$result.tmp", $result);
        }
    }
    unlink $request;
}

sub apply_node_profile {
    my $path = "$main::root/local/dxregia-node.json";
    open(my $file, '<:raw', $path) or return;
    my $bytes = "";
    read($file, $bytes, 8193);
    close($file);
    return if !$bytes || length($bytes) > 8192 || $bytes eq $applied_profile;
    my $config = eval { JSON::decode_json($bytes) };
    return unless ref($config) eq 'HASH' && ($config->{node} || '') eq $main::mycall &&
                  ($config->{sysop} || '') eq $main::myalias;
    my $grid = $config->{locator};
    return unless defined($grid) && $grid =~ /\A[A-R]{2}[0-9]{2}(?:[A-X]{2})?\z/;

    my ($lat, $lon) = ($config->{latitude}, $config->{longitude});
    if (!defined($lat) || !defined($lon)) {
        my @grid = split(//, $grid);
        $lon = (ord($grid[0]) - ord('A')) * 20 + $grid[2] * 2 - 180;
        $lat = (ord($grid[1]) - ord('A')) * 10 + $grid[3] - 90;
        $lon += @grid == 6 ? (ord($grid[4]) - ord('A') + .5) / 12 : 1;
        $lat += @grid == 6 ? (ord($grid[5]) - ord('A') + .5) / 24 : .5;
    }
    return unless $lat >= -90 && $lat <= 90 && $lon >= -180 && $lon <= 180;
    $main::myname = $config->{name} || $main::myalias;
    $main::myqth = $config->{qth} || $grid;
    $main::myemail = $config->{email} || '';
    ($main::mylocator, $main::mylatitude, $main::mylongitude) = ($grid, $lat, $lon);
    for my $call ($main::mycall, $main::myalias) {
        my $user = DXUser::get_current($call) or next;
        $user->name($main::myname);
        $user->qth($main::myqth);
        $user->email($main::myemail);
        $user->qra($grid);
        $user->lat($lat);
        $user->long($lon);
        $user->put;
    }
    $applied_profile = $bytes;
}

sub register_partner_account {
    my ($user, $call) = @_;
    $user->registered(1);
    $user->put;
    my $channel = DXChannel::get($call);
    $channel->registered(1) if $channel;
}

# Existing CC Cluster peers keep their type and local credentials.
sub existing_ccluster_peer {
    my ($user) = @_;
    return $user && $user->sort eq 'L' && ($user->priv || 0) <= 1 &&
        !$user->lockout && $user->passwd;
}

sub provision_partners {
    my $dir = "$main::local_data/dxregia-partners";
    opendir(my $dh, $dir) or return;
    my @pending = grep { /\A[A-Z0-9]+(?:-[0-9]+)?[.]enable\z/ } readdir($dh);
    closedir($dh);
    for my $file (@pending) {
        my ($call) = $file =~ /\A([A-Z0-9-]+)[.]enable\z/;
        my $result;
        open(my $request, '<:raw', "$dir/$file") or next;
        my $password = '';
        read($request, $password, 65);
        close($request);
        my $unlock = $password =~ s/\Aunlock-outgoing:(?=[0-9a-f]{48}\z)/outgoing-only:/;
        my $outgoing = $password =~ /\Aoutgoing-only:[0-9a-f]{48}\z/;
        if (!$outgoing && $password !~ /\A[0-9a-f]{48}\z/) {
            $result = 'Richiesta non valida';
        } elsif (length($call) > 12 || $call eq $main::mycall || $call eq $main::myalias) {
            $result = 'Indicativo riservato';
        } else {
            my $user = DXUser::get_current($call);
            if ($unlock && (!$user || $user->sort ne 'S' || ($user->priv || 0) > 1 || !$user->lockout)) {
                $result = 'Account Mojo esistente da verificare';
            } elsif ($user && $user->sort eq 'S' && ($user->priv || 0) <= 1 && $user->lockout && !$unlock) {
                $result = 'Nodo bloccato in Mojo';
            } elsif ($outgoing && existing_ccluster_peer($user)) {
                $result = 'Pronto in uscita';
            } elsif ($user && (($user->lockout && !$unlock) || $user->sort ne 'S' || (!$user->passwd && !$outgoing))) {
                $result = 'Account Mojo esistente da verificare';
            } else {
                if (!$user) {
                    $user = DXUser->new($call);
                    $user->registered(1);
                    $user->sort('S');
                    $user->homenode($call);
                    $user->priv(1);
                    $user->passwd($outgoing ? '{DXREGIA-BCRYPT}' . $password : $password);
                    $user->close();
                    $result = $outgoing ? 'Pronto in uscita' : 'Abilitato in Mojo';
                } else {
                    if ($outgoing) {
                        $user->lockout(0) if $unlock;
                        # Account senza password (es. creato con set/register): resta valido
                        # per il solo collegamento in uscita, senza abilitare l'accesso in ingresso.
                        $user->passwd('{DXREGIA-BCRYPT}' . $password) unless $user->passwd;
                        register_partner_account($user, $call) unless $user->registered;
                        $user->put if $unlock || $user->passwd eq ('{DXREGIA-BCRYPT}' . $password);
                        $result = $user->passwd =~ /\A\{DXREGIA-BCRYPT\}outgoing-only:/ ? 'Pronto in uscita' : 'Abilitato in Mojo';
                    } elsif ($user->passwd =~ /\A\{DXREGIA-BCRYPT\}outgoing-only:/) {
                        $user->passwd($password);
                        register_partner_account($user, $call);
                        $result = 'Abilitato in Mojo';
                    } else {
                        if ($user->passwd eq $password) {
                            register_partner_account($user, $call) unless $user->registered;
                            $result = 'Abilitato in Mojo';
                        } else {
                            $result = 'Account Mojo esistente da verificare';
                        }
                    }
                }
            }
        }
        my $target = "$dir/$call.result";
        open(my $ack, '>:raw', "$target.tmp") or next;
        print $ack $result or next;
        close($ack) or next;
        rename("$target.tmp", $target) or next;
        DXDebug::dbg("DXRegia partner $call $result") if DXDebug::isdbg('connect');
        unlink "$dir/$file";
    }
}

# Un nodo eliminato dal pannello viene scollegato e il suo account Mojo viene cancellato,
# così la vecchia password non permette più di entrare.
sub remove_partners {
    my $dir = "$main::local_data/dxregia-partners";
    opendir(my $dh, $dir) or return;
    my @pending = grep { /\A[A-Z0-9]+(?:-[0-9]+)?[.]remove\z/ } readdir($dh);
    closedir($dh);
    for my $file (@pending) {
        my ($call) = $file =~ /\A([A-Z0-9-]+)[.]remove\z/;
        next unless unlink "$dir/$file";    # ogni richiesta si esegue una sola volta
        next if length($call) > 12 || $call eq $main::mycall || $call eq $main::myalias;
        eval {
            my $channel = DXChannel::get($call);
            $channel->disconnect if $channel;
            my $user = DXUser::get($call);
            $user->del if $user && $user->sort eq 'S';
        };
    }
}

sub partner_connect_script {
    my ($host, $port, $local_call, $remote_call, $password) = @_;
    my $script = "timeout 15\nconnect telnet $host $port\n'login' '$local_call'\n";
    $script .= "'password:' '$password'\n" if length($password);
    return $script . "client $remote_call telnet\n";
}

sub connect_partners {
    my $dir = "$main::local_data/dxregia-partners";
    opendir(my $dh, $dir) or return;
    my @pending = grep { /\A[A-Z0-9]+(?:-[0-9]+)?[.]connect\z/ } readdir($dh);
    closedir($dh);
    foreach my $file (@pending) {
        my ($call) = $file =~ /\A([A-Z0-9-]+)[.]connect\z/;
        open(my $request, '<:raw', "$dir/$file") or next;
        my $data = '';
        read($request, $data, 1024);
        close($request);
        my ($host, $port, $password) = split(/\n/, $data, 3);
        my $user = DXUser::get($call);
        my $result = 'Account Mojo non abilitato';
        if (length($call) <= 12 && $call ne $main::mycall && $call ne $main::myalias &&
            (($user && $user->sort eq 'S' && !$user->lockout && $user->passwd) || existing_ccluster_peer($user))) {
            if (DXChannel::get($call) || Msg->conns($call)) {
                $result = 'Sessione gia presente';
            } elsif (!defined($password) || $host !~ /\A[A-Za-z0-9.:-]{1,253}\z/ ||
                     $port !~ /\A[0-9]{1,5}\z/ || $port < 1 || $port > 65535 ||
                     $password !~ /\A[A-Za-z0-9]{0,72}\z/) {
                $result = 'Parametri non validi';
            } else {
                my $script = "$dir/$call.script";
                if (open(my $chat, '>:raw', $script)) {
                    chmod 0600, $script;
                    my $written = print $chat partner_connect_script($host, $port, $main::mycall, $call, $password);
                    if (close($chat) && $written) {
                        eval { ExtMsg::start_connect($call, $script); };
                        $result = $@ ? 'Avvio non riuscito' : 'Tentativo avviato';
                    }
                    unlink $script;
                }
            }
        }
        my $target = "$dir/$call.connect-result";
        open(my $ack, '>:raw', "$target.tmp") or next;
        print $ack $result or next;
        close($ack) or next;
        rename("$target.tmp", $target) or next;
        DXDebug::dbg("DXRegia partner $call $result") if DXDebug::isdbg('connect');
        unlink "$dir/$file";
    }
}

1;
