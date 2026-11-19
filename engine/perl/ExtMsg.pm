#
# This class is the internal subclass that deals with the external port
# communications for Msg.pm
#
# This is where the cluster handles direct connections coming both in
# and out
#
#
# Copyright (c) 2001 - Dirk Koopman G1TLH
#
#	Modified Jan 2006 by John Wiseman G8BPQ to support connections to BPQ32 node,
#		and fix pattern matching on 'chat' abort handling
#

package ExtMsg;

use strict;
use Msg;
use DXVars;
use DXUtil;
use DXDebug;
use DXTimer;

use IO::File;
use IO::Socket;
use JSON ();
use Socket ();
use IPC::Open3;

# A configured CC Cluster peer must use its exact TCP source address.
sub dxregia_peer_ip_match
{
	my ($conn, $call, $user) = @_;
	return undef if length($call) > 12 || $call !~ /\A[A-Z0-9]+(?:-[0-9]+)?\z/;
	my $path = "$main::local_data/dxregia-partners/$call.inbound-ip";
	return undef unless -e $path;
	return 0 unless $user && $user->sort eq 'L' && !$user->lockout &&
		($user->priv || 0) <= 1 && $call ne $main::mycall && $call ne $main::myalias;
	open(my $file, '<:raw', $path) or return 0;
	my $data = '';
	read($file, $data, 260);
	close $file;
	my $trusted = eval { JSON::decode_json($data) };
	my $source = eval { $conn->{sock}->handle->peerhost };
	return 0 unless defined($trusted) && !ref($trusted) && defined($source);
	$source =~ s/^::ffff://i;
	my @addresses = ($trusted);
	if ($trusted !~ /:/ && $trusted !~ /\A(?:[0-9]{1,3}[.]){3}[0-9]{1,3}\z/) {
		my $cache = "$main::local_data/dxregia-partners/$call.inbound-dns";
		open(my $dns, '<:raw', $cache) or return 0;
		my $json = '';
		read($dns, $json, 2048);
		close $dns;
		my $resolved = eval { JSON::decode_json($json) };
		return 0 unless ref($resolved) eq 'HASH' && ($resolved->{Host} // '') eq $trusted &&
			($resolved->{Expires} // 0) > time && ref($resolved->{Addresses}) eq 'ARRAY';
		@addresses = @{$resolved->{Addresses}};
	}
	for my $address (@addresses) {
		next if !defined($address) || ref($address);
		my $family = $address =~ /:/ ? Socket::AF_INET6() : Socket::AF_INET();
		my ($expected, $actual) = eval { map { Socket::inet_pton($family, $_) } ($address, $source) };
		return 1 if defined($expected) && defined($actual) && $expected eq $actual;
	}
	return 0;
}

use vars qw(@ISA $deftimeout);

@ISA = qw(Msg);
$deftimeout = 60;

# DXRegia 2026-11-18: verify managed sysop hashes while accepting legacy peer credentials.
sub supports_dxregia_bcrypt
{
	my ($hash) = @_;
	my $probe = eval { crypt('DXRegia-bcrypt-probe', $hash) };
	return defined($probe) && $probe =~ /\A\$2a\$10\$[.\/A-Za-z0-9]{53}\z/;
}

sub dxregia_password_matches
{
	my ($password, $stored) = @_;
	return 0 unless defined($stored);
	return $password eq $stored unless $stored =~ s/\A\{DXREGIA-BCRYPT\}//;
	return 0 unless $stored =~ /\A\$2a\$10\$[.\/A-Za-z0-9]{53}\z/ &&
		length($password) <= 72 && $password !~ /[^!-~]/;
	my $actual = eval { crypt($password, $stored) };
	return 0 unless defined($actual) && length($actual) == length($stored);
	my $different = 0;
	for my $i (0 .. length($stored) - 1) {
		$different |= ord(substr($actual, $i, 1)) ^ ord(substr($stored, $i, 1));
	}
	return $different == 0;
}

sub login
{
	goto &main::login;        # save some writing, this was the default
}

sub enqueue
{
	my ($conn, $msg) = @_;
	unless ($msg =~ /^[ABZ]/) {
		if ($msg =~ m{^E[-\w\/]+\|([01])} && $conn->{csort} eq 'telnet') {
			$conn->{echo} = $1;
			if ($1) {
#				$conn->send_raw("\xFF\xFC\x01");
			} else {
#				$conn->send_raw("\xFF\xFB\x01");
			}
		} else {
			$msg =~ s{^[-\w\/]+\|}{};
			push (@{$conn->{outqueue}}, $msg . $conn->{lineend});
		}
	}
}

sub send_raw
{
	my ($conn, $msg) = @_;
	dbg((ref $conn) . " connect $conn->{cnum}: $msg") if $conn->{state} ne 'C' && isdbg('connect');
	$conn->SUPER::send_raw($msg);
}

sub echo
{
	my $conn = shift;
	$conn->{echo} = shift;
}

sub _rcv
{
    my $conn = shift; # $rcv_now complement of $flush
	my $msg = shift;
    my $sock = $conn->{sock};
    return unless defined($sock);
	return if $conn->{disconnecting};

	if ($conn->{state} eq 'WL' && $conn->{sort} =~ /^I/ && $msg =~ /^PROXY/) {
		my $echo = $conn->{echo};
		$conn->{echo} = 0;
		$conn->SUPER::_rcv($msg);
		$conn->{echo} = $echo;
	} else {
		$conn->SUPER::_rcv($msg);
	}
}

sub dequeue
{
	my $conn = shift;
	my $msg;

	if ($conn->ax25 && exists $conn->{msg}) {
		$conn->{msg} =~ s/\cM/\cJ/g;
	}
	if ($conn->{state} eq 'WC') {
		if (exists $conn->{cmd}) {
			if (@{$conn->{cmd}}) {
				dbg("connect $conn->{cnum}: $conn->{msg}") if isdbg('connect');
				$conn->_docmd($conn->{msg});
			} 
		}
		if ($conn->{state} eq 'WC' && exists $conn->{cmd} && @{$conn->{cmd}} == 0) {
			$conn->to_connected($conn->{call}, 'O', $conn->{csort});
		}
	} elsif ($conn->{msg} =~ /\cJ/) {
		my @lines =  $conn->{msg} =~ /([^\cM\cJ]*)\cM?\cJ/g;
		if ($conn->{msg} =~ /\cJ$/) {
			delete $conn->{msg};
		} else {
			$conn->{msg} =~ s/([^\cM\cJ]*)\cM?\cJ//g;
		}
		$conn->{linesin} += @lines;
		$Msg::total_lines_in += @lines;
		while (defined ($msg = shift @lines)) {
			# DXRegia, 2026-11-18: exclude incoming peer passwords from connection debug logs.
			dbg("connect $conn->{cnum}: $msg") if $conn->{state} ne 'C' && $conn->{state} ne 'WP' && isdbg('connect');
		
			$msg =~ s/\xff\xfa.*\xff\xf0|\xff[\xf0-\xfe].//g; # remove telnet options
#			$msg =~ s/[\x00-\x08\x0a-\x19\x1b-\x1f\x80-\x9f]/./g;         # immutable CSI sequence + control characters
			
			if ($conn->{state} eq 'C') {
				&{$conn->{rproc}}($conn, "I$conn->{call}|$msg");
			} elsif ($conn->{state} eq 'WL' ) {
				$msg = uc $msg;
				if ($conn->{sort} =~ /^I/ && (my ($ip, $from) = $msg =~ /^PROXY TCP[46] ([\da-fA-F:\.]+) ([\da-fA-F:\.]+)/) ) {
					# SOMEONE appears to have affixed an HA Proxy to my connection
					$ip =~ s|^::ffff:||; # chop off leading pseudo IPV6 stuff on dual stack listeners
					$from =~ s|^::ffff:||;
					if ($from eq $conn->{peerhost}) {
						dbg("ExtMsg: connect - PROXY IP change from '$conn->{peerhost}' -> '$ip'");
						$conn->{peerhost} = $ip;
					} else {
						dbg("ExtMsg: connect - PROXY someone ($from) is trying to spoof '$ip'");
						$conn->send_now("Sorry $msg is an invalid callsign");
						$conn->disconnect;
					}
				} elsif (is_callsign($msg)) {
					my $call = normalise_call($msg);
					if ($main::allowslashcall || $msg !~ m|/|) {
						my $sort = $conn->{csort};
						$sort = 'local' if $conn->{peerhost} =~ /127\.\d+\.\d+\.\d+$/ || $conn->{peerhost} eq '::1';
						my $uref = DXUser::get_current($call);
						my $ip_match = dxregia_peer_ip_match($conn, $call, $uref);
						if (defined($ip_match)) {
							if ($ip_match) {
								$conn->to_connected($call, 'A', $sort);
							} else {
								$conn->send_now("Connection rejected: unexpected source address");
								$conn->disconnect;
							}
						} elsif ($main::passwdreq || $uref && $uref->passwd) {
							$conn->conns($call);
							$conn->{state} = 'WP';
							$conn->{decho} = $conn->{echo};
							$conn->{echo} = 0;
							$conn->send_raw('password: ');
						} else {
							$conn->to_connected($msg, 'A', $sort);
						}
					} else {
						$conn->send_now("Sorry $msg is an invalid callsign");
						$conn->disconnect;
					}
				} else {
					$conn->send_now("Sorry $msg is an invalid callsign");
					$conn->disconnect;
				}
			} elsif ($conn->{state} eq 'WP' ) {
				my $uref = DXUser::get_current($conn->{call});
				$msg =~ s/[\r\n]+$//;
				if ($uref && dxregia_password_matches($msg, $uref->passwd)) {
					my $sort = $conn->{csort};
					$conn->{echo} = $conn->{decho};
					delete $conn->{decho};
					$sort = 'local' if $conn->{peerhost} =~ /127\.\d+\.\d+\.\d+$/ || $conn->{peerhost} eq '::1';
					$conn->{usedpasswd} = 1;
					$conn->to_connected($conn->{call}, 'A', $sort);
				} else {
					$conn->_dxregia_log('password-rejected');
					$conn->send_now("Sorry");
					$conn->disconnect;
				}
			} elsif ($conn->{state} eq 'WC') {
				if (exists $conn->{cmd} && @{$conn->{cmd}}) {
					$conn->_docmd($msg);
					if ($conn->{state} eq 'WC' && exists $conn->{cmd} &&  @{$conn->{cmd}} == 0) {
						$conn->to_connected($conn->{call}, 'O', $conn->{csort});
					}
				}
			}
		}
	}
}

sub to_connected
{
	my ($conn, $call, $dir, $sort) = @_;
	$conn->{state} = 'C';
	$conn->conns($call);
	delete $conn->{cmd};
	$conn->{timeout}->del if $conn->{timeout};
	delete $conn->{timeout};
	$conn->{csort} = $sort;
	&{$conn->{rproc}}($conn, "$dir$call|$sort");
	$conn->_send_file(localdata("connected")) unless $conn->{outgoing};
}

# DXRegia, 2026-11-19: use the saved issue translation before the login prompt.
sub dxregia_issue
{
	my ($conn) = @_;
	my $host = $conn->peerhost;
	my $v4 = Socket::inet_pton(Socket::AF_INET(), $host);
	$host =~ s/%[A-Za-z0-9_.-]+$//;
	my $ip = $v4 ? (("\0" x 10) . "\xff\xff" . $v4) : Socket::inet_pton(Socket::AF_INET6(), $host);
	my $lang;
	my $local = $v4
		? $host =~ /\A(?:10[.]|127[.]|169[.]254[.]|192[.]168[.]|172[.](?:1[6-9]|2[0-9]|3[01])[.])/
		: defined($ip) && ($ip eq (("\0" x 15) . "\1") || (ord(substr($ip, 0, 1)) & 254) == 252 ||
			substr(unpack('H*', $ip), 0, 3) =~ /\Afe[89ab]\z/);
	if ($local) {
		my %language = (85 => 'it', 32 => 'de', 129 => 'de', 42 => 'fr', 34 => 'es', 23 => 'pt', 142 => 'pt');
		my ($dxcc) = Prefix::cty_data($main::mycall);
		$lang = $language{$dxcc};
	} elsif (defined $ip && open(my $file, '<:raw', "$main::local_data/dxregia-issue-countries.bin")) {
		my @stat = stat($file);
		my ($low, $high) = (0, int(($stat[7] // 0) / 34) - 1);
		while ($low <= $high) {
			my $middle = int(($low + $high) / 2);
			last unless seek($file, $middle * 34, 0) && (read($file, my $record, 34) // 0) == 34;
			if ($ip lt substr($record, 0, 16)) {
				$high = $middle - 1;
			} elsif ($ip gt substr($record, 16, 16)) {
				$low = $middle + 1;
			} else {
				$lang = substr($record, 32, 2);
				last;
			}
		}
		close $file;
	}
	if (defined($lang) && $lang =~ /\A(?:it|de|fr|es|pt)\z/) {
		my $translated = localdata("issue.$lang");
		return $translated if -f $translated;
	}
	return localdata("issue");
}

sub new_client {
	
	my $server_conn = shift;
	my $client = shift;
	my $conn = $server_conn->SUPER::new_client($client);
	# send login prompt
	$conn->{state} = 'WL';
	$conn->_send_file($conn->dxregia_issue);
	$conn->send_raw("login: ");
	$conn->_dotimeout(60);
#	$conn->{echo} = 1;
}

sub start_connect
{
	my $call = shift;
	my $fn = shift;
	my $conn = ExtMsg->new(\&main::new_channel); 
	$conn->{outgoing} = 1;
	$conn->conns($call);
	
	my $f = new IO::File $fn;
	push @{$conn->{cmd}}, <$f>;
	$f->close;
	$conn->{state} = 'WC';
	$conn->_dotimeout($deftimeout);
	$conn->_docmd;
}

sub _docmd
{
	my $conn = shift;
	my $msg = shift;
	my $cmd;

	while ($cmd = shift @{$conn->{cmd}}) {
		chomp $cmd;
		next if $cmd =~ /^\s*\#/o;
		next if $cmd =~ /^\s*$/o;
		$conn->_doabort($1) if $cmd =~ /^\s*a\w*\s+(.*)/i;
		$conn->_dotimeout($1) if $cmd =~ /^\s*t\w*\s+(\d+)/i;
		$conn->_dolineend($1) if $cmd =~ /^\s*[Ll]\w*\s+\'((?:\\[rn])+)\'/i;
		if ($cmd =~ /^\s*co\w*\s+(\w+)\s+(.*)$/i) {
			unless ($conn->_doconnect($1, $2)) {
				$conn->disconnect;
				@{$conn->{cmd}} = [];    # empty any further commands
				last;
			}  
		}
		if ($cmd =~ /^\s*\'([^\']*)\'\s+\'([^\']*)\'/) {
			# DXRegia 2026-11-18: run the next directive after a matched prompt.
			last unless $conn->_dochat($cmd, $msg, $1, $2);
			undef $msg;
			next;
		}
		if ($cmd =~ /^\s*cl\w+\s+(.*)/i) {
			$conn->_doclient($1);
			last;
		}
		last if $conn->{state} eq 'E';
	}
}

sub _doconnect
{
	my ($conn, $sort, $line) = @_;
	my $r;

	$sort = lc $sort;
	dbg("CONNECT $conn->{cnum} sort: $sort command: $line") if isdbg('connect');
	if ($sort eq 'telnet') {
		# this is a straight network connect
		my ($host, $port) = split /\s+/, $line;
		$port = 23 if !$port;
		$r = $conn->connect($host, $port);
		if ($r) {
			dbg("Connected $conn->{cnum} to $host $port") if isdbg('connect');
		} else {
			dbg("***Connect $conn->{cnum} Failed to $host $port $!") if isdbg('connect');
		}
	} elsif ($sort eq 'agw') {
		# turn it into an AGW object
		bless $conn, 'AGWMsg';
		$r = $conn->connect($line);
	} elsif ($sort eq 'bpq') {
		# turn it into an BPQ object
		bless $conn, 'BPQMsg';
		$r = $conn->connect($line);
	} elsif ($sort eq 'ax25' || $sort eq 'prog') {
		$r = $conn->start_program($line, $sort);
	} else {
		dbg("invalid type of connection ($sort)");
	}
	$conn->disconnect unless $r;
	return $r;
}

sub _doabort
{
	my $conn = shift;
	my $string = shift;
	dbg("connect $conn->{cnum}: abort $string") if isdbg('connect');
	$conn->{abort} = $string;
}

sub _dotimeout
{
	my $conn = shift;
	my $val = shift;
	dbg("connect $conn->{cnum}: timeout set to $val") if isdbg('connect');
	$conn->{timeout}->del if $conn->{timeout};
	$conn->{timeval} = $val;
	$conn->{timeout} = DXTimer->new($val, sub{ &_timedout($conn) });
}

sub _dolineend
{
	my $conn = shift;
	my $val = shift;
	dbg("connect $conn->{cnum}: lineend set to $val ") if isdbg('connect');
	$val =~ s/\\r/\r/g;
	$val =~ s/\\n/\n/g;
	$conn->{lineend} = $val;
}

sub _dochat
{
	my $conn = shift;
	my $cmd = shift;
	my $line = shift;
	my $expect = shift;
	my $send = shift;
		
	if ($line) {
		if ($expect) {
			dbg("connect $conn->{cnum}: expecting: \"$expect\" received: \"$line\"") if isdbg('connect');
			if ($conn->{abort} && $line =~ /$conn->{abort}/i) {
				dbg("connect $conn->{cnum}: aborted on /$conn->{abort}/") if isdbg('connect');
				$conn->disconnect;
				delete $conn->{cmd};
				return;
			}
			if ($line =~ /\Q$expect/i) {
				if (length $send) {
					$conn->_dxregia_log('login-sent') if $expect eq 'login';
					$conn->_dxregia_log('password-sent') if $expect eq 'password:';
					# DXRegia, 2026-11-18: omit outgoing login passwords from connection debug logs.
					dbg("connect $conn->{cnum}: got: \"$expect\" sending: \"$send\"") if isdbg('connect') && $expect !~ /assword/i;
					$conn->send_later("D$conn->{call}|$send");
				}
				delete $conn->{msg}; # get rid any input if a match
				return 1;
			}
		}
	}
	$conn->{state} = 'WC';
	unshift @{$conn->{cmd}}, $cmd;
	return 0;
}

sub _timedout
{
	my $conn = shift;
	my $stage = $conn->{outgoing} ? 'remote-prompt-timeout' : $conn->{state} eq 'WP' ? 'password-timeout' : 'login-timeout';
	$conn->_dxregia_log($stage, $conn->{timeval});
	$conn->disconnect;
}

# DXRegia, 2026-11-19: log stages and validated metadata, without protocol payloads or credentials.
sub _dxregia_log
{
	my ($conn, $event, $seconds) = @_;
	return unless isdbg('connect');
	my $call = $conn->{call} // '';
	$call = '-' unless $call =~ /\A[A-Z0-9]+(?:[-\/][A-Z0-9]+)*\z/ && length($call) <= 16;
	my $host = $conn->{peerhost} // '';
	$host = '-' unless $host =~ /\A[A-Za-z0-9.:\-]{1,253}\z/;
	my $direction = $conn->{outgoing} ? 'outgoing' : 'incoming';
	$seconds ||= 0;
	dbg("DXRegia session $conn->{cnum} $direction $call $host $event $seconds");
}

# handle callsign and connection type firtling
sub _doclient
{
	my $conn = shift;
	my $line = shift;
	my @f = split /\s+/, $line;
	my $call = uc $f[0] if $f[0];
	$conn->conns($call);
	$conn->{csort} = $f[1] if $f[1];
	$conn->{state} = 'C';
	eval {$conn->{peerhost} = $conn->{sock}->handle->peerhost} unless $conn->ax25;
	$conn->_dxregia_log('protocol-start');
	&{$conn->{rproc}}($conn, "O$call|$conn->{csort}");
	delete $conn->{cmd};
	$conn->{timeout}->del if $conn->{timeout};
}

sub _send_file
{
	my $conn = shift;
	my $fn = shift;
	
	if (-e $fn) {
		my $f = new IO::File $fn;
		if ($f) {
			while (<$f>) {
				chomp;
				my $l = $_;
				dbg("connect $conn->{cnum}: $l") if isdbg('connll');
				$conn->send_raw($l . $conn->{lineend});
			}
			$f->close;
		}
	}
}
