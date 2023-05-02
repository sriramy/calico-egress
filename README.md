# calico-egress
Egress SNAT controller based on [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)

# Function test
We use [xcluster](https://github.com/Nordix/xcluster) to run function tests.

## Setup
It is recommended to run the tests in netns, so add a namespace to start with
```
export XCLUSTER_DIR=$HOME/xc/xcluster
export CALICO_EGRESS_DIR=$HOME/code/calico-egress
cd $XCLUSTER_DIR
. Envsettings.k8s
xc nsadd 1
exit
```

Enter the namespace and set default paths
```
ip netns exec ${USER}_xcluster1 bash
export XCLUSTER_DIR=$HOME/xc/xcluster
export CALICO_EGRESS_DIR=$HOME/code/calico-egress
```

Initialize xcluster and add calico-egress to ovl path
```
cd $XCLUSTER_DIR
. Envsettings.k8s
export XCLUSTER_OVLPATH=$(readlink -f .)/ovl:$CALICO_EGRESS_DIR/test/ovl
cdo calico-egress
```

To run with a specific kernel version
```
source ~/.kubectl.bash
export __kver=linux-5.18.1
export __kbin=$XCLUSTER_WORKSPACE/xcluster/bzImage-$__kver
export __kobj=$XCLUSTER_WORKSPACE/xcluster/obj-$__kver
export __kcfg=$XCLUSTER_DIR/config/$__kver
```

## Run tests
```
cdo calico-egress
./calico-egress.sh build
./calico-egress.sh test
```
