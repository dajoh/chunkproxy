chunkproxy
==========

A scalable reverse caching proxy for immutable data. Caches data in chunks both in memory and on disk, and seamlessly rearranges them for the client.

Use Cases
---------

Chunkproxy is ideal for serving video content, as it supports range requests and only caches immutable data. A common case with video is that a client only watches a video partially. As chunkproxy only retrieves parts of a file from the backing store and not the entire file this saves bandwidth and time. Since it's cached in chunks the parts of the video that have already been watched will get evicted from the cache once new content is needed.

Live Scaling
------------

Chunkproxy supports live scaling of clusters, with no configuration needed. When nodes are added/removed, other nodes are notified and the load is rebalanced. This requires an [etcd](https://github.com/coreos/etcd) server/cluster set up.

How It Works
------------

When chunkproxy gets a request from a client (either a normal one or a range request), it figures out the first chunk it needs to get of a file. The lookup of a chunk consists of these steps:

1. Check if the chunk is in memory because we are either the owner of it or if it's super hot
2. Check if the owner of the chunk in the cluster has it, if so get it from them
3. If we are the owner, and it doesn't reside in memory, check if we have it on disk
4. If we don't have it on disk, get the chunk from the backing store

Chunkproxy uses [groupcache](https://github.com/golang/groupcache) internally for caching chunks in memory and coordinating with cluster peers.

How to Get It
-------------

```
$ go get github.com/dajoh/chunkproxy
$ chunkproxy -help
```

Setting up a static cluster
---------------------------

Setting up a static cluster is done with the `-poolpeers` flag.

```
$ chunkproxy -dir cache1 -pooladdr 127.0.0.1:7171 -webaddr :8181 -poolpeers 127.0.0.1:7171,127.0.0.1:7272
$ chunkproxy -dir cache2 -pooladdr 127.0.0.1:7272 -webaddr :8282 -poolpeers 127.0.0.1:7171,127.0.0.1:7272
```

Setting up a dynamic cluster
----------------------------

Setting up a dynamic cluster is done with the `-coord` and `-cluster` flags.

* `-coord` sets the etcd server URL, e.g. `http://etcd.dajoh.io`
* `-cluster` sets the etcd directory name, e.g. `clusters/banana`

```
$ chunkproxy -dir cache1 -pooladdr 127.0.0.1:7171 -webaddr :8181 -coord http://etcd.dajoh.io -cluster clusters/banana
$ chunkproxy -dir cache2 -pooladdr 127.0.0.1:7272 -webaddr :8282 -coord http://etcd.dajoh.io -cluster clusters/banana
```

License
-------

If you want to use chunkproxy in your application or on your website, please contact me.
