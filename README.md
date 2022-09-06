# calico-egress
Egress SNAT controller

# Function test
We use [xcluster](https://github.com/Nordix/xcluster) to run function tests.

## Setup
It is recommended to run the tests in netns, so add a namespace to start with
'''
export XCLUSTER_DIR=$HOME/xc/xcluster
export CALICO_EGRESS_DIR=$HOME/code/calico-egress
xc nsadd 1
ip netns exec sriramy_xcluster1 bash
'''

Initialize xcluster and add calico-egress to ovl path
'''
cd $XCLUSTER_DIR
. Envsettings.k8s
export XCLUSTER_OVLPATH=$(readlink -f .)/ovl:$CALICO_EGRESS_DIR/test/ovl
cd $CALICO_EGRESS_DIR/test/ovl/calico-egress
'''

To run with a specific kernel version
'''
source ~/.kubectl.bash
export __kver=linux-5.18.1
export __kbin=$XCLUSTER_WORKSPACE/xcluster/bzImage-$__kver
export __kobj=$XCLUSTER_WORKSPACE/xcluster/obj-$__kver
export __kcfg=$XCLUSTER_DIR/config/$__kver
'''

## Run tests
'''
cd test/ovl/calico-egress
./calico-egress build
./calico-egress test
'''