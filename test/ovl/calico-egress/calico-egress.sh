#! /bin/sh
##
## calico-egress.sh --
##
##   Help script for the xcluster ovl/calico-egress.
##
## Commands;
##

prg=$(basename $0)
dir=$(dirname $0); dir=$(readlink -f $dir)
me=$dir/$prg
tmp=/tmp/${prg}_$$

die() {
	echo "ERROR: $*" >&2
	rm -rf $tmp
	exit 1
}
help() {
	grep '^##' $0 | cut -c3-
	rm -rf $tmp
	exit 0
}
test -n "$1" || help
echo "$1" | grep -qi "^help\|-h" && help

log() {
	echo "$prg: $*" >&2
}
dbg() {
	test -n "$__verbose" && echo "$prg: $*" >&2
}

##  env
##    Print environment.
##
cmd_env() {

	if test "$cmd" = "env"; then
		set | grep -E '^(__.*)='
		retrun 0
	fi

	test -n "$XCLUSTER" || die 'Not set [$XCLUSTER]'
	test -x "$XCLUSTER" || die "Not executable [$XCLUSTER]"
	eval $($XCLUSTER env)
}

##   test --list
##   test [--xterm] [--no-stop] [test...] > logfile
##
cmd_test() {
	if test "$__list" = "yes"; then
		grep '^test_' $me | cut -d'(' -f1 | sed -e 's,test_,,'
		return 0
	fi

	cmd_env
	start=starts
	test "$__xterm" = "yes" && start=start
	rm -f $XCLUSTER_TMP/cdrom.iso

	if test -n "$1"; then
		for t in $@; do
			test_$t
		done
	else
		test_start
	fi		

	now=$(date +%s)
	tlog "Xcluster test ended. Total time $((now-begin)) sec"

}

test_start() {
	. ./network-topology/Envsettings

	# Test with k8s-xcluster;
	__image=$XCLUSTER_HOME/hd-k8s-xcluster-$__k8sver.img
	test -r $__image || __image=$XCLUSTER_HOME/hd-k8s-xcluster.img
	export __image
	test -r $__image || die "Not readable [$__image]"
	export XCTEST_HOOK=$($XCLUSTER ovld k8s-xcluster)/xctest-hook
	export xcluster_FIRST_WORKER=2

	$($XCLUSTER ovld images)/images.sh lreg_preload k8s-cni-calico
	xcluster_start k8s-cni-calico iptools network-topology usrsctp cert-manager calico-egress $@

	otc 1 check_namespaces
	otc 1 check_nodes
	otc 1 deploy_cert_manager
	otc 1 deploy_controller

	otc 221 start_server

	otc 201 vip_ecmp_route
	otc 202 "vip_ecmp_route 2"

	otc 221 "start_tcpdump eth1"
	otc 221 "start_tcpdump eth2"

	otc 2 deploy_client_pods
	otc 2 "start_tcpdump_proc_ns usrsctpt"

	tlog "Sleep for 60 seconds for the client to finish"
	sleep 60
	#otc 1 start_client_interactive

	otc 2 stop_all_tcpdump
	otc 221 stop_all_tcpdump

	sleep 10

	rcp 2 /var/log/*.pcap captures/
	rcp 221 /var/log/*.pcap captures/
}

. $($XCLUSTER ovld test)/default/usr/lib/xctest
indent=''

# Get the command
cmd=$1
shift
grep -q "^cmd_$cmd()" $0 $hook || die "Invalid command [$cmd]"

while echo "$1" | grep -q '^--'; do
	if echo $1 | grep -q =; then
		o=$(echo "$1" | cut -d= -f1 | sed -e 's,-,_,g')
		v=$(echo "$1" | cut -d= -f2-)
		eval "$o=\"$v\""
	else
		o=$(echo "$1" | sed -e 's,-,_,g')
		eval "$o=yes"
	fi
	shift
done
unset o v
long_opts=`set | grep '^__' | cut -d= -f1`

# Execute command
trap "die Interrupted" INT TERM
cmd_$cmd "$@"
status=$?
rm -rf $tmp
exit $status
