# github.com/sriramy/calico-egress - ovl/calico-egress

Function tests for the
[calico-egress](https://github.com/sriramy/calico-egress).

## Usage

Basic tests;
```
test -n "$log" || log=/tmp/$USER/xc-calico-egress.log
. ./network-topology/Envsettings
./calico-egress.sh test > $log
```

It is *recommended* to setup `xcluster` in a
[netns](https://github.com/Nordix/xcluster/blob/master/doc/netns.md)
for these tests.

Setup xcluster;
```
XC_PATH=$HOME/xc/xcluster
GIT_TOP=$(git rev-parse --show-toplevel)

cd $XC_PATH
. ./Envsettings.k8s
. ~/.kubectl.bash
export __kver=linux-5.18.1
export __kbin=$XCLUSTER_WORKSPACE/xcluster/bzImage-$__kver
export __kobj=$XCLUSTER_WORKSPACE/xcluster/obj-$__kver
export __kcfg=$(git rev-parse --show-toplevel)/config/$__kver

export XCLUSTER_OVLPATH=$(readlink -f .)/ovl:$GIT_TOP/test/ovl
cd -
```
