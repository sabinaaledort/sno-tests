# PTP T-GM Suite

## Prerequisites

- You need a running SNO with a WPC NIC and ICE driver installed.
- You need to install the ptp-operator and set a grandmaster PtpConfig.

## Running tests directly from git source tree

First, git clone the source tree.

Install Ginkgo and Gomega:
```
go install github.com/onsi/ginkgo/v2/ginkgo
go get github.com/onsi/gomega/...
```

- Set the `KUBECONFIG` variable.
- Set the `TGM_TESTING_PORT` variable to one of the WPC NIC ports.
- Optionally set the `TESTS_REPORTS_PATH` variable.

Run tests:
```
TGM_TESTING_PORT=ens1f0 make test-tgm
```

This describes two roles which will be used for the PTP tests: "gm" role and "tester" role. The "gm" role should be straightforward - a T-GM. The "tester" role needs some explanation - just like the "gm", the "tester" also has a WPC NIC installed with GNSS connected. The "tester" can be running in "free run" mode and uses its GPS to check the T-GM PTP timing accuracy. The "tester" can also be running the same functionality as a normal "slave" clock and sync up to the "gm" via PTP.

Each role has some ports connecting to other roles. Under the "gm", `toTester` means which port on the "gm" is used to connect the "tester"; under the "tester", `toGM` means which port on the "tester" is used to connect the "gm".

If the test environment does not have a second server with a GNSS connection, that's fine - certain test cases will simply be skipped.

If the test environment has only one single SNO and still want to test the WPC GNSS function on that single SNO, it can be done with a `topology.yaml` like the sample below,
```
ptp:
  gm:
    node: node12
    toTester: ens7f3    # A WPC port is needed even if the ethernet port is not connected
```

In the above topology file, there is one single gm role for the single SNO. In order to test the GNSS, `toTester` is still specified, even though no "tester" is defined. The `toTester` ethernet port does not need to be connected or up. The Ginkgo script uses the `toTester` field to find the intended WPC PCI address and query the GNSS module.  

With the above files under the `/testconfig` directory, the Ginkgo test suite can be triggered,
```
 ginkgo -v
 ```

## Running tests from a test container

#todo