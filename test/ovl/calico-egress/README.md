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



