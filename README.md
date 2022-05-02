# github.com/liumingmin/gocd
golang continue deploy

cdserver
cdnode(node manage  creds)
cdscript(script template)
cdservice(name, pkgUrl, targetPath, runCmd , envVar, cdscript)

cdserver.deploy(cdservice, ip)

tips:
binary from aws s3(minio)

next:
cdserviceversion?  cdservicenode?

jenkins concurrent run job may get same taskId(queueId), 
so create same as num executor jobs to run