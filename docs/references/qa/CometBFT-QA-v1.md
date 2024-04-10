---
order: 1
parent:
  title: QA results for CometBFT v1.x
  description: This is a report on the results obtained when running CometBFT v1.x on testnets
  order: 5
---

# QA results for CometBFT v1.x

We run this iteration of the QA tests on CometBFT `v1.0.0-alpha.2`, the second tag of the backport
branch `v1.x` from the CometBFT repository. The previous QA tests were performed on
`v0.38.0-alpha.2` from May 21, 2023, which we use here as a baseline for comparison. There are many
changes with respect to the baseline, including `TO COMPLETE`. For the full list of changes, check
out the [CHANGELOG](https://github.com/cometbft/cometbft/blob/v1.0.0-alpha.2/CHANGELOG.md).

The main goal of the QA process is to validate that there are no meaningful, substantial regressions
from the previous version. We consider that there is a regression if we find a difference bigger
than 10% in the results. After having performed the experiments, we conclude that there are no
significant differences with respect to the baseline. Therefore version `v1.0.0-alpha.2` has passed
the QA tests. 

In the rest of this document we present and analyse the obtained results. The main steps of the QA
process are the following:
- [Saturation point](#saturation-point): On a 200-nodes network, find its saturation point, that is,
  the transaction load on which the system begins to show a degraded performance. On the rest of the
  QA experiments we will use subject the system to a load slightly under the saturation point.
- [200-nodes test](#200-nodes-test): During a fixed amount of time, inject on the 200-nodes network
  a constant load of transactions. Then collect metrics and block data to compute latencies, and
  compare them against the results of the baseline.
- [Rotating-nodes test](#rotating-nodes-test): Run initially 10 validators and 3 seed nodes. Then
  start a full node, wait until it is block-synced, and stop it. Repeat these steps 25 times while
  checking the nodes are able to catch up to the latest height of the network.

## Latency emulation (LE)

For the first time in the QA process we can additionally run the experiments using latency emulation
(LE). We typically deploy all the nodes of the testnet in the same region of a DigitalOcean
data center. This keeps the costs of running the tests low, but it makes the communication between
nodes unrealistic, as there is almost no latency. While still deploying the testnet in one region,
we now can emulate latency by adding random delays to outgoing messages. 

This is how we emulate latency:
- [This table][aws-latencies] has real data collected from AWS and containing the average latencies
  between different AWS data centers in the world.
- When we define the testnet, we randomly assign a "zone" to each node, that is, one of the regions
  in the latency table.
- Before starting CometBFT in each node, we run [this script][latency-emulator-script] to set the
  added delays between the current node and each of the other zones, as defined in the table. The
  script calls the `tc` utility for controlling the network traffic at the kernel level.

Until now all of our QA results were obtained without latency emulation. In order to analyze the
obtained results under similar configurations, we will make the analysis in a two-step comparison.
First, we will compare the QA results of `v0.38` (the baseline) to those of `v1` without latency
emulation. Then, we will compare results of `v1` with and without latency emulation.

Note that in this report we are not using the results with latency emulation to assess whether
`v1.0.0-alpha.2` passes or not the QA tests. The goal is to have a baseline for comparison for the
next QA tests to be performed for a future release.

## Storage optimizations

We have conducted several experiments aimed to address concerns regarding storage efficiency and
performance of CometBFT. These experiments focused on various aspects, including the effectiveness
of the pruning mechanism, the impact of different database key layouts, and the performance of
alternative database engines like PebbleDB. Check out the full report [here](../storage/README.md).

The experiments were performed on different versions of CometBFT. Of interest for this report are
the those where we targeted a version based on `v1.0.0-alpha.1`. The main difference with
`v1.0.0-alpha.2` is PBTS, which does not affect storage performance. Therefore, we consider that the
obtained results equally apply to `v1.0.0-alpha.2`. In particular, both versions contain the data
companion API, background pruning, compaction, and support for different key layouts. 

Briefly, the results relevant to `v1` indicate that:
- While pruning alone was ineffective in controlling storage growth, combining pruning with forced
  compaction proved to be an effective strategy.
- Experiments reveal mixed results regarding the impact of different database key layouts on
  performance, with some scenarios showing improvements in block processing times and storage
  efficiency, particularly when utilizing the new key layout. However, further analysis suggests
  that the benefits of the new layout were not consistently realized across different environments,
  prompting us to designate the new key layout as purely experimental.
- Tests with PebbleDB showcased promising performance improvements, with superior handling of
  compaction without the need for manual intervention. 

## Table of Contents
- [Saturation point](#saturation-point)
- [200-nodes test](#200-nodes-test)
  - [Latencies](#latencies)
  - [Metrics](#metrics)
  - [Results](#results)
- [Rotating-nodes test](#rotating-nodes-test)

## Saturation point

The first step of our QA process is to find the saturation point of the testnet. As in other
iterations of our QA process, we have used a network of 200 nodes as testbed, plus one node to send
the transaction load and another to collect metrics. The experiment consists of several iterations,
each of 90 seconds, with different load configurations. A configuration is defined by:
- `c`, the number of connections from the load runner process to the target node, and
- `r`, the rate or number of transactions issued per second. Each connection sends `r` transactions
  per second. 

For more details on the methodology to identify the saturation point, see [here](CometBFT-QA-34.md#saturation-point).

The following figure shows the obtained values for v1 and v0.38 (the baseline). Note that
configurations that have the same amount of transaction load, for example `c=1,r=400` and
`c=2,r=200`, are considered equivalent, and plotted in the same x-axis value corresponding to their
total rate, that is, to the equivalent configurations with `c=1`.

![saturation-plot](imgs/v1/saturation/saturation_v1_v038.png) 

We observe in the figure that until a rate of 400 txs/s, the obtained values are equal or very close
to the expected number of processed transactions (35600 txs). After this point, the system is not
able to process all the transactions that it receives, so some transactions are dropped, and we say
that the system is saturated. The expected number of processed transactions is `c * r * 89 s = 35600
txs`. (Note that we use 89 out of 90 seconds of the experiment because the last transaction batch
coincides with the end of the experiment and is thus not sent.) 

The complete results from which the figure was generated can be found in the file
[`v1_report_tabbed.txt`](imgs/v1/200nodes/metrics/v1_report_tabbed.txt). The following table
summarizes them. (These values are plotted in the figure.) We can see the saturation point in the
diagonal defined by `c=1,r=400` and `c=2,r=200`.

| r    | c=1       | c=2       | c=4   |
| ---: | --------: | --------: | ----: |
| 200  | 17800     | **34600** | 50464 |
| 400  | **31200** | 54706     | 49463 |
| 800  | 51146     | 51917     | 41376 |
| 1600 | 50889     | 47732     | 45530 |

For comparison, this is the table obtained on the baseline version, with the same saturation point.

| r    | c=1       | c=2       | c=4   |
| ---: | --------: | --------: | ----: |
| 200  | 17800     | **33259** | 33259 |
| 400  | **35600** | 41565     | 41384 |
| 800  | 36831     | 38686     | 40816 |
| 1600 | 40600     | 45034     | 39830 |

In conclusion, we chose `c=1,r=400` as the transaction load that we will use in the rest of QA
process. This is the same value used in the previous QA tests.

#### With latency emulation

For this comparison we run a new set of experiments with different transaction loads: we use only
one connection, and for the rate we use values in the range [100,1000] in intervals of 100
txs/second.

![v1_saturation](imgs/v1/saturation/saturation_v1_LE.png) 

These are the actual values from which the figure was generated:
| r    | v1 without LE | v1 with LE   | 
| ---: | ----: | ----: |
| 100  |  8900 |  8900 |
| 200  | 17800 | 17800 |
| 300  | 26053 | 26700 |
| 400  | 28800 | 35600 |
| 500  | 32513 | 34504 |
| 600  | 30455 | 42169 |
| 700  | 33077 | 38916 |
| 800  | 32191 | 38004 |
| 900  | 30688 | 34332 |
| 1000 | 32395 | 36948 |

## 200-nodes test

This experiment consist in running 200 nodes, injecting a load of 400 txs/s during 90 seconds, and
collect the metrics. The network is composed of 175 validator nodes, 20 full nodes, and 5 seed
nodes. Another node sends the load to only one of the validators.

### Latencies

The following figures show the latencies of the experiment carried out with the configuration
`c=1,r=400`. Each dot represents a block: at which time it was created (x axis) and the average
latency of its transactions (y axis).

| v0.38 | v1 (without LE / with LE) 
|:--------------:|:--------------:|
| ![latency-1-400-v38](img38/200nodes/e_de676ecf-038e-443f-a26a-27915f29e312.png) | ![latency-1-400-v1](imgs/v1/200nodes/latencies/e_8e4e1e81-c171-4879-b86f-bce96ee2e861.png) 
| | ![latency-1-400-v1-le](imgs/v1/200nodes_with_latency_emulation/latencies/e_8190e83a-9135-444b-92fb-4efaeaaf2b52.png)

In both cases, most latencies are around or below 4 seconds. On v0.38 there are peaks reaching 10
seconds, while on v1 (without LE) the only peak reaches 8 seconds. In general, the images are similar; then, from
this small experiment we infer that the latencies measured on the version under test is not worse
than those of the baseline. With latency emulation, the latencies are considerably higher, as expected.

### Metrics

We further examine key metrics extracted from Prometheus data on the experiment with configuration
`c=1,r=400`.

Note that the experiments with latency emulation have a duration of 180 seconds instead of the 90
seconds of the experiments without latency emulation.

#### Mempool size

<!-- The mempool size, a count of the number of transactions in the mempool, was shown to be stable and
homogeneous at all full nodes. It did not exhibit any unconstrained growth.  -->

<!-- The following figures show the evolution over time of the cumulative number of transactions inside
all full nodes' mempools at a given time. -->

<!-- | v0.38 | v1 (without LE / with LE)
| :--------------:|:--------------:|
| ![mempool-cumulative-baseline](img38/200nodes/mempool_size.png) | ![mempoool-cumulative](imgs/v1/200nodes/metrics/mempool_size.png)
| | ![mempoool-cumulative-le](imgs/v1/200nodes_with_latency_emulation/metrics/mempool_size.png)

`TODO`: fix scale in y axis on LE image. -->

The following figures show the evolution of the average and maximum mempool size over all full
nodes. On v1, the average mostly stays below 1000 outstanding transactions except for a peak above
2000, coinciding with the moment the system reached round number 1 (see below); this is better than
the baseline, which oscilates between 1000 and 2500.

| v0.38 | v1 (without LE / with LE) 
| :--------------:|:--------------:|
| ![mempool-avg-baseline](img38/200nodes/avg_mempool_size.png) | ![mempool-avg](imgs/v1/200nodes/metrics/avg_mempool_size.png)
| | ![mempool-avg-le](imgs/v1/200nodes_with_latency_emulation/metrics/avg_mempool_size.png)

With latency emulation, the average mempool size stays mostly above 2000 outstanding transactions
with peaks almost reaching the maximum mempool size of 5000 transactions.

The maximum mempool size show us when one or more nodes reached the maximem mempool capacity in
terms of number of transactions. On v0.38, we see that most of the time there is at least one node
that is dropping incoming transactions, while on v1 this happens less often, particularly after
reaching round 1 (see below). On v1 with latency emulation, there always a node that has its mempool
saturated.

| v0.38 | v1 (without LE / with LE)
| :--------------:|:--------------:|
| ![mempool-cumulative-baseline](img38/200nodes/mempool_size_max.png) | ![mempoool-cumulative](imgs/v1/200nodes/metrics/mempool_size_max.png)
| | ![mempoool-cumulative-le](imgs/v1/200nodes_with_latency_emulation/metrics/mempool_size_max.png)

#### Peers

The number of peers was stable at all nodes. As expected, the seed nodes have more peers (around
125) than the rest (between 20 and 70 for most nodes). The red dashed line denotes the average
value.

| v0.38 | v1 (without LE / with LE) 
|:--------------:|:--------------:|
| ![peers](img38/200nodes/peers.png) | ![peers](imgs/v1/200nodes/metrics/peers.png)
| | ![peers](imgs/v1/200nodes_with_latency_emulation/metrics/peers.png)

Just as in the baseline, the fact that non-seed nodes reach more than 50 peers is due to [\#9548].

#### Consensus rounds

Most blocks took just one round to reach consensus, except for a few cases when it was needed a
second round. For these specific runs, the baseline required an extra round more times.

With latency emulation, the performance is significantly worse; on multiple times needing an extra
round and even reaching three rounds.

| v0.38 | v1 (without LE / with LE) 
|:--------------:|:--------------:|
| ![rounds](img38/200nodes/rounds.png) | ![rounds](imgs/v1/200nodes/metrics/rounds.png)
| | ![rounds](imgs/v1/200nodes_with_latency_emulation/metrics/rounds.png)

#### Blocks produced per minute and transactions processed per minute

These figures show the rate in which blocks were created, from the point of view of each node. That
is, they shows when each node learned that a new block had been agreed upon. For most of the time
when load was being applied to the system, most of the nodes stayed around 20 blocks/minute. The
spike to more than 100 blocks/minute is due to a slow node catching up. The baseline experienced a
similar behavior. With latency emulation, the performance is degraded, going from around 30
blocks/min (without LE) to around 10 blocks/min.

| v0.38 | v1 (without LE / with LE)
|:--------------:|:--------------:|
| ![heights-baseline](img38/200nodes/block_rate.png) | ![heights](imgs/v1/200nodes/metrics/block_rate.png)
| | ![heights](imgs/v1/200nodes_with_latency_emulation/metrics/block_rate.png)

| v0.38 | v1 (without LE / with LE)
|:--------------:|:--------------:|
| ![total-txs-baseline](img38/200nodes/total_txs_rate.png) | ![total-txs](imgs/v1/200nodes/metrics/total_txs_rate.png)
| | ![total-txs](imgs/v1/200nodes_with_latency_emulation/metrics/total_txs_rate.png)

The collective spike on the right of the graph marks the end of the load injection, when blocks
become smaller (empty) and impose less strain on the network. This behavior is reflected in the
following graph, which shows the number of transactions processed per minute.

#### Memory resident set size

The following graphs show the Resident Set Size of all monitored processes. Most nodes use less than
0.9 GB of memory, and a maximum of 1.3GB. In all cases, the memory usage in this run is less than
the baseline. On all processes, the memory usage went down as the load was being removed, showing no
signs of unconstrained growth.

| v0.38 | v1 (without LE / with LE) 
|:--------------:|:--------------:|
|![rss](img38/200nodes/memory.png) | ![rss](imgs/v1/200nodes/metrics/memory.png)
| | ![rss](imgs/v1/200nodes_with_latency_emulation/metrics/memory.png)

#### CPU utilization

The best metric from Prometheus to gauge CPU utilization in a Unix machine is `load1`, as it usually
appears in the [output of
`top`](https://www.digitalocean.com/community/tutorials/load-average-in-linux). In this case, the
load is contained below 4 on most nodes, with the baseline showing a similar behavior.

| v0.38 | v1 (without LE / with LE) 
|:--------------:|:--------------:|
| ![load1-baseline](img38/200nodes/cpu.png) | ![load1](imgs/v1/200nodes/metrics/cpu.png)
| | ![load1](imgs/v1/200nodes_with_latency_emulation/metrics/cpu.png)

### Test Results

We have shown that there is no regressions when comparing CometBFT `v1.0.0-alpha.2` against the
results obtained for `v0.38`. The observed results are equal or sometimes slightly better than the
baseline. We therefore conclude that this version of CometBFT has passed the test.

| Scenario  | Date       | Version                                                   | Result |
| --------- | ---------- | --------------------------------------------------------- | ------ |
| 200-nodes | 2024-03-21 | v1 (without LE / with LE).0.0-alpha.2 (4ced46d3d742bdc6093050bd67d9bbde830b6df2) | Pass   |


## Rotating Nodes Testnet

As done in past releases, we use `c=1,r=400` as load, as the saturation point in the 200-node test
has not changed from `v0.38.x` (see the corresponding [section](#saturation-point) above).
Further, although latency emulation is now available, we decided to run this test case
without latency emulation (LE); this choice may change in future releases.

The baseline considered in this test case is `v0.38.0-alpha.2`, as described
in the [introduction](#qa-results-for-cometbft-v1x) above.

### Latencies

The following two plots show latencies for the whole experiment.
We see the baseline (`v0.38.0-alpha.2`) on th left, and the current version
on the right.

We can appreciate that most latencies are under 4 seconds in both cases,
and both graphs have a comparable amount of outliers between 4 seconds and 11 seconds.

|                               v0.38.0                               |                    v1.0.0 (without LE)                    |
| :-----------------------------------------------------------------: | :-------------------------------------------------------: |
| ![rotating-all-latencies-bl](img38/rotating/rotating_latencies.png) | ![rotating-all-latencies](imgs/v1/rotating/latencies.png) |


### Prometheus Metrics

This section shows relevant metrics both on `v1.0.0` and the baseline (`v0.38.0`).
In general, all metrics roughly match those seen on the baseline
for the rotating node experiment.

#### Blocks and Transactions per minute

The following two plots show the blocks produced per minute. In both graphs, most nodes stabilize around 40 blocks per minute.

|                            v0.38.0                             |                          v1.0.0 (without LE)                          |
| :------------------------------------------------------------: | :-------------------------------------------------------------------: |
| ![rotating-heights-bl](img38/rotating/rotating_block_rate.png) | ![rotating-heights](imgs/v1/rotating/metrics/rotating_block_rate.png) |

The following plots show only the heights reported by ephemeral nodes, both when they were blocksyncing
and when they were running consensus.

|                               v0.38.0                                |                        v1.0.0 (without LE)                         |
| :------------------------------------------------------------------: | :----------------------------------------------------------------: |
| ![rotating-heights-ephe-bl](img38/rotating/rotating_eph_heights.png) | ![rotating-heights-ephe](imgs/v1/rotating/metrics/rotating_eph_heights.png) |

In both cases, we see the main loop of the `rotating` test case repeat itself a number of times.
Ephemeral nodes are stopped, their persisted state is wiped out, their config is transferred over
from the orchestrating node, they are started, we wait for all of them to catch up via blocksync,
and the whole cycle starts over. All these steps are carried out via `ansible` playbooks.

We see that there are less cycles in `v1.0.0`. The reason is the following.
All `ansible` steps are currently run from the orchestrating node (i.e., the engineer's laptop).
The orchestrating node when running the rotating nodes test case for `v1.0.0`
was connected to a network connection that was slower than when the equivalent test was run for `v0.38.x`.
This caused the steps reconfiguring ephemeral nodes at the end of each cycle to be somewhat slower.
This can be noticed in the graphs when comparing the width (in x-axis terms) of the gaps without metric
from the end of a cycle to the beginning of the next one.

If we focus on the _width_ of periods when ephemeral nodes are blocksynching, we see their are slightly narrower
in `v1.0.0`. This is likely due to the improvements introduced as part of the following issues
[#1283](https://github.com/cometbft/cometbft/issues/1283),
[#2379](https://github.com/cometbft/cometbft/issues/2379), and
[#2465](https://github.com/cometbft/cometbft/issues/2465).

The following plots show the transactions processed per minute.

|                            v0.38.0                             |                          v1.0.0 (without LE)                          |
| :------------------------------------------------------------: | :-------------------------------------------------------------------: |
| ![rotating-total-txs-bl](img38/rotating/rotating_txs_rate.png) | ![rotating-total-txs](imgs/v1/rotating/metrics/rotating_txs_rate.png) |

They seem similar, except for an outlier in the `v1.0.0` plot.

#### Peers

The plots below show the evolution of the number of peers throughout the experiment.

|                         v0.38.0                         |                      v1.0.0 (without LE)                       |
| :-----------------------------------------------------: | :------------------------------------------------------------: |
| ![rotating-peers-bl](img38/rotating/rotating_peers.png) | ![rotating-peers](imgs/v1/rotating/metrics/rotating_peers.png) |

The plotted values and their evolution show the same dynamics in both plots.
Nevertheless, all nodes seem to acquire more peers when ephemeral node are catching up in the `v1.0.0` experiment.

For further explanations on these plots, see the [this section](TMCore-QA-34.md#peers-1).

#### Memory Resident Set Size

These plots show the average Resident Set Size (RSS) over all processes.
They are comparable in both releases.

|                            v0.38.0                             |                          v1.0.0 (without LE)                          |
| :------------------------------------------------------------: | :-------------------------------------------------------------------: |
| ![rotating-rss-avg-bl](img38/rotating/rotating_avg_memory.png) | ![rotating-rss-avg](imgs/v1/rotating/metrics/rotating_avg_memory.png) |

#### CPU utilization

The plots below show metric `load1` for all nodes for `v1.0.0-alpha.2` and for the baseline (`v0.38.0`).

|                        v0.38.0                        |                 v1.0.0 (without LE)                 |
| :---------------------------------------------------: | :-------------------------------------------------: |
| ![rotating-load1-bl](img38/rotating/rotating_cpu.png) | ![rotating-load1](imgs/v1/rotating/metrics/rotating_cpu.png) |

In both cases, it is contained under 5 most of the time, which is considered normal load.

### Test Result

| Scenario | Date       | Version                                                   | Result |
| -------- | ---------- | --------------------------------------------------------- | ------ |
| Rotating | 2024-04-03 | v1.0.0-alpha.2 (e42f62b681a2d0b05607a61d834afea90f73d366) | Pass   |

[aws-latencies]: https://github.com/cometbft/cometbft/blob/v1.0.0-alpha.2/test/e2e/pkg/latency/aws-latencies.csv
[latency-emulator-script]: https://github.com/cometbft/cometbft/blob/v1.0.0-alpha.2/test/e2e/pkg/latency/latency-setter.py 
[\#9548]: https://github.com/tendermint/tendermint/issues/9548
[end-to-end]: https://github.com/cometbft/cometbft/tree/main/test/e2e
