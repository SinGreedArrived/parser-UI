#/usr/bin/perl 

BEGIN { $/ = "\n"; $\ = "\n"; }
LINE:while (defined($_ = <ARGV>)) {
    chomp $_;
    our @F = split(/\t/, $_, 0);
    print "$F[0]$F[1]";
    push @urls, $F[2] if $F[2] =~ /http/;
}
{
    foreach $e (@urls) {
        system "bash firefox $e";
    }
}
