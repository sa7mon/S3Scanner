# Storage Providers

### Global Endpoints

When checking if a bucket exists, S3Scanner will send a request to the storage provider's "global endpoint" if the provider has one. These endpoints are not region-specific and will respond with a redirect to the correct region URL if the bucket does exist. In this way, we can verify both that the bucket exists and which region it belongs to with only 1 request.

If the storage provider does not have a "global endpoint", S3Scanner relies on an internal list of regions for each provider and will check each one thus resulting in a much slower scanning process.

## Supported Providers with Global Endpoints

* AWS - https://s3.amazonaws.com
* GCP - https://storage.googleapis.com


## Supported

These providers have full support can be scanned using the `-provider` flag.

| Service                                                                                         | Insecure | Address Style |
|-------------------------------------------------------------------------------------------------|----------|---------------|
| [Amazon Web Services S3]()                                                                      | No       | path / vhost  |
| [Digital Ocean Spaces](https://www.digitalocean.com/products/spaces/)                           | No       | path          |
| [Dreamhost Objects](https://www.dreamhost.com/cloud/storage/)                                   | Yes      | path / vhost  |
| [Google Cloud Storage Buckets](https://cloud.google.com/storage/docs/request-endpoints#xml-api) | No       | path / vhost  |
| [Linode Object Storage](https://www.linode.com/products/object-storage/)                        | No       | vhost         |
| [Scaleway Object Storage](https://www.scaleway.com/en/object-storage/)                          | No       | path          |
| [Wasabi S3-compatible Cloud Storage](https://wasabi.com/s3-compatible-cloud-storage/)           | Yes      | path          |

## Tested Providers

These are storage providers that have been manually tested and confirmed to work but do not have built-in support via the `-provider` flag yet. Use `-provider custom` and provide a config file to scan these providers.

| Service                                                                                                          | Insecure | Address Style | Notes                                                                                                                                                                                                                                              |
|------------------------------------------------------------------------------------------------------------------|----------|---------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [Alibaba Cloud Object Storage](https://www.alibabacloud.com/product/oss)                                         | No       | path          | [Global Endpoint](https://oss.aliyuncs.com), [All endpoints](https://www.alibabacloud.com/help/en/oss/user-guide/regions-and-endpoints)                                                                                                            |
| [IBM Cloud Object Storage](https://cloud.ibm.com/docs/cloud-object-storage?topic=cloud-object-storage-endpoints) | No       | path          | [All Endpoints](https://cloud.ibm.com/docs/cloud-object-storage?topic=cloud-object-storage-endpoints)                                                                                                                                              | 
| [Tencent Cloud Object Storage](https://intl.cloud.tencent.com/document/product/436/6224#sample)                  | No       | path          | [Regions and Access Endpoints](https://intl.cloud.tencent.com/document/product/436/6224) S3Scanner should be given the bucket name in the form `bucket-APPID`. More info [here](https://intl.cloud.tencent.com/document/product/436/6224#examples) |
| [Vultr](https://www.vultr.com/docs/vultr-object-storage)                                                         | No       | path          |                                                                                                                                                                                                                                                    |

## Untested Providers

Connector to test with: https://github.com/gaul/s3proxy

| Service                                                                                                                  | Notes |
|--------------------------------------------------------------------------------------------------------------------------|-------|
| Apache CloudStack                                                                                                        |       |
| [Apache Hadoop Ozone](https://hadoop.apache.org/ozone/docs/0.3.0-alpha/s3.html)                                          |       | 
| Backblaze B2                                                                                                             |       |
| [Basho Technologies Riak CS](http://basho.com/riak-cloud-storage/)                                                       |       |
| [BriteSky Cloud Storage](http://www.britesky.ca/managed-cloud-services/off-site-cloud-backup/)                           |       |
| Ceph with RADOS gateway                                                                                                  |       |
| Cloudflare R2                                                                               | Not yet available as of 2021-10-18 |
| [Cloudian](http://www.cloudian.com/colt-cloud-storage-service-en.htm)                                                    |       |
| Cloudian HyperStore                                                                                                      |       |
| Connectria's Cloud Storage                                                                                               |       |
| [Cynny Space](http://www.cynnyspace.com)                                                                                 |       |
| [DCP (Distributed Cloud Platform)](https://www.gig.tech/evault-long-term-storage-service-lts2)                           |       |
| DDN Web Object Scaler (WOS) for on-premise cloud storage                                                                 |       |
| [EMC ECS](http://www.emc.com/storage/ecs/index.htm)                                                                      |       |
| Eucalyptus                                                                                                               |       |
| [FlexVault: Cloud Object Storage](https://www.flexvault.de/eternus-cd/)                                                  |       |
| [Garage](https://news.ycombinator.com/item?id=30256753) | |
| [HDS Sapphire Lake](https://www.hds.com/)                                                                                |       |
| [HGST ActiveScale](http://www.hgst.com/products/systems/activescale-system-content-platform.html)                        |       |
| [Huawei Cloud OBS (Object Storage Service)](http://www.hwclouds.com/product/obs.html)                                    |       |
| [Huawei UDS Massive Storage System](http://www.huawei.com/en/storage/massive-storage/fusionstorage)                      |       |
| [IBM Cloud Object Storage](https://www.ibm.com/cloud-computing/products/storage/object-storage/)                         |       |
| Minio                                                                                                                    | Might use [certificates](https://blog.min.io/certificate-based-authentication-with-s3/)       |
| [NEC Cloud IaaS](http://jpn.nec.com/cloud/service/platform_service/iaas.html)                                            |       |
| NetApp StorageGRID                                                                                                       |       |
| [NetApp® StorageGRID® Webscale](http://www.netapp.com/us/)                                                               |       |
| [NIFTY Cloud Object Storage](http://cloud.nifty.com/service/obj_storage.htm)                                             |       |
| [NooBaa](https://www.noobaa.com)                                                                                         |       |
| OpenIO                                                                                                                   |       |
| Openstack Swift                                                                                                          |       |
| Oracle Cloud                                                                                                             |       |
| OwnCloud                                                                                                                 |Find a hosted instance for testing      |
| [Peak 10 Object Storage](http://www.peak10.com/products-services/cloud-services/object-storage/)                         |       |
| [Pure Storage – FlashBlade](https://www.purestorage.com/products/flashblade.html)                                        |       |
| [Pure Storage – ObjectEngine](https://www.purestorage.com/products/objectengine.html)                                    |       |
| [QingStor](https://www.qingcloud.com/index.aspx)                                                                         |       |
| [Red Hat Ceph Storage](https://www.redhat.com/en/technologies/storage/ceph)                                              |       |
| [Revera Vault](http://www.revera.co.nz/solutions/homeland-cloud-iaas/vault-quick-viewsuse-enterprise-storage/)           |       |
| Riak CS                                                                                                                  |       |
| RSTOR Space                                                                                                              |       |
| [Scality RING](http://www.scality.com/suse-enterprise-storage/)                                                          |       |
| [Storj DCS Object Storage](https://docs.storj.io/dcs/getting-started/gateway-mt)                                         | May require credentials to scan |
| [SwiftStack](https://www.swiftstack.com/solutions/backup-and-recovery)                                                   |       |
| [ThinkOn Canada S3 Cloud Storage](http://www.thinkon.com)                                                                |       |
| TrueNAS Core                                                                                                             |       |
| [vCloud® Air™](http://vcloud.vmware.com/)                                                                                |       |
| [XSKY Object Storage (XEOS and XEDP)](https://www.xsky.com/en/)                                                          |       |