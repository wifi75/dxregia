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
    provision_partners();
    connect_partners();
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
        };
    }
    my $target = "$main::local_data/dxregia-presence.json";
    my $tmp = "$target.tmp";
    open(my $file, '>:raw', $tmp) or die "DXRegia snapshot: $!";
    print $file JSON::encode_json({ At => int($now), PartnerConnect => JSON::true, Channels => \@channels }) or die "DXRegia snapshot: $!";
    close($file) or die "DXRegia snapshot: $!";
    rename($tmp, $target) or die "DXRegia snapshot: $!";
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
        read($request, $password, 49);
        close($request);
        if ($password !~ /\A[0-9a-f]{48}\z/) {
            $result = 'Richiesta non valida';
        } elsif (length($call) > 12 || $call eq $main::mycall || $call eq $main::myalias) {
            $result = 'Indicativo riservato';
        } else {
            my $user = DXUser::get($call);
            if ($user && ($user->lockout || $user->sort ne 'S' || !$user->passwd)) {
                $result = 'Account Mojo esistente da verificare';
            } else {
                if (!$user) {
                    $user = DXUser->new($call);
                    $user->sort('S');
                    $user->homenode($call);
                    $user->priv(1);
                    $user->passwd($password);
                    $user->close();
                    $result = 'Abilitato in Mojo';
                } else {
                    $result = $user->passwd eq $password ? 'Abilitato in Mojo' : 'Account Mojo esistente da verificare';
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
            $user && $user->sort eq 'S' && !$user->lockout && $user->passwd) {
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
                    print $chat "timeout 15\nconnect telnet $host $port\n'ogin:' '$main::mycall'\n";
                    print $chat "'assword:' '$password'\n" if length($password);
                    if (close($chat)) {
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
        unlink "$dir/$file";
    }
}

1;
