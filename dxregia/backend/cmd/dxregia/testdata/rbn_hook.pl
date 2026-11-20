use strict;
use warnings;
use JSON::PP ();
use Test::More;

BEGIN {
    package JSON;
    sub encode_json { JSON::PP::encode_json($_[0]) }
    sub decode_json { JSON::PP::decode_json($_[0]) }
    sub is_bool { JSON::PP::is_bool($_[0]) }
    sub true () { JSON::PP::true }
    sub false () { JSON::PP::false }
    $INC{'JSON.pm'} = __FILE__;
}

package RBN;
use constant { ROrigin => 0, RQrg => 1, RCall => 2, RMode => 3, RStrength => 4, RUtz => 6, CData => 2 };
our $native_reports = 0;
sub send_dx_spot { ++$native_reports }

package Prefix;
sub cty_data { return $_[0] eq 'I2ZZZ' ? (85, 0, 0, '', '', 'Italy-I') : (666) }

package DXDebug;
sub dbg { die @_ }

package DXUser;
our $account;
sub get_current { $account }
sub new { $account = bless {}, shift }
sub sort { $_[0]->{sort} = $_[1] }
sub name { @_ > 1 ? $_[0]->{name} = $_[1] : $_[0]->{name} }
sub passwd { @_ > 1 ? $_[0]->{passwd} = $_[1] : $_[0]->{passwd} }
sub homenode { $_[0]->{home} = $_[1] }
sub close {}
sub is_rbn { $_[0]->{sort} eq 'N' }
sub lockout { 0 }
sub priv { 0 }

package DXChannel;
our @active;
sub get {
    my ($call) = @_;
    return (grep { $_->{call} eq $call } @active)[0];
}
sub get_all { @active }
sub is_rbn { 1 }
sub conn { $_[0]->{conn} }
sub disconnect { @active = () }

package Msg;
sub conns { undef }

package ExtMsg;
our ($destination, $profile, $attempts);
sub start_connect {
    my ($call, $path) = @_;
    ++$attempts;
    $destination = $call;
    open my $script, '<:raw', $path or die $!;
    local $/;
    $profile = <$script>;
    close $script;
}

package main;
our ($root, $local_data, $mycall, $myalias);
my $hook;
($root, $hook) = @ARGV;
$local_data = "$root/local_data";
$mycall = 'I1ZZZ-7';
$myalias = 'I1ZZZ';
require $hook;
Local::init();

my $now = time;
my $snapshot = "$local_data/dxregia-rbn.json";
sub state {
    open my $file, '<:raw', $snapshot or die $!;
    local $/;
    return JSON::PP::decode_json(<$file>);
}

Local::process_rbn($now);
is($ExtMsg::destination, 'DX0RBN', 'creates an RBN account reserved for the feed');
like($ExtMsg::profile, qr/connect telnet telnet\.reversebeacon\.net 7000\n'call:' 'I1ZZZ-7'/, 'connects only to CW/RTTY using the node callsign');
like($DXUser::account->passwd, qr/\A\{DXREGIA-BCRYPT\}\$2a\$10\$/, 'requires a random credential for inbound use of the reserved account');
is(state()->{State}, 'connecting', 'waits for the feed to connect');
Local::process_rbn($now + 5);
is($ExtMsg::attempts, 1, 'waits before retrying the connection');

my $active = bless { call => 'DX0RBN', conn => { outgoing => 1, peerport => 7000 } }, 'DXChannel';
push @DXChannel::active, $active;
RBN::send_dx_spot($active, 'Q:3*', [
    $now, 3, ['DL1ZZZ-#', 14020.5, 'I2ZZZ', 'RTT', 12, '', $now],
    ['DL2ZZZ-#', 14020.7, 'I2ZZZ', 'RTT', 19, '', $now],
]);
RBN::send_dx_spot($active, 'Q:2', [$now, 2, ['DL3ZZZ', 14074, 'I3ZZZ', 'FT8', 10, '', $now]]);
my $inbound = bless { conn => { outgoing => 0, peerport => 7000 } }, 'DXChannel';
RBN::send_dx_spot($inbound, 'Q:3', [$now, 3, ['DL4ZZZ', 7020, 'I4ZZZ', 'CW', 12, '', $now]]);
Local::process_rbn($now + 10);
is(state()->{State}, 'connected', 'reports an active CW/RTTY feed');
is_deeply(state()->{Spots}, [{
    Call => 'I2ZZZ', Skimmer => 'DL2ZZZ', Mode => 'RTTY', Frequency => 14020.5, SNR => 19, Quality => 3, At => $now, DXCC => 85, DXCCPrefix => 'I',
}], 'exports one report per curated group and ignores inbound and FT8 reports');
is($RBN::native_reports, 3, 'keeps the native delivery path intact');

open my $config, '>:raw', "$root/local/dxregia-rbn.json" or die $!;
print $config '{"Enabled":false}';
close $config;
Local::process_rbn($now + 15);
is(scalar @DXChannel::active, 0, 'closes the managed connection when disabled');
done_testing();
